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
    "addon",
    "aka",
    "an",
    "and",
    "as",
    "at",
    "beta",
    "but",
    "by",
    "cd-rip",
    "crack",
    "cracks",
    "demo",
    "docs",
    "documentation",
    "fix",
    "for",
    "if",
    "in",
    "installer",
    "is",
    "of",
    "on",
    "or",
    "so",
    "the",
    "to",
    "final",
    "from",
    "patch",
    "part",
    "prerelease",
    "pre-release",
    "preview",
    "proper",
    "release",
    "repack",
    "rip",
    "trainer",
    "v1",
    "v2",
    "v3",
    "v4",
    "v5",
    "v6",
    "v7",
    "v8",
    "v9",
    "v10",
    "vs",
    "with",
  ];
  const uppers = [
    "2d",
    "3d",
    "4d",
    "abc",
    "ad&d",
    "bbs",
    "cd",
    "cga",
    "dos",
    "dox",
    "ega",
    "ehq",
    "f1",
    "ftp",
    "hq",
    "id",
    "iso",
    "la",
    "ls",
    "mbl",
    "ms",
    "nascar",
    "nba",
    "ncaa",
    "nfl",
    "nfo",
    "nhl",
    "nt",
    "oem",
    "os",
    "pc",
    "psx",
    "usa",
    "ushq",
    "uss",
    "vga",
    "whq",
    "wwf",
    "xp",
  ];
  text = text.trim();
  // Replace all underscores with spaces
  text = text.replace(/_/g, " ");
  // Remove suffix (1) (2) (3) etc. (a) (b) (c) etc.
  text = text.replace(/ \([0-9a-z]\)/g, "");
  // Insert a space after a colon following an alphanumeric string
  text = text.replace(/([a-zA-Z0-9]): /g, "$1 : ");
  const wordCount = text.split(" ").length;
  text = text
    .split(" ")
    .map((word, index) => {
      const x = romanFix(word);
      if (index > 0 && Number.isInteger(x)) {
        return x;
      }
      const y = replacementFix(word);
      if (index > 0 && y !== word) {
        return y;
      }
      const z = tailFix(word, index, wordCount);
      if (z !== word) {
        return z;
      }
      if (uppers.includes(word.toLowerCase())) {
        return word.toUpperCase();
      }
      // Capitalize the first word and any word not in the common adverbs list
      // Convert all other words to lowercase
      if (index === 0 || !commonAdverbs.includes(word.toLowerCase())) {
        return word.charAt(0).toUpperCase() + word.slice(1).toLowerCase();
      }
      return word.toLowerCase();
    })
    .join(" ")
    .trim();
  return text;
}

/**
 * Removes the last word in a string if it is a common suffix.
 * @param {string} word - The word to be checked.
 * @param {number} index - The index of the word in the string.
 * @param {number} length - The total number of words in the string.
 * @returns {string} - The word with the suffix removed if it is a common suffix.
 * @example
 * tailFix("cracktro", 0, 2) // returns "cracktro"
 * tailFix("cracktro", 1, 2) // returns ""
 */
export function tailFix(word, index, length) {
  if (index != length - 1) {
    return word;
  }
  const removers = [
    "cheat",
    "cheater",
    "cracktro",
    "loader",
    "installer",
    "trainer",
    "version",
  ];
  if (removers.includes(word.toLowerCase())) {
    return "";
  }
  return word;
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
 * Converts elite words and symbols to English or numerals.
 * @param {string} word - The string to be converted.
 * @returns {string} - The converted string.
 * @example
 * replacementFix("][") // returns "2"
 */
export function replacementFix(word) {
  switch (word) {
    case "cdrip":
      return "cd-rip";
    case "&":
      return "and";
    case "ad+d":
      return "AD&D";
    case "V1.0":
      return "v1";
    case "V2.0":
      return "v2";
    case "V3.0":
      return "v3";
    case "V4.0":
      return "v4";
    case "V5.0":
      return "v5";
    case "V6.0":
      return "v6";
    case "V7.0":
      return "v7";
    case "V8.0":
      return "v8";
    case "V9.0":
      return "v9";
    case "V10.0":
      return "v10";
    case ("][", "||"):
      return "2";
    case "]|[":
      return "3";
    case "]||[":
      return "4";
    case " I:":
      return " 1 :";
    case " II:":
      return " 2 :";
    case " III:":
      return " 3 :";
    case "war-craft":
      return "Warcraft";
    default:
      return word;
  }
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
