# Files Referencing or Using htmx in `server`

Complete inventory of every file in the repository that uses, serves, tests, builds, or references **htmx** (including `hx-*` attributes, `htmx:*` events, `htmx.*` JS APIs, `Hx-*` HTTP headers, and the `handler/htmx` / `view/htmx` packages).

---

## 1. Frontend JavaScript Source (`assets/js/`)

| File | How it uses or references htmx |
| :--- | :--- |
| [assets/js/editor-artifact.js](../assets/js/editor-artifact.js) | Calls `htmx.trigger(saveRels, 'submit-releasers')` (lines 348, 369) when modifying or clearing releasers in the artifact editor. |
| [assets/js/editor-assets.js](../assets/js/editor-assets.js) | Calls `progress(...)` for download/preview upload forms (lines 18–19); listens for `htmx:afterRequest` on `document.body` (line 312) and inspects `event.detail` / `xhr` in `afterLinkRequest`, `afterRefreshRequest`, and `afterFormRequest`. |
| [assets/js/htmx-response-targets.js](../assets/js/htmx-response-targets.js) | Source copy of the `response-targets` extension (`htmx.defineExtension("response-targets", ...)`), hooking `htmx:beforeSwap` to route error status codes via `hx-target-error` / `hx-target-*`. |
| [assets/js/layout-htmx-search.mjs](../assets/js/layout-htmx-search.mjs) | Manages live search input, clear button, results container, and `.htmx-request` indicator class for `#search-htmx-*`. |
| [assets/js/layout-htmx.mjs](../assets/js/layout-htmx.mjs) | Registers global `htmx:beforeRequest` and `htmx:afterRequest` listeners on `document.body` (lines 17, 37) to handle search feedback, uploader modal feedback, and artifact editor inline updates/errors. |
| [assets/js/layout-keyboard.mjs](../assets/js/layout-keyboard.mjs) | References htmx live search results in keyboard navigation comments/handling (line 47). |
| [assets/js/layout.js](../assets/js/layout.js) | Imports and initializes `htmxEvents()` from `layout-htmx.mjs` (lines 9, 20). |
| [assets/js/uploader-submitter.mjs](../assets/js/uploader-submitter.mjs) | Listens for `htmx:beforeRequest`, `htmx:beforeSwap`, and `htmx:afterRequest` (lines 55, 61, 70) to handle Demozoo and Pouët production lookups and error messages. |
| [assets/js/uploader.js](../assets/js/uploader.js) | Entry point for uploader modals; wires `submitter(...)` (`uploader-submitter.mjs`) and `progress(...)` (`uploader.mjs`) for all 6 upload forms. |
| [assets/js/uploader.mjs](../assets/js/uploader.mjs) | Implements `progress(formId, elementId)` using `htmx.on('#...', 'htmx:xhr:progress', ...)` and `htmx.find(...)` (lines 119–137). |

---

## 2. HTML Templates (`view/`)

### Application Views (`view/app/`)

