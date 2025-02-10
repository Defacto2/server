/**
 * @module helper
 * This module provides functions for handling common tasks.
 */

/**
 * Intercepts the paste event and formats the pasted text.
 * @param {Event} evt - The paste event.
 */
export function formatPaste(evt) {
  // Prevent the default paste action
  evt.preventDefault();

  // Get the pasted text from the clipboard
  const pastedText = (evt.clipboardData || window.Clipboard).getData("text");
  const title = this;
  const start = title.selectionStart;
  const end = title.selectionEnd;
  title.value =
    title.value.slice(0, start) + pastedText + title.value.slice(end);
  title.setSelectionRange(start + pastedText.length, start + pastedText.length);

  const formatted = titleize(title.value);
  if (title.value != formatted) {
    console.log(
      `Formatted input text "%s" is formatted to "%s".`,
      title.value,
      formatted
    );
    title.value = formatted;
  }
}

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
    "betas",
    "but",
    "by",
    "compatibility",
    "crack",
    "cracks",
    "demo",
    "demos",
    "doc",
    "docs",
    "documentation",
    "fix",
    "fixes",
    "for",
    "final",
    "from",
    "if",
    "in",
    "installer",
    "installers",
    "is",
    "hint",
    "hints",
    "map",
    "maps",
    "mod",
    "mods",
    "of",
    "on",
    "or",
    "patch",
    "patches",
    "part",
    "prerelease",
    "prereleases",
    "preview",
    "previews",
    "proper",
    "release",
    "releases",
    "repack",
    "repacks",
    "rip",
    "rips",
    "so",
    "solve",
    "solves",
    "the",
    "trainer",
    "trainers",
    "to",
    "update",
    "updates",
    "unprotect",
    "unprotects",
    "vs",
    "with",
  ];
  const uppers = [
    ".com",
    ".exe",
    "2d",
    "3d",
    "4d",
    "abc",
    "ad&d",
    "api",
    "bbs",
    "bios",
    "bsa",
    "cd",
    "cga",
    "dos",
    "dox",
    "dvd",
    "ega",
    "ehq",
    "f1",
    "fbi",
    "ftp",
    "hd",
    "hq",
    "ibm",
    "id",
    "iso",
    "la",
    "lego",
    "ls",
    "masm",
    "mbl",
    "ms",
    "mtv",
    "nascar",
    "nba",
    "ncaa",
    "nfl",
    "nsa",
    "nfo",
    "nhl",
    "nt",
    "oem",
    "os",
    "pc",
    "pcb",
    "pga",
    "ppe",
    "pfs",
    "pkarc",
    "pkzip",
    "psx",
    "rac",
    "rom",
    "sdk",
    "sfx",
    "spa",
    "tv",
    "ufo",
    "usa",
    "ushq",
    "uss",
    "vga",
    "whq",
    "ww1",
    "ww2",
    "ww3",
    "wwf",
    "xp",
    "ys",
  ];
  text = text.trim();
  // Replace all underscores with spaces
  text = text.replace(/_/g, " ");
  // if text contains 4 or more periods, place all periods with spaces
  if (text.match(/\./g) && text.match(/\./g).length > 4) {
    text = text.replace(/\./g, " ");
  }
  // Replace all single quotes, double quotes and graves with nothing
  text = text.replace(/['"`]/g, "");
  // Remove suffix (1) (2) (3) etc. (a) (b) (c) etc.
  text = text.replace(/ \([0-9a-z]\)/g, "");
  // Insert a space after a colon following an alphanumeric string
  text = text.replace(/([a-zA-Z0-9]): /g, "$1 : ");
  // Special fixes for specific strings
  text = text.replace(/([xX][-| ][Mm]en)/g, "X-Men"); // avoid roman numeral conversion
  // Insert temporary spaces between parentheses
  text = text.replace(/\(/g, "( ");
  text = text.replace(/\)/g, " )");
  const wordCount = text.split(" ").length;
  text = text
    .split(" ")
    .map((word, index) => {
      var edit = word;
      if (versionMatch(edit) === true) {
        return edit.toLowerCase();
      }
      if (uppers.includes(edit.toLowerCase())) {
        return edit.toUpperCase();
      }
      const y = replacementFix(edit);
      if (y !== "") {
        return y;
      }
      const x = romanFix(edit);
      if (index > 0 && Number.isInteger(x)) {
        edit = `${x}`;
      }
      const z = tailFix(edit, index, wordCount);
      if (z !== edit) {
        edit = z;
      }
      // return any edits
      if (word !== edit) {
        return edit;
      }
      // Capitalize the first word and any word not in the common adverbs list
      // Convert all other words to lowercase
      if (index === 0 || !commonAdverbs.includes(edit.toLowerCase())) {
        return edit.charAt(0).toUpperCase() + edit.slice(1).toLowerCase();
      }
      return word.toLowerCase();
    })
    .join(" ")
    .trim();
  // replace word pairs
  text = text.replace(/(Lotus 123)/g, "Lotus 1-2-3");
  text = text.replace(/(Falcon at )/g, "Falcon AT ");
  text = text.replace(/(the Games)/g, "The Games");
  // Move "Unprotect for" to the suffix if it is the prefix
  text = text.replace(/^(Unprotect for )(.+)/, "$2 unprotect");
  text = text.replace(/^(Unprotecting )(.+)/, "$2 unprotect");
  text = text.replace(/^(Unprotect )(.+)/, "$2 unprotect");
  // replace formatting quirks
  text = text.replace(/( : a)/g, " : A");
  text = text.replace(/( - a)/g, " - A");
  text = text.replace(/( : t)/g, " : T");
  text = text.replace(/( - t)/g, " - T");
  text = text.replace(/(f-)/g, "F-");
  text = text.replace(/(3-d)/g, "3D");
  text = text.replace(/(Pfs-)/g, "PFS-");
  text = text.replace(/(Mean-18)/g, "Mean 18");
  // replace v1.0 with v1, v2.0 with v2 etc.
  text = text.replace(/(v\d+)\.0/g, "$1");
  // remove temporary space between parentheses
  text = text.replace(/\( /g, "(");
  text = text.replace(/ \)/g, ")");
  // lowercase all text between square brackets
  text = text.replace(/\[([^)]+)\]/g, function (match) {
    return match.toUpperCase();
  });

  return text;
}

