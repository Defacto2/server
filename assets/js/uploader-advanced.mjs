/**
 * @module uploader-advanced
 * This module provides functions for handling file uploads advanced UI.
 */
import { validYear, validMonth, validDay } from "./helper.mjs";
import { getElmById } from "./helper.mjs";
import { checkAdvanced as mime } from "./uploader-mime.mjs";
import {
  checkDuplicate,
  checkErrors,
  checkSize,
  checkDay,
  checkMonth,
  checkYear,
  checkValue,
  hiddenDetails,
  submitError,
  resetInput,
} from "./uploader.mjs";

export default submit;

const formId = `uploader-advanced-form`,
  invalid = "is-invalid",
  none = "d-none";

const form = getElmById(formId),
  alert = getElmById("uploader-advanced-alert"),
  category = getElmById("uploader-advanced-category"),
  classification = getElmById("uploader-advanced-classification-help"),
  day = getElmById("uploader-advanced-day"),
  fileInput = getElmById("uploader-advanced-file"),
  lastMod = getElmById("uploader-advanced-last-modified"),
  list1 = getElmById("uploader-advanced-list-1"),
  list2 = getElmById("uploader-advanced-list-2"),
  magic = getElmById("uploader-advanced-magic"),
  month = getElmById("uploader-advanced-month"),
  os = getElmById("uploader-advanced-operating-system"),
  releaser1 = getElmById("uploader-advanced-releaser-1"),
  results = getElmById("uploader-advanced-results"),
  year = getElmById("uploader-advanced-year"); //,
//youtube = getElmById("uploader-advanced-youtube");

form.addEventListener("reset", function () {
  lastMod.value = "";
  magic.value = "";
  resetForm();
});

fileInput.addEventListener("change", checkFile);
releaser1.addEventListener("input", checkValue);
year.addEventListener("input", checkYear);
month.addEventListener("input", checkMonth);
day.addEventListener("input", checkDay);
category.addEventListener("change", checkValue);
os.addEventListener("change", checkValue);

/**
 * After performing input validations this submits the form when the specified element is clicked.
 * @param {string} elementId - The ID of the element that triggers the form submission, e.g. a button element type.
 */
export function submit(elementId) {
  const element = getElmById(elementId);
  element.addEventListener("click", function () {
    let pass = true;
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
    if (validDay(day.value) == false) {
      day.classList.add(invalid);
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
    if (os.value == "") {
      os.classList.add(invalid);
      pass = false;
    }
    if (category.value == "") {
      category.classList.add(invalid);
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
  if (mime(file.type)) {
    return `The chosen file mime type ${file.type} is probably not suitable for an upload.`;
  }
  return ``;
}

function resetForm() {
  list1.innerHTML = "";
  list2.innerHTML = "";
  results.innerHTML = "";
  classification.innerHTML = "";
  results.classList.add(none);
  alert.innerText = "";
  alert.classList.add(none);
  year.classList.remove(invalid);
  month.classList.remove(invalid);
  day.classList.remove(invalid);
  releaser1.classList.remove(invalid);
  fileInput.classList.remove(invalid);
  os.classList.remove(invalid);
  category.classList.remove(invalid);
}
