// layout.js

import {
  keyboardShortcuts as layoutKeys,
  pagination,
} from "./layout-keyboard.mjs";
import htmxLoader from "./layout-htmx.mjs";

(() => {
  "use strict";
  htmxLoader();
  layoutKeys();
  pagination("paginationRange");
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
})();
