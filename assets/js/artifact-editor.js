/**
 * @file artifact-editor.js
 * This script is the entry point for the artifact editor page.
 */
import {
  date as validateDate,
  releaser as validateReleaser,
  repository as validateGitHub,
  color16 as validate16color,
} from "./artifact-validate.mjs";

(() => {
  "use strict";

  const osLabel = document.getElementById("artifact-editor-os-label");
  if (osLabel === null) {
    throw new Error("The operating system label is missing.");
  }
  const osInput = document.getElementById("artifact-editor-operating-system");
  if (osInput === null) {
    throw new Error("The operating system input is missing.");
  }
  osInput.addEventListener("input", relabelOS);
  const catInput = document.getElementById("artifact-editor-category");
  if (catInput === null) {
    throw new Error("The category input is missing.");
  }
  catInput.addEventListener("input", relabelCat);
  relabelOS();
  relabelCat();
  /**
   * Relabels the operating system label based on the selected option in the dropdown.
   */
  function relabelOS() {
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
  /**
   * Relabels the category input based on the selected index.
   */
  function relabelCat() {
    const index = catInput.selectedIndex;
    if (index == 0) {
      catInput.classList.remove("is-valid");
      catInput.classList.add("is-invalid");
    }
  }

  const classifications = document.getElementsByName("reset-classifications");
  if (classifications.length === 0) {
    throw new Error("The reset classifications are missing.");
  }
  for (let i = 0; i < classifications.length; i++) {
    undoClassification(i);
  }
  /**
   * Undo the classification for a given element.
   *
   * @param {number} i - The index of the element in the classifications array.
   */
  function undoClassification(i) {
    const elm = classifications[i];
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
      relabelOS();
      relabelCat();
    });
  }

  const name = "artifact-editor-filename";
  const nameInput = document.getElementById(name);
  if (nameInput === null) {
    throw new Error("The filename input is missing.");
  }
  nameInput.addEventListener("input", (e) => {
    e.target.classList.remove("is-valid");
    e.target.classList.remove("is-invalid");
    if (e.target.value.trim().length === 0) {
      e.target.classList.add("is-invalid");
    }
  });
  const nameReset = document.getElementById(name + "-reset");
  if (nameReset === null) {
    throw new Error("The filename reset is missing.");
  }
  const nameResetter = document.getElementsByName(name + "-resetter");
  if (nameResetter.length === 0) {
    throw new Error("The filename resetter is missing.");
  }
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

  const rel = "artifact-editor-releaser";
  const rel1Input = document.getElementById(rel + "-1");
  if (rel1Input === null) {
    throw new Error("The releaser 1 input is missing.");
  }
  rel1Input.addEventListener("input", (e) => validateReleaser(e.target));
  const rel2Input = document.getElementById(rel + "-2");
  if (rel2Input === null) {
    throw new Error("The releaser 2 input is missing.");
  }
  rel2Input.addEventListener("input", (e) => validateReleaser(e.target));
  const relsReset = document.getElementById(rel + "-reset");
  if (relsReset === null) {
    throw new Error("The releasers reset is missing.");
  }
  relsReset.addEventListener("click", undoRels);
  function undoRels() {
    const revert1 = rel1Input.getAttribute("data-reset-rel1");
    if (revert1 === null) {
      throw new Error("data-reset-rel1 attribute is required for rel1Input.");
    }
    rel1Input.value = revert1;
    validateReleaser(rel1Input);
    const revert2 = rel2Input.getAttribute("data-reset-rel2");
    if (revert2 === null) {
      throw new Error("data-reset-rel2 attribute is required for rel2Input.");
    }
    rel2Input.value = revert2;
    validateReleaser(rel2Input);
  }

  const title = "artifact-editor-title";
  const titleInput = document.getElementById(title);
  if (titleInput === null) {
    throw new Error("The title input is missing.");
  }
  titleInput.addEventListener("input", (e) => {
    e.target.classList.remove("is-valid");
  });
  const titleReset = document.getElementById(title + "-reset");
  if (titleReset === null) {
    throw new Error("The title reset is missing.");
  }
  const titleResetter = document.getElementsByName(title + "-resetter");
  if (titleResetter.length === 0) {
    throw new Error("The title resetter is missing.");
  }
  titleReset.addEventListener("click", () => {
    titleInput.classList.remove("is-valid");
    if (titleResetter.length === 0) {
      throw new Error("The title resetter is missing.");
    }
    titleInput.value = titleResetter[0].value;
    titleInput.classList.add("is-valid");
  });

  const credit = "artifact-editor-credit";
  const creTextInput = document.getElementById(credit + "-text");
  if (creTextInput === null) {
    throw new Error("The creator text input is missing.");
  }
  creTextInput.addEventListener("input", (e) => {
    e.target.classList.remove("is-valid");
  });
  const creIllInput = document.getElementById(credit + "-ill");
  if (creIllInput === null) {
    throw new Error("The creator illustrator input is missing.");
  }
  creIllInput.addEventListener("input", (e) => {
    e.target.classList.remove("is-valid");
  });
  const creProgInput = document.getElementById(credit + "-prog");
  if (creProgInput === null) {
    throw new Error("The creator programmer input is missing.");
  }
  creProgInput.addEventListener("input", (e) => {
    e.target.classList.remove("is-valid");
  });
  const creAudioInput = document.getElementById(credit + "-audio");
  if (creAudioInput === null) {
    throw new Error("The creator audio input is missing.");
  }
  creAudioInput.addEventListener("input", (e) => {
    e.target.classList.remove("is-valid");
  });
  const creResetter = document.getElementById(credit + "-resetter");
  if (creResetter === null) {
    throw new Error("The creator resetter is missing.");
  }
  const creReset = document.getElementById(credit + "-reset");
  if (creReset === null) {
    throw new Error("The creator reset is missing.");
  }
  creReset.addEventListener("click", () => {
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
  if (cmmtInput === null) {
    throw new Error("The comment input is missing.");
  }
  cmmtInput.addEventListener("input", (e) => {
    e.target.classList.remove("is-valid");
  });
  const cmmtReset = document.getElementById("artifact-editor-comment-reset");
  if (cmmtReset === null) {
    throw new Error("The comment reset is missing.");
  }
  const cmmtResetter = document.getElementById(
    "artifact-editor-comment-resetter"
  );
  if (cmmtResetter === null) {
    throw new Error("The comment resetter is missing.");
  }
  cmmtReset.addEventListener("click", () => {
    cmmtInput.classList.remove("is-valid");
    cmmtInput.value = cmmtResetter.value;
  });

  const virustotalInput = document.getElementById("artifact-editor-virustotal");
  if (virustotalInput === null) {
    throw new Error("The virustotal input is missing.");
  }
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
  if (yearInput === null) {
    throw new Error("The year input is missing.");
  }
  yearInput.addEventListener("input", () => {
    validateDate(yearInput, monthInput, dayInput);
  });
  const monthInput = document.getElementById("artifact-editor-month");
  if (monthInput === null) {
    throw new Error("The month input is missing.");
  }
  monthInput.addEventListener("input", () => {
    validateDate(yearInput, monthInput, dayInput);
  });
  const dayInput = document.getElementById("artifact-editor-day");
  if (dayInput === null) {
    throw new Error("The day input is missing.");
  }
  dayInput.addEventListener("input", () => {
    validateDate(yearInput, monthInput, dayInput);
  });

  const dateReset = document.getElementById("artifact-editor-date-reset");
  if (dateReset === null) {
    throw new Error("The date reset is missing.");
  }
  const dateResetter = document.getElementById("artifact-editor-date-resetter");
  if (dateResetter === null) {
    throw new Error("The date resetter is missing.");
  }
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

  const dateLastMod = document.getElementById("artifact-editor-date-lastmod");
  if (dateLastMod === null) {
    throw new Error("The date last mod input is missing.");
  }
  const dateLastModder = document.getElementById(
    "artifact-editor-date-lastmodder"
  );
  if (dateLastModder === null) {
    throw new Error("The date last modder input is missing.");
  }
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

  const github = document.getElementById("artifact-editor-github");
  if (github === null) {
    throw new Error("The GitHub input is missing.");
  }
  github.addEventListener("input", (e) => validateGitHub(e.target));

  const colors16 = document.getElementById("artifact-editor-16colors");
  if (colors16 === null) {
    throw new Error("The 16colors input is missing.");
  }
  colors16.addEventListener("input", (e) => validate16color(e.target));

  const youtube = document.getElementById("artifact-editor-youtube");
  if (youtube === null) {
    throw new Error("The YouTube input is missing.");
  }
  youtube.addEventListener("input", (e) => {
    e.target.classList.remove("is-valid", "is-invalid");
    const value = e.target.value.trim();
    const required = 11;
    if (value.length > 0 && value.length != required) {
      e.target.classList.add("is-invalid");
    }
  });
})();
