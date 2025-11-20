/**
 * @file canvas-readme.js
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
    pre437 = getElmById(cp437),
    utf8 = "readmeUTF8",
    preUTF8 = getElmById(utf8)

  const cascadia = document.getElementById(`monoFont`);
  if (cascadia !== null) {
    cascadia.addEventListener("click", useCascadia);
  }
  function useCascadia() {
    pre437.classList.add(none);
    preLatin1.classList.add(none);
    preUTF8.classList.add("font-cascadia-mono", ...blackBG);
    preUTF8.classList.remove(none);
  }

  const topaz = document.getElementById(`topazFont`);
  if (topaz !== null) {
    topaz.addEventListener("click", useAmiga);
  }
  function useAmiga() {
    pre437.classList.add(none);
    preUTF8.classList.add(none);
    preLatin1.classList.add("font-amiga", ...blackBG);
    preLatin1.classList.remove(none);
  }

  const vga = document.getElementById(`vgaFont`);
  if (vga !== null) {
    vga.addEventListener("click", useIBM);
  }
  function useIBM() {
    preLatin1.classList.add(none);
    preUTF8.classList.add(none);
    pre437.classList.add("font-dos", ...blackBG);
    pre437.classList.remove(none);
  }

  const copier = getElmById(`artifact-copy-textbody`);
  if (typeof navigator.clipboard === `undefined`) {
    copier.classList.add(none);
  } else {
    copier.addEventListener(`click`, copyText);
  }
  function copyText() {
    if (topaz !== null && topaz.checked) clipText(latin);
    else if (vga !== null && vga.checked) clipText(cp437);
    else clipText(utf8);
  }
})();
