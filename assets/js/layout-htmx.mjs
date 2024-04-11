// layout-htmx.mjs

import { releaserInit, releaser } from "./layout-htmx-search.mjs";

export default htmxLoader;

/**
 * Initializes the htmx event listeners.
 */
export function htmxLoader() {
  //htmx.logAll();
  releaserInit();

  // This event is triggered after an AJAX request has finished.
  // https://htmx.org/events/#htmx:afterRequest
  document.body.addEventListener("htmx:afterRequest", function (event) {
    releaser(event);
  });
}
