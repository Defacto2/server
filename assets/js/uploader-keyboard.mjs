/**
 * @module uploader-keyboard
 * This module provides keyboard shortcuts for handling file uploads and jump to searches.
 */
import { focusModalById } from "./uploader.mjs";
export default keyboardShortcuts;

const pouetModal = focusModalById("uploader-pouet-modal", "pouet-submission");
const demozooModal = focusModalById(
  "uploader-demozoo-modal",
  "demozoo-submission"
);
const introModal = focusModalById(
  "uploader-intro-modal",
  "uploader-intro-file"
);
const trainerModal = focusModalById(
  "uploader-trainer-modal",
  "uploader-trainer-file"
);
const textModal = focusModalById("uploader-text-modal", "uploader-text-file");
const graphicModal = focusModalById(
  "uploader-image-modal",
  "uploader-image-file"
);
const magazineModal = focusModalById(
  "uploader-magazine-modal",
  "uploader-magazine-file"
);
const advancedModal = focusModalById(
  "uploader-advanced-modal",
  "uploader-advanced-file"
);

const bs = "Backspace",
  enter = "Enter",
  demozoo = "d",
  pouet = "p",
  intro = "i",
  trainer = "t",
  nfo = "n",
  graphic = "g",
  magazine = "m",
  advanced = "a",
  releasers = "r",
  files = "f",
  descriptions = "x";

/**
 * Binds keyboard shortcuts to specific actions.
 *
 * Key values for keyboard events:
 * https://developer.mozilla.org/en-US/docs/Web/API/UI_Events/Keyboard_event_key_values
 *
 */
export function keyboardShortcuts() {
  document.addEventListener("keydown", function (event) {
    if (!event.ctrlKey || !event.altKey) {
      return;
    }
    const gotoRecord = document.getElementById("go-to-the-new-artifact-record");
    const card1 = document.getElementById("artifact-card-link-1");
    const card2 = document.getElementById("artifact-card-link-2");
    const card3 = document.getElementById("artifact-card-link-3");
    const card4 = document.getElementById("artifact-card-link-4");
    const card5 = document.getElementById("artifact-card-link-5");
    const card6 = document.getElementById("artifact-card-link-6");
    const card7 = document.getElementById("artifact-card-link-7");
    const card8 = document.getElementById("artifact-card-link-8");
    const card9 = document.getElementById("artifact-card-link-9");
    switch (event.key) {
      case "1":
        event.preventDefault();
        if (card1) {
          openNewTab(card1.href);
        }
        break;
      case "2":
        event.preventDefault();
        if (card2) {
          openNewTab(card2.href);
        }
        break;
      case "3":
        event.preventDefault();
        if (card3) {
          openNewTab(card3.href);
        }
        break;
      case "4":
        event.preventDefault();
        if (card4) {
          openNewTab(card4.href);
        }
        break;
      case "5":
        event.preventDefault();
        if (card5) {
          openNewTab(card5.href);
        }
        break;
      case "6":
        event.preventDefault();
        if (card6) {
          openNewTab(card6.href);
        }
        break;
      case "7":
        event.preventDefault();
        if (card7) {
          openNewTab(card7.href);
        }
        break;
      case "8":
        event.preventDefault();
        if (card8) {
          openNewTab(card8.href);
        }
        break;
      case "9":
        event.preventDefault();
        if (card9) {
          openNewTab(card9.href);
        }
        break;
      case bs:
        event.preventDefault();
        demozooModal.hide();
        pouetModal.hide();
        introModal.hide();
        trainerModal.hide();
        textModal.hide();
        graphicModal.hide();
        magazineModal.hide();
        advancedModal.hide();
        break;
      case enter:
        event.preventDefault();
        if (gotoRecord) {
          gotoRecord.click();
        }
        break;
      case demozoo:
        demozooModal.show();
        break;
      case pouet:
        pouetModal.show();
        break;
      case intro:
        introModal.show();
        break;
      case trainer:
        trainerModal.show();
        break;
      case nfo:
        textModal.show();
        break;
      case graphic:
        graphicModal.show();
        break;
      case magazine:
        magazineModal.show();
        break;
      case advanced:
        advancedModal.show();
        break;
      case releasers:
        event.preventDefault();
        window.location.href = "/search/releaser";
        break;
      case files:
        event.preventDefault();
        window.location.href = "/search/file";
        break;
      case descriptions:
        event.preventDefault();
        window.location.href = "/search/desc";
        break;
    }
  });
}

function openNewTab(url) {
  const newTab = window.open(url, "_blank");
  newTab.focus();
}
