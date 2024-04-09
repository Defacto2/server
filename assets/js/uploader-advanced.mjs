import { validYear, validMonth, validDay } from "./uploader.mjs";

export default advancedUploader;

export function advancedUploader(uploaderId) {
  const element = document.getElementById(uploaderId);
  if (element == null) {
    throw new Error(`The ${uploaderId} element is null.`);
  }
  element.addEventListener("click", function () {
    const choose = "Choose...";
    let pass = true;
    advReset();
    if (advFile.value == "") {
      advFile.classList.add(invalid);
      pass = false;
    }
    if (advOS.value == "" || advOS.value == choose) {
      advOS.classList.add(invalid);
      pass = false;
    }
    if (advCat.value == "" || advCat.value == choose) {
      advCat.classList.add(invalid);
      pass = false;
    }
    if (advTitl.value == "") {
      advTitl.classList.add(invalid);
      pass = false;
    }
    if (advRels.value == "") {
      advRels.classList.add(invalid);
      pass = false;
    }
    if (validYear(advYear.value) == false) {
      advYear.classList.add(invalid);
      pass = false;
    }
    if (validMonth(advMonth.value) == false) {
      advMonth.classList.add(invalid);
      pass = false;
    }
    if (validDay(advDay.value) == false) {
      advDay.classList.add(invalid);
      pass = false;
    }
    if (pass == true) {
      advFrm.submit();
    }
  });
}

const invalid = "is-invalid";

const advFrm = document.getElementById("advancedUploader");
// Event listener for the advanced reset button
advFrm.addEventListener("reset", advReset);
function advReset() {
  advFile.classList.remove(invalid);
  advOS.classList.remove(invalid);
  advCat.classList.remove(invalid);
  advTitl.classList.remove(invalid);
  advRels.classList.remove(invalid);
  advYear.classList.remove(invalid);
  advMonth.classList.remove(invalid);
  advDay.classList.remove(invalid);
}

// Elements for the advanced uploader
const advFile = document.getElementById("advFile");
const advOS = document.getElementById("advSelOS");
const advCat = document.getElementById("advSelCat");
const advTitl = document.getElementById("advTitle");
const advRels = document.getElementById("releasersAdv");
const advYear = document.getElementById("advYear");
const advMonth = document.getElementById("advMonth");
const advDay = document.getElementById("advDay");
