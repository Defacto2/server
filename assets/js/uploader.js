/**
 * This module handles the uploader functionality for the website.
 * It contains functions to validate client input and show/hide modals.
 * @module uploader
 * @requires bootstrap
 */
import { advancedUploader } from "./uploader-advanced.mjs";
import { imageSubmit } from "./uploader-image.mjs";
import { introSubmit } from "./uploader-intro.mjs";
import { magazineSubmit } from "./uploader-magazine.mjs";
import { textSubmit } from "./uploader-text.mjs";
import { pagination } from "./uploader.mjs";
import { keyboardShortcuts } from "./uploader-keyboard.mjs";

(() => {
  "use strict";
  keyboardShortcuts();
  pagination("paginationRange");
  advancedUploader(`advSubmit`);
  imageSubmit(`imageSubmit`);
  introSubmit(`introSubmit`);
  magazineSubmit(`magSubmit`);
  textSubmit(`textSubmit`, `textUploader`);
})();