| File | How it uses or references htmx |
| :--- | :--- |
| [view/app/areacodes.tmpl](../view/app/areacodes.tmpl) | Live telephone area code search input using `hx-post="/areacodes"`, `hx-trigger="input changed delay:150ms, search"`, `hx-target="#areacodes-results"`, and `hx-indicator="#areacodes-indicator"`. |
| [view/app/artifactedit.tmpl](../view/app/artifactedit.tmpl) | Heavy htmx usage (~165 occurrences): inline `hx-patch` editors for classifications, dates, releasers, title, filename, VirusTotal, credits, comments, and external links; `hx-delete` with `hx-ext="response-targets"` and `hx-target-error` for permanent record deletion. |
| [view/app/artifacteditjsdos.tmpl](../view/app/artifacteditjsdos.tmpl) | Inline `hx-patch` controls (`hx-trigger="change"`, `hx-on:change`, `hx-swap="innerHTML"`, `hx-include`) for DOSBox / js-dos emulation settings (broken toggle, run program, machine, CPU, sound, RAM). |
| [view/app/artifactfile.tmpl](../view/app/artifactfile.tmpl) | Asset management controls using `hx-post`, `hx-patch`, and `hx-delete` (with `hx-encoding="multipart/form-data"`, `hx-confirm`, `hx-validate`, `hx-on:click`, `hx-swap`, `hx-indicator`) for replacing downloads, previews, thumbnails, pixelating images, and extracting README/DIZ/HLP files. |
| [view/app/artifactlock.tmpl](../view/app/artifactlock.tmpl) | Approve button using `hx-get="/editor/online/true/{{$modID}}"` and `hx-target="#record-is-hidden"` (line 26). |
| [view/app/configurations.tmpl](../view/app/configurations.tmpl) | Admin diagnostics buttons using `hx-get="/editor/configurations/dbconns"` and `hx-get="/editor/configurations/pings"` with child `.htmx-indicator` spinners (lines 28–34). |
| [view/app/fixers.tmpl](../view/app/fixers.tmpl) | Fix buttons using `hx-post="/editor/fixers/fix/{{ .ObfuscatedID }}"`, `hx-target="closest .list-group-item"`, and `hx-swap="outerHTML"` (line 25). |
| [view/app/layout.tmpl](../view/app/layout.tmpl) | Loads `<script src="{{ htmx }}">` and `<script defer src="{{ htmxRespTargets }}">` with SRI integrity hashes (lines 384–386). |
| [view/app/searchhtmx.tmpl](../view/app/searchhtmx.tmpl) | Live search page template for releasers and IDs (`hx-post`, `hx-trigger="input changed delay:500ms, search"`, `hx-target="#search-htmx-results"`, `hx-indicator="#search-htmx-indicator"`). |
| [view/app/uploader_modal.tmpl](../view/app/uploader_modal.tmpl) | Upload modals (~95 occurrences): Demozoo and Pouët lookup forms (`hx-ext="response-targets"`, `hx-get`, `hx-target-error`, `hx-validate`) and 6 file upload forms (`hx-post`, `hx-encoding="multipart/form-data"`, `hx-on::after-swap`, plus `hx-patch` releaser/magazine datalist typeaheads). |

### Partial / Fragment Views (`view/htmx/`)

| File | How it uses or references htmx |
| :--- | :--- |
| [view/htmx/datalistreleasers.tmpl](../view/htmx/datalistreleasers.tmpl) | Partial HTML `<option>` list returned for htmx releaser typeahead datalists. |
| [view/htmx/layout.tmpl](../view/htmx/layout.tmpl) | Empty layout wrapper template used when rendering partial HTML responses for htmx AJAX requests. |
| [view/htmx/searchids.tmpl](../view/htmx/searchids.tmpl) | Partial HTML fragment rendering ID/UUID search results for `#search-htmx-results`. |
| [view/htmx/searchreleasers.tmpl](../view/htmx/searchreleasers.tmpl) | Partial HTML fragment rendering releaser search results for `#search-htmx-results`. |

---

## 3. Go Backend Handlers & Routers (`handler/`)

### Top-Level Router & Handler Wiring (`handler/`)

| File | How it uses or references htmx |
| :--- | :--- |
| [handler/handler.go](../handler/handler.go) | Imports `github.com/Defacto2/server/handler/htmx` and delegates area code lookups (`htmx.AreaCodeLookup`) and public/hidden toggles (`htmx.RecordPublic`). |
| [handler/router.go](../handler/router.go) | Imports `handler/htmx` and registers htmx routes via `serv.routesHtmx(sl, e)`. |
| [handler/routerhtmx.go](../handler/routerhtmx.go) | Defines public and read-only htmx routes (`/areacodes`, `/demozoo/production`, `/pouet/production`, `/search/releaser`, `/search/id`, `/uploader/releaser/:id`, `/uploader/magazine`). |
| [handler/routerlock.go](../handler/routerlock.go) | Defines authenticated `/editor/*` and `/uploader/*` htmx routes (`GET`, `POST`, `PUT`, `PATCH`, `DELETE`) calling into the `handler/htmx` package. |

