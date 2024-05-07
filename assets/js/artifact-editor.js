/**
 * artifact-editor.js
 * This script is the entry point for the artifact editor page.
 */
(() => {
  "use strict";

  const resets = document.getElementsByName("reset-classifications");
  const osLabel = document.getElementById("artifact-editor-os-label");
  const osInput = document.getElementById("artifact-editor-operating-system");
  const catInput = document.getElementById("artifact-editor-category");
  const name = "artifact-editor-filename";
  const nameInput = document.getElementById(name);
  const nameReset = document.getElementById(name + "-reset");
  const nameResetter = document.getElementsByName(name + "-resetter");
  const rel = "artifact-editor-releaser";
  const rel1Input = document.getElementById(rel + "-1");
  const rel2Input = document.getElementById(rel + "-2");
  const relsReset = document.getElementById(rel + "-reset");
  const title = "artifact-editor-title";
  const titleInput = document.getElementById(title);
  const titleReset = document.getElementById(title + "-reset");
  const titleResetter = document.getElementsByName(title + "-resetter");

  const creTextInput = document.getElementById("artifact-editor-credit-text");
  creTextInput.addEventListener("input", (e) => {
    e.target.classList.remove("is-valid");
  });
  const creIllInput = document.getElementById("artifact-editor-credit-ill");
  creIllInput.addEventListener("input", (e) => {
    e.target.classList.remove("is-valid");
  });
  const creProgInput = document.getElementById("artifact-editor-credit-prog");
  creProgInput.addEventListener("input", (e) => {
    e.target.classList.remove("is-valid");
  });
  const creAudioInput = document.getElementById("artifact-editor-credit-audio");
  creAudioInput.addEventListener("input", (e) => {
    e.target.classList.remove("is-valid");
  });
  const creResetter = document.getElementById(
    "artifact-editor-credit-resetter"
  );
  const creReset = document.getElementById("artifact-editor-credit-reset");
  creReset.addEventListener("click", () => {
    console.log("resetting credits");
    if (creResetter.length === 0) {
      throw new Error("The creator resetter is missing.");
    }
    const creators = creResetter.value.split(";");
    if (creators.length != 4) {
      throw new Error("The creator resetter values are invalid.");
    }
    const text = creators[0];
    const ill = creators[1];
    const prog = creators[2];
    const audio = creators[3];
    creTextInput.value = text;
    creIllInput.value = ill;
    creProgInput.value = prog;
    creAudioInput.value = audio;
  });

  const cmmtInput = document.getElementById("artifact-editor-comment");
  cmmtInput.addEventListener("input", (e) => {
    e.target.classList.remove("is-valid");
  });
  const cmmtReset = document.getElementById("artifact-editor-comment-reset");
  const cmmtResetter = document.getElementById(
    "artifact-editor-comment-resetter"
  );
  cmmtReset.addEventListener("click", () => {
    cmmtInput.classList.remove("is-valid");
    cmmtInput.value = cmmtResetter.value;
  });

  const virustotalInput = document.getElementById("artifact-editor-virustotal");
  virustotalInput.addEventListener("input", (e) => {
    e.target.classList.remove("is-valid", "is-invalid");
    const value = e.target.value.trim();
    if (value.length != 0) {
      if (!value.startsWith("https://www.virustotal.com/")) {
        e.target.classList.add("is-invalid");
      }
    }
  });

  const yearInput = document.getElementById("artifact-editor-year");
  const monthInput = document.getElementById("artifact-editor-month");
  const dayInput = document.getElementById("artifact-editor-day");
  yearInput.addEventListener("input", validateDate);
  monthInput.addEventListener("input", validateDate);
  dayInput.addEventListener("input", validateDate);
  const dateReset = document.getElementById("artifact-editor-date-reset");
  const dateResetter = document.getElementById("artifact-editor-date-resetter");
  const dateLastMod = document.getElementById("artifact-editor-date-lastmod");
  const dateLastModder = document.getElementById(
    "artifact-editor-date-lastmodder"
  );
  dateReset.addEventListener("click", () => {
    yearInput.classList.remove("is-invalid", "is-valid");
    monthInput.classList.remove("is-invalid", "is-valid");
    dayInput.classList.remove("is-invalid", "is-valid");
    const value = dateResetter.value;
    const values = value.split("-");
    if (values.length != 3) {
      throw new Error("The date resetter values are invalid.");
    }
    yearInput.value = values[0];
    monthInput.value = values[1];
    dayInput.value = values[2];
  });

  dateLastMod.addEventListener("click", () => {
    yearInput.classList.remove("is-invalid", "is-valid");
    monthInput.classList.remove("is-invalid", "is-valid");
    dayInput.classList.remove("is-invalid", "is-valid");
    const value = dateLastModder.value;
    const values = value.split("-");
    if (values.length != 3) {
      throw new Error("The date last modder values are invalid.");
    }
    yearInput.value = values[0];
    monthInput.value = values[1];
    dayInput.value = values[2];
  });

  function validateDate() {
    yearInput.classList.remove("is-invalid", "is-valid");
    monthInput.classList.remove("is-invalid", "is-valid");
    dayInput.classList.remove("is-invalid", "is-valid");

    const year = parseInt(yearInput.value, 10);
    if (isNaN(year)) {
      yearInput.value = "0";
    } else {
      yearInput.value = year; // remove leading zeros
    }
    const month = parseInt(monthInput.value, 10);
    if (isNaN(month)) {
      monthInput.value = "0";
    } else {
      monthInput.value = month;
    }
    const day = parseInt(dayInput.value, 10);
    if (isNaN(day)) {
      dayInput.value = "0";
    } else {
      dayInput.value = day;
    }

    const none = 0;
    const currentYear = new Date().getFullYear();
    const validYear = year >= 1980 && year <= currentYear;
    // use greater than instead of != none to avoid a isNaN condition
    if (year > none && !validYear) {
      yearInput.classList.add("is-invalid");
    }
    const validMonth = month >= 1 && month <= 12;
    if (month > none && !validMonth) {
      monthInput.classList.add("is-invalid");
    }
    const validDay = day >= 1 && day <= 31;
    if (day > none && !validDay) {
      dayInput.classList.add("is-invalid");
    }
    if (isNaN(year) && (validMonth || validDay)) {
      yearInput.classList.add("is-invalid");
    }
    if ((month == none || isNaN(month)) && validDay) {
      monthInput.classList.add("is-invalid");
    }
  }

  updateLabelOS();
  updateLabelCat();
  for (let i = 0; i < resets.length; i++) {
    resetClassifications(i);
  }
  osInput.addEventListener("input", updateLabelOS);
  catInput.addEventListener("input", updateLabelCat);
  rel1Input.addEventListener("input", (e) => validateReleaser(e.target));
  rel2Input.addEventListener("input", (e) => validateReleaser(e.target));
  relsReset.addEventListener("click", resetRleasers);
  titleInput.addEventListener("input", (e) => {
    e.target.classList.remove("is-valid");
  });
  titleReset.addEventListener("click", () => {
    titleInput.classList.remove("is-valid");
    if (titleResetter.length === 0) {
      throw new Error("The title resetter is missing.");
    }
    titleInput.value = titleResetter[0].value;
    titleInput.classList.add("is-valid");
  });
  nameInput.addEventListener("input", (e) => {
    e.target.classList.remove("is-valid");
    e.target.classList.remove("is-invalid");
    if (e.target.value.trim().length === 0) {
      e.target.classList.add("is-invalid");
    }
  });
  nameReset.addEventListener("click", () => {
    nameInput.classList.remove("is-valid");
    if (nameResetter.length === 0) {
      throw new Error("The filename resetter is missing.");
    }
    nameInput.value = nameResetter[0].value;
    nameInput.classList.add("is-valid");
    nameInput.classList.remove("is-invalid");
    if (nameInput.value.trim().length === 0) {
      nameInput.classList.add("is-invalid");
    }
  });

  function resetClassifications(i) {
    const elm = resets[i];
    const os = elm.getAttribute("data-reset-os");
    if (os === null) {
      throw new Error("data-reset-os attribute is required for ${elm.id}.");
    }
    const cat = elm.getAttribute("data-reset-cat");
    if (cat === null) {
      throw new Error("data-reset-cat attribute is required for ${elm.id}.");
    }
    elm.addEventListener("click", (e) => {
      e.preventDefault();
      osInput.value = os;
      osInput.classList.remove("is-invalid");
      catInput.value = cat;
      catInput.classList.remove("is-invalid");
      updateLabelOS();
      updateLabelCat();
    });
  }

  // todo, move to a mjs file.

  function resetRleasers() {
    const revert1 = rel1Input.getAttribute("data-reset-rel1");
    const revert2 = rel2Input.getAttribute("data-reset-rel2");
    if (revert1 === null) {
      throw new Error("data-reset-rel1 attribute is required for rel1Input.");
    }
    if (revert2 === null) {
      throw new Error("data-reset-rel2 attribute is required for rel2Input.");
    }
    rel1Input.value = revert1;
    validateReleaser(rel1Input);
    rel2Input.value = revert2;
    validateReleaser(rel2Input);
  }

  function validateReleaser(elm) {
    if (elm == null) {
      throw new Error("The element of the releaser validator is null.");
    }
    elm.classList.remove("is-valid", "is-invalid");

    let value = elm.value.trim().toUpperCase();
    value = value.replace(/[^A-ZÀ-ÖØ-Þ0-9\-,& ]/g, "");
    elm.value = value;

    const min = elm.getAttribute("minlength");
    const max = elm.getAttribute("maxlength");
    const req = elm.getAttribute("required");
    if (min === null) {
      throw new Error(`The minlength attribute is required for ${elm.id}.`);
    }
    if (max === null) {
      throw new Error(`The maxlength attribute is required for ${elm.id}.`);
    }

    const error = document.getElementById("artifact-editor-releasers-error");
    if (error === null) {
      throw new Error("The releasers error element is null.");
    }

    const requireBounds = value.length < min || value.length > max;
    if (req != null && requireBounds) {
      elm.classList.add("is-invalid");
      if (elm.id === "-1") {
        error.classList.add("d-block");
      }
      return;
    }
    const emptyBounds =
      value.length > 0 && (value.length < min || value.length > max);
    if (req == null && emptyBounds) {
      elm.classList.add("is-invalid");
      return;
    }
    elm.classList.remove("is-invalid");
    error.classList.remove("d-block");
  }

  function updateLabelOS() {
    const index = osInput.selectedIndex;
    const sel = osInput.options[index];
    let group = sel.parentNode.label;

    if (typeof group == "undefined" || group == "") {
      osInput.classList.add("is-invalid");
      osInput.classList.remove("is-valid");
      group = `Operating system`;
    }
    osLabel.textContent = `${group}`;
  }
  function updateLabelCat() {
    const index = catInput.selectedIndex;
    if (index == 0) {
      catInput.classList.remove("is-valid");
      catInput.classList.add("is-invalid");
    }
  }
})();
