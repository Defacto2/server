// uploader-submitter.mjs

import { getElmById } from "./helper.mjs";
export default submitter;

const none = "d-none";

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
    alert.innerText = "";
    alert.classList.add(none);
    results.innerHTML = "";
  }

  document.body.addEventListener("htmx:beforeRequest", function () {
    beforeReset(alert, results);
  });

  document.body.addEventListener("htmx:afterRequest", function (event) {
    if (event.detail.elt === null || event.detail.elt.id !== `${elementId}`) {
      return;
    }
    if (event.detail.successful) {
      return successful(alert);
    }
    const xhr = event.detail.xhr;
    if (event.detail.failed && xhr) {
      if (xhr.status === 404) {
        return error404(alert, results, api);
      }
      return errorXhr(alert, xhr);
    }
    errorBrowser(alert);
  });
}

function beforeReset(alert, results) {
  results.innerHTML = "";
  alert.innerText = "";
  alert.classList.add(none);
}

function successful(alert) {
  alert.classList.add(none);
  alert.innerText = "";
}

function error404(alert, results, api) {
  results.innerText = `Production not found on ${api}.`;
  alert.classList.add(none);
  alert.innerText = "";
}

function errorBrowser(alert) {
  alert.innerText = `Something with the browser is not working, please refresh the page.`;
  alert.classList.remove(none);
}

function errorXhr(alert, xhr) {
  alert.innerText = `Something went wrong, ${xhr.status} status: ${xhr.responseText}.`;
  alert.classList.remove(none);
}
