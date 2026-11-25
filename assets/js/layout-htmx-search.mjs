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
  const input = document.getElementById(searchInput);
  const clear = document.getElementById(searchReset);
  if (input === null || clear === null) {
    return;
  }

  function syncClear() {
    clear.disabled = input.value === "";
    if (input.value.trim() === "") {
      if (typeof htmx !== "undefined") {
        htmx.trigger(input, "htmx:abort");
      }
      const results = document.getElementById("search-htmx-results");
      if (results !== null) {
        results.innerHTML = "";
      }
    }
  }

  input.addEventListener("input", syncClear);
  input.addEventListener("search", syncClear);
  input.addEventListener("keydown", (event) => {
    if (event.key === "Escape" && input.value !== "") {
      event.preventDefault();
      clearer();
    }
  });
  clear.addEventListener("click", clearer);
  document.body.addEventListener("htmx:before:swap", (event) => {
    const ctx = event.detail?.ctx;
    const elt = ctx?.sourceElement ?? event.detail?.elt ?? null;
    if (elt !== input) return;
    if (input.value.trim() === "") {
      event.preventDefault();
    }
  });
  window.addEventListener("pageshow", syncClear);
  syncClear();
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
    throw new Error(`The htmx search results element is null`);
  }
  if (typeof htmx !== "undefined") {
    htmx.trigger(input, "htmx:abort");
  }
  input.value = "";
  input.focus();
  alert.classList.add("d-none");
  alert.innerText = "";
  indicator.classList.remove("htmx-request");
  indicator.style.removeProperty("opacity");
  results.innerHTML = "";
  const clear = document.getElementById(searchReset);
  if (clear !== null) {
    clear.disabled = true;
  }
}
