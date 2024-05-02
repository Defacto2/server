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
  const rel1Input = document.getElementById("artifact-editor-releaser-1");
  const rel2Input = document.getElementById("artifact-editor-releaser-2");
  const relsReset = document.getElementById("artifact-editor-releaser-reset");
  const titleInput = document.getElementById("artifact-editor-title");
  const titleReset = document.getElementById("artifact-editor-title-reset");
  const titleResetter = document.getElementsByName(
    "artifact-editor-title-resetter"
  );

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
      if (elm.id === "artifact-editor-releaser-1") {
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
