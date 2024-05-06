// layout-htmx.mjs

import { releaserEvents } from "./layout-htmx-search.mjs";

export default htmxEvents;

/**
 * Initializes the htmx event listeners.
 */
export function htmxEvents() {
  //htmx.logAll();
  releaserEvents();

  document.body.addEventListener("htmx:beforeRequest", function (event) {
    removeSelectsValid(event, `artifact-editor-reset-classifications`);
    removeSelectsValid(event, `artifact-editor-text-for-dos`);
    removeSelectsValid(event, `artifact-editor-text-for-amiga`);
    removeSelectsValid(event, `artifact-editor-proof-of-release`);
    removeSelectsValid(event, `artifact-editor-intro-for-dos`);
    removeSelectsValid(event, `artifact-editor-intro-for-win`);
    removeSelectsValid(event, `artifact-editor-intro-for-bbs`);
    removeSelectsValid(event, `artifact-editor-trainer-for-dos`);
    removeSelectsValid(event, `artifact-editor-trainer-for-win`);
    removeSelectsValid(event, `artifact-editor-ansi-for-bbs`);
    removeSelectsValid(event, `artifact-editor-magazine-for-text`);
    removeSelectsValid(event, `artifact-editor-magazine-for-dos`);
  });
  // This event is triggered after an AJAX request has finished.
  // https://htmx.org/events/#htmx:afterRequest
  document.body.addEventListener("htmx:afterRequest", function (event) {
    // search releaser.
    afterRequest(event, `search-releaser-input`, `search-releaser-alert`);
    // image uploader.
    const alertImg = `uploader-image-alert`;
    afterRequest(event, `uploader-image-form`, alertImg);
    afterRequest(event, `uploader-image-releaser-1`, alertImg);
    afterRequest(event, `uploader-image-releaser-2`, alertImg);
    // intro uploader.
    const alertIntro = `uploader-intro-alert`;
    afterRequest(event, `uploader-intro-form`, alertIntro);
    afterRequest(event, `uploader-intro-releaser-1`, alertIntro);
    afterRequest(event, `uploader-intro-releaser-2`, alertIntro);
    // text uploader.
    const alertText = `uploader-text-alert`;
    afterRequest(event, `uploader-text-form`, alertText);
    afterRequest(event, `uploader-text-releaser-1`, alertText);
    afterRequest(event, `uploader-text-releaser-2`, alertText);
    // record toggle.
    afterRecord(event, `artifact-editor-hidden`, `artifact-editor-public`);
    afterRecord(event, `artifact-editor-public`, `artifact-editor-hidden`);
    // record classification.
    afterUpdate(event, `artifact-editor-operating-system`);
    afterUpdate(event, `artifact-editor-category`);
    afterClassifications(event, `artifact-editor-text-for-dos`);
    afterClassifications(event, `artifact-editor-text-for-amiga`);
    afterClassifications(event, `artifact-editor-proof-of-release`);
    afterClassifications(event, `artifact-editor-intro-for-dos`);
    afterClassifications(event, `artifact-editor-intro-for-win`);
    afterClassifications(event, `artifact-editor-intro-for-bbs`);
    afterClassifications(event, `artifact-editor-trainer-for-dos`);
    afterClassifications(event, `artifact-editor-trainer-for-win`);
    afterClassifications(event, `artifact-editor-ansi-for-bbs`);
    afterClassifications(event, `artifact-editor-magazine-for-text`);
    afterClassifications(event, `artifact-editor-magazine-for-dos`);
    // record releaser.
    afterUpdate(event, `artifact-editor-releaser-reset`);
    afterUpdateRels(event, `artifact-editor-releaser-update`);
    // record title.
    afterUpdate(event, `artifact-editor-title`);
    afterReset(event, `artifact-editor-title-reset`, `artifact-editor-title`);
    // record filename.
    afterUpdate(event, `artifact-editor-filename`);
    afterReset(
      event,
      `artifact-editor-filename-reset`,
      `artifact-editor-filename`
    );
    // record virustotal.
    afterUpdate(event, `artifact-editor-virustotal`);
    // record date.
    afterUpdate(event, `artifact-editor-date-reset`);
    afterUpdate(event, `artifact-editor-date-lastmod`);
    afterUpdateDate(event, `artifact-editor-date-update`);
    // record creators.
    afterUpdate(event, `artifact-editor-credit-text`);
    afterUpdate(event, `artifact-editor-credit-ill`);
    afterUpdate(event, `artifact-editor-credit-prog`);
    afterUpdate(event, `artifact-editor-credit-audio`);
    afterCreators(event, `artifact-editor-credit-resetter`);
    // record comment.
    afterUpdate(event, `artifact-editor-comment`);
    afterReset(
      event,
      `artifact-editor-comment-reset`,
      `artifact-editor-comment`
    );
    // record links.
    afterUpdate(event, `artifact-editor-youtube`);
    afterUpdate(event, `artifact-editor-demozoo`);
    afterUpdate(event, `artifact-editor-pouet`);
    afterUpdate(event, `artifact-editor-16colors`);
    afterUpdate(event, `artifact-editor-github`);
    afterUpdate(event, `artifact-editor-relations`);
    afterUpdate(event, `artifact-editor-websites`);
  });
}

