/**
 * @module layout-htmx
 * This module provides beforeRequest and afterRequest handlers to htmx submissions.
 */
import { searchEvents } from './layout-htmx-search.mjs';

export default htmxEvents;

/**
 * Initializes the htmx event listeners.
 */
export function htmxEvents() {
  //htmx.config.logAll = true; // uncomment, to debug but spam

  searchEvents();

  document.body.addEventListener('htmx:before:request', function (event) {
    removeSelectsValid(event, `artifact-editor-classifications-undo`);
    removeSelectsValid(event, `artifact-editor-text-for-dos`);
    removeSelectsValid(event, `artifact-editor-text-for-amiga`);
    removeSelectsValid(event, `artifact-editor-proof-of-release`);
    removeSelectsValid(event, `artifact-editor-intro-for-dos`);
    removeSelectsValid(event, `artifact-editor-intro-for-win`);
    removeSelectsValid(event, `artifact-editor-intro-for-bbs`);
    removeSelectsValid(event, `artifact-editor-trainer-for-dos`);
    removeSelectsValid(event, `artifact-editor-trainer-for-win`);
    removeSelectsValid(event, `artifact-editor-ansi-for-bbs`);
    removeSelectsValid(event, `artifact-editor-ansi-for-text`);
    removeSelectsValid(event, `artifact-editor-magazine-for-text`);
    removeSelectsValid(event, `artifact-editor-magazine-for-dos`);
    removeSelectsValid(event, `artifact-editor-dox`);
    removeSelectsValid(event, `artifact-editor-inst-for-dos`);
    removeSelectsValid(event, `artifact-editor-inst-for-win`);
  });
  // This event is triggered after an AJAX request has finished.
  // https://four.htmx.org/reference/events/htmx-after-request/
  document.body.addEventListener('htmx:after:request', function (event) {
    // searches.
    afterRequest(event, `search-htmx-input`, `search-htmx-alert`);
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
    afterClassifications(event, `artifact-editor-ansi-for-text`);
    afterClassifications(event, `artifact-editor-magazine-for-text`);
    afterClassifications(event, `artifact-editor-magazine-for-dos`);
    afterClassifications(event, `artifact-editor-dox`);
    afterClassifications(event, `artifact-editor-inst-for-dos`);
    afterClassifications(event, `artifact-editor-inst-for-win`);
    // record releaser.
    afterUpdateRels(event, `artifact-editor-releaser-undo`);
    afterUpdateRels(event, `artifact-editor-releaser-update`);
    // record title.
    afterUpdate(event, `artifact-editor-title`);
    afterReset(event, `artifact-editor-title-undo`, `artifact-editor-title`);
    // record filename.
    afterUpdate(event, `artifact-editor-filename`);
    afterReset(
      event,
      `artifact-editor-filename-undo`,
      `artifact-editor-filename`
    );
    // record virustotal.
    afterUpdate(event, `artifact-editor-virustotal`);
    // record date.
    afterUpdateDate(event, `artifact-editor-date-undo`);
    afterUpdateDate(event, `artifact-editor-date-lastmod`);
    afterUpdateDate(event, `artifact-editor-date-update`);
    // record creators.
    afterUpdate(event, `artifact-editor-credit-text`);
    afterUpdate(event, `artifact-editor-credit-ill`);
    afterUpdate(event, `artifact-editor-credit-prog`);
    afterUpdate(event, `artifact-editor-credit-audio`);
    afterCreators(event, `artifact-editor-credit-undo`);
    // record comment.
    afterUpdate(event, `artifact-editor-comment`);
    afterReset(
      event,
      `artifact-editor-comment-undo`,
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
    afterLinks(event, `artifact-editor-links-undo`);
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
    // trainer uploader.
    const alertTrainer = `uploader-trainer-alert`;
    afterRequest(event, `uploader-trainer-form`, alertTrainer);
    afterRequest(event, `uploader-trainer-releaser-1`, alertTrainer);
    afterRequest(event, `uploader-trainer-releaser-2`, alertTrainer);
  });
}

function getEditorAlert(elt) {
  const modalAlert = elt
    ?.closest('.modal-content')
    ?.querySelector('.artifact-editor-alert');
  if (modalAlert) return modalAlert;
  const liveAlert = document.querySelector('.artifact-editor-alert');
  if (typeof liveAlert === 'undefined' || liveAlert === null) {
    throw new Error('The htmx alert element .artifact-editor-alert is null');
  }
  return liveAlert;
}

function afterLinks(event, buttonId) {
  const { elt, isOk, isFailed } = parseHtmxEvent(event);
  if (elt === null) return;
  if (elt.id !== `${buttonId}`) return;

  const liveAlert = getEditorAlert(elt);
  if (isOk) {
    console.log('okay');
    return;
  }
  if (isFailed) {
    console.log('error');
    return;
  }
  errorBrowser(liveAlert);
}

function afterCreators(event, buttonId) {
  const { elt, text, isOk, isFailed } = parseHtmxEvent(event);
  if (elt === null) return;
  if (elt.id !== `${buttonId}`) return;

  const liveAlert = getEditorAlert(elt);
  if (isOk) {
    updateSuccess(liveAlert, `artifact-editor-credit-text`);
    updateSuccess(liveAlert, `artifact-editor-credit-ill`);
    updateSuccess(liveAlert, `artifact-editor-credit-prog`);
    return updateSuccess(liveAlert, `artifact-editor-credit-audio`);
  }
  if (isFailed) {
    updateError(text, `artifact-editor-credit-text`, liveAlert);
    updateError(text, `artifact-editor-credit-ill`, liveAlert);
    updateError(text, `artifact-editor-credit-prog`, liveAlert);
    return updateError(text, `artifact-editor-credit-audio`, liveAlert);
  }
  errorBrowser(liveAlert);
}

function afterUpdateDate(event, buttonId) {
  const { elt, text, isOk, isFailed } = parseHtmxEvent(event);
  if (elt === null) return;
  if (elt.id !== `${buttonId}`) return;

  const liveAlert = getEditorAlert(elt);
  const year = 'artifact-editor-year';
  const month = 'artifact-editor-month';
  const day = 'artifact-editor-day';
  if (isOk) {
    updateSuccess(liveAlert, year);
    updateSuccess(liveAlert, month);
    return updateSuccess(liveAlert, day);
  }
  if (isFailed) {
    updateError(text, year, liveAlert);
    updateError(text, month, liveAlert);
    return updateError(text, day, liveAlert);
  }
  errorBrowser(liveAlert);
}

function afterUpdateRels(event, buttonId) {
  const { elt, text, isOk, isFailed } = parseHtmxEvent(event);
  if (elt === null) return;
  if (elt.id !== `${buttonId}`) return;

  const liveAlert = getEditorAlert(elt);

  const rel1 = 'artifact-editor-releaser-1';
  const rel2 = 'artifact-editor-releaser-2';
  if (isOk) {
    updateSuccess(liveAlert, rel1);
    return updateSuccess(liveAlert, rel2);
  }
  if (isFailed) {
    updateError(text, rel1, liveAlert);
    return updateError(text, rel2, liveAlert);
  }
  errorBrowser(liveAlert);
}

function removeSelectsValid(event, buttonId) {
  const { elt } = parseHtmxEvent(event);
  if (elt === null) return;
  if (elt.id !== `${buttonId}`) return;

  const select1Id = 'artifact-editor-operating-system';
  const select2Id = 'artifact-editor-category';
  const elm1 = document.getElementById(select1Id);
  if (typeof elm1 === 'undefined' || elm1 === null) {
    return;
  }
  elm1.classList.remove('is-valid');
  const elm2 = document.getElementById(select2Id);
  if (typeof elm2 === 'undefined' || elm2 === null) {
    return;
  }
  elm2.classList.remove('is-valid');
}

function afterUpdate(event, inputId) {
  const { elt, text, isOk, isFailed } = parseHtmxEvent(event);
  if (elt === null) return;
  if (elt.id !== `${inputId}`) return;

  const liveAlert = getEditorAlert(elt);

  if (isOk) {
    return updateSuccess(liveAlert, inputId);
  }
  if (isFailed) {
    return updateError(text, inputId, liveAlert);
  }
  errorBrowser(liveAlert);
}

function afterClassifications(event, buttonId) {
  const { elt, text, isOk, isFailed } = parseHtmxEvent(event);
  if (elt === null) return;
  if (elt.id !== `${buttonId}`) return;

  const select1Id = 'artifact-editor-operating-system';
  const select2Id = 'artifact-editor-category';

  const liveAlert = getEditorAlert(elt);

  if (isOk) {
    updateSuccess(liveAlert, select1Id);
    updateSuccess(liveAlert, select2Id);
    return;
  }
  if (isFailed) {
    return updateError(text, null, liveAlert);
  }
  errorBrowser(liveAlert);
}

function afterReset(event, buttonId, inputId) {
  const { elt, text, isOk, isFailed } = parseHtmxEvent(event);
  if (elt === null) return;
  if (elt.id !== `${buttonId}`) return;

  const liveAlert = getEditorAlert(elt);
  console.log(`afterReset ${buttonId} ${inputId}`, event.detail);
  if (isOk) {
    console.log(isOk, `sucessful`);
    return updateSuccess(liveAlert, inputId);
  }
  if (isFailed) {
    console.log(isFailed, `failed`);
    return updateError(text, inputId, liveAlert);
  }
  errorBrowser(liveAlert);
}

function updateSuccess(alertElm, successId) {
  console.log(`updateSuccess ${successId}`);
  alertElm.innerText = '';
  alertElm.classList.add('d-none');
  if (typeof successId === 'undefined' || successId === null) {
    return;
  }
  const elm = document.getElementById(successId);
  if (typeof elm === 'undefined' || elm === null) {
    return;
  }
  elm.classList.remove('is-invalid');
  elm.classList.add('is-valid');
}

function updateError(responseText, inputId, alertElm) {
  alertElm.innerText = `${timeNow()} Could not update the database record, ${responseText}.`;
  alertElm.classList.remove('d-none');
  if (inputId !== null) {
    const inputElm = document.getElementById(inputId);
    inputElm.classList.remove('is-valid');
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
  const { elt, text, isOk, isFailed } = parseHtmxEvent(event);
  if (elt === null) return;
  if (elt.id !== `${inputId}`) return;

  const liveAlert = getEditorAlert(elt);

  if (isOk) {
    return recordSuccess(text, inputId, liveAlert);
  }
  if (isFailed) {
    return recordError(text, revertId, liveAlert);
  }
  errorBrowser(liveAlert);
}

/**
 * Handles the successful event.
 *
 * @param {string} responseText - The response body text.
 * @param {string} inputId - The ID of the input element.
 * @param {HTMLElement} alertElm - The alert element.
 */
function recordSuccess(responseText, inputId, alertElm) {
  alertElm.classList.add('d-none');
  alertElm.innerText = '';
  const elm = document.getElementById(`artifact-editor-modal-header`);
  const statusElm = document.getElementById(`artifact-editor-modal-status`);
  if (!elm) return;

  const state = responseText?.trim();
  const isOnline =
    state === 'online' ||
    (state !== 'offline' && inputId === 'artifact-editor-public');

  if (isOnline) {
    elm.classList.add('border-success');
    elm.classList.remove('border-danger');
    if (statusElm) {
      statusElm.classList.add('text-bg-success');
      statusElm.classList.remove('text-bg-danger');
      statusElm.textContent = 'Public';
    }
    return;
  }
  if (state === 'offline' || inputId === 'artifact-editor-hidden') {
    elm.classList.remove('border-success');
    elm.classList.add('border-danger');
    if (statusElm) {
      statusElm.classList.remove('text-bg-success');
      statusElm.classList.add('text-bg-danger');
      statusElm.textContent = 'Hidden';
    }
    return;
  }
  console.error(`The record success ${inputId} is not supported.`);
}

/**
 * Handles the error response from a request.
 *
 * @param {string} responseText - The response body text.
 * @param {string} revertId - The ID of the radio button to revert to.
 * @param {HTMLElement} alertElm - The alert element to display the error message.
 */
function recordError(responseText, revertId, alertElm) {
  alertElm.innerText = `${timeNow()} Could not update the database record, ${responseText}.`;
  alertElm.classList.remove('d-none');
  document.getElementById(revertId).checked = true;
}

function timeNow() {
  let now = new Date();
  let hours = now.getHours();
  let minutes = now.getMinutes();
  let seconds = now.getSeconds();
  minutes = (minutes < 10 ? '0' : '') + minutes;
  seconds = (seconds < 10 ? '0' : '') + seconds;
  return hours + ':' + minutes + ':' + seconds;
}

function parseHtmxEvent(event) {
  const ctx = event.detail?.ctx;
  const elt = ctx?.sourceElement ?? event.detail?.elt ?? null;
  const status = ctx?.response?.status ?? event.detail?.xhr?.status ?? 0;
  const text = ctx?.text ?? event.detail?.xhr?.responseText ?? '';
  const hasResponse = Boolean(ctx?.response || event.detail?.xhr);
  const isOk =
    ctx?.response !== undefined
      ? status >= 200 && status < 400
      : Boolean(event.detail?.successful);
  const isFailed =
    ctx?.response !== undefined
      ? status >= 400
      : Boolean(event.detail?.failed && event.detail?.xhr);
  return { elt, status, text, hasResponse, isOk, isFailed };
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
  const { elt, status, text, isOk, isFailed } = parseHtmxEvent(event);
  if (elt === null) return;
  if (elt.id !== `${inputId}`) return;

  const liveAlert = document.getElementById(alertId);
  if (typeof liveAlert === 'undefined' || liveAlert === null) {
    throw new Error(`The htmx alert element ${alertId} is null`);
  }

  if (isOk) {
    return successful(event, liveAlert);
  }
  if (isFailed) {
    return errorXhr(liveAlert, status, text);
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
  alertElm.classList.add('d-none');
  alertElm.innerText = '';
  const { elt } = parseHtmxEvent(event);
  const match = '-form',
    id = (elt ?? event.target).id;
  const suffix = id.slice(-match.length);
  if (suffix == match) {
    const select = id.replace(match, '-file');
    resetFile(elt ?? event.target, `#${select}`);

    // Show the results div by removing d-none class
    const resultsId = id.replace(match, '-results');
    const resultsElm = document.getElementById(resultsId);
    if (resultsElm) {
      resultsElm.classList.remove('d-none');
    }
  }
}

/**
 * Resets the value of the file input element.
 * @param {HTMLElement} formElm - The form element.
 * @param {string} selector - The selector of the file input element.
 */
function resetFile(formElm, selector) {
  const input = formElm.querySelector(selector);
  if (input) {
    input.value = '';
    input.innerText = '';
    return;
  }
  console.error(`The reset file ${selector} element is null`);
}

/**
 * Handles the error response from a request.
 *
 * @param {HTMLElement} alertElm - The alert element to display the error message.
 * @param {number} status - The HTTP response status code.
 * @param {string} text - The HTTP response body text.
 */
function errorXhr(alertElm, status, text) {
  alertElm.innerText = `Our server is not working at the moment.`;
  console.log(`Request did not work, ${status} status: ${text}.`);
  alertElm.classList.remove('d-none');
}

/**
 * Displays an error message usually caused by the browser.
 * @param {HTMLElement} alertElm - The alert element where the error message will be displayed.
 */
function errorBrowser(alertElm) {
  alertElm.innerText =
    'Something with the browser is not working, please try again or refresh the page.';
  alertElm.classList.remove('d-none');
}
