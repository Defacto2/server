/**
 * @file readme.js
 * Provides functions for handling readme and NFO file rendering within a pre element.
 */
import { clipText, getElmById } from "./helper.mjs";
(() => {
  "use strict";
  const none = `d-none`;
  const wrap = "pre-wrap";
  const blackBG = ["reader-invert", "border", "border-black", "rounded-1"];

  const latin = "readmeLatin1",
    preLatin1 = getElmById(latin),
    cp437 = "readmeCP437",
    pre437 = getElmById(cp437);

  const openSans = document.getElementById(`openSansFont`);
  if (openSans !== null) {
    openSans.addEventListener("click", useBrowser);
  }
  function useBrowser() {
    preLatin1.classList.remove(none, "font-amiga", ...blackBG);
    preLatin1.classList.add("font-opensans", wrap);
    pre437.classList.add(none);
  }

  const topaz = document.getElementById(`topazFont`);
  if (topaz !== null) {
    topaz.addEventListener("click", useAmiga);
  }
  function useAmiga() {
    preLatin1.classList.remove(none, "font-opensans", wrap);
    preLatin1.classList.add("font-amiga", ...blackBG);
    pre437.classList.add(none);
  }

  const vga = document.getElementById(`vgaFont`);
  if (vga !== null) {
    vga.addEventListener("click", useIBM);
  }
  function useIBM() {
    preLatin1.classList.add(none);
    pre437.classList.remove(none);
    pre437.classList.add(...blackBG);
  }

  const copier = getElmById(`copyReadme`);
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