### Application Package (`handler/app/`)

| File | How it uses or references htmx |
| :--- | :--- |
| [handler/app/app.go](../handler/app/app.go) | Computes and stores Subresource Integrity (SRI) hashes for `Htmx` and `HtmxRespTargets` (`s.Htmx`, `s.HtmxRespTargets`). |
| [handler/app/apps_internal_test.go](../handler/app/apps_internal_test.go) | Tests rendered radio buttons for `hx-swap="none"` and `hx-sync="closest .row:replace"` attributes (lines 213–216). |
| [handler/app/asset.go](../handler/app/asset.go) | Defines asset enum constants `Htmx` (`/js/htmx.v2.min.js`) and `HtmxRespTargets` (`/js/htmx.v2.ext-response-targets.min.js`). |
| [handler/app/context.go](../handler/app/context.go) | Defines and renders the `searchhtmx` template for `/search/desc` and `/search/id` pages. |
| [handler/app/template.go](../handler/app/template.go) | Registers `searchhtmx.tmpl`, template helpers `htmx`, `htmxRespTargets`, `sri_htmx`, `sri_htmxRespTargets`, and generates `hx-patch`/`hx-include`/`hx-swap`/`hx-sync` radio markup for `#artifact-editor-public` and `#artifact-editor-hidden`. |
| [handler/app/template_test.go](../handler/app/template_test.go) | Tests `searchhtmx` and `areacodes` template rendering for expected `hx-post` and `hx-trigger` attributes. |

### Dedicated `htmx` Handler Package (`handler/htmx/`)

| File | How it uses or references htmx |
| :--- | :--- |
| [handler/htmx/artifact.go](../handler/htmx/artifact.go) | Core artifact editor htmx endpoint handlers (classifications, VirusTotal, online/offline state, `DeleteForever`, and `PageReload` setting `Hx-Refresh: true`). |
| [handler/htmx/artifact_test.go](../handler/htmx/artifact_test.go) | Unit tests for `handler/htmx/artifact.go`, `artifactfs.go`, and `artifacttx.go`. |
| [handler/htmx/artifactfs.go](../handler/htmx/artifactfs.go) | Filesystem asset mutation handlers for htmx editor actions (copy/remove DIZ, HLP, README, repack zip, clear cache, crop/align/thumb/pixelate preview images). |
| [handler/htmx/artifacttx.go](../handler/htmx/artifacttx.go) | Database transaction handlers for inline `hx-patch` artifact edits (credits, Demozoo, Pouët, 16colors, GitHub, relations, websites, YouTube, filename, title, comment, releasers, dates, and `HTMLLinkTo` previews). |
| [handler/htmx/htmx.go](../handler/htmx/htmx.go) | General htmx handlers for area code lookup, datalist typeaheads (`DatalistReleasers`, `DatalistMagazines`), live searches (`SearchReleaser`, `SearchByID`), DB connection stats, and route pings. |
| [handler/htmx/htmx_test.go](../handler/htmx/htmx_test.go) | Unit tests for `handler/htmx/htmx.go`. |
| [handler/htmx/submit.go](../handler/htmx/submit.go) | Handles multipart file uploads submitted by the 6 htmx uploader modal forms and artifact editor download/preview forms. |
| [handler/htmx/submit_test.go](../handler/htmx/submit_test.go) | Unit tests for `handler/htmx/submit.go`, `submiterr.go`, and `submitprod.go`. |
| [handler/htmx/submiterr.go](../handler/htmx/submiterr.go) | Formats HTML error fragments and HTTP status codes returned to htmx uploader targets. |
| [handler/htmx/submitprod.go](../handler/htmx/submitprod.go) | Fetches Demozoo and Pouët API data and renders HTML preview cards containing `<button hx-put="..." hx-indicator="..." hx-target="..." hx-swap="innerHTML" hx-trigger="click once delay:500ms" hx-target-error="...">`. |
| [handler/htmx/template.go](../handler/htmx/template.go) | Configures HTML template rendering for the `view/htmx/` partial templates. |

