/**
 * @file content-binary.js
 * Provides functions for handling ansi and binary text rendering within a pre element.
 */
import { clipText, getElmById } from "./helper.mjs";
(() => {
  "use strict";
  const elementId = "precontentBinary",
    preElm = getElmById(elementId);

  const zoom = document.getElementById(`textcontBtnZoom`);
  if (zoom !== null) {
    zoom.addEventListener("mouseover", useZoom);
  }
  const standard = document.getElementById(`textcontBtn16px`);
  if (standard !== null) {
    standard.addEventListener("mouseover", useStandard);
  }
  const hires = document.getElementById(`textcontBtn08px`)
  if (hires !== null) {
    hires.addEventListener("mouseover", useHires)
  }

  function useZoom() {
    zoom.classList.add("active");
    hires.classList.remove("active");
    standard.classList.remove("active");
    use16px(true)
  }
  function useStandard() {
    standard.classList.add("active");
    zoom.classList.remove("active");
    hires.classList.remove("active");
    use16px(false)
  }
  function use16px(largefont) {
    preElm.classList.replace("font-ansi", "font-dos");
    preElm.classList.replace("reader-hires", "reader");
    if (largefont == true) {
      preElm.classList.add("font-large");
    } else {
      preElm.classList.remove("font-large");
    }
  }
  function useHires() {
    preElm.classList.replace("font-dos", "font-ansi");
    preElm.classList.replace("reader", "reader-hires");
    preElm.classList.remove("font-large");
    hires.classList.add("active");
    zoom.classList.remove("active");
    standard.classList.remove("active");
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

