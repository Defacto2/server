/**
 * @module uploader-submitter
 * This module provides functions for handling file upload submissions.
 */
import { getElmById, validId } from "./helper.mjs";
export default submitter;

/**
 * To test some error handling, use the following IDs:
 *
 * Pouët ID: 16, deleted from the database
 * Pouët ID: 15, not suitable for Defacto2
 */

const invalid = "is-invalid",
  none = "d-none";

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
  close.addEventListener("click", reset);
  const clear = getElmById(`${elementId}-clear`);
  clear.addEventListener("click", reset);

  function reset() {
    input.value = "";
    input.focus();
    input.classList.remove(invalid);
    alert.innerText = "";
    alert.classList.add(none);
    results.innerHTML = "";
  }

  const demozooSanity = 450000,
    pouetSanity = 200000;
  switch (elementId) {
    case "demozoo-submission":
      validate(input, demozooSanity);
      break;
    case "pouet-submission":
      validate(input, pouetSanity);
      break;
  }

  // The htmx:beforeRequest event is triggered before the request is made.
  document.body.addEventListener("htmx:beforeRequest", function () {
    beforeReset(alert, results);
  });

  // The htmx:beforeSwap event is triggered before the content is swapped.
  // This is the best place to check the status of the request and display an error message.
  document.body.addEventListener("htmx:beforeSwap", function (evt) {
    const badRequest = 400;
    if (evt.detail.xhr.status >= badRequest) {
      alert.classList.remove(none);
    }
  });

  // The htmx:afterRequest event is triggered after the request is completed.
  // Multiple requests can be made, so we need to check if the request is the one we are interested in.
  document.body.addEventListener("htmx:afterRequest", function (evt) {
    if (evt.detail.elt === null || evt.detail.elt.id !== `${elementId}`) {
      return;
    }
    if (evt.detail.successful) {
      return successful(input);
    }
    const xhr = evt.detail.xhr;
    if (evt.detail.failed && xhr) {
      if (xhr.status === 404) {
        return error404(alert, results, api);
      }
      return errorXhr(alert, xhr);
    }
    errorBrowser(alert);
  });
}

function validate(input, sanity) {
  input.addEventListener("input", function () {
    if (!validId(input.value, sanity)) {
      input.classList.add(invalid);
      return;
    }
    input.classList.remove(invalid);
  });
}

function beforeReset(alert, results) {
  alert.innerText = "";
  alert.classList.add(none);
}

function successful(input) {
  input.focus();
}

function error404(alert, results, api) {
  results.innerText = `Production not found on ${api}.`;
}

function errorBrowser(alert) {
  alert.innerText = `Something with the browser is not working, please refresh the page.`;
  alert.classList.remove(none);
}

function errorXhr(alert, xhr) {
  alert.innerText = `Something went wrong, ${xhr.status} status: ${xhr.responseText}.`;
  alert.classList.remove(none);
}
