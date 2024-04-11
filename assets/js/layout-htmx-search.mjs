// layout-htmx-search.mjs

const releaserId = "search-releaser-alert",
  releaserReset = "search-releaser-clear",
  releaserIndi = "search-releaser-indicator",
  releaserInput = "search-releaser-input";

const alert = document.getElementById(releaserId);

export function releaserInit() {
  const clear = document.getElementById(releaserReset);
  if (clear !== null) {
    clear.addEventListener("click", function () {
      clearer(releaserId, releaserInput, releaserIndi);
    });
  }
}

/**
 * Clears the input field, hides the alert, and resets the search results.
 * @throws {Error} If any of the required elements are null.
 */
export function clearer() {
  const input = document.getElementById(releaserInput);
  if (input === null) {
    throw new Error(`The ${releaserInput} element is null`);
  }
  const alert = document.getElementById(releaserId);
  if (alert === null) {
    throw new Error(`The htmx alert element ${releaserId} is null`);
  }
  const indicator = document.getElementById(releaserIndi);
  if (indicator === null) {
    throw new Error(
      `The releaser search indicator element ${releaserIndi} is null`
    );
  }
  const results = document.getElementById("search-releaser-results");
  if (results === null) {
    throw new Error(`The releaser search indicator element is null`);
  }
  input.value = "";
  input.focus();
  alert.setAttribute("hidden", "true");
  indicator.style.opacity = 0;
  results.innerHTML = "";
}

/**
 * Handles the search a releaser input event.
 *
 * @param {Event} event - The event object.
 * @throws {Error} If the htmx alert element is null.
 */
export function releaser(event) {
  if (event.detail.elt === null || event.detail.elt.id !== `${releaserInput}`) {
    return;
  }
  if (typeof alert === "undefined" || alert === null) {
    throw new Error(
      `The htmx alert element ${releaserId} for releaser search is null`
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
  alert.setAttribute("hidden", "true");
  alert.innerText = "";
}

function errorXhr(alert, event) {
  const xhr = event.detail.xhr;
  alert.innerText = `Something on the server is not working, ${xhr.status} status: ${xhr.responseText}.`;
  alert.removeAttribute("hidden");
}

function errorBrowser(alert) {
  alert.innerText =
    "Something with the browser is not working, please refresh the page.";
  alert.removeAttribute("hidden");
}
