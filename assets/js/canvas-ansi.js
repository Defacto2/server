/**
 * @file canvas-ansi.js
 * Provides functions for handling ansi and binary text rendering within a pre element.
 */
import { clipText, getElmById } from "./helper.mjs";
//import { getElmById } from "./helper.mjs";
(() => {
  "use strict";
  //const none = `d-none`;
  //const blackBG = ["reader-invert", "border", "border-black", "rounded-1"];

  const ansiMe = "ansiMe",
    preANSI = getElmById(ansiMe);

  const vgaStd = document.getElementById(`ansiVGA`);
  if (vgaStd !== null) {
    vgaStd.addEventListener("mouseover", useDOS);
  }
  const vga50 = document.getElementById(`ansiVGA50`)
  if (vga50 !== null) {
    vga50.addEventListener("mouseover", useANSI)
  }

  function useANSI() {
    preANSI.classList.replace("font-dos", "font-ansi");
    preANSI.classList.replace("reader", "reader-hires");
    vga50.classList.add("active");
    vgaStd.classList.remove("active");
  }
  function useDOS() {
    preANSI.classList.replace("font-ansi", "font-dos");
    preANSI.classList.replace("reader-hires", "reader");
    vga50.classList.remove("active");
    vgaStd.classList.add("active");
  }

  const copier = getElmById(`artifact-copy-textbody`);
  if (typeof navigator.clipboard === `undefined`) {
    copier.classList.add(none);
  } else {
    copier.addEventListener(`click`, copyText);
  }
  function copyText() {
    clipText(ansiMe);
  }
})();

