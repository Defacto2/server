/**
 * @module uploader
 * This module provides functions for handling file uploads.
 */
import { getElmById, validYear, validMonth, validDay } from "./helper.mjs";

const invalid = "is-invalid",
  none = "d-none",
  percentage = 100,
  megabyte = 1024 * 1024,
  sizeLimit = 100 * megabyte;

/**
 * Retrieves a modal object by its element ID.
 *
 * @param {string} elementId - The ID of the element representing the modal.
 * @returns {Object} - The modal object.
 * @throws {Error} - If the bootstrap object is undefined.
 */
export function getModalById(elementId) {
  if (elementId == null) {
    throw new Error(`The elementId value of getModalById is null.`);
  }
  if (bootstrap === undefined) {
    throw new Error(`The bootstrap object is undefined.`);
  }
  const element = getElmById(elementId);
  const modal = new bootstrap.Modal(element, {
    keyboard: true,
  });
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
  if (elementId == null) {
    throw new Error(`The elementId value of focusModalById is null.`);
  }
  if (submissionId == null) {
    throw new Error(`The submissionId value of focusModalById is null.`);
  }
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
  const modal = new bootstrap.Modal(element, {
    keyboard: true,
  });
  return modal;
}

/**
 * Checks the SHA-384 hash of a file by sending it to the server.
 * This function is a client convenience to save time and bandwidth.
 * If the client browser does not support the required APIs,
 * it does not matter as the hash is rechecked on the server after uploading.
 * @param {File} file - The file to be hashed.
 * @returns {Promise<string>} - A promise that returns an ID if the file is already in the database, otherwise an empty string.
 */
export async function checkSHA(file) {
  if (file == null) {
    throw new Error(`The file value of checkSHA is null.`);
  }

  const hash = await sha384(file);
  const response = await fetch(`/uploader/sha384/${hash}`, {
    method: "PATCH",
    headers: {
      "Content-Type": "text/plain",
    },
    body: hash,
  });
  if (!response.ok) {
    throw new Error(
      `Hashing is not possible, server response: ${response.status}`
    );
  }
  const responseText = await response.text();
  return responseText;
}

/**
 * Calculates the SHA-384 hash of a given file.
 *
 * @param {File} file - The file to calculate the hash for.
 * @returns {Promise<string>} A promise that resolves with the SHA-384 hash as a hexadecimal string.
 * @throws {Error} If the arrayBuffer or crypto.subtle APIs are not available.
 */
async function sha384(file) {
  if (file == null) {
    throw new Error(`The file value of sha384 is null.`);
  }
  try {
    const buffer = await file.arrayBuffer();
    const hashBuffer = await crypto.subtle.digest("SHA-384", buffer);
    return Array.from(new Uint8Array(hashBuffer))
      .map((b) => b.toString(16).padStart(2, "0"))
      .join("");
  } catch (e) {
    throw new Error(`Could not use arrayBuffer or crypto.subtle: ${e}`);
  }
}

/**
 * Updates the progress bar based on the upload progress.
 * Based on the htmx:xhr:progress example from https://htmx.org/examples/file-upload/.
 */
export function progress(formId, elementId) {
  if (formId == null) {
    throw new Error(`The formId value of progress is null.`);
  }
  if (elementId == null) {
    throw new Error(`The elementId value of progress is null.`);
  }
  htmx.on(`#${formId}`, "htmx:xhr:progress", function (event) {
    if (event.target.id != `${formId}`) return;
    htmx
      .find(`#${elementId}`)
      .setAttribute(
        "value",
        (event.detail.loaded / event.detail.total) * percentage
      );
  });
}

/**
 * Checks if the size of the file is within the specified limit.
 *
 * @param {File} file - The file to check the size of.
 * @returns {string} - An error message if the file size exceeds the limit, otherwise an empty string.
 * @throws {Error} - If the file parameter is null.
 */
export function checkSize(file) {
  if (file == null) {
    throw new Error(`The file value of checkSize is null.`);
  }
  if (file.size > sizeLimit) {
    const errSize = Math.round(file.size / megabyte);
    return `The chosen file is too big at ${errSize}MB, maximum size is ${sizeLimit / megabyte}MB.`;
  }
  return ``;
}

/**
 * Checks if the value of the element represents a valid day.
 * @returns {boolean} Returns true if the value is a valid day, false otherwise.
 */
export function checkDay() {
  if (validDay(this.value) == false) {
    this.classList.add(invalid);
    return false;
  }
  this.classList.remove(invalid);
  return true;
}

/**
 * Checks if the month value is valid.
 * @returns {boolean} Returns true if the month value is valid, false otherwise.
 */
export function checkMonth() {
  console.log(`The month value is ${this.value}.`);
  if (validMonth(this.value) == false) {
    this.classList.add(invalid);
    return false;
  }
  this.classList.remove(invalid);
  return true;
}

/**
 * Checks if the value of the input field represents a valid year.
 * @returns {boolean} Returns true if the year is valid, false otherwise.
 */
export function checkYear() {
  if (validYear(this.value) == false) {
    this.classList.add(invalid);
    return false;
  }
  this.classList.remove(invalid);
  return true;
}

/**
 * Checks for errors and updates the UI accordingly.
 *
 * @param {Array} errors - The array of errors.
 * @param {HTMLElement} alert - The alert element to display the errors.
 * @param {HTMLElement} fileInput - The file input element.
 * @param {HTMLElement} results - The results element.
 * @throws {Error} If any of the parameters are null.
 */
