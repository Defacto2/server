// layout-htmx.mjs

import { releaserInit, releaser } from "./layout-htmx-search.mjs";

export default htmxLoader;

/**
 * Initializes the htmx event listeners.
 */
export function htmxLoader() {
  //htmx.logAll();

  releaserInit();
  uploadProgress();

  // Triggered when new content is added to the DOM.
  // https://htmx.org/events/#htmx:load
  // document.body.addEventListener("htmx:load", function () {
  //   uploadProgress();
  // });

  // This event is triggered after an AJAX request has finished.
  // https://htmx.org/events/#htmx:afterRequest
  document.body.addEventListener("htmx:afterRequest", function (event) {
    releaser(event);
  });
}

function uploadProgress() {
  htmx.on("#uploader-intro-form", "htmx:xhr:progress", function (event) {
    console.log("htmx:xhr:progress event fired");
    htmx
      .find("#uploader-intro-progress")
      .setAttribute("value", (event.detail.loaded / event.detail.total) * 100);
  });
}
