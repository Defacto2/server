import { validYear, validMonth, validDay } from "./helper.mjs";
export default magazineSubmit;
// Elements for the magazine uploader
const magFile = document.getElementById("magFile");
const magTitl = document.getElementById("magTitle");
const magIssu = document.getElementById("magIssue");
const magYear = document.getElementById("magYear");
const magMonth = document.getElementById("magMonth");
const magDay = document.getElementById("magDay");
const invalid = "is-invalid";
function magReset() {
  magFile.classList.remove(invalid);
  magTitl.classList.remove(invalid);
  magIssu.classList.remove(invalid);
  magYear.classList.remove(invalid);
  magMonth.classList.remove(invalid);
  magDay.classList.remove(invalid);
}
const magFrm = document.getElementById("magUploader");

export function magazineSubmit(elementId) {
  const element = document.getElementById(elementId);
  if (element == null) {
    throw new Error(`The ${elementId} element is null.`);
  }
  element.addEventListener("click", function () {
    let pass = true;
    magReset();
    if (magFile.value == "") {
      magFile.classList.add(invalid);
      pass = false;
    }
    if (magTitl.value == "") {
      magTitl.classList.add(invalid);
      pass = false;
    }
    if (magIssu.value == "") {
      magIssu.classList.add(invalid);
      pass = false;
    }
    if (validYear(magYear.value) == false) {
      magYear.classList.add(invalid);
      pass = false;
    }
    if (validMonth(magMonth.value) == false) {
      magMonth.classList.add(invalid);
      pass = false;
    }
    if (validDay(magDay.value) == false) {
      magDay.classList.add(invalid);
      pass = false;
    }
    if (pass == true) {
      magFrm.submit();
    }
  });
}
