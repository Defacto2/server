/**
 * @module helper
 * This module provides functions for handling common tasks.
 */

/**
 * Titleizes a string of text but keeps common adverbs in lowercase.
 *
 * @param {string} text - The input text to be titleized.
 * @returns {string} - The titleized text.
 */
export function titleize(text) {
  const commonAdverbs = [
    "a",
    "an",
    "and",
    "as",
    "but",
    "for",
    "if",
    "of",
    "or",
    "so",
    "the",
    "to",
  ];
  return text
    .split(" ")
    .map((word, index) => {
      const x = romanFix(word);
      if (index > 0 && Number.isInteger(x)) {
        return x;
      }
      // Capitalize the first word and any word not in the common adverbs list
      if (index === 0 || !commonAdverbs.includes(word.toLowerCase())) {
        return word.charAt(0).toUpperCase() + word.slice(1).toLowerCase();
      }
      return word.toLowerCase();
    })
    .join(" ");
}

/**
 * Converts an I,V or X based Roman numeral to an integer or returns the original string if it is not a valid Roman numeral.
 * @param {string} word - The Roman numeral to be converted.
 * @returns {number|string} - The integer equivalent of the Roman numeral.
 */
export function romanFix(word) {
  const romanNumeralMap = {
    I: 1,
    V: 5,
    X: 10,
  };
  let total = 0;
  let prevValue = 0;
  for (let i = word.length - 1; i >= 0; i--) {
    const currentValue = romanNumeralMap[word[i].toUpperCase()];
    if (currentValue < prevValue) {
      total -= currentValue;
    } else {
      total += currentValue;
    }
    prevValue = currentValue;
  }
  if (total == 0) {
    return word;
  }
  return total;
}

/**
 * Copies the text content of an HTML element to the clipboard.
 * @async
 * @function clipText
 * @param {string} elementId - The ID of the HTML element to copy the text from.
 * @throws {Error} Throws an error if the specified element is missing.
 * @returns {Promise<void>} A Promise that resolves when the text has been copied to the clipboard.
 */
export async function clipText(elementId) {
  const oneSecond = 1000;
  const element = getElmById(elementId);
  element.focus(); // select the element to avoid NotAllowedError: Clipboard write is not allowed in this context
  await navigator.clipboard.writeText(`${element.textContent}`).then(
    function () {
      console.log(
        `Copied ${humanFilesize(element.textContent.length)} to the clipboard`
      );
      const button = document.getElementById(`artifact-copy-readme-body`);
      if (button === null) return;
      const save = button.textContent;
      button.textContent = `✓ Copied`;
      window.setTimeout(() => {
        button.textContent = `${save}`;
      }, oneSecond);
    },
    function (err) {
      console.error(`could not save any text to the clipboard: ${err}`);
    }
  );
}

/**
 * Copies the value of an HTML element to the clipboard.
 * @async
 * @function clipValue
 * @param {string} elementId - The ID of the HTML element to copy the value from.
 * @throws {Error} Throws an error if the specified element is missing.
 * @returns {Promise<void>} A Promise that resolves when the value has been copied to the clipboard.
 */
export async function clipValue(elementId) {
  const element = getElmById(elementId);
  element.focus(); // select the element to avoid NotAllowedError: Clipboard write is not allowed in this context
  await navigator.clipboard.writeText(`${element.value}`).then(
    function () {
      console.log(
        `Copied ${humanFilesize(element.value.length)} to the clipboard`
      );
    },
    function (err) {
      console.error(`could not save any text to the clipboard: ${err}`);
    }
  );
}

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
    console.error(`The ${elementId} for getElmById() element is null.`);
    return;
    //    throw new Error(`The ${elementId} for getElmById() element is null.`);
  }
  return element;
}

/**
 * Converts a file size in bytes to a human-readable format.
 *
 * @param {number} size - The file size in bytes.
 * @returns {string} A human-readable string representation of the file size.
 */
export function humanFilesize(size = 0) {
  const three = 3,
    round = 100,
    kB = 1000,
    MB = Math.pow(kB, 2),
    GB = Math.pow(kB, three);
  if (size > GB)
    return `${(Math.round((size * round) / GB) / round).toFixed(2)} GB`;
  if (size > MB)
    return `${(Math.round((size * round) / MB) / round).toFixed(1)} MB`;
  if (size > kB)
    return `${(Math.round((size * round) / kB) / round).toFixed()} kB`;
  return `${Math.round(size).toFixed()} bytes`;
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

// Validate a database ID.
// @param {string} id - The ID to be validated.
// @returns {boolean} - Returns true if the ID is valid, false otherwise.
export function validId(id, sanity) {
  if (id == "") {
    return true;
  }
  const nid = Number(id),
    max = Number(sanity);
  if (!Number.isInteger(max) || max < 1) {
    throw new Error(`The ID sanity value is invalid: ${max}`);
  }
  return Number.isInteger(nid) && nid > 0 && nid <= max;
}