### Internal Helpers (`handler/internal/`)

| File | How it uses or references htmx |
| :--- | :--- |
| [handler/internal/filerecord/filerecord.go](../handler/internal/filerecord/filerecord.go) | Generates classification `<button>` elements with `hx-indicator`, `hx-target`, and `hx-patch` attributes (lines 266–273). |

---

## 4. Build Tooling, Linting & Static Assets (`runner/`, `public/`, Root)

| File | How it uses or references htmx |
| :--- | :--- |
| [eslint.config.mjs](../eslint.config.mjs) | Declares `htmx: "readonly"` in ESLint browser globals for `assets/js/**/*.js` and `assets/js/**/*.mjs` (line 21). |
| [runner/runner.go](../runner/runner.go) | Minifies `assets/js/htmx-response-targets.js` into `public/js/htmx-response-targets.min.js` (lines 33, 172–175, 230). |
| [public/UPDATE.md](../public/UPDATE.md) | Documents the `curl` command for updating `public/js/htmx.v2.min.js` from jsDelivr (lines 17–21). |
| [public/js/htmx.v2.min.js](../public/js/htmx.v2.min.js) | Active bundled htmx v2.0.10 library served in production. |
| [public/js/htmx.v2.ext-response-targets.min.js](../public/js/htmx.v2.ext-response-targets.min.js) | Active bundled htmx v2 `response-targets` extension served in production. |
| [public/js/htmx-response-targets.min.js](../public/js/htmx-response-targets.min.js) | Minified output generated by `runner/runner.go` from `assets/js/htmx-response-targets.js`. |
| [public/js/htmx.v1.min.js](../public/js/htmx.v1.min.js) | Legacy htmx v1 minified bundle retained in `public/js/`. |
| [public/js/htmx.v1-response-targets.min.js](../public/js/htmx.v1-response-targets.min.js) | Legacy htmx v1 response-targets bundle retained in `public/js/`. |
| [public/js/editor-artifact.min.js](../public/js/editor-artifact.min.js) | Compiled & minified bundle containing `assets/js/editor-artifact.js`. |
| [public/js/editor-assets.min.js](../public/js/editor-assets.min.js) | Compiled & minified bundle containing `assets/js/editor-assets.js` and `uploader.mjs`. |
| [public/js/layout.min.js](../public/js/layout.min.js) | Compiled & minified bundle containing `layout.js`, `layout-htmx.mjs`, and `layout-htmx-search.mjs`. |
| [public/js/uploader.min.js](../public/js/uploader.min.js) | Compiled & minified bundle containing `uploader.js`, `uploader.mjs`, and `uploader-submitter.mjs`. |
| [handler/coverage.out](../handler/coverage.out) | Generated Go test coverage output referencing `handler/routerhtmx.go`. |
| [handler/htmx/coverage.out](../handler/htmx/coverage.out) | Generated Go test coverage output for `handler/htmx/*`. |

---

## 5. Documentation (`docs/`)

| File | Description |
| :--- | :--- |
| [docs/README.md](README.md) | Mentions HTMX in the Architecture overview (line 42). |
| [docs/htmx-list.md](htmx-list.md) | This file inventory of all htmx references across `server`. |
| [docs/htmx4-review.md](htmx4-review.md) | Review of htmx 4 features, architectural benefits, and release strategy for `server`. |
| [docs/htmx4-update.md](htmx4-update.md) | Detailed breaking changes audit and `event.detail.ctx` migration cheat sheet for upgrading from htmx v2 to v4. |
| [docs/htmx4-upgrade-check.txt](htmx4-upgrade-check.txt) | Output of `npx htmx.org@4.0.0 --no-color --ext .tmpl --ext .mjs --ext .go ./view ./assets/js ./handler`. |

---

## 6. Non-htmx False Positive

* [handler/html3/ext/ext.go](../handler/html3/ext/ext.go) (lines 55, 100): Defines a local constant `htmx = ".html"` alongside `htm = ".htm"` for matching `.html` file extensions; unrelated to the htmx library.
