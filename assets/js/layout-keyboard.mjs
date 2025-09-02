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
  end = document.getElementById("paginationEnd"),
  srchP = document.getElementById("layout-search-program"),
  srchN = document.getElementById("layout-search-filename"),
  srchG = document.getElementById("layout-search-groups");
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
    if (event.altKey && event.shiftKey) {
      // note: because of the live search results with htmx, we need to update the elements on each keypress.
      const srch1 = document.getElementById("search-result-1"),
        srch2 = document.getElementById("search-result-2"),
        srch3 = document.getElementById("search-result-3"),
        srch4 = document.getElementById("search-result-4"),
        srch5 = document.getElementById("search-result-5"),
        srch6 = document.getElementById("search-result-6"),
        srch7 = document.getElementById("search-result-7"),
        srch8 = document.getElementById("search-result-8"),
        srch9 = document.getElementById("search-result-9"),
        srch10 = document.getElementById("search-result-0"),
        srch11 = document.getElementById("search-result--"),
        srch12 = document.getElementById("search-result-="),
        srch13 = document.getElementById("search-result-["),
        srch14 = document.getElementById("search-result-]");
      // note the follow keys are in use by editor-assets.js: e, d, r, v
      switch (event.key) {
        case "P":
        case "p":
          if (srchP) {
            event.preventDefault();
            srchP.click();
          }
          break;
        case "G":
        case "g":
          if (srchG) {
            event.preventDefault();
            srchG.click();
          }
          break;
        case "N":
        case "n":
          if (srchN) {
            event.preventDefault();
            srchN.click();
          }
          break;
        case "1":
        case "!":
          if (srch1) {
            event.preventDefault();
            window.open(srch1.href, "srch1");
          }
          break;
        case "2":
        case "@":
          if (srch2) {
            event.preventDefault();
            window.open(srch2.href, "srch2");
          }
          break;
        case "3":
        case "#":
          if (srch3) {
            event.preventDefault();
            window.open(srch3.href, "srch3");
          }
          break;
        case "4":
        case "$":
          if (srch4) {
            event.preventDefault();
            window.open(srch4.href, "srch4");
          }
          break;
        case "5":
        case "%":
          if (srch5) {
            event.preventDefault();
            window.open(srch5.href, "srch5");
          }
          break;
        case "6":
        case "^":
          if (srch6) {
            event.preventDefault();
            window.open(srch6.href, "srch6");
          }
          break;
        case "7":
        case "&":
          if (srch7) {
            event.preventDefault();
            window.open(srch7.href, "srch7");
          }
          break;
        case "8":
        case "*":
          if (srch8) {
            event.preventDefault();
            window.open(srch8.href, "srch8");
          }
          break;
        case "9":
        case "(":
          if (srch9) {
            event.preventDefault();
            window.open(srch9.href, "srch9");
          }
          break;
        case "0":
        case ")":
          if (srch10) {
            event.preventDefault();
            window.open(srch10.href, "srch10");
          }
          break;
        case "-":
        case "_":
          if (srch11) {
            event.preventDefault();
            window.open(srch11.href, "srch11");
          }
          break;
        case "+":
        case "=":
          if (srch12) {
            event.preventDefault();
            window.open(srch12.href, "srch12");
          }
          break;
        case "[":
        case "{":
          if (srch13) {
            event.preventDefault();
            window.open(srch13.href, "srch13");
          }
          break;
        case "]":
        case "}":
          if (srch14) {
            event.preventDefault();
            window.open(srch14.href, "srch14");
          }
          break;
      }
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
