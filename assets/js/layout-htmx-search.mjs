// layout-htmx-search.mjs

const releaserAlert = "search-releaser-alert",
  releaserReset = "search-releaser-clear",
  releaserIndi = "search-releaser-indicator",
  releaserInput = "search-releaser-input";

export function releaserInit() {
  const clear = document.getElementById(releaserReset);
  if (clear !== null) {
    clear.addEventListener("click", function () {
      clearer(releaserAlert, releaserInput, releaserIndi);
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
  const alert = document.getElementById(releaserAlert);
  if (alert === null) {
    throw new Error(`The htmx alert element ${releaserAlert} is null`);
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
