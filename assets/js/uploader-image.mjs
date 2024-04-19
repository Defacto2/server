// uploader-image.mjs
import { getElmById } from "./helper.mjs";
import { checkImage as mime } from "./uploader-mime.mjs";
import {
  checkDuplicate,
  checkErrors,
  checkSize,
  checkMonth,
  checkYear,
  checkReleaser,
  hiddenDetails,
  submitError,
  resetInput,
} from "./uploader.mjs";
import { validYear, validMonth } from "./helper.mjs";
export default submit;

const formId = `uploader-image-form`,
  invalid = "is-invalid",
  none = "d-none";

const form = getElmById(formId),
  alert = getElmById("uploader-image-alert"),
  fileInput = getElmById("uploader-image-file"),
  lastMod = getElmById("uploader-image-last-modified"),
  list1 = getElmById("uploader-image-list-1"),
  list2 = getElmById("uploader-image-list-2"),
  magic = getElmById("uploader-image-magic"),
  month = getElmById("uploader-image-month"),
  releaser1 = getElmById("uploader-image-releaser-1"),
  results = getElmById("uploader-image-results"),
  year = getElmById("uploader-image-year");

form.addEventListener("reset", function () {
  lastMod.value = "";
  magic.value = "";
  resetForm();
});

fileInput.addEventListener("change", checkFile);
releaser1.addEventListener("input", checkReleaser);
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
  list2.innerHTML = "";
  results.innerHTML = "";
  results.classList.add(none);
  alert.innerText = "";
  alert.classList.add(none);
  year.classList.remove(invalid);
  month.classList.remove(invalid);
  releaser1.classList.remove(invalid);
  fileInput.classList.remove(invalid);
}
