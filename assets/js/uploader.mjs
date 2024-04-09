export default getModalById;

export function getModalById(uploaderId) {
  const element = document.getElementById(uploaderId);
  if (element == null) {
    throw new Error(`The ${uploaderId} element is null.`);
  }
  if (bootstrap === undefined) {
    throw new Error(`The bootstrap object is undefined.`);
  }
  const modal = new bootstrap.Modal(element);
  return modal;
}

export function focusModalById(uploaderId, submissionId) {
  const element = document.getElementById(uploaderId);
  if (element == null) {
    throw new Error(`The ${uploaderId} element is null.`);
  }
  const input = document.getElementById(submissionId);
  if (input == null) {
    throw new Error(`The ${submissionId} element is null.`);
  }
  element.addEventListener("shown.bs.modal", function () {
    input.focus();
  });
  if (bootstrap === undefined) {
    throw new Error(`The bootstrap object is undefined.`);
  }
  const modal = new bootstrap.Modal(element);
  return modal;
}

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
  const currentYear = new Date().getFullYear();
  if (year < 1980 || year > currentYear) {
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
  if (month < 1 || month > 12) {
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
  if (day < 1 || day > 31) {
    return false;
  }
  return true;
}
