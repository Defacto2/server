// uploader-text.mjs
import { getElmById, validYear, validMonth } from "./uploader.mjs";
export default textSubmit;

/**
 * Submits the text form when the specified element is clicked.
 * @param {string} elementId - The ID of the element that triggers the form submission.
 * @param {string} formId - The ID of the form to be submitted.
 * @throws {Error} If the specified element or form is null.
 */
export function textSubmit(elementId, formId) {
  const element = document.getElementById(elementId);
  if (element == null) {
    throw new Error(`The ${elementId} element is null.`);
  }
  const form = document.getElementById(formId);
  if (form == null) {
    throw new Error(`The ${formId} form element is null.`);
  }
  form.addEventListener("reset", reset);
  element.addEventListener("click", function () {
    if (validate() == true) {
      form.submit();
    }
  });
}

const file = getElmById("textFile"),
  title = getElmById("textTitle"),
  releasers = getElmById("textReleasers"),
  year = getElmById("textYear"),
  month = getElmById("textMonth"),
  invalid = "is-invalid";

function reset() {
  file.classList.remove(invalid);
  title.classList.remove(invalid);
  releasers.classList.remove(invalid);
  year.classList.remove(invalid);
  month.classList.remove(invalid);
}

function validate() {
  let pass = true;
  reset();
  if (file.value == "") {
    file.classList.add(invalid);
    pass = false;
  }
  if (title.value == "") {
    title.classList.add(invalid);
    pass = false;
  }
  if (releasers.value == "") {
    releasers.classList.add(invalid);
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
  return pass;
}
