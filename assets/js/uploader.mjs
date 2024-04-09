export default getElmById;

/**
 * Retrieves an element from the DOM using its ID.
 *
 * @param {string} elementId - The ID of the element to retrieve.
 * @returns {HTMLElement} - The retrieved element.
 * @throws {Error} - If the element is not found in the DOM.
 */
export function getElmById(elementId) {
  const element = document.getElementById(elementId);
  if (element == null) {
    throw new Error(`The ${elementId} element is null.`);
  }
  return element;
}

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

/**
 * Checks if a given year is valid, i.e. between 1980 and the current year.
 * @param {number} year - The year to be validated.
 * @returns {boolean} - Returns true if the year is valid, false otherwise.
 */
export function validYear(year) {
  if (`${year}` == "") {
    return true;
  }
  const epochYear = 1980;
  const currentYear = new Date().getFullYear();
  if (year < epochYear || year > currentYear) {
    return false;
  }
  return true;
}

/**
 * Checks if a given month is valid.
 * @param {number} month - The month to be validated.
 * @returns {boolean} - Returns true if the month is valid, false otherwise.
 */
export function validMonth(month) {
  if (`${month}` == "") {
    return true;
  }
  const jan = 1,
    dec = 12;
  if (month < jan || month > dec) {
    return false;
  }
  return true;
}

/**
 * Checks if a given day is valid.
 * @param {number} day - The day to be checked.
 * @returns {boolean} - Returns true if the day is valid, false otherwise.
 */
export function validDay(day) {
  if (`${day}` == "") {
    return true;
  }
  const first = 1,
    last = 31;
  if (day < first || day > last) {
    return false;
  }
  return true;
}
