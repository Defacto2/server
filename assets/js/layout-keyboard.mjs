/**
 * @module layout-keyboard
 * This module provides keyboard shortcuts for the website layout.
 */
const right = "ArrowRight",
  left = "ArrowLeft";

const start = document.getElementById("paginationStart"),
  previous = document.getElementById("paginationPrev"),
  previousPair = document.getElementById("paginationPrev2"),
  next = document.getElementById("paginationNext"),
  nextPair = document.getElementById("paginationNext2"),
  end = document.getElementById("paginationEnd");

/**
 * Binds keyboard shortcuts to specific actions.
 */
export function keyboardShortcuts() {
  document.addEventListener("keydown", function (event) {
    if (event.ctrlKey && event.key == left) {
      if (start != null) start.click();
      return;
    }
    if (event.ctrlKey && event.key == right) {
      if (end != null) end.click();
      return;
    }
    if (event.shiftKey && event.key == left) {
      if (previousPair != null) previousPair.click();
      return;
    }
    if (event.shiftKey && event.key == right) {
      if (nextPair != null) nextPair.click();
      return;
    }
    if (event.key == left) {
      if (previous != null) previous.click();
      return;
    }
    if (event.key == right) {
      if (next != null) next.click();
      return;
    }
  });
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
    changePage(pageRange.value);
  });
  const label = `paginationRangeLabel`;
  const pageRangeLabel = document.getElementById(label);
  if (pageRangeLabel === null) {
    throw new Error(`The ${label} for pagination() element is null.`);
  }
  pageRange.addEventListener("input", function () {
    pageRangeLabel.textContent = "Jump to page " + pageRange.value;
  });
}

function changePage(range) {
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
}
