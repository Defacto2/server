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
    // record toggle.
    afterRecord(
      event,
      `artifact-editor-hidden`,
      `artifact-editor-public`,
      `artifact-editor-alert`
    );
    afterRecord(
      event,
      `artifact-editor-public`,
      `artifact-editor-hidden`,
      `artifact-editor-alert`
    );
    // record classification.
    afterClassification(
      event,
      `artifact-editor-operating-system`,
      `artifact-editor-alert`
    );
    afterClassification(
      event,
      `artifact-editor-category`,
      `artifact-editor-alert`
    );
  });
}

function afterClassification(event, inputId, alertId) {
  if (event.detail.elt === null) return;
  if (event.detail.elt.id !== `${inputId}`) return;

  const liveAlert = document.getElementById(alertId);
  if (typeof liveAlert === "undefined" || liveAlert === null) {
    throw new Error(`The htmx alert element ${alertId} is null`);
  }

  if (event.detail.successful) {
    return;
  }
  if (event.detail.failed && event.detail.xhr) {
    return classificationError(event, liveAlert);
  }
  errorBrowser(liveAlert);
}

function classificationError(event, alertElm) {
  const xhr = event.detail.xhr;
  alertElm.innerText = `${timeNow()} Could not update the database record, ${xhr.responseText}.`;
  alertElm.classList.remove("d-none");
}

/**
 * Handles the logic after a record event.
 *
 * @param {Event} event - The event object.
 * @param {string} inputId - The ID of the input element.
 * @param {string} revertId - The ID of the revert element.
 * @param {string} alertId - The ID of the alert element.
 * @returns {void}
 * @throws {Error} If the htmx alert element is null.
 */
function afterRecord(event, inputId, revertId, alertId) {
  if (event.detail.elt === null) return;
  if (event.detail.elt.id !== `${inputId}`) return;

  const liveAlert = document.getElementById(alertId);
  if (typeof liveAlert === "undefined" || liveAlert === null) {
    throw new Error(`The htmx alert element ${alertId} is null`);
  }

  if (event.detail.successful) {
    return recordSuccess(event, inputId, liveAlert);
  }
  if (event.detail.failed && event.detail.xhr) {
    return recordError(event, revertId, liveAlert);
  }
  errorBrowser(liveAlert);
}

/**
 * Handles the successful event.
 *
 * @param {Event} event - The event object.
 * @param {HTMLElement} alertElm - The alert element.
 */
function recordSuccess(event, inputId, alertElm) {
  alertElm.classList.add("d-none");
  alertElm.innerText = "";
  const elm = document.getElementById(`artifact-editor-modal-header`);
  switch (inputId) {
    case "artifact-editor-hidden":
      elm.classList.remove("bg-success-subtle");
      elm.classList.add("bg-danger-subtle");
      break;
    case "artifact-editor-public":
      elm.classList.add("bg-success-subtle");
      elm.classList.remove("bg-danger-subtle");
      break;
    default:
      console.error(`The record success ${inputId} is not supported.`);
  }
}

/**
 * Handles the error response from an XHR request.
 *
 * @param {CustomEvent} event - The event object containing the XHR details.
 * @param {HTMLElement} alertElm - The alert element to display the error message.
 */
function recordError(event, revertId, alertElm) {
  const xhr = event.detail.xhr;
  alertElm.innerText = `${timeNow()} Could not update the database record, ${xhr.responseText}.`;
  alertElm.classList.remove("d-none");
  document.getElementById(revertId).checked = true;
}

function timeNow() {
  let now = new Date();
  let hours = now.getHours();
  let minutes = now.getMinutes();
  let seconds = now.getSeconds();
  minutes = (minutes < 10 ? "0" : "") + minutes;
  seconds = (seconds < 10 ? "0" : "") + seconds;
  return hours + ":" + minutes + ":" + seconds;
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

  const liveAlert = document.getElementById(alertId);
  if (typeof liveAlert === "undefined" || liveAlert === null) {
    throw new Error(`The htmx alert element ${alertId} is null`);
  }

  if (event.detail.successful) {
    return successful(event, liveAlert);
  }
  if (event.detail.failed && event.detail.xhr) {
    return errorXhr(liveAlert, event);
  }
  errorBrowser(liveAlert);
}

/**
 * Handles the successful event.
 *
 * @param {Event} event - The event object.
 * @param {HTMLElement} alertElm - The alert element.
 */
function successful(event, alertElm) {
  alertElm.classList.add("d-none");
  alertElm.innerText = "";
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
 * @param {HTMLElement} alertElm - The alert element to display the error message.
 * @param {CustomEvent} event - The event object containing the XHR details.
 */
function errorXhr(alertElm, event) {
  const xhr = event.detail.xhr;
  alertElm.innerText = `Something on the server is not working, ${xhr.status} status: ${xhr.responseText}.`;
  alertElm.classList.remove("d-none");
}

/**
 * Displays an error message usually caused by the browser.
 * @param {HTMLElement} alertElm - The alert element where the error message will be displayed.
 */
function errorBrowser(alertElm) {
  alertElm.innerText =
    "Something with the browser is not working, please try again or refresh the page.";
  alertElm.classList.remove("d-none");
}
