# htmx v2 to v4 Migration Audit

Audit of potential breaking changes and risks when upgrading the site from **htmx v2** to **htmx 4.0.0** (released 2026-08-28, reference: https://four.htmx.org/announcements/2026-08-28-htmx-4.0.0-is-released/).

No code changes have been made; this document records what needs attention before or during an upgrade.

---

## 1. Event Names & `event.detail.xhr` Rewrite (`XMLHttpRequest` → `fetch()`)

In htmx 4, `XMLHttpRequest` is replaced by `fetch()`:

* Lifecycle events are renamed to `htmx:phase:action` (`htmx:beforeRequest` → `htmx:before:request`, `htmx:afterRequest` → `htmx:after:request`, `htmx:beforeSwap` → `htmx:before:swap`, `htmx:afterSwap` → `htmx:after:swap`).
* Flat properties (`event.detail.elt`, `event.detail.xhr`, `event.detail.successful`, `event.detail.failed`) move under the unified request context `event.detail.ctx` (e.g., `ctx.sourceElement`, `ctx.response`, `ctx.target`, `ctx.text`).

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
* **Likely safe without changes (verify during QA):**
  * **Explicit inheritance (`:inherited`):** No global `hx-headers`, `hx-confirm`, or `hx-target` wrappers were found in `view/` outside `hx-ext="response-targets"`.
  * **`hx-delete` inside forms:** The `hx-delete` buttons in `view/app/artifactedit.tmpl` (line `77`) and `view/app/artifactfile.tmpl` (lines `405`, `429`) encode IDs in the URL path rather than relying on enclosing form inputs.

---

## 6. Notes on `htmx.org@4` `upgrade-check` Output

Running the v4 upgrade checker initially produced `docs/htmx-upgrade-check.txt` (`Scanning 19 file(s)...`):

```text
./assets/js/editor-assets.js:312: [old-event] old event name "htmx:afterRequest" → "htmx:after:request"
./assets/js/htmx-response-targets.js:75: [old-api] htmx.defineExtension() is removed → use htmx.registerExtension()
./assets/js/htmx-response-targets.js:100: [old-event] old event name "htmx:beforeSwap" → "htmx:before:swap"
```

That run only scanned 19 default files (`.js`/`.html`) and skipped the repository's **15 `.mjs` files**, **65 `.tmpl` files**, and **`.go` handlers**. To include all relevant files:

```bash
npx htmx.org@4.0.0 upgrade-check --ext .tmpl --ext .mjs --ext .go ./view ./assets/js ./handler/htmx
```

Even with all extensions included, static checks will still miss `event.detail.xhr` / `event.detail.elt` / `successful` / `failed` usage inside JS handlers and runtime default changes (`noSwap`, `defaultTimeout`, `htmx:xhr:progress` removal).
