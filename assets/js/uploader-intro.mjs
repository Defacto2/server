import { validYear, validMonth } from "./helper.mjs";
export default introSubmit;

const introFrm = document.getElementById("introUploader");

// Elements for the intro uploader
const introFile = document.getElementById("introFile");
const introTitl = document.getElementById("releaseTitle");
const introRels = document.getElementById("introReleasers");
const introYear = document.getElementById("introYear");
const introMonth = document.getElementById("introMonth");

const invalid = "is-invalid";
/**
 * Resets the input fields for the intro section of the uploader form.
 */
function introReset() {
  introFile.classList.remove(invalid);
  introTitl.classList.remove(invalid);
  introRels.classList.remove(invalid);
  introYear.classList.remove(invalid);
  introMonth.classList.remove(invalid);
}
introFrm.addEventListener("reset", introReset);

export function introSubmit(elementId) {
  const element = document.getElementById(elementId);
  if (element == null) {
    throw new Error(`The ${elementId} element is null.`);
  }
  element.addEventListener("click", function () {
    let pass = true;
    introReset();
    if (introFile.value == "") {
      introFile.classList.add(invalid);
      pass = false;
    }
    if (introTitl.value == "") {
      introTitl.classList.add(invalid);
      pass = false;
    }
    if (introRels.value == "") {
      introRels.classList.add(invalid);
      pass = false;
    }
    if (validYear(introYear.value) == false) {
      introYear.classList.add(invalid);
      pass = false;
    }
    if (validMonth(introMonth.value) == false) {
      introMonth.classList.add(invalid);
      pass = false;
    }
    if (pass == true) {
      introFrm.submit();
    }
  });
}
