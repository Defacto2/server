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

  updateLabelOS();
  for (let i = 0; i < resets.length; i++) {
    resetClassifications(i);
  }
  osInput.addEventListener("input", updateLabelOS);
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
      catInput.value = cat;
      updateLabelOS();
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
      group = `Operating system`;
    }
    osLabel.textContent = `${group}`;
  }
})();