export function checkErrors(errors, alert, fileInput, results) {
  if (errors == null) {
    throw new Error(`The errors value of checkErrors is null.`);
  }
  if (alert == null) {
    throw new Error(`The alert value of checkErrors is null.`);
  }
  if (fileInput == null) {
    throw new Error(`The fileInput value of checkErrors is null.`);
  }
  if (results == null) {
    throw new Error(`The results value of checkErrors is null.`);
  }
  errors = errors.filter((error) => error.trim() != "");
  if (errors.length <= 0) {
    return;
  }
  alert.innerText = errors.join(" ");
  alert.classList.remove(none);
  fileInput.innerText = "";
  fileInput.classList.add(invalid);
  results.classList.add(none);
}

/**
 * Checks if a file is a duplicate and performs necessary actions.
 *
 * @param {File} file - The file to check for duplication.
 * @param {HTMLElement} alert - The element to display the alert message.
 * @param {HTMLElement} fileInput - The element to clear the file input value.
 * @param {HTMLElement} results - The element to hide the results.
 * @throws {Error} If any of the parameters are null.
 */
export async function checkDuplicate(file, alert, fileInput, results) {
  if (file == null) {
    throw new Error(`The file value of checkDuplicate is null.`);
  }
  if (alert == null) {
    throw new Error(`The alert value of checkDuplicate is null.`);
  }
  if (fileInput == null) {
    throw new Error(`The fileInput value of checkDuplicate is null.`);
  }
  if (results == null) {
    throw new Error(`The results value of checkDuplicate is null.`);
  }

  const alerter = () => {
    alert.classList.remove(none);
    fileInput.innerText = "";
    fileInput.classList.add(invalid);
    results.classList.add(none);
  };

  let uriID = ``;
  try {
    const alreadyExists = await checkSHA(file);
    if (alreadyExists == "") {
      alert.innerText = ``;
      return;
    }
    uriID = alreadyExists;
  } catch (e) {
    console.log(`${e}`);
    alert.innerText = `${e}`;
    alerter();
    return;
  }

  alert.innerText = `The chosen file already exists in the database: `;
  const anchor = document.createElement("a");
  anchor.href = `/f/${uriID}`;
  anchor.innerText = `${file.name}`;
  alert.appendChild(anchor);
  alerter();
}

/**
 * Updates the hidden details based on the provided file information.
 *
 * @param {File} file1 - The file object.
 * @param {HTMLInputElement} lastMod - The input element for the last modified value.
 * @param {HTMLInputElement} magic - The input element for the magic value.
 * @throws {Error} If any of the parameters are null.
 */
export function hiddenDetails(file1, lastMod, magic) {
  if (file1 == null) {
    throw new Error(`The file1 value of hiddenDetails is null.`);
  }
  if (lastMod == null) {
    throw new Error(`The lastMod value of hiddenDetails is null.`);
  }
  if (magic == null) {
    throw new Error(`The magic value of hiddenDetails is null.`);
  }

  const lastModified = file1.lastModified,
    currentTime = new Date().getTime(),
    oneHourMs = 60 * 60 * 1000;
  const underOneHour = currentTime - lastModified < oneHourMs;
  if (!underOneHour) {
    lastMod.value = lastModified;
  }
  if (file1.type != "") {
    magic.value = file1.type;
  }
}

/**
 * Checks if the value of an input element is empty.
 * @returns {boolean} Returns true if the value is not empty, false otherwise.
 */
export function checkValue() {
  if (this.value == "") {
    this.classList.add(invalid);
    return false;
  }
  this.classList.remove(invalid);
  return true;
}

/**
 * Checks if the value of an input field is a valid YouTube video ID.
 * @returns {boolean} Returns true if the value is empty or a valid YouTube video ID, otherwise returns false.
 */
export function checkYouTube() {
  if (this.value == "") {
    this.classList.remove(invalid);
    return true;
  }
  const re = new RegExp(/^[a-zA-Z0-9_-]{11}$/);
  if (re.test(this.value) == false) {
    this.classList.add(invalid);
    return false;
  }
  this.classList.remove(invalid);
  return true;
}

/**
 * Resets the input fields and elements associated with file uploading.
 *
 * @param {HTMLElement} fileInput - The file input element.
 * @param {HTMLElement} alert - The alert element.
 * @param {HTMLElement} results - The results element.
 * @throws {Error} If any of the input parameters are null.
 */
export function resetInput(fileInput, alert, results) {
  if (fileInput == null) {
    throw new Error(`The fileInput value of resetInput is null.`);
  }
  if (alert == null) {
    throw new Error(`The alert value of resetInput is null.`);
  }
  if (results == null) {
    throw new Error(`The results value of resetInput is null.`);
  }
  fileInput.innerText = "";
  fileInput.classList.remove(invalid);
  alert.innerText = "";
  alert.classList.add(none);
  results.innerText = "";
  results.classList.add(none);
}

/**
 * Updates the error message and results display based on the provided alert and results elements.
 * @param {HTMLElement} alert - The alert element to update.
 * @param {HTMLElement} results - The results element to update.
 * @throws {Error} If the alert or results value is null.
 */
export function submitError(alert, results) {
  if (alert == null) {
    throw new Error(`The alert value of submitError is null.`);
  }
  if (results == null) {
    throw new Error(`The results value of submitError is null.`);
  }
  alert.innerText = "Please correct the problems with the form.";
  alert.classList.remove(none);
  results.innerText = "";
  results.classList.add(none);
}
