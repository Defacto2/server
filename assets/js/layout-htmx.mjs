// layout-htmx.mjs

import { releaserInit } from "./layout-htmx-search.mjs";

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
    afterReleaser(event, `search-releaser-input`, `search-releaser-alert`);
    afterReleaser(event, `uploader-intro-releaser-1`, `uploader-intro-alert`);
    afterReleaser(event, `uploader-intro-releaser-2`, `uploader-intro-alert`);
  });
}

/**
 * Handles the after request, search a releaser input event.
 *
 * @param {Event} event - The event object.
 * @throws {Error} If the htmx alert element is null.
 */
export function afterReleaser(event, inputId, alertId) {
  if (event.detail.elt === null) return;
  if (event.detail.elt.id !== `${inputId}`) return;

  const alert = document.getElementById(alertId);
  if (typeof alert === "undefined" || alert === null) {
    throw new Error(
      `The htmx alert element ${alertId} for releaser search is null`
    );
  }

  if (event.detail.successful) {
    return successful(alert);
  }
  if (event.detail.failed && event.detail.xhr) {
    return errorXhr(alert, event);
  }
  errorBrowser(alert);
}

function successful(alert) {
  alert.classList.add("d-none");
  alert.innerText = "";
}

function errorXhr(alert, event) {
  const xhr = event.detail.xhr;
  alert.innerText = `Something on the server is not working, ${xhr.status} status: ${xhr.responseText}.`;
  alert.classList.remove("d-none");
}

function errorBrowser(alert) {
  alert.innerText =
    "Something with the browser is not working, please refresh the page.";
  alert.classList.remove("d-none");
}
