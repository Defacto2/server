// uploader.js

import { advancedUploader } from "./uploader-advanced.mjs";
import { imageSubmit } from "./uploader-image.mjs";
import { introSubmit } from "./uploader-intro.mjs";
import { keyboardShortcuts } from "./uploader-keyboard.mjs";
import { magazineSubmit } from "./uploader-magazine.mjs";
import { pagination } from "./uploader.mjs";
import { textUploader } from "./uploader-text.mjs";

(() => {
  "use strict";
  keyboardShortcuts();
  pagination("paginationRange");
  advancedUploader(`advSubmit`);
  imageSubmit(`imageSubmit`);
  introSubmit(`introSubmit`);
  magazineSubmit(`magSubmit`);
  textUploader(`textSubmit`, `textUploader`);
})();
