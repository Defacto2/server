/**
 * This module handles the uploader functionality for the website.
 * It contains functions to validate client input and show/hide modals.
 * @module uploader
 * @requires bootstrap
 */
(() => {
  "use strict";

  const invalid = "is-invalid";

  // Poeut modal elements
  const pouetM = document.getElementById("uploader-pouet");
  if (pouetM == null) {
    throw new Error("The uploader-pouet element is null.");
  }
  pouetM.addEventListener("shown.bs.modal", function () {
    poeutInput.focus();
  });
  const pouetModal = new bootstrap.Modal(pouetM);
  const poeutInput = document.getElementById("pouet-submission");
  if (poeutInput == null) {
    throw new Error("The pouet-submission element is null.");
  }

  // Demozoo modal elements
  const demozooM = document.getElementById("uploader-demozoo");
  if (demozooM == null) {
    throw new Error("The uploader-demozoo element is null.");
  }
  demozooM.addEventListener("shown.bs.modal", function () {
    demozooInput.focus();
  });
  const demozooModal = new bootstrap.Modal(demozooM);
  const demozooInput = document.getElementById("demozoo-submission");
  if (demozooInput == null) {
    throw new Error("The demozoo-submission element is null.");
  }

  const introM = document.getElementById("uploader-intro");
  if (introM == null) {
    throw new Error("The uploader-intro element is null.");
  }
  introM.addEventListener("shown.bs.modal", function () {
    introInput.focus();
  });
  const introModal = new bootstrap.Modal(introM);
  const introInput = document.getElementById("uploader-intro-file");
  if (introInput == null) {
    throw new Error("The uploader-intro-file element is null.");
  }

  const txtM = document.getElementById("uploaderText");
  const imgM = document.getElementById("uploaderImg");
  const magM = document.getElementById("uploaderMag");
  const advM = document.getElementById("uploaderAdv");
  const glossM = document.getElementById("termsModal"); // not part of uploader but still a modal

  // Modal objects
  const txtModal = new bootstrap.Modal(txtM);
  const imgModal = new bootstrap.Modal(imgM);
  const magModal = new bootstrap.Modal(magM);
  const advModal = new bootstrap.Modal(advM);
  const glossModal = new bootstrap.Modal(glossM);

  // Pagination button elements
  const pageS = document.getElementById("paginationStart");
  const pageP = document.getElementById("paginationPrev");
  const pageP2 = document.getElementById("paginationPrev2");
  const pageN = document.getElementById("paginationNext");
  const pageN2 = document.getElementById("paginationNext2");
  const pageE = document.getElementById("paginationEnd");

  const pageRange = document.getElementById("paginationRange");
  if (typeof pageRange !== "undefined" && pageRange != null) {
    pageRange.addEventListener("change", function () {
      const range = pageRange.value;
      const url = new URL(window.location.href);
      const path = url.pathname;
      const paths = path.split("/");
      const page = paths[paths.length - 1];
      if (!isNaN(page) && typeof Number(page) === "number") {
        paths[paths.length - 1] = range;
      } else {
        paths.push(range);
      }
      url.pathname = paths.join("/");
      window.location.href = url.href;
    });
    const pageRangeLabel = document.getElementById("paginationRangeLabel");
    pageRange.addEventListener("input", function () {
      pageRangeLabel.textContent = "Jump to page " + pageRange.value;
    });
  }

  // Keyboard shortcuts event listener
  document.addEventListener("keydown", function (event) {
    const demozoo = "d",
      pouet = "p",
      intro = "i",
      nfo = "n",
      graphic = "g",
      magazine = "m",
      advanced = "a",
      glossaryOfTerms = "t";
    if (event.ctrlKey && event.altKey) {
      switch (event.key) {
        case demozoo:
          demozooModal.show();
          break;
        case pouet:
          pouetModal.show();
          break;
        case intro:
          introModal.show();
          break;
        case nfo:
          txtModal.show();
          break;
        case graphic:
          imgModal.show();
          break;
        case magazine:
          magModal.show();
          break;
        case advanced:
          advModal.show();
          break;
        case glossaryOfTerms:
          glossModal.show();
          break;
      }
    }

    const right = "ArrowRight",
      left = "ArrowLeft";
    // Ctrl + Left arrow key to go to the start page
    if (event.ctrlKey && event.key == left) {
      if (pageS != null) pageS.click();
      return;
    }
    // Ctrl + Right arrow key to go to the end page
    if (event.ctrlKey && event.key == right) {
      if (pageE != null) pageE.click();
      return;
    }
    // Shift + Left arrow key to go to the start page
    if (event.shiftKey && event.key == left) {
      if (pageP2 != null) pageP2.click();
      return;
    }
    // Shift + Right arrow key to go to the end page
    if (event.shiftKey && event.key == right) {
      if (pageN2 != null) pageN2.click();
      return;
    }
    // Left arrow key to go to the previous page
    if (event.key == left) {
      if (pageP != null) pageP.click();
      return;
    }
    // Right arrow key to go to the next page
    if (event.key == right) {
      if (pageN != null) pageN.click();
      return;
    }
  });

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

  // Uploader forms
  const introFrm = document.getElementById("introUploader");
  const txtFrm = document.getElementById("textUploader");
  const imgFrm = document.getElementById("imageUploader");
  const magFrm = document.getElementById("magUploader");
  const advFrm = document.getElementById("advancedUploader");

  // Elements for the intro uploader
  const introFile = document.getElementById("introFile");
  const introTitl = document.getElementById("releaseTitle");
  const introRels = document.getElementById("introReleasers");
  const introYear = document.getElementById("introYear");
  const introMonth = document.getElementById("introMonth");

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

  // Event listener for the intro submit button
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
  // Event listener for the intro reset button
  introFrm.addEventListener("reset", introReset);

  // Elements for the text uploader
  const txtFile = document.getElementById("textFile");
  const txtTitl = document.getElementById("textTitle");
  const txtRels = document.getElementById("textReleasers");
  const txtYear = document.getElementById("textYear");
  const txtMonth = document.getElementById("textMonth");

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

  // Event listener for the text submit button
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
  // Event listener for the text reset button
  txtFrm.addEventListener("reset", txtReset);

  // Elements for the image uploader
  const imgFile = document.getElementById("imageFile");
  const imgTitl = document.getElementById("imageTitle");
  const imgRels = document.getElementById("imageReleasers");
  const imgYear = document.getElementById("imageYear");
  const imgMonth = document.getElementById("imageMonth");

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

  // Event listener for the image submit button
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
  // Event listener for the image reset button
  imgFrm.addEventListener("reset", imgReset);

  // Elements for the magazine uploader
  const magFile = document.getElementById("magFile");
  const magTitl = document.getElementById("magTitle");
  const magIssu = document.getElementById("magIssue");
  const magYear = document.getElementById("magYear");
  const magMonth = document.getElementById("magMonth");
  const magDay = document.getElementById("magDay");

  /**
   * Resets the form fields for a magazine upload.
   */
  function magReset() {
    magFile.classList.remove(invalid);
    magTitl.classList.remove(invalid);
    magIssu.classList.remove(invalid);
    magYear.classList.remove(invalid);
    magMonth.classList.remove(invalid);
    magDay.classList.remove(invalid);
  }

  // Event listener for the magazine submit button
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
  });
  // Event listener for the magazine reset button
  magFrm.addEventListener("reset", magReset);

  // Elements for the advanced uploader
  const advFile = document.getElementById("advFile");
  const advOS = document.getElementById("advSelOS");
  const advCat = document.getElementById("advSelCat");
  const advTitl = document.getElementById("advTitle");
  const advRels = document.getElementById("releasersAdv");
  const advYear = document.getElementById("advYear");
  const advMonth = document.getElementById("advMonth");
  const advDay = document.getElementById("advDay");

  /**
   * Resets the form by removing the "invalid" class from all form elements.
   */
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

  // Event listener for the advanced submit button
  document.getElementById("advSubmit").addEventListener("click", function () {
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
  // Event listener for the advanced reset button
  advFrm.addEventListener("reset", advReset);
})();
