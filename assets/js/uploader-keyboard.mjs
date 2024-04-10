import { getModalById, focusModalById } from "./uploader.mjs";
export default keyboardShortcuts;

const pouetModal = focusModalById("uploader-pouet", "pouet-submission");
const demozooModal = focusModalById("uploader-demozoo", "demozoo-submission");
const introModal = focusModalById("uploader-intro", "uploader-intro-file");
const textModal = getModalById("uploaderText");
const graphicModal = getModalById("uploaderImg");
const magazineModal = getModalById("uploaderMag");
const advancedModal = getModalById("uploaderAdv");

const demozoo = "d",
  pouet = "p",
  intro = "i",
  nfo = "n",
  graphic = "g",
  magazine = "m",
  advanced = "a";

/**
 * Binds keyboard shortcuts to specific actions.
 */
export function keyboardShortcuts() {
  document.addEventListener("keydown", function (event) {
    if (!event.ctrlKey || !event.altKey) {
      return;
    }
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
    }
  });
}
