(() => {
  "use strict";

  const invalid = "is-invalid";

  /**
   * Checks if a given year is valid, i.e. between 1980 and the current year.
   * @param {number} year - The year to be validated.
   * @returns {boolean} - Returns true if the year is valid, false otherwise.
   */
  function validYear(year) {
    if (`${year}` == "") {
      return true;
    }
    const currentYear = new Date().getFullYear();
    if (year < 1980 || year > currentYear) {
      return false;
    }
    return true;
  }

  /**
   * Checks if a given month is valid.
   * @param {number} month - The month to be validated.
   * @returns {boolean} - Returns true if the month is valid, false otherwise.
   */
  function validMonth(month) {
    if (`${month}` == "") {
      return true;
    }
    if (month < 1 || month > 12) {
      return false;
    }
    return true;
  }

  /**
   * Checks if a given day is valid.
   * @param {number} day - The day to be checked.
   * @returns {boolean} - Returns true if the day is valid, false otherwise.
   */
  function validDay(day) {
    if (`${day}` == "") {
      return true;
    }
    if (day < 1 || day > 31) {
      return false;
    }
    return true;
  }

  const introFrm = document.getElementById("introUploader");
  const introFile = document.getElementById("introFile");
  const introTitl = document.getElementById("releaseTitle");
  const introRels = document.getElementById("introReleasers");
  const introYear = document.getElementById("introYear");
  const introMonth = document.getElementById("introMonth");

  function introReset() {
    introFile.classList.remove(invalid);
    introTitl.classList.remove(invalid);
    introRels.classList.remove(invalid);
    introYear.classList.remove(invalid);
    introMonth.classList.remove(invalid);
  }

  document.getElementById("introSubmit").addEventListener("click", function () {
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
  introFrm.addEventListener("reset", introReset);

  const txtFrm = document.getElementById("textUploader");
  const txtFile = document.getElementById("textFile");
  const txtTitl = document.getElementById("textTitle");
  const txtRels = document.getElementById("textReleasers");
  const txtYear = document.getElementById("textYear");
  const txtMonth = document.getElementById("textMonth");

  function txtReset() {
    txtFile.classList.remove(invalid);
    txtTitl.classList.remove(invalid);
    txtRels.classList.remove(invalid);
    txtYear.classList.remove(invalid);
    txtMonth.classList.remove(invalid);
  }

  document.getElementById("textSubmit").addEventListener("click", function () {
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
  txtFrm.addEventListener("reset", txtReset);

  const imgFrm = document.getElementById("imageUploader");
  const imgFile = document.getElementById("imageFile");
  const imgTitl = document.getElementById("imageTitle");
  const imgRels = document.getElementById("imageReleasers");
  const imgYear = document.getElementById("imageYear");
  const imgMonth = document.getElementById("imageMonth");

  function imgReset() {
    imgFile.classList.remove(invalid);
    imgTitl.classList.remove(invalid);
    imgRels.classList.remove(invalid);
    imgYear.classList.remove(invalid);
    imgMonth.classList.remove(invalid);
  }

  document.getElementById("imageSubmit").addEventListener("click", function () {
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
  imgFrm.addEventListener("reset", imgReset);

  const magFrm = document.getElementById("magUploader");
  const magFile = document.getElementById("magFile");
  const magTitl = document.getElementById("magTitle");
  const magIssu = document.getElementById("magIssue");
  const magYear = document.getElementById("magYear");
  const magMonth = document.getElementById("magMonth");
  const magDay = document.getElementById("magDay");

  function magReset() {
    magFile.classList.remove(invalid);
    magTitl.classList.remove(invalid);
    magIssu.classList.remove(invalid);
    magYear.classList.remove(invalid);
    magMonth.classList.remove(invalid);
    magDay.classList.remove(invalid);
  }

  document.getElementById("magSubmit").addEventListener("click", function () {
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
    magFrm.addEventListener("reset", magReset);
  });
})();
