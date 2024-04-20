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
