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
    switch (event.key) {
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
