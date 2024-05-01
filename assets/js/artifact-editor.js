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

  updateLabelOS();

  for (let i = 0; i < resets.length; i++) {
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
  osInput.addEventListener("input", updateLabelOS);

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
