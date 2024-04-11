import { validYear, validMonth } from "./helper.mjs";
import { getElmById } from "./helper.mjs";
export default submit;

const formId = `uploader-intro-form`,
  invalid = "is-invalid",
  none = "d-none",
  megabyte = 1024 * 1024,
  percentage = 100;

const form = getElmById(formId);
const file = getElmById("uploader-intro-file");
const year = getElmById("uploader-intro-year");
const month = getElmById("uploader-intro-month");
const releaser1 = getElmById("uploader-intro-releaser-1");
const list1 = getElmById("uploader-intro-list-1"),
  list2 = getElmById("uploader-intro-list-2");
const youtube = getElmById("uploader-intro-youtube");
const alert = getElmById("uploader-intro-alert");
const results = getElmById("uploader-intro-results");

form.addEventListener("reset", reset);
releaser1.addEventListener("input", validateRel1);
year.addEventListener("input", validateY);
month.addEventListener("input", validateM);
youtube.addEventListener("input", validateYT);

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
      console.error("Submit failed. Please check the form.");
      return;
    }
    reset();
    results.classList.remove(none);
  });
}

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
  file.addEventListener("change", function () {
    const file1 = this.files[0],
      removeSelection = "";
    alert.innerText = "";
    alert.classList.add(none);
    if (file1.size > 10 * megabyte) {
      errSize = Math.round(file1.size / megabyte);
      alert.innerText = `The chosen file is too big at ${errSize}MB, maximum size is 100MB.`;
      alert.classList.remove(none);
      this.value = removeSelection;
    }
  });
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
