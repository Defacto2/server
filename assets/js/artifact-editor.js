/**
 * artifact-editor.js
 * This script is the entry point for the artifact editor page.
 */
(() => {
  "use strict";
  const elms = document.getElementsByName("reset-classifications");
  const osInput = document.getElementById("artifact-editor-operating-system");
  const catInput = document.getElementById("artifact-editor-category");
  const osLabel = document.getElementById("artifact-editor-os-label");
  for (let i = 0; i < elms.length; i++) {
    const elm = elms[i];
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

      console.log(osInput.parentNode);
      // const selected = e.target.options[e.target.selectedIndex];
      // let group = selected.parentNode.label;
      // console.log(`Selected OS: ${os}, Group: ${group}`);
    });
  }
  //.addEventListener(
  //   "mouseup",
  //   handleMouseUp,
  //   passiveSupported ? { passive: true } : false,
  // );
  osInput.addEventListener("", (e) => {
    console.log(`honkee: ${e.type}`);
  });
  osInput.addEventListener("change", (e) => {
    console.log(`honkee: ${e.type}`);
    console.log(`HONK z`);
    e.preventDefault();
  });
  osInput.addEventListener(
    "input",
    (e) => {
      console.log(`honkee: ${e.type}`);
      console.log(`HONK y`);
      e.preventDefault();
      console.log(`Selected OS: ${e.target.value}`);
      const os = e.target.value;
      const selected = e.target.options[e.target.selectedIndex];
      let group = selected.parentNode.label;
      if (typeof group == "undefined" || group == "") {
        group = `Operating system`;
      }
      console.log(`Selected OS: ${os}, Group: ${group}`);
      osLabel.textContent = `${group}`;
    },
    { capture: true, once: false, passive: true }
  );
})();
