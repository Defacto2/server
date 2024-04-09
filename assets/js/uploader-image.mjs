import { validYear, validMonth } from "./helper.mjs";
export default imageSubmit;

const imgFile = document.getElementById("imageFile");
const imgTitl = document.getElementById("imageTitle");
const imgRels = document.getElementById("imageReleasers");
const imgYear = document.getElementById("imageYear");
const imgMonth = document.getElementById("imageMonth");

const imgFrm = document.getElementById("imageUploader");
const invalid = "is-invalid";

/**
 * Resets the input fields for image upload.
 */
function imgReset() {
  imgFile.classList.remove(invalid);
  imgTitl.classList.remove(invalid);
  imgRels.classList.remove(invalid);
  imgYear.classList.remove(invalid);
  imgMonth.classList.remove(invalid);
}
// Event listener for the image reset button
imgFrm.addEventListener("reset", imgReset);

export function imageSubmit(elementId) {
  const element = document.getElementById(elementId);
  if (element == null) {
    throw new Error(`The ${elementId} element is null.`);
  }
  element.addEventListener("click", function () {
    let pass = true;
    imgReset();
    if (imgFile.value == "") {
      imgFile.classList.add(invalid);
      pass = false;
    }
    if (imgTitl.value == "") {
      imgTitl.classList.add(invalid);
      pass = false;
    }
    if (imgRels.value == "") {
      imgRels.classList.add(invalid);
      pass = false;
    }
    if (validYear(imgYear.value) == false) {
      imgYear.classList.add(invalid);
      pass = false;
    }
    if (validMonth(imgMonth.value) == false) {
      imgMonth.classList.add(invalid);
      pass = false;
    }
    if (pass == true) {
      imgFrm.submit();
    }
  });
}
