/**
 * @file readme.js
 * Provides functions for handling readme and NFO file rendering within a pre element.
 */
import { clipText, getElmById } from "./helper.mjs";
(() => {
  "use strict";
  const none = `d-none`;
  const blackBG = ["reader-invert", "border", "border-black", "rounded-1"];

  const latin = "readmeLatin1",
    preLatin1 = getElmById(latin),
    cp437 = "readmeCP437",
    pre437 = getElmById(cp437);

  const defaultFont = document.querySelector(
    `input[name="readme-base"]:checked`
  );
  const cascadiaMono = document.getElementById(`monoFont`);
  if (cascadiaMono !== null) {
    cascadiaMono.addEventListener("click", useMono);
  }
  function useMono() {
    switch (defaultFont.id) {
      case `vgaFont`:
        preLatin1.classList.add(none);
        pre437.classList.remove(none, "font-dos", ...blackBG);
        pre437.classList.add("font-cascadia-mono");
        break;
      case `topazFont`:
        pre437.classList.add(none);
        preLatin1.classList.remove(none, "font-amiga", ...blackBG);
        preLatin1.classList.add("font-cascadia-mono");
        break;
      default:
        throw new Error(`no default font found for the readme`);
    }
  }

  const topaz = document.getElementById(`topazFont`);
  if (topaz !== null) {
    topaz.addEventListener("click", useAmiga);
  }
  function useAmiga() {
    pre437.classList.add(none);
    preLatin1.classList.remove(none, "font-cascadia-mono");
    preLatin1.classList.add("font-amiga", ...blackBG);
  }

  const vga = document.getElementById(`vgaFont`);
  if (vga !== null) {
    vga.addEventListener("click", useIBM);
  }
  function useIBM() {
    preLatin1.classList.add(none);
    pre437.classList.remove(none, "font-cascadia-mono");
    pre437.classList.add("font-dos", ...blackBG);
  }

  const copier = getElmById(`artifact-copy-readme-body`);
  if (typeof navigator.clipboard === `undefined`) {
    copier.classList.add(none);
  } else {
    copier.addEventListener(`click`, copyText);
  }
  function copyText() {
    if (topaz !== null && topaz.checked) clipText(latin);
    else if (vga !== null && vga.checked) clipText(cp437);
    else clipText(cp437);
  }
})();
