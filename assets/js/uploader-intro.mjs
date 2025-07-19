/**
 * @module uploader-intro
 * This module provides functions for handling file uploads intro UI.
 */
import { formatPaste, getElmById, validYear, validMonth } from "./helper.mjs";
import { checkIntro as mime } from "./uploader-mime.mjs";
import {
  checkDuplicate,
  checkErrors,
  checkSize,
  checkMonth,
  checkValue,
  checkYouTube,
  hiddenDetails,
  submitError,
  resetInput,
} from "./uploader.mjs";
export default submit;

const formId = `uploader-intro-form`,
  invalid = "is-invalid",
  none = "d-none";

const form = getElmById(formId),
  alert = getElmById("uploader-intro-alert"),
  fileInput = getElmById("uploader-intro-file"),
  lastMod = getElmById("uploader-intro-last-modified"),
  list1 = getElmById("uploader-intro-list-1"),
  list2 = getElmById("uploader-intro-list-2"),
  magic = getElmById("uploader-intro-magic"),
  month = getElmById("uploader-intro-month"),
  releaser1 = getElmById("uploader-intro-releaser-1"),
  results = getElmById("uploader-intro-results"),
  title = getElmById("uploader-intro-title"),
  year = getElmById("uploader-intro-year"),
  youtube = getElmById("uploader-intro-youtube");

form.addEventListener("reset", function () {
  lastMod.value = "";
  magic.value = "";
  resetForm();
});

fileInput.addEventListener("change", checkFile);
title.addEventListener("paste", formatPaste);
releaser1.addEventListener("input", checkValue);
month.addEventListener("input", checkMonth);
youtube.addEventListener("input", checkYouTube);
year.addEventListener("input", function () {
  const currentYear = new Date().getFullYear();
  // year values of 80-99 will automatically be prefixed with a 19, aka 1980-1999.
  const comp19 = parseInt(this.value) + 1900;
  if (comp19 >= 1980 && comp19 < 2000 && year.value.length == 2) {
    year.value = comp19;
    year.classList.remove(invalid);
    return;
  }
  // year values of 00 through to the current year will automatically be prefixed with a 20, aka 2000-2025.
  // however, year values of 19(xx) and 20(xx) are ignored as they create a weird UI situation when users edit year values.
  const comp20 = parseInt(this.value) + 2000;
  if (
    comp20 >= 2000 &&
    comp20 <= currentYear &&
    comp20 != 2019 &&
    comp20 != 2020 &&
    year.value.length == 2
  ) {
    year.value = comp20;
    year.classList.remove(invalid);
    return;
  }
  if (validYear(this.value) == false) {
    this.classList.add(invalid);
  }
  this.classList.remove(invalid);
});

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
  if (errors[0] === "" && errors[1] === "") {
    document
      .getElementById("uploader-intro-submit")
      .focus({ focusVisible: true });
  }
}

function checkMime(file) {
  if (!mime(file.type)) {
    return `The chosen file mime type ${file.type} might not be suitable for an intro.`;
  }
  return ``;
}

function resetForm() {
  list1.innerHTML = "";
  list2.innerHTML = "";
  results.innerHTML = "";
  results.classList.add(none);
  alert.innerText = "";
  alert.classList.add(none);
  year.classList.remove(invalid);
  month.classList.remove(invalid);
  releaser1.classList.remove(invalid);
  youtube.classList.remove(invalid);
  fileInput.classList.remove(invalid);
}
