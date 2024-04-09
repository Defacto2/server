import { getModalById, focusModalById } from "./uploader.mjs";
export default keyboardShortcuts;

export function keyboardShortcuts() {
  document.addEventListener("keydown", function (event) {
    if (event.ctrlKey && event.altKey) {
      switch (event.key) {
        case demozoo:
          demozooModal.show();
          break;
        case pouet:
          pouetModal.show();
          break;
        case intro:
          introModal.show();
          break;
        case nfo:
          txtModal.show();
          break;
        case graphic:
          imgModal.show();
          break;
        case magazine:
          magModal.show();
          break;
        case advanced:
          advModal.show();
          break;
        case glossaryOfTerms:
          glossModal.show();
          break;
      }
    }
    // Ctrl + Left arrow key to go to the start page
    if (event.ctrlKey && event.key == left) {
      if (pageS != null) pageS.click();
      return;
    }
    // Ctrl + Right arrow key to go to the end page
    if (event.ctrlKey && event.key == right) {
      if (pageE != null) pageE.click();
      return;
    }
    // Shift + Left arrow key to go to the start page
    if (event.shiftKey && event.key == left) {
      if (pageP2 != null) pageP2.click();
      return;
    }
    // Shift + Right arrow key to go to the end page
    if (event.shiftKey && event.key == right) {
      if (pageN2 != null) pageN2.click();
      return;
    }
    // Left arrow key to go to the previous page
    if (event.key == left) {
      if (pageP != null) pageP.click();
      return;
    }
    // Right arrow key to go to the next page
    if (event.key == right) {
      if (pageN != null) pageN.click();
      return;
    }
  });
}

const pouetModal = focusModalById("uploader-pouet", "pouet-submission");
const demozooModal = focusModalById("uploader-demozoo", "demozoo-submission");
const introModal = focusModalById("uploader-intro", "uploader-intro-file");
const txtModal = getModalById("uploaderText");
const imgModal = getModalById("uploaderImg");
const magModal = getModalById("uploaderMag");
const advModal = getModalById("uploaderAdv");
const glossModal = getModalById("termsModal"); // TODO: move to layout.js or main.js

const demozoo = "d",
  pouet = "p",
  intro = "i",
  nfo = "n",
  graphic = "g",
  magazine = "m",
  advanced = "a",
  glossaryOfTerms = "t";

const pageS = document.getElementById("paginationStart");
const pageP = document.getElementById("paginationPrev");
const pageP2 = document.getElementById("paginationPrev2");
const pageN = document.getElementById("paginationNext");
const pageN2 = document.getElementById("paginationNext2");
const pageE = document.getElementById("paginationEnd");
const right = "ArrowRight",
  left = "ArrowLeft";

// Keyboard shortcuts event listener
