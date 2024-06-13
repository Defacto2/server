/**
 * @module uploader-magazine
 * This module provides functions for handling file uploads magazine UI.
 */
import { getElmById } from "./helper.mjs";
import { checkMagazine as mime } from "./uploader-mime.mjs";
import {
  checkDuplicate,
  checkErrors,
  checkSize,
  checkMonth,
  checkYear,
  checkValue,
  hiddenDetails,
  submitError,
  resetInput,
} from "./uploader.mjs";
import { validYear, validMonth } from "./helper.mjs";
export default submit;

const formId = `uploader-magazine-form`,
  invalid = "is-invalid",
  none = "d-none";

const form = getElmById(formId),
  alert = getElmById("uploader-magazine-alert"),
  fileInput = getElmById("uploader-magazine-file"),
  lastMod = getElmById("uploader-magazine-last-modified"),
  list1 = getElmById("uploader-magazine-list-1"),
  magic = getElmById("uploader-magazine-magic"),
  month = getElmById("uploader-magazine-month"),
  releaser1 = getElmById("uploader-magazine-releaser-1"),
  results = getElmById("uploader-magazine-results"),
  title = getElmById("uploader-magazine-title"),
  year = getElmById("uploader-magazine-year");

form.addEventListener("reset", function () {
  lastMod.value = "";
  magic.value = "";
  resetForm();
});

fileInput.addEventListener("change", checkFile);
releaser1.addEventListener("input", checkValue);
year.addEventListener("input", checkYear);
month.addEventListener("input", checkMonth);

/**
 * After performing input validations this submits the form when the specified element is clicked.
 * @param {string} elementId - The ID of the element that triggers the form submission, e.g. a button element type.
 */
export function submit(elementId) {
  const element = getElmById(elementId);
  element.addEventListener("click", function () {
    let pass = true;
    if (title.value == "") {
      title.classList.add(invalid);
      pass = false;
    }
    if (releaser1.value == "") {
      releaser1.classList.add(invalid);
      pass = false;
    }
    if (validYear(year.value) == false) {
      year.classList.add(invalid);
      pass = false;
    }
    if (validMonth(month.value) == false) {
      month.classList.add(invalid);
      pass = false;
    }
    if (month.value != "" && year.value == "") {
      year.classList.add(invalid);
      pass = false;
    }
    if (fileInput.value == "") {
      fileInput.classList.add(invalid);
      pass = false;
    }
    if (pass == false) {
      return submitError(alert, results);
    }
    resetForm();
    results.innerText = "...";
    results.classList.remove(none);
  });
}

async function checkFile() {
  resetInput(fileInput, alert, results);
  const file1 = this.files[0];
  let errors = [checkSize(file1), checkMime(file1)];
  checkErrors(errors, alert, fileInput, results);
  checkDuplicate(file1, alert, fileInput, results);
  hiddenDetails(file1, lastMod, magic);
}

function checkMime(file) {
  if (!mime(file.type)) {
    return `The chosen file mime type ${file.type} is probably not suitable for an image.`;
  }
  return ``;
}

function resetForm() {
  list1.innerHTML = "";
  results.innerHTML = "";
  results.classList.add(none);
  alert.innerText = "";
  alert.classList.add(none);
  year.classList.remove(invalid);
  month.classList.remove(invalid);
  releaser1.classList.remove(invalid);
  fileInput.classList.remove(invalid);
  title.classList.remove(invalid);
}
