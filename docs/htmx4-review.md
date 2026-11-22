# htmx 4 Potential Benefits Review for Defacto2 Server

A review of the architectural and feature changes introduced in **htmx 4** ([What's New in htmx 4](https://four.htmx.org/docs/whats-new-in-htmx-4/), [htmx 4.0.0 Release Announcement](https://four.htmx.org/announcements/2026-08-28-htmx-4.0.0-is-released/), [htmx 4: What's New, What Breaks, Why It's Not latest](https://morello.dev/blog/htmx-4)) and how they specifically benefit the **Defacto2 `server`** codebase.

For details on breaking changes, file audits, and migration requirements, see [htmx4-update.md](htmx4-update.md) and [htmx4-upgrade-check.txt](htmx4-upgrade-check.txt).

---

## Executive Summary

Upgrading from htmx 2 to htmx 4 offers tangible benefits for `server`:
1. **Reduces pipeline complexity**: Drops the bundled third-party `response-targets` extension and its custom build/SRI steps in favor of native `hx-status`.
2. **Improves interactive editor UX**: Built-in Idiomorph support (`innerMorph` / `outerMorph`) preserves DOM focus and scroll position across metadata edits.
3. **Simplifies multi-element responses**: `<hx-partial>` enables Go handlers to update multiple distinct page targets cleanly in a single HTTP response without client-side event duct tape.
4. **Fixes history desync**: Removes `localStorage` DOM caching, preventing stale Bootstrap modal and dropdown state when navigating back/forward.
5. **Accelerates page browsing**: First-party `hx-preload` enables instant hover-based prefetching across file archives and directory lists.

---

## Detailed Benefits for `server`

### 1. Native Status Error Routing (`hx-status`) Replaces External Extension

* **Current State in `server`**:
  * `server` currently vendors and compiles [assets/js/htmx-response-targets.js](../assets/js/htmx-response-targets.js) into `public/js/htmx.v2.ext-response-targets.min.js`.
  * The build step in [runner/runner.go](../runner/runner.go) and integrity calculation in [handler/app/asset.go](../handler/app/asset.go) exist primarily to support `hx-target-error` in templates ([view/app/artifactedit.tmpl](../view/app/artifactedit.tmpl), [view/app/uploader_modal.tmpl](../view/app/uploader_modal.tmpl), and [handler/htmx/submitprod.go](../handler/htmx/submitprod.go)).
* **In htmx 4**:
  * Status code routing is built directly into core using `hx-status:4xx`, `hx-status:5xx`, or specific HTTP codes (`hx-status:422="target:#error-box"`).
* **Impact**:
  * Eliminates `assets/js/htmx-response-targets.js` entirely.
  * Removes an asset compilation target and SRI entry.
  * Eliminates runtime monkeypatching of htmx swap internals.

---

### 2. Built-in Idiomorph Swapping (`innerMorph` / `outerMorph`)

* **Current State in `server`**:
  * In the artifact editor ([view/app/artifactedit.tmpl](../view/app/artifactedit.tmpl)), inline updates to titles, dates, classifications, and releasers use standard `innerHTML` swaps.
  * Full element replacement can cause micro-flicker, drop active input focus, or reset scroll position.
* **In htmx 4**:
  * The Idiomorph algorithm is bundled into the core library, accessible with `hx-swap="innerMorph"` or `hx-swap="outerMorph"`.
* **Impact**:
  * Metadata edits, tags, and badge updates preserve active input focus, text cursor positions, and UI scroll state automatically, without loading external morphing libraries.

---

### 3. Clean Multi-Target Updates via `<hx-partial>`

* **Current State in `server`**:
  * Updating multiple elements in response to a single action (e.g., updating record metadata while simultaneously refreshing the header status badge, public/hidden pill, or timestamp) currently requires either out-of-band swaps (`hx-swap-oob`) or custom JavaScript hooks in [assets/js/layout-htmx.mjs](../assets/js/layout-htmx.mjs) (`recordSuccess`, `recordError`).
* **In htmx 4**:
  * Go handlers can return multiple `<hx-partial>` tags:
    ```html
    <hx-partial hx-target="#artifact-badge" hx-swap="outerHTML">
        <span class="badge bg-success">Online</span>
    </hx-partial>
    <hx-partial hx-target="#artifact-editor-date" hx-swap="innerMorph">
        <!-- updated date fields -->
    </hx-partial>
    ```
* **Impact**:
  * Each `<hx-partial>` explicitly controls its own target and swap strategy.
  * Eliminates imperative DOM manipulation in client JavaScript.

---

### 4. Elimination of `localStorage` History State Bugs

* **Current State in `server`**:
  * htmx 2 snapshots and serializes the DOM into browser `localStorage`.
  * For pages with active Bootstrap components (such as the new responsive navigation dropdowns, filter menus, and upload modals), restoring from `localStorage` on back navigation can leave orphan modal backdrops, frozen menus, or detached event listeners.
* **In htmx 4**:
  * `localStorage` DOM caching is dropped entirely.
  * Back/forward navigation re-fetches the page or restores cleanly into `<body>` / `[hx-history-elt]`.
* **Impact**:
  * Eliminates stale DOM state and Bootstrap backdrop/dropdown synchronization bugs.

---

### 5. Instant Browsing with First-Party `hx-preload`

* **In htmx 4**:
  * An official `hx-preload` extension is included in core extensions to preload content on mouse hover or touchstart:
    ```html
    <div hx-preload>
        <a href="/f/abc123" hx-boost="true">View File</a>
    </div>
    ```
* **Impact**:
  * For Defacto2's file listings, category indexes, groups, and people directories, links are prefetched before the user finishes clicking, making navigation feel near-instant.

---

### 6. Native `fetch()` & Streaming Architecture

* **In htmx 4**:
  * Replaces legacy `XMLHttpRequest` with modern `fetch()`.
  * Provides a consistent request context object (`event.detail.ctx`) across all lifecycle events.
  * Native error routing sends thrown exceptions and network failures to `console.error` with stack traces automatically.
* **Impact**:
  * Better aligns with modern browser networking (HTTP/2 and HTTP/3 multiplexing).
  * Lays the groundwork for streaming responses (SSE or multipart chunked HTML rendering).

---

### 7. Constructable Stylesheets for `.htmx-indicator` (Better CSP Security)

* **Current State in `server`**:
  * `server` uses `.htmx-indicator` spinners across [view/app/searchhtmx.tmpl](../view/app/searchhtmx.tmpl), [view/app/areacodes.tmpl](../view/app/areacodes.tmpl), [view/app/uploader_modal.tmpl](../view/app/uploader_modal.tmpl), [view/app/artifactfile.tmpl](../view/app/artifactfile.tmpl), and [view/app/configurations.tmpl](../view/app/configurations.tmpl).
  * In htmx 2, htmx injects an inline `<style>` tag into `<head>` for `.htmx-indicator`, which requires `'unsafe-inline'` or a style nonce under a Content Security Policy (CSP).
* **In htmx 4**:
  * Indicator CSS is applied via browser **Constructable Stylesheets** (`document.adoptedStyleSheets`), which do not require inline `<style>` tags or CSP nonces.
* **Impact**:
  * Makes it significantly easier to enforce a strict `style-src` Content Security Policy on `server`.

---

### 8. Safer & Simpler Swap Styles (`textContent` and `delete`)

* **In htmx 4**:
  * `hx-swap="textContent"` writes plain text directly to a target node without invoking the HTML parser.
  * `hx-swap="delete"` removes the target element cleanly on a successful response without relying on an empty body + `outerHTML`.
* **Impact**:
  * Endpoints in [handler/htmx/artifact.go](../handler/htmx/artifact.go) that return plain text strings (such as counts, status strings, or validation messages) can use `textContent` for faster updates and built-in XSS safety.
  * Item deletion actions (such as in [view/app/fixers.tmpl](../view/app/fixers.tmpl) or asset removals in [view/app/artifactfile.tmpl](../view/app/artifactfile.tmpl)) can use `hx-swap="delete"`.

---

### 9. Phased Upgrade Path via `htmx-2-compat`

* **In htmx 4**:
  * Ships an official `htmx-2-compat` core extension that temporarily restores htmx 2 defaults (`implicitInheritance = true`, `noSwap = [204, 304, '4xx', '5xx']`, `defaultTimeout = 0`) and maps legacy camelCase event names (`htmx:afterRequest` → `htmx:after:request`).
* **Impact**:
  * Allows `server` to perform a lower-risk, staged migration rather than an all-at-once cutover (though `htmx:xhr:progress` and `event.detail.ctx` still require direct code updates).

---

### 10. HTTP `QUERY` Support via `hx-query` (RFC 10008) & Consolidated `htmx:error`

* **`hx-query` for Body-Based Read Queries**:
  * htmx 4 adds first-class support for the new HTTP `QUERY` method (RFC 10008)—a safe, idempotent read request that carries a request body like `POST`.
  * In [view/app/searchhtmx.tmpl](../view/app/searchhtmx.tmpl) (line 43) and [view/app/areacodes.tmpl](../view/app/areacodes.tmpl) (line 22), `server` currently uses `hx-post` for read-only live searches to send form payloads in the request body. Once Echo and upstream proxies adopt `QUERY`, `hx-query` fits this exact pattern with proper read/caching semantics.
* **Consolidated Error Event (`htmx:error`)**:
  * Four separate error events in htmx 2 (`htmx:sendError`, `htmx:swapError`, `htmx:targetError`, and `htmx:timeout`) are collapsed into a single `htmx:error` event (alongside `htmx:response:error` for HTTP status errors), simplifying global client error handling.

---

### 11. Single-File `htmax.js` Distribution & Official LLM Migration Skills

* **`htmax.js` All-in-One Bundle**:
  * In addition to standard `htmx.min.js`, htmx 4 ships **`htmax.js`** (`htmax.min.js`), which bundles htmx core together with the most popular first-party extensions (`hx-preload`, `hx-live`, etc.) into a single file.
  * Currently, [handler/app/asset.go](../handler/app/asset.go), [handler/app/app.go](../handler/app/app.go), and [view/app/layout.tmpl](../view/app/layout.tmpl) manage two separate script files and SRI hashes (`Htmx` and `HtmxRespTargets`). Whether adopting `htmx.min.js` or `htmax.min.js`, `server` can simplify its asset and SRI configuration down to a single script tag.
* **Official htmx 4 Agent / LLM Skills**:
  * The htmx team publishes official LLM skills (`htmx-upgrade-from-htmx2`, `htmx-guidance`, `htmx-debugging`, and `htmx-extension-authoring`) along with `llms.txt` / `llms-full.txt` on `four.htmx.org` to assist with automated migration and debugging.
* **Minor Syntax & Header Enhancements**:
  * **Multi-selector `from:` quoting**: In [view/app/artifactedit.tmpl](../view/app/artifactedit.tmpl) (lines 587–593), the 7 repeated `input from:#...` triggers can optionally be condensed using HCON single-quoted selector lists (`hx-trigger="input from:'#artifact-editor-youtube, #artifact-editor-demozoo, ...'"`).
  * **New `HX-Request-Type` & `HX-Source` request headers**: Requests now send `HX-Request-Type: "partial" | "full"` and `HX-Source: tagName#id`.

---

## Release Strategy & Support Timeline (No Urgency to Migrate)

* **Indefinite htmx 2 Support**:
  * The htmx maintainers explicitly commit to supporting and maintaining **htmx 2 indefinitely** ("100-year web services" philosophy). There is no EOL clock or maintenance pressure forcing `server` off htmx 2.
* **npm Dist-Tag (`@next` vs `@latest`)**:
  * On npm, `htmx.org@latest` remains pinned to `2.x` (`2.0.10`) until early 2027 so unpinned CDN users are not broken. Installing v4 via npm before early 2027 requires the `next` dist-tag (`npm install htmx.org@next`) or downloading the explicit `4.0.0` release asset.

---

## Migration Considerations & Trade-offs

Before upgrading, the following items documented in [htmx4-update.md](htmx4-update.md) must be addressed:

| Item | Challenge in htmx 4 | Proposed Solution for `server` |
| :--- | :--- | :--- |
| **Upload Progress** | `htmx:xhr:progress` removed because `fetch()` lacks upload progress events. | Replace `<progress>` bars in upload modals with indeterminate loading indicators, or use a dedicated small XHR upload script. |
| **Event Properties** | Flat `event.detail.xhr` / `event.detail.elt` moved under `event.detail.ctx`. | Refactor listeners in `layout-htmx.mjs`, `uploader-submitter.mjs`, and `editor-assets.js` to read from `event.detail.ctx`. |
| **Default Swap Behavior** | `4xx` and `5xx` responses swap into targets by default. | Set `htmx.config.noSwap = [204, 304, '4xx', '5xx']` globally to preserve v2 behavior, or route errors explicitly with `hx-status`. |
| **Default Timeout** | Requests timeout at 60s by default (was 0/no timeout). | Set `htmx.config.defaultTimeout = 0` (or higher) to allow large file uploads on slower uplinks. |
| **Removed JS Helpers** | `htmx.logAll()` (commented out in `layout-htmx.mjs:13`) and `htmx.defineExtension()` removed. | Use `htmx.config.logAll = true` for debug logging and drop `htmx-response-targets.js`. |

