/**
 * @module layout-htmx-search
 * This module provides functions for htmx, search events.
 */
const searchAlert = "search-htmx-alert",
  searchReset = "search-htmx-clear",
  searchIndic = "search-htmx-indicator",
  searchInput = "search-htmx-input";

/**
 * Handles search specific events such as the Clear button.
 */
export function searchEvents() {
  const clear = document.getElementById(searchReset);
  if (clear !== null) {
    clear.addEventListener("click", function () {
      clearer(searchAlert, searchInput, searchIndic);
    });
  }
}

/**
 * Clears the input field, hides the alert, and resets the search results.
 * @throws {Error} If any of the required elements are null.
 */
export function clearer() {
  const input = document.getElementById(searchInput);
  if (input === null) {
    throw new Error(`The ${searchInput} for clearer() element is null`);
  }
  const alert = document.getElementById(searchAlert);
  if (alert === null) {
    throw new Error(`The htmx alert element ${searchAlert} is null`);
  }
  const indicator = document.getElementById(searchIndic);
  if (indicator === null) {
    throw new Error(`The htmx search indicator element ${searchIndic} is null`);
  }
  const results = document.getElementById("search-htmx-results");
  if (results === null) {
    throw new Error(`The htmx search indicator element is null`);
  }
  input.value = "";
  input.focus();
  alert.setAttribute("hidden", "true");
  indicator.style.opacity = 0;
  results.innerHTML = "";
}
