// uploader.mjs

import { getElmById } from "./helper.mjs";

/**
 * Retrieves a modal object by its element ID.
 *
 * @param {string} elementId - The ID of the element representing the modal.
 * @returns {Object} - The modal object.
 * @throws {Error} - If the bootstrap object is undefined.
 */
export function getModalById(elementId) {
  if (bootstrap === undefined) {
    throw new Error(`The bootstrap object is undefined.`);
  }
  const element = getElmById(elementId);
  const modal = new bootstrap.Modal(element);
  return modal;
}

/**
 * Focuses the modal by its element ID and submission ID.
 * @param {string} elementId - The ID of the modal element.
 * @param {string} submissionId - The ID of the submission element.
 * @returns {Object} - The modal object.
 * @throws {Error} - If the submission element is null or if the bootstrap object is undefined.
 */
export function focusModalById(elementId, submissionId) {
  const input = document.getElementById(submissionId);
  if (input == null) {
    throw new Error(`The ${submissionId} element is null.`);
  }
  const element = getElmById(elementId);
  element.addEventListener("shown.bs.modal", function () {
    input.focus();
  });
  if (bootstrap === undefined) {
    throw new Error(`The bootstrap object is undefined.`);
  }
  const modal = new bootstrap.Modal(element);
  return modal;
}

/**
 * Adds pagination functionality to an element.
 * @param {string} elementId - The ID of the element to add pagination to.
 */
export function pagination(elementId) {
  const pageRange = document.getElementById(elementId);
  if (typeof pageRange === "undefined" || pageRange === null) {
    return;
  }
  pageRange.addEventListener("change", function () {
    const range = pageRange.value;
    const url = new URL(window.location.href);
    const path = url.pathname;
    const paths = path.split("/");
    const page = paths[paths.length - 1];
    if (!isNaN(page) && typeof Number(page) === "number") {
      paths[paths.length - 1] = range;
    } else {
      paths.push(range);
    }
    url.pathname = paths.join("/");
    window.location.href = url.href;
  });
  const label = `paginationRangeLabel`;
  const pageRangeLabel = document.getElementById(label);
  if (pageRangeLabel === null) {
    throw new Error(`The ${label} element is null.`);
  }
  pageRange.addEventListener("input", function () {
    pageRangeLabel.textContent = "Jump to page " + pageRange.value;
  });
}
