/**
 * @file canvas-ansi.js
 * Provides functions for handling ansi and binary text rendering within a pre element.
 */
import { clipText, getElmById } from "./helper.mjs";
(() => {
  "use strict";
  const elementId = "precontentBinary",
    preElm = getElmById(elementId);

  const standard = document.getElementById(`ansiVGA`);
  if (standard !== null) {
    standard.addEventListener("mouseover", useDOS);
  }
  const hires = document.getElementById(`ansiVGA50`)
  if (hires !== null) {
    hires.addEventListener("mouseover", useANSI)
  }

  function useANSI() {
    preElm.classList.replace("font-dos", "font-ansi");
    preElm.classList.replace("reader", "reader-hires");
    preElm.classList.toggle("font-large");
    hires.classList.add("active");
    standard.classList.remove("active");
  }
  function useDOS() {
    preElm.classList.replace("font-ansi", "font-dos");
    preElm.classList.replace("reader-hires", "reader");
    preElm.classList.toggle("font-large");
    hires.classList.remove("active");
    standard.classList.add("active");
  }

  const copier = getElmById(`artifact-copy-textbody`);
  if (typeof navigator.clipboard === `undefined`) {
    copier.classList.add(none);
  } else {
    copier.addEventListener(`click`, copyText);
  }
  function copyText() {
    clipText(elementId);
  }
})();

