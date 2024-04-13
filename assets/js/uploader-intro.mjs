import { validYear, validMonth } from "./helper.mjs";
import { getElmById } from "./helper.mjs";
import { intro as mime } from "./uploader-mime.mjs";
export default submit;

const formId = `uploader-intro-form`,
  invalid = "is-invalid",
  none = "d-none",
  megabyte = 1024 * 1024,
  sizeLimit = 100 * megabyte,
  percentage = 100;

const form = getElmById(formId),
  alert = getElmById("uploader-intro-alert"),
  file = getElmById("uploader-intro-file"),
  list1 = getElmById("uploader-intro-list-1"),
  list2 = getElmById("uploader-intro-list-2"),
  month = getElmById("uploader-intro-month"),
  releaser1 = getElmById("uploader-intro-releaser-1"),
  results = getElmById("uploader-intro-results"),
  year = getElmById("uploader-intro-year"),
  youtube = getElmById("uploader-intro-youtube");

form.addEventListener("reset", reset);

file.addEventListener("change", checks);

releaser1.addEventListener("input", validateRel1);
year.addEventListener("input", validateY);
month.addEventListener("input", validateM);
youtube.addEventListener("input", validateYT);

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
    if (file.value == "") {
      file.classList.add(invalid);
      pass = false;
    }
    if (pass == false) {
      alert.innerText = "Please correct the problems with the form.";
      alert.classList.remove(none);
      return;
    }
    reset();
    results.classList.remove(none);
  });
}

/**
 * Updates the progress bar based on the upload progress.
 * Based on the htmx:xhr:progress example from https://htmx.org/examples/file-upload/.
 */
export function progress() {
  htmx.on(`#${formId}`, "htmx:xhr:progress", function (event) {
    if (event.target.id != `${formId}`) return;
    htmx
      .find("#uploader-intro-progress")
      .setAttribute(
        "value",
        (event.detail.loaded / event.detail.total) * percentage
      );
  });
}

class checks {
  constructor() {
    const file1 = this.files[0],
      removeSelection = "";

    file.classList.remove(invalid);
    alert.innerText = "";
    alert.classList.add(none);

    let errors = [checkSize(file1), checkMime(file1)];
    errors = errors.filter((error) => error != "");

    if (errors.length > 0) {
      alert.innerText = errors.join(" ");
      alert.classList.remove(none);
      this.value = removeSelection;
      this.classList.add(invalid);
      return;
    }
  }
}

function checkSize(file) {
  if (file.size > sizeLimit) {
    const errSize = Math.round(file.size / megabyte);
    return `The chosen file is too big at ${errSize}MB, maximum size is ${sizeLimit / megabyte}MB.`;
  }
  return ``;
}

function checkMime(file) {
  if (!mime(file.type)) {
    return `The chosen file mime type ${file.type} is probably not suitable for an intro.`;
  }
  return ``;
}

function reset() {
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
  file.classList.remove(invalid);
}

function validateRel1() {
  if (releaser1.value == "") {
    releaser1.classList.add(invalid);
    return false;
  }
  releaser1.classList.remove(invalid);
  return true;
}

function validateM() {
  if (validMonth(month.value) == false) {
    month.classList.add(invalid);
    return false;
  }
  month.classList.remove(invalid);
  return true;
}

function validateY() {
  if (validYear(year.value) == false) {
    year.classList.add(invalid);
    return false;
  }
  year.classList.remove(invalid);
  return true;
}

function validateYT() {
  if (youtube.value == "") {
    youtube.classList.remove(invalid);
    return true;
  }
  const re = new RegExp(/^[a-zA-Z0-9_-]{11}$/);
  if (re.test(youtube.value) == false) {
    youtube.classList.add(invalid);
    return false;
  }
  youtube.classList.remove(invalid);
  return true;
}
