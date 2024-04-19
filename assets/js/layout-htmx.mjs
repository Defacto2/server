// layout-htmx.mjs

import { releaserEvents } from "./layout-htmx-search.mjs";

export default htmxEvents;

/**
 * Initializes the htmx event listeners.
 */
export function htmxEvents() {
  //htmx.logAll();
  releaserEvents();

  // This event is triggered after an AJAX request has finished.
  // https://htmx.org/events/#htmx:afterRequest
  document.body.addEventListener("htmx:afterRequest", function (event) {
    // search releaser.
    afterRequest(event, `search-releaser-input`, `search-releaser-alert`);
    // image uploader.
    afterRequest(event, `uploader-image-form`, `uploader-image-alert`);
    afterRequest(event, `uploader-image-releaser-1`, `uploader-image-alert`);
    afterRequest(event, `uploader-image-releaser-2`, `uploader-image-alert`);
    // intro uploader.
    afterRequest(event, `uploader-intro-form`, `uploader-intro-alert`);
    afterRequest(event, `uploader-intro-releaser-1`, `uploader-intro-alert`);
    afterRequest(event, `uploader-intro-releaser-2`, `uploader-intro-alert`);
    // text uploader.
    afterRequest(event, `uploader-text-form`, `uploader-text-alert`);
    afterRequest(event, `uploader-text-releaser-1`, `uploader-text-alert`);
    afterRequest(event, `uploader-text-releaser-2`, `uploader-text-alert`);
  });
}

/**
 * Handles the response after an htmx request.
 * Any error messages are displayed in the alert element.
 *
 * @param {Event} event - The htmx event object.
 * @param {string} inputId - The ID of the input element.
 * @param {string} alertId - The ID of the alert element.
 * @throws {Error} If the htmx alert element is null.
 */
function afterRequest(event, inputId, alertId) {
  if (event.detail.elt === null) return;
  if (event.detail.elt.id !== `${inputId}`) return;

  const alert = document.getElementById(alertId);
  if (typeof alert === "undefined" || alert === null) {
    throw new Error(`The htmx alert element ${alertId} is null`);
  }

  if (event.detail.successful) {
    return successful(event, alert);
  }
  if (event.detail.failed && event.detail.xhr) {
    return errorXhr(alert, event);
  }
  errorBrowser(alert);
}

/**
 * Handles the successful event.
 *
 * @param {Event} event - The event object.
 * @param {HTMLElement} alert - The alert element.
 */
function successful(event, alert) {
  alert.classList.add("d-none");
  alert.innerText = "";
  const match = "-form",
    id = event.target.id;
  const suffix = id.slice(-match.length);
  if (suffix == match) {
    const select = id.replace(match, "-file");
    resetFile(event, `#${select}`);
  }
}

/**
 * Resets the value of the file input element.
 * @param {Event} event - The event object.
 * @param {string} selector - The selector of the file input element.
 */
function resetFile(event, selector) {
  const input = event.target.querySelector(selector);
  if (input) {
    input.value = "";
    input.innerText = "";
    return;
  }
  console.error(`The reset file ${selector} element is null`);
}

/**
 * Handles the error response from an XHR request.
 *
 * @param {HTMLElement} alert - The alert element to display the error message.
 * @param {CustomEvent} event - The event object containing the XHR details.
 */
function errorXhr(alert, event) {
  const xhr = event.detail.xhr;
  alert.innerText = `Something on the server is not working, ${xhr.status} status: ${xhr.responseText}.`;
  alert.classList.remove("d-none");
}

/**
 * Displays an error message usually caused by the browser.
 * @param {HTMLElement} alert - The alert element where the error message will be displayed.
 */
function errorBrowser(alert) {
  alert.innerText =
    "Something with the browser is not working, please try again or refresh the page.";
  alert.classList.remove("d-none");
}
