import { validYear, validMonth } from "./uploader.mjs";
export default textSubmit;

const txtFile = document.getElementById("textFile");
const txtTitl = document.getElementById("textTitle");
const txtRels = document.getElementById("textReleasers");
const txtYear = document.getElementById("textYear");
const txtMonth = document.getElementById("textMonth");

const txtFrm = document.getElementById("textUploader");

const invalid = "is-invalid";
/**
 * Resets the input fields for file, title, release date, and year by removing the 'invalid' class.
 */
function txtReset() {
  txtFile.classList.remove(invalid);
  txtTitl.classList.remove(invalid);
  txtRels.classList.remove(invalid);
  txtYear.classList.remove(invalid);
  txtMonth.classList.remove(invalid);
}
txtFrm.addEventListener("reset", txtReset);

export function textSubmit(elementId) {
  const element = document.getElementById(elementId);
  if (element == null) {
    throw new Error(`The ${elementId} element is null.`);
  }
  element.addEventListener("click", function () {
    let pass = true;
    txtReset();
    if (txtFile.value == "") {
      txtFile.classList.add(invalid);
      pass = false;
    }
    if (txtTitl.value == "") {
      txtTitl.classList.add(invalid);
      pass = false;
    }
    if (txtRels.value == "") {
      txtRels.classList.add(invalid);
      pass = false;
    }
    if (validYear(txtYear.value) == false) {
      txtYear.classList.add(invalid);
      pass = false;
    }
    if (validMonth(txtMonth.value) == false) {
      txtMonth.classList.add(invalid);
      pass = false;
    }
    if (pass == true) {
      txtFrm.submit();
    }
  });
}