// Matches a version string.
// @param {string} word - The word to be checked.
// @returns {boolean} - The string matching the version syntax.
export function versionMatch(word) {
  // regex that matches V1.0, V1.1, V11, V2.11, V2.11a etc.
  const regex = /^[vV]\d+(\.\d+)?[a-z]?$/i;
  return regex.test(word);
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
 * Converts brands or elite words and symbols to English or numerals.
 * @param {string} word - The string to be converted.
 * @returns {string} - The converted string.
 * @example
 * replacementFix("][") // returns "2"
 */
export function replacementFix(word) {
  const s = word.toLowerCase();
  switch (s) {
    case "dr.":
      return "Dr";
    case "jr.":
      return "Jr";
    case "ms.":
      return "Ms";
    case "u.s.a.":
      return "USA";
    case "abcs":
      return "ABCs";
    case "cdrip":
    case "cd-rip":
      return "CD RIP";
    case "pre-release":
      return "prerelease";
    case "&":
      return "and";
    case "ad+d":
      return "AD&D";
    case "][":
    case "||":
      return "2";
    case "]|[":
    case "]I[":
      return "3";
    case "]||[":
    case "]II[":
      return "4";
    case " I:":
      return " 1 :";
    case " II:":
      return " 2 :";
    case " III:":
      return " 3 :";
    case "at&t":
      return "AT&T";
    case "dbase":
      return "dBase";
    case "doubledos":
      return "DoubleDOS";
    case "fastback":
      return "FastBack";
    case "loadit":
      return "LoadIt";
    case "memoryshift":
      return "Memory Shift";
    case "multilink":
      return "MultiLink";
    case "paperboy":
      return "PaperBoy";
    case "pc-draw":
      return "PC-Draw";
    case "pcjr":
      return "PCjr";
    case "prokey":
      return "ProKey";
    case "rbase":
    case "r:base":
      return "R:Base";
    case "sidekick":
      return "SideKick";
    case "visicalc":
      return "VisiCalc";
    case "war-craft":
      return "Warcraft";
    case "wordstar":
      return "WordStar";
    default:
      return "";
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
