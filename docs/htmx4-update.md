# htmx v2 to v4 Migration Audit

Audit of potential breaking changes and risks when upgrading the site from **htmx v2** to **htmx 4.0.0** (released 2026-08-28).

### Official Reference Links
* [htmx 4.0.0 Release Announcement](https://four.htmx.org/announcements/2026-08-28-htmx-4.0.0-is-released/)
* [What's New in htmx 4](https://four.htmx.org/docs/whats-new-in-htmx-4/)
* [htmx 4 Events Guide (`ctx` Reference)](https://four.htmx.org/docs/htmx-events-guide)
* [Official Migration Skill (`htmx-upgrade-from-htmx2.md`)](https://raw.githubusercontent.com/bigskysoftware/htmx/v4.0.0/dist/skills/htmx-upgrade-from-htmx2.md)
* [Official Guidance Skill (`htmx-guidance.md`)](https://raw.githubusercontent.com/bigskysoftware/htmx/v4.0.0/dist/skills/htmx-guidance.md)

No code changes have been made; this document records what needs attention before or during an upgrade. For potential benefits of upgrading, see [htmx4-review.md](htmx4-review.md).

---

## 1. Event Names & `event.detail.xhr` Rewrite (`XMLHttpRequest` → `fetch()`)

In htmx 4, `XMLHttpRequest` is replaced by `fetch()`:

* Lifecycle events are renamed to `htmx:phase:action` (`htmx:beforeRequest` → `htmx:before:request`, `htmx:afterRequest` → `htmx:after:request`, `htmx:beforeSwap` → `htmx:before:swap`, `htmx:afterSwap` → `htmx:after:swap`).
* Error events collapse into `htmx:response:error` (for HTTP `>= 400`) and `htmx:error` (for network/fetch/swap exceptions).
* Flat properties (`event.detail.elt`, `event.detail.xhr`, `event.detail.successful`, `event.detail.failed`) move under the unified request context `event.detail.ctx`.

### `event.detail` → `event.detail.ctx` Cheat Sheet for `server`

| htmx 2 (`event.detail`) | htmx 4 (`event.detail.ctx`) | Notes |
| :--- | :--- | :--- |
| `event.detail.elt` | `event.detail.ctx.sourceElement` | The DOM element that triggered the request |
| `event.detail.target` | `event.detail.ctx.target` | The swap target element (writable in `htmx:after:request`) |
| `event.detail.xhr` | `event.detail.ctx.response` | Object with `{ raw: Response, status: number, headers: Headers }` |
| `event.detail.xhr.status` | `event.detail.ctx.response.status` | HTTP status code (e.g., `200`, `302`, `400`, `404`, `500`) |
| `event.detail.xhr.responseText` | `event.detail.ctx.text` | Raw response body string (populated before `htmx:after:request`) |
| `event.detail.successful` | `event.detail.ctx.response?.status < 400` (or `ctx.response?.raw?.ok`) | True when HTTP response succeeded |
| `event.detail.failed && event.detail.xhr` | `event.detail.ctx.response?.status >= 400` | True when server returned an HTTP `4xx` / `5xx` status (also fires `htmx:response:error`) |
| Browser / network error fallback | Listen to `htmx:error` (or `!event.detail.ctx.response` in `htmx:finally:request`) | Network failures where `fetch()` rejects fire `htmx:error` and `htmx:finally:request` |
| `hx-on::after-swap="..."` | `hx-on:htmx:after:swap="..."` (or `hx-on::after:swap="..."`) | Inside `hx-on`, `ctx` is directly in scope |

### Example Helper Pattern for `layout-htmx.mjs`, `editor-assets.js`, and `uploader-submitter.mjs`

To avoid repetitive property lookups when migrating the three JS files, a small helper can normalize the v4 event context:

```javascript
function parseHtmxEvent(event) {
  const ctx = event.detail?.ctx;
  const elt = ctx?.sourceElement ?? event.detail?.elt ?? null;
  const status = ctx?.response?.status ?? 0;
  const text = ctx?.text ?? '';
  const hasResponse = Boolean(ctx?.response);
  const successful = hasResponse && status >= 200 && status < 400;
  const failed = hasResponse && status >= 400;
  return { ctx, elt, status, text, hasResponse, successful, failed };
}
```

### Affected Files

* **`assets/js/layout-htmx.mjs`**
  * Lines `17` & `37`: listens on `document.body` for `htmx:beforeRequest` and `htmx:afterRequest`.
  * Lines `137-476`: `afterLinks`, `afterCreators`, `afterUpdateDate`, `afterUpdateRels`, `removeSelectsValid`, `afterUpdate`, `afterClassifications`, `afterReset`, `updateError`, `afterRecord`, `recordSuccess`, `recordError`, `afterRequest`, and `errorXhr` all read `event.detail.elt`, `event.detail.successful`, `event.detail.failed`, `event.detail.xhr`, `xhr.status`, and `xhr.responseText`.
  * Without updating (or loading `htmx-2-compat`), search alerts, uploader alerts, artifact editor validation states (`.is-valid`), and public/hidden header/status badge updates (`recordSuccess`/`recordError`) will stop firing.
* **`assets/js/editor-assets.js`**
  * Line `312`: listens for `htmx:afterRequest`.
  * Lines `370-492`: `afterLinkRequest`, `afterRefreshRequest`, `afterFormRequest`, `successful`, and `errorXhr` inspect `event.detail.elt`, `event.detail.successful`, `event.detail.failed`, `event.detail.xhr`, `xhr.status`, and `xhr.responseText` for file/preview uploads, pixelate reloads, link deletion, and download-content copy buttons.
* **`assets/js/uploader-submitter.mjs`**
  * Lines `55`, `61-66`, `70-85`: listens for `htmx:beforeRequest`, `htmx:beforeSwap` (`evt.detail.xhr.status >= 400`), and `htmx:afterRequest` (`evt.detail.elt`, `evt.detail.successful`, `evt.detail.failed`, `evt.detail.xhr`, `xhr.status`, `xhr.responseText`) for Demozoo and Pouët lookup/submission modals.
* **`view/app/uploader_modal.tmpl`**
  * Lines `184`, `277`, `377`, `462`, `548`, `664`: six `#uploader-*-results` `<div>` elements use `hx-on::after-swap="this.classList.remove('d-none')"`. Verify whether this must change to `hx-on:htmx:after:swap` (or `hx-on::after:swap`) without `htmx-2-compat`, since htmx 4 renames `htmx:afterSwap` to `htmx:after:swap`.

---

## 2. Upload Progress Bars (`htmx:xhr:progress`) Will Stop Working

Because `fetch()` has no upload progress event equivalent, `htmx:xhr:*` events (`htmx:xhr:loadstart`, `htmx:xhr:progress`, `htmx:xhr:abort`) are **removed** in htmx 4. Even `htmx-2-compat` cannot restore `htmx:xhr:progress` for `fetch()`-based requests.

### Affected Files

* **`assets/js/uploader.mjs`** (lines `121-137`): `progress(formId, elementId)` uses `htmx.on('#${formId}', "htmx:xhr:progress", ...)` reading `event.detail.loaded / event.detail.total` (and `htmx.find`).
* **`assets/js/uploader.js`** (lines `23-38`): calls `progress(...)` for all 6 uploader modal forms (`uploader-image-form`, `uploader-intro-form`, `uploader-trainer-form`, `uploader-magazine-form`, `uploader-text-form`, `uploader-advanced-form`).
* **`assets/js/editor-assets.js`** (lines `18-19`): calls `progress(...)` for `artifact-editor-dl-form` and `artifact-editor-preview-form`.

**Action needed:** Upload `<progress>` bars will stay at `0` unless upload progress is handled with a dedicated `XMLHttpRequest` uploader or replaced with an indeterminate indicator.

---

## 3. `response-targets` Extension & `hx-ext` Replaced by `hx-status`

htmx 4 removes the `hx-ext` attribute from core (extensions are loaded via `<script>` tags or `htmx-2-compat`) and replaces the `response-targets` extension with native `hx-status` attributes (e.g., `hx-status:4xx="target:#..."`, `hx-status:5xx="target:#..."`).

### Affected Files

* **`assets/js/htmx-response-targets.js`**
  * Line `75`: uses `htmx.defineExtension(...)` (removed in v4; replaced by `htmx.registerExtension()`).
  * Lines `100-130`: hooks `htmx:beforeSwap` and reads `evt.detail.xhr`, `evt.detail.requestConfig.elt`, `evt.detail.shouldSwap`, `evt.detail.target`, `evt.detail.isError`, plus v2 internal API helpers (`api.getClosestAttributeValue`, `api.querySelectorExt`, `api.findThisElement`).
* **Asset/build wiring:**
  * `runner/runner.go` (lines `33`, `172-175`, `230`): builds `htmx-response-targets.min.js`.
  * `handler/app/asset.go` (lines `23-24`, `52-53`) & `handler/app/app.go` (lines `1050-1051`, `1163-1170`): serves `/js/htmx.v2.min.js` and `/js/htmx.v2.ext-response-targets.min.js` with SRI hashes.
  * `view/app/layout.tmpl` (lines `365-366`) & `public/UPDATE.md` (line `21`).
* **Templates & Go HTML fragments using `hx-ext="response-targets"` / `hx-target-error`:**
  * `view/app/artifactedit.tmpl` (lines `66-80`): `#artifact-delete-forever` + `hx-target-error="#artifact-delete-forever-errors"`.
  * `view/app/uploader_modal.tmpl` (lines `30-48`, `71-87`): Demozoo & Pouët modal bodies + `hx-target-error="#demozoo-submission-error"` / `"#pouet-submission-error"`.
  * `handler/htmx/submitprod.go` (lines `195-223`): generated `hx-put` buttons with `hx-target-error`.

---

## 4. `4xx` / `5xx` Responses Swap into Targets by Default

In htmx 2, `4xx` and `5xx` responses do not swap by default; in htmx 4, **all HTTP status codes except `204` and `304` swap by default**.

### Affected Flows

* **Live search (`view/app/searchhtmx.tmpl` lines `49-52`):** If `SearchReleaser` (`handler/htmx/htmx.go` lines `381-390`) returns `503` plain text on a DB error, htmx 4 will swap that error text directly into `#search-htmx-results` instead of leaving results untouched while `#search-htmx-alert` displays the error.
* **Artifact editor (`view/app/artifactedit.tmpl` lines `147-430`):** Controls with `hx-patch` that do not set `hx-swap="none"` (e.g., classifications `#artifact-editor-classification-help`, date `Save`/`Undo`, title/comment/releasers) will swap `400 Bad Request` plain text from `badRequest` (`handler/htmx/artifact.go` lines `273-280`) directly into buttons/targets.
* **Upload modals (`view/app/uploader_modal.tmpl` lines `112-185` etc.):** Upload forms target `#uploader-*-results`, while `handler/htmx/submiterr.go` (lines `45-78`) returns `400`/`500`/`503` HTML.

**Mitigation:** Set `htmx.config.noSwap = [204, 304, '4xx', '5xx']` globally to keep v2 error-skipping behavior, or handle error statuses explicitly with `hx-status`.

---

## 5. Smaller / Behavioral Items to Verify

* **60s default request timeout (`htmx.config.defaultTimeout = 60000`):** htmx 2 defaults to `0` (no timeout). Large file uploads (up to 100 MB per `assets/js/uploader.mjs` line `11`) or slow remote Demozoo/Pouët downloads (`handler/htmx/submitprod.go` lines `195-223`) can hit 60s on slow connections unless `htmx.config.defaultTimeout = 0` (or higher) is configured.
* **`PageReload` (`302 Found` + `Hx-Refresh: true`):** `PageReload` (`handler/htmx/artifact.go` lines `103-108`, called in `handler/htmx/artifactfs.go` lines `85`, `146`, `174`, `201`, `231`, `337`, `408`, `483`, `504`) returns HTTP `302` without a `Location` header alongside `Hx-Refresh: true`, and `assets/js/editor-assets.js` line `385` checks `xhr.status === 200 || xhr.status === 302`. Test these asset-editor actions under `fetch()` to verify the reload and feedback behavior.
* **`hx-swap="swap:50ms"` shorthand (`view/app/artifactfile.tmpl` line `370`):** Uses `swap:50ms` without specifying `innerHTML` first; verify htmx 4 accepts the timing modifier alone or update to `innerHTML swap:50ms`.
* **Explicit Inheritance (`:inherited`) Caught by `upgrade-check`:**
  * **`view/app/artifactedit.tmpl` (line `585`):** The `<form hx-patch="/editor/links" hx-target="#artifact-editor-link-examples">` encloses 7 `<input>` elements (`#artifact-editor-demozoo`, `#artifact-editor-pouet`, `#artifact-editor-16colors`, `#artifact-editor-github`, `#artifact-editor-youtube`, `#artifact-editor-relations`, `#artifact-editor-websites`) that each fire `hx-patch="/editor/..."` and return `HTMLLinkTo` preview HTML (`handler/htmx/artifacttx.go` line `200`) without their own `hx-target`. In htmx 2, they implicitly inherit `hx-target="#artifact-editor-link-examples"` from `<form>`. In htmx 4, line `595` **must** change to `hx-target:inherited="#artifact-editor-link-examples"` (and line `594` can use `hx-include:inherited="[name='artifact-editor-key']"` to avoid repeating `hx-include` on all 8 child controls).
  * **`view/app/uploader_modal.tmpl` (lines `111`, `200`, `293`, `393`, `478`, `564`):** The 6 `<form id="uploader-*-form">` elements set `hx-encoding="multipart/form-data"`, `hx-target="#uploader-*-results"`, and `hx-swap="innerHTML"`, while child typeahead inputs (`hx-patch="/uploader/releaser/1"`, etc.) already specify their own `hx-target="#uploader-*-list-1"` and use the default `innerHTML` swap. Those child typeahead requests do **not** need `multipart/form-data`, so leaving `:inherited` off the uploader `<form>` attributes is actually cleaner in htmx 4.
* **Likely safe without changes (verify during QA):**
  * **`hx-delete` inside forms:** The `hx-delete` buttons in `view/app/artifactedit.tmpl` (line `77`) and `view/app/artifactfile.tmpl` (lines `405`, `429`) encode IDs in the URL path rather than relying on enclosing form inputs.

---

## 6. Notes on `htmx.org@4` `upgrade-check` Output

Because the `htmx.org@4.0.0` npm package sets its default binary to the `upgrade-check` script directly, pass options straight to `npx htmx.org@4.0.0` (omitting the literal subcommand word `upgrade-check`, which would otherwise be parsed as a file path):

```bash
npx htmx.org@4.0.0 --no-color --ext .tmpl --ext .mjs --ext .go ./view ./assets/js ./handler
```

Running the full scan across all **238 files** (saved in [htmx4-upgrade-check.txt](htmx4-upgrade-check.txt)) reports **38 issues across 7 files**:
* `assets/js/editor-assets.js` (1 issue: `htmx:afterRequest`)
* `assets/js/htmx-response-targets.js` (2 issues: `htmx.defineExtension`, `htmx:beforeSwap`)
* `assets/js/layout-htmx.mjs` (4 issues: `htmx.logAll`, `htmx:beforeRequest`, `htmx:afterRequest`)
* `assets/js/uploader-submitter.mjs` (6 issues: `htmx:beforeRequest`, `htmx:beforeSwap`, `htmx:afterRequest`)
* `assets/js/uploader.mjs` (2 issues: `htmx:xhr:progress`)
* `view/app/artifactedit.tmpl` (3 issues: `hx-ext`, plus `:inherited` on `hx-include` and `hx-target` at line 585)
* `view/app/uploader_modal.tmpl` (20 issues: `hx-ext` at lines 30 & 71, plus `:inherited` notices on the 6 upload forms)

Even with all extensions included, static checks will still miss `event.detail.xhr` / `event.detail.elt` / `successful` / `failed` usage inside JS handlers and runtime default changes (`noSwap`, `defaultTimeout`, `htmx:xhr:progress` removal).

---

## 7. 6-Stage Browser-Testable Migration Plan

Because htmx 2 and htmx 4 cannot be loaded simultaneously on the same page (both bind `window.htmx` and scan `document.body` for `hx-*` attributes, which would fire every request twice), the cleanest way to manually test each UX area in the browser one by one—leaving the Uploader modals until last—is to switch the core script with v2-safe configuration defaults first, then migrate and verify each UX feature in order from simplest to most complex.

### Stage 1 — Swap Core Library & Test Declarative Pages (Zero Custom JS)

1. **Changes**:
   * Download `htmx.v4.min.js` (`htmx.org@4.0.0`) into `public/js/` and wire it in [handler/app/asset.go](../handler/app/asset.go) and [view/app/layout.tmpl](../view/app/layout.tmpl).
   * Add `<meta name="htmx-config" content='{"noSwap":[204,304,"4xx","5xx"],"defaultTimeout":0}'>` to `<head>` in [view/app/layout.tmpl](../view/app/layout.tmpl).
2. **Browser Verification**:
   * **Area Codes** (`/areacodes` — [view/app/areacodes.tmpl](../view/app/areacodes.tmpl)): Type an area code (`212`, `415`) and verify live results and spinner.
   * **Configurations** (`/editor/configurations` — [view/app/configurations.tmpl](../view/app/configurations.tmpl)): Click **Refresh the connections** and **Ping the routes**.
   * **Fixers** (`/editor/fixers` — [view/app/fixers.tmpl](../view/app/fixers.tmpl)): Test the **Fix** button swap.
   * **Approve Banner** ([view/app/artifactlock.tmpl](../view/app/artifactlock.tmpl)): Click **Approve** on a hidden artifact page. *(Note: `TxPublicByKey` in [handler/htmx/artifacttx.go](../handler/htmx/artifacttx.go) was updated from returning plain text `"Record is visible to the public."` to returning the Bootstrap `<span class="badge text-bg-success">Public</span>` fragment matching `artifactlock.tmpl`.)*

### Stage 2 — Live Search (`/search/releaser`, `/search/desc`, `/search/id`)

1. **Changes**:
   * In [assets/js/layout-htmx.mjs](../assets/js/layout-htmx.mjs), add the `parseHtmxEvent(event)` helper and update the listener from `htmx:afterRequest` → `htmx:after:request` for `search-htmx-input` / `search-htmx-alert`.
   * Recompile `public/js/layout.min.js` via `runner/runner.go`.
2. **Browser Verification**:
   * Open `/search/releaser`, `/search/desc`, and `/search/id`.
   * Type queries, test the **Clear** button (`#search-htmx-clear`), keyboard navigation, loading spinner, and error alert handling.

### Stage 3 — Artifact Editor: DOSBox / js-dos Emulation Modal

1. **Changes**:
   * [view/app/artifacteditjsdos.tmpl](../view/app/artifacteditjsdos.tmpl) uses standard `hx-patch` and `hx-on:change` (native DOM event), requiring no template edits.
2. **Browser Verification**:
   * Open a DOS artifact → **Site editor → Emulation settings**.
   * Toggle broken status, change machine/CPU/audio/RAM dropdowns, and verify the saved feedback text updates.

### Stage 4 — Artifact Editor: Metadata Modal & Delete Forever

1. **Changes**:
   * In [assets/js/layout-htmx.mjs](../assets/js/layout-htmx.mjs), update `htmx:beforeRequest` → `htmx:before:request` and update the artifact editor handlers (`afterRecord`, `afterUpdate`, `afterClassifications`, `afterUpdateRels`, `afterReset`, `afterUpdateDate`, `afterCreators`, `afterLinks`) to read from `event.detail.ctx`.
   * In [view/app/artifactedit.tmpl](../view/app/artifactedit.tmpl) (line `595`), change `hx-target="#artifact-editor-link-examples"` to `hx-target:inherited="#artifact-editor-link-examples"`.
   * In [view/app/artifactedit.tmpl](../view/app/artifactedit.tmpl) (lines `66–80`, "Delete forever" modal), replace `hx-ext="response-targets"` + `hx-target-error="#artifact-delete-forever-errors"` with `hx-status:4xx="target:#artifact-delete-forever-errors swap:innerHTML"` and `hx-status:5xx="target:#artifact-delete-forever-errors swap:innerHTML"`.
   * Recompile `public/js/layout.min.js` via `runner/runner.go`.
2. **Browser Verification**:
   * Open an artifact → **Edit data**:
     * Toggle **Public / Hidden** radio buttons (verify modal header border & badge change between green `Public` and red `Hidden`).
     * Click quick classification buttons & OS/Category dropdowns (verify `.is-valid` green checkmarks).
     * Edit and Undo title, filename, releasers, release date, credits, comment, and external links (verify `#artifact-editor-link-examples` updates live).

### Stage 5 — Artifact Editor: File & Image Assets Modal

1. **Changes**:
   * In [assets/js/editor-assets.js](../assets/js/editor-assets.js) (line `312`), update `htmx:afterRequest` → `htmx:after:request` and migrate `afterLinkRequest`, `afterRefreshRequest`, and `afterFormRequest` to read `event.detail.ctx.sourceElement`, `ctx.response.status`, and `ctx.text`.
   * In [view/app/artifactfile.tmpl](../view/app/artifactfile.tmpl) (line `370`), verify/update `hx-swap="swap:50ms"` to `hx-swap="innerHTML swap:50ms"`.
   * Recompile `public/js/editor-assets.min.js` via `runner/runner.go`.
2. **Browser Verification**:
   * Open an artifact → **Manage assets**:
     * Test README/DIZ/HLP copy, disable, and delete buttons.
     * Test image crop, align, thumbnail, and **Pixelate NSFW images** (verifying the `302` + `Hx-Refresh: true` page reload).

### Stage 6 (Last) — Uploader Modals, Demozoo/Pouët & Progress Bars

1. **Changes**:
   * **Demozoo & Pouët**: In [view/app/uploader_modal.tmpl](../view/app/uploader_modal.tmpl) (lines `30–88`) and [handler/htmx/submitprod.go](../handler/htmx/submitprod.go) (lines `140`, `163`), replace `hx-ext="response-targets"` + `hx-target-error` with `hx-status:4xx` / `hx-status:5xx`, and update [assets/js/uploader-submitter.mjs](../assets/js/uploader-submitter.mjs) events (`htmx:before:request`, `htmx:before:swap`, `htmx:after:request`) + `ctx`.
   * **6 File Upload Modals**: In [view/app/uploader_modal.tmpl](../view/app/uploader_modal.tmpl), update the 6 `hx-on::after-swap` attributes to `hx-on:htmx:after:swap`.
   * **Upload Progress Bars**: In [assets/js/uploader.mjs](../assets/js/uploader.mjs), replace `htmx:xhr:progress` with an indeterminate progress/spinner state (or dedicated XHR upload helper).
   * **Cleanup**: Delete `assets/js/htmx-response-targets.js` and remove `HtmxRespTargets` from [runner/runner.go](../runner/runner.go), [handler/app/asset.go](../handler/app/asset.go), [handler/app/app.go](../handler/app/app.go), [handler/app/template.go](../handler/app/template.go), and [view/app/layout.tmpl](../view/app/layout.tmpl).
2. **Browser Verification**:
   * Test Demozoo & Pouët valid and invalid IDs (`15`, `16`, `99999999`).
   * Test releaser/magazine typeahead datalists and file uploads across all 6 uploader modals.

