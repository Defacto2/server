// layout.js

import {
  keyboardShortcuts as layoutKeys,
  pagination,
} from "./layout-keyboard.mjs";
import htmxEvents from "./layout-htmx.mjs";

(() => {
  "use strict";

  htmxEvents();
  layoutKeys();
  pagination("paginationRange");
  toolTips();
})();

/**
 * Initializes tooltips for elements with the data-bs-toggle="tooltip" attribute.
 * @throws {Error} If tooltip trigger list is not found or if Bootstrap Tooltip is undefined.
 */
function toolTips() {
  const tooltipTriggerList = document.querySelectorAll(
    '[data-bs-toggle="tooltip"]'
  );
  if (tooltipTriggerList === null) {
    throw new Error("Tooltip trigger list not found");
  }
  if (typeof bootstrap.Tooltip === "undefined") {
    throw new Error("Bootstrap Tooltip is undefined");
  }
  // eslint-disable-next-line no-unused-vars
  const tooltipList = [...tooltipTriggerList].map(
    (tooltipTriggerEl) => new bootstrap.Tooltip(tooltipTriggerEl)
  );
}
