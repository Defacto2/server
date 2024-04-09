// readme.js

import { clipText, getElmById } from "./helper.mjs";

(() => {
  "use strict";
  const hide = `d-none`;
  const blackBG = ["reader-invert", "border", "border-black", "rounded-1"];
  const latin = "readmeLatin1",
    cp437 = "readmeCP437";
  const preLatin1 = getElmById(latin),
    pre437 = getElmById(cp437);

  /**
   * Event listener for the Open Sans font radio button.
   * @function
   */
  const openSans = document.getElementById(`openSansFont`);
  if (openSans !== null) {
    openSans.addEventListener("click", () => {
      preLatin1.classList.remove(hide, "font-amiga", ...blackBG);
      preLatin1.classList.add("font-opensans");
      pre437.classList.add(hide);
    });
  }

  /**
   * Event listener for the Topaz font radio button.
   * @function
   */
  const topaz = document.getElementById(`topazFont`);
  if (topaz !== null) {
    topaz.addEventListener("click", () => {
      preLatin1.classList.remove(hide, "font-opensans");
      preLatin1.classList.add("font-amiga", ...blackBG);
      pre437.classList.add(hide);
    });
  }

  /**
   * Event listener for the VGA font radio button.
   * @function
   */
  const vga = document.getElementById(`vgaFont`);
  if (vga !== null) {
    vga.addEventListener("click", () => {
      preLatin1.classList.add(hide);
      pre437.classList.remove(hide);
    });
  }

  const copier = getElmById(`copyReadme`);
  if (typeof navigator.clipboard === `undefined`) {
    copier.classList.add(hide);
  } else {
    copier.addEventListener(`click`, () => {
      if (topaz !== null && topaz.checked) clipText(latin);
      else if (vga !== null && vga.checked) clipText(cp437);
      else clipText(cp437);
    });
  }
})();