function afterCreators(event, buttonId) {
  if (event.detail.elt === null) return;
  if (event.detail.elt.id !== `${buttonId}`) return;

  const alertId = "artifact-editor-alert";
  const liveAlert = document.getElementById(alertId);
  if (typeof liveAlert === "undefined" || liveAlert === null) {
    throw new Error(`The htmx alert element ${alertId} is null`);
  }
  if (event.detail.successful) {
    updateSuccess(liveAlert, `artifact-editor-credit-text`);
    updateSuccess(liveAlert, `artifact-editor-credit-ill`);
    updateSuccess(liveAlert, `artifact-editor-credit-prog`);
    return updateSuccess(liveAlert, `artifact-editor-credit-audio`);
  }
  if (event.detail.failed && event.detail.xhr) {
    updateError(event, `artifact-editor-credit-text`, liveAlert);
    updateError(event, `artifact-editor-credit-ill`, liveAlert);
    updateError(event, `artifact-editor-credit-prog`, liveAlert);
    return updateError(event, `artifact-editor-credit-audio`, liveAlert);
  }
  errorBrowser(liveAlert);
}

function afterUpdateDate(event, buttonId) {
  if (event.detail.elt === null) return;
  if (event.detail.elt.id !== `${buttonId}`) return;

  const alertId = "artifact-editor-alert";
  const liveAlert = document.getElementById(alertId);
  if (typeof liveAlert === "undefined" || liveAlert === null) {
    throw new Error(`The htmx alert element ${alertId} is null`);
  }
  const year = "artifact-editor-year";
  const month = "artifact-editor-month";
  const day = "artifact-editor-day";
  if (event.detail.successful) {
    updateSuccess(liveAlert, year);
    updateSuccess(liveAlert, month);
    return updateSuccess(liveAlert, day);
  }
  if (event.detail.failed && event.detail.xhr) {
    updateError(event, year, liveAlert);
    updateError(event, month, liveAlert);
    return updateError(event, day, liveAlert);
  }
  errorBrowser(liveAlert);
}

function afterUpdateRels(event, buttonId) {
  if (event.detail.elt === null) return;
  if (event.detail.elt.id !== `${buttonId}`) return;

  const alertId = "artifact-editor-alert";
  const liveAlert = document.getElementById(alertId);
  if (typeof liveAlert === "undefined" || liveAlert === null) {
    throw new Error(`The htmx alert element ${alertId} is null`);
  }

  const rel1 = "artifact-editor-releaser-1";
  const rel2 = "artifact-editor-releaser-2";
  if (event.detail.successful) {
    updateSuccess(liveAlert, rel1);
    return updateSuccess(liveAlert, rel2);
  }
  if (event.detail.failed && event.detail.xhr) {
    updateError(event, rel1, liveAlert);
    return updateError(event, rel2, liveAlert);
  }
  errorBrowser(liveAlert);
}

