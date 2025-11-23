/**
 * @file content-text.js
 * Provides functions for handling readme and NFO file rendering within a pre element.
 */
import { clipText, getElmById } from "./helper.mjs";
(() => {
  "use strict";
  const none = `d-none`;
  const blackBG = ["reader-invert", "border", "border-black", "rounded-1"];

  const latinId = "precontentLatin1",
    preLatin1 = getElmById(latinId),
    cp437Id = "precontentCP437",
    preCP437 = getElmById(cp437Id),
    utf8Id = "precontentUTF8",
    preUTF8 = getElmById(utf8Id)

  const cascadia = document.getElementById(`textcontBtnWeb`);
  if (cascadia !== null) {
    cascadia.addEventListener("click", useCascadia);
  }
  function useCascadia() {
    preCP437.classList.add(none);
    preLatin1.classList.add(none);
    preUTF8.classList.add("font-cascadia-mono", ...blackBG);
    preUTF8.classList.remove(none);
  }

  const topaz = document.getElementById(`textcontBtnAmiga`);
  if (topaz !== null) {
    topaz.addEventListener("click", useAmiga);
  }
  function useAmiga() {
    preCP437.classList.add(none);
    preUTF8.classList.add(none);
    preLatin1.classList.add("font-amiga", ...blackBG);
    preLatin1.classList.remove(none);
  }

  const vga = document.getElementById(`textcontBtnDOS`);
  if (vga !== null) {
    vga.addEventListener("click", useIBM);
  }
  function useIBM() {
    preLatin1.classList.add(none);
    preUTF8.classList.add(none);
    preCP437.classList.add("font-dos", ...blackBG);
    preCP437.classList.remove(none);
  }

  const copier = getElmById(`artifact-copy-textbody`);
  if (typeof navigator.clipboard === `undefined`) {
    copier.classList.add(none);
  } else {
    copier.addEventListener(`click`, copyText);
  }
  function copyText() {
    if (topaz !== null && topaz.checked) clipText(latinId);
    else if (vga !== null && vga.checked) clipText(cp437Id);
    else clipText(utf8Id);
  }
})();
