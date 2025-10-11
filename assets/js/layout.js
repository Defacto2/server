/**
 * @file layout.js
 * This script is the entry point for the website layout.
 */
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
  fluidColumns();
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

/**
  * Fluid columns between a fix number of fluid number of artifact columns.
  */
function fluidColumns() {
  const fbn = "fluid-button"
  const key = "fluid"
  const btn = document.getElementById(fbn)
  const box = "box-container"
  document.addEventListener('DOMContentLoaded', () => {
    const el = document.querySelector('#fluid-artifacts');
    if (el) {
      const fluid = localStorage.getItem(key);
      const bc = document.getElementById(box)
      if (fluid === "1" && box) {
        bc.classList.toggle('container-xxl');
        bc.classList.toggle('container-fluid');
        btn.textContent = 'Fix columns'
      }
    }
  });
  if (btn) {
    btn.addEventListener('click', () => {
      document.getElementById(box).classList.toggle('container-xxl');
      document.getElementById(box).classList.toggle('container-fluid');
      const item = localStorage.getItem(key);
      if (item === "1") {
        localStorage.removeItem(key)
        document.getElementById(fbn).textContent = 'Unlock columns'
      } else {
        localStorage.setItem(key, "1");
        document.getElementById(fbn).textContent = 'Fix columns'
      }
    });
  }
}