function removeSelectsValid(event, buttonId) {
  if (event.detail.elt === null) return;
  if (event.detail.elt.id !== `${buttonId}`) return;

  const select1Id = "artifact-editor-operating-system";
  const select2Id = "artifact-editor-category";
  const elm1 = document.getElementById(select1Id);
  if (typeof elm1 === "undefined" || elm1 === null) {
    return;
  }
  elm1.classList.remove("is-valid");
  const elm2 = document.getElementById(select2Id);
  if (typeof elm2 === "undefined" || elm2 === null) {
    return;
  }
  elm2.classList.remove("is-valid");
}

function afterUpdate(event, inputId) {
  if (event.detail.elt === null) return;
  if (event.detail.elt.id !== `${inputId}`) return;

  const alertId = "artifact-editor-alert";
  const liveAlert = document.getElementById(alertId);
  if (typeof liveAlert === "undefined" || liveAlert === null) {
    throw new Error(`The htmx alert element ${alertId} is null`);
  }

  if (event.detail.successful) {
    return updateSuccess(liveAlert, inputId);
  }
  if (event.detail.failed && event.detail.xhr) {
    return updateError(event, inputId, liveAlert);
  }
  errorBrowser(liveAlert);
}

function afterClassifications(event, buttonId) {
  if (event.detail.elt === null) return;
  if (event.detail.elt.id !== `${buttonId}`) return;

  const alertId = "artifact-editor-alert";
  const select1Id = "artifact-editor-operating-system";
  const select2Id = "artifact-editor-category";

  const liveAlert = document.getElementById(alertId);
  if (typeof liveAlert === "undefined" || liveAlert === null) {
    throw new Error(`The htmx alert element ${alertId} is null`);
  }

  if (event.detail.successful) {
    updateSuccess(liveAlert, select1Id);
    updateSuccess(liveAlert, select2Id);
    return;
  }
  if (event.detail.failed && event.detail.xhr) {
    return updateError(event, null, liveAlert);
  }
  errorBrowser(liveAlert);
}

function afterReset(event, buttonId, inputId) {
  if (event.detail.elt === null) return;
  if (event.detail.elt.id !== `${buttonId}`) return;

  const alertId = "artifact-editor-alert";
  const liveAlert = document.getElementById(alertId);
  if (typeof liveAlert === "undefined" || liveAlert === null) {
    throw new Error(`The htmx alert element ${alertId} is null`);
  }
  console.log(`afterReset ${buttonId} ${inputId}`, event.detail);
  if (event.detail.successful) {
    console.log(event.detail.successful, `sucessful`);
    return updateSuccess(liveAlert, inputId);
  }
  if (event.detail.failed && event.detail.xhr) {
    console.log(event.detail.failed, `failed`);
    return updateError(event, inputId, liveAlert);
  }
  errorBrowser(liveAlert);
}

function updateSuccess(alertElm, successId) {
  console.log(`updateSuccess ${successId}`);
  alertElm.innerText = "";
  alertElm.classList.add("d-none");
  if (typeof successId === "undefined" || successId === null) {
    return;
  }
  const elm = document.getElementById(successId);
  if (typeof elm === "undefined" || elm === null) {
    return;
  }
  elm.classList.add("is-valid");
}

function updateError(event, inputId, alertElm) {
  const xhr = event.detail.xhr;
  alertElm.innerText = `${timeNow()} Could not update the database record, ${xhr.responseText}.`;
  alertElm.classList.remove("d-none");
  if (inputId !== null) {
    const inputElm = document.getElementById(inputId);
    inputElm.classList.remove("is-valid");
  }
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
function afterRecord(event, inputId, revertId) {
  if (event.detail.elt === null) return;
  if (event.detail.elt.id !== `${inputId}`) return;

  const alertId = "artifact-editor-alert";
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
