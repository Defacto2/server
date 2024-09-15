/**
 * @module artifact-validate
 * Module contains functions for validating input element values.
 */

/**
 * Validates the date input values.
 *
 * @param {HTMLInputElement} yearInput - The input element for the year.
 * @param {HTMLInputElement} monthInput - The input element for the month.
 * @param {HTMLInputElement} dayInput - The input element for the day.
 * @throws {Error} If any of the input elements are null.
 */
export function date(yearInput, monthInput, dayInput) {
  if (yearInput == null) {
    throw new Error("The year input element is null.");
  }
  if (monthInput == null) {
    throw new Error("The month input element is null.");
  }
  if (dayInput == null) {
    throw new Error("The day input element is null.");
  }
  yearInput.classList.remove("is-invalid", "is-valid");
  monthInput.classList.remove("is-invalid", "is-valid");
  dayInput.classList.remove("is-invalid", "is-valid");

  const year = parseInt(yearInput.value, 10);
  if (isNaN(year)) {
    yearInput.value = "0";
  } else {
    yearInput.value = year; // remove leading zeros
  }
  const month = parseInt(monthInput.value, 10);
  if (isNaN(month)) {
    monthInput.value = "0";
  } else {
    monthInput.value = month;
  }
  const day = parseInt(dayInput.value, 10);
  if (isNaN(day)) {
    dayInput.value = "0";
  } else {
    dayInput.value = day;
  }

  const none = 0;
  const currentYear = new Date().getFullYear();
  const validYear = year >= 1980 && year <= currentYear;
  // use greater than instead of != none to avoid a isNaN condition
  if (year > none && !validYear) {
    yearInput.classList.add("is-invalid");
  }
  const validMonth = month >= 1 && month <= 12;
  if (month > none && !validMonth) {
    monthInput.classList.add("is-invalid");
  }
  const validDay = day >= 1 && day <= 31;
  if (day > none && !validDay) {
    dayInput.classList.add("is-invalid");
  }
  if (isNaN(year) && (validMonth || validDay)) {
    yearInput.classList.add("is-invalid");
  }
  if ((month == none || isNaN(month)) && validDay) {
    monthInput.classList.add("is-invalid");
  }
}

export function repository(elm) {
  return urlPath(elm, true);
}

export function color16(elm) {
  return urlPath(elm, false);
}

/**
 * Validates the URL path and updates the element's classList accordingly.
 *
 * @param {HTMLInputElement} elm - The repository URL element.
 * @param {boolean} github - Indicates whether the URL element value is for a GitHub repository.
 * @throws {Error} If the repository URL element is null or if the maxlength attribute is missing.
 */
function urlPath(elm, github) {
  if (elm == null) {
    throw new Error("The repository URL element is null.");
  }
  elm.classList.remove("is-valid", "is-invalid");

  let value = elm.value.trim();
  if (value.length === 0) {
    return;
  }
  // valid characters were determined by this document,
  // https://docs.github.com/en/get-started/using-git/dealing-with-special-characters-in-branch-and-tag-names#naming-branches-and-tags
  if (github == true && value.startsWith("refs/")) {
    elm.classList.add("is-invalid");
    return;
  }
  const rawURL = "://";
  if (value.includes(rawURL)) {
    elm.classList.add("is-invalid");
    return;
  }
  const permittedChrs = /[^A-Za-z0-9-._/]/g;
  value = value.replace(permittedChrs, "");
  value = value.replaceAll("//", "/");
  const regLeadSeparators = /^\//;
  value = value.replace(regLeadSeparators, "");
  elm.value = value;

  const max = elm.getAttribute("maxlength");
  if (max === null) {
    throw new Error(`The maxlength attribute is required for ${elm.id}.`);
  }
  if (value.length > max) {
    elm.classList.add("is-invalid");
    return;
  }
}

/**
 * Validates and updates the releaser element.
 *
 * @param {HTMLElement} elm - The releaser element to validate.
 * @throws {Error} If the element is null or if the minlength or maxlength attributes are missing.
 */
export function releaser(elm) {
  if (elm == null) {
    throw new Error("The element of the releaser validator is null.");
  }
  elm.classList.remove("is-valid", "is-invalid");

  let value = elm.value.trim().toUpperCase();
  value = value.replace(/[^ A-ZÀ-ÖØ-Þ0-9\-,&]/g, "");
  elm.value = value;

  const min = elm.getAttribute("minlength");
  const max = elm.getAttribute("maxlength");
  const req = elm.getAttribute("required");
  if (min === null) {
    throw new Error(`The minlength attribute is required for ${elm.id}.`);
  }
  if (max === null) {
    throw new Error(`The maxlength attribute is required for ${elm.id}.`);
  }

  const error = document.getElementById("artifact-editor-releasers-error");
  if (error === null) {
    throw new Error("The releasers error element is null.");
  }

  const requireBounds = value.length < min || value.length > max;
  if (req != null && requireBounds) {
    elm.classList.add("is-invalid");
    if (elm.id === "-1") {
      error.classList.add("d-block");
    }
    return;
  }
  const emptyBounds =
    value.length > 0 && (value.length < min || value.length > max);
  if (req == null && emptyBounds) {
    elm.classList.add("is-invalid");
    return;
  }
  elm.classList.remove("is-invalid");
  error.classList.remove("d-block");
}

export function youtube(elm) {
  if (elm == null) {
    throw new Error("The element of the releaser validator is null.");
  }
  elm.classList.remove("is-valid", "is-invalid");
  const value = elm.value.trim();
  const required = 11;
  if (value.length > 0 && value.length != required) {
    elm.classList.add("is-invalid");
  }
}

export function number(elm, max) {
  if (elm == null) {
    throw new Error("The element of the number validator is null.");
  }
  elm.classList.remove("is-valid", "is-invalid");
  const value = parseInt(elm.value, 10);
  if (isNaN(value)) {
    elm.classList.add("is-invalid");
  }
  if (value > max || value < 0) {
    elm.classList.add("is-invalid");
  }
}
