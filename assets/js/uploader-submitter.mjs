/**
 * @module uploader-submitter
 * This module provides functions for handling file upload submissions.
 */
import { getElmById, validId } from './helper.mjs';
export default submitter;

/**
 * To test some error handling, use the following IDs:
 *
 * Pouët ID: 16, deleted from the database
 * Pouët ID: 15, not suitable for Defacto2
 */

const invalid = 'is-invalid',
  none = 'd-none';

/**
 * Submits the number input and handles the response from a remote API.
 *
 * @param {string} elementId - The ID of the submitter element.
 * @param {string} api - The title of the API endpoint, e.g., "Demozoo" or "Pouët".
 */
export function submitter(elementId, api) {
  const input = getElmById(elementId);
  const alert = getElmById(`${elementId}-error`);
  const results = getElmById(`${elementId}-results`);

  const close = getElmById(`${elementId}-close`);
  close.addEventListener('click', reset);
  const clear = getElmById(`${elementId}-clear`);
  clear.addEventListener('click', reset);

  function reset() {
    htmx.trigger(input, 'htmx:abort');
    const submitBtn = results.querySelector('[hx-put]');
    if (submitBtn) {
      htmx.trigger(submitBtn, 'htmx:abort');
    }
    input.value = '';
    input.focus();
    input.classList.remove(invalid);
    alert.innerText = '';
    alert.classList.add(none);
    results.innerHTML = '';
  }

  const demozooSanity = 450000,
    pouetSanity = 200000;
  switch (elementId) {
    case 'demozoo-submission':
      validate(input, demozooSanity);
      break;
    case 'pouet-submission':
      validate(input, pouetSanity);
      break;
  }

  // The htmx:before:request event is triggered before the request is made.
  document.body.addEventListener('htmx:before:request', function () {
    beforeReset(alert, results);
  });

  // The htmx:before:swap event is triggered before the content is swapped.
  // This is the best place to check the status of the request and display an error message.
  document.body.addEventListener('htmx:before:swap', function (evt) {
    const badRequest = 400;
    const status = evt.detail?.ctx?.response?.status ?? 0;
    if (status >= badRequest) {
      alert.classList.remove(none);
    }
  });

  // The htmx:after:request event is triggered after the request is completed.
  // Multiple requests can be made, so we need to check if the request is the one we are interested in.
  document.body.addEventListener('htmx:after:request', function (evt) {
    const ctx = evt.detail?.ctx;
    const elt = ctx?.sourceElement ?? null;
    if (elt === null || elt.id !== `${elementId}`) {
      return;
    }
    const response = ctx?.response;
    const status = response?.status ?? 0;
    const text = ctx?.text ?? '';
    if (response && status >= 200 && status < 400) {
      return successful(input);
    }
    if (response && status >= 400) {
      if (status === 404) {
        return error404(alert, results, api);
      }
      return errorXhr(alert, results, status, text);
    }
    errorBrowser(alert, results);
  });
}

function validate(input, sanity) {
  input.addEventListener('input', function () {
    if (!validId(input.value, sanity)) {
      input.classList.add(invalid);
      return;
    }
    input.classList.remove(invalid);
  });
}

function beforeReset(alert) {
  alert.innerText = '';
  alert.classList.add(none);
  // results.innerHTML = ""; // Uncomment and inject results as an argument, if we want to clear results too
}

function successful(input) {
  input.focus();
}

function error404(alert, results, api) {
  results.innerText = `Production not found on ${api}.`;
}

function errorBrowser(alert, results) {
  results.innerHTML = '';
  alert.innerText = `Something with the browser is not working, please refresh the page.`;
  alert.classList.remove(none);
}

function errorXhr(alert, results, status, text) {
  results.innerHTML = '';
  alert.innerText = `Something went wrong, ${status} status: ${text}.`;
  alert.classList.remove(none);
}
