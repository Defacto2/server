// uploader.mjs

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
 * @returns {Promise<boolean>} - A promise that resolves to true if the server confirms the hash, false otherwise.
 */
export async function checkSHA(file) {
  if (file == null) {
    throw new Error(`The file value of checkSHA is null.`);
  }
  try {
    const hash = await sha384(file);
    const response = await fetch(`/uploader/sha384/${hash}`, {
      method: "PUT",
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
    return responseText == "true";
  } catch (e) {
    console.log(`Hashing is not possible: ${e}`);
  }
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

export function checkReleaser() {
  if (this.value == "") {
    this.classList.add(invalid);
    return false;
  }
  this.classList.remove(invalid);
  return true;
}

export function checkDay() {
  if (validDay(this.value) == false) {
    this.classList.add(invalid);
    return false;
  }
  this.classList.remove(invalid);
  return true;
}

export function checkMonth() {
  console.log(`The month value is ${this.value}.`);
  if (validMonth(this.value) == false) {
    this.classList.add(invalid);
    return false;
  }
  this.classList.remove(invalid);
  return true;
}

export function checkYear() {
  if (validYear(this.value) == false) {
    this.classList.add(invalid);
    return false;
  }
  this.classList.remove(invalid);
  return true;
}

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
  errors = errors.filter((error) => error != "");
  if (errors.length <= 0) {
    return;
  }
  alert.innerText = errors.join(" ");
  alert.classList.remove(none);
  fileInput.innerText = "";
  fileInput.classList.add(invalid);
  results.classList.add(none);
}

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

  const alreadyExists = await checkSHA(file);
  if (alreadyExists == false) {
    return;
  }
  alert.innerText = `The chosen file already exists in the database: ${file.name}`;
  alert.classList.remove(none);
  fileInput.innerText = "";
  fileInput.classList.add(invalid);
  results.classList.add(none);
}

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

export function checkValue() {
  if (this.value == "") {
    this.classList.add(invalid);
    return false;
  }
  this.classList.remove(invalid);
  return true;
}
