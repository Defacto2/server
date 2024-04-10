// uploader.js

import { advancedUploader } from "./uploader-advanced.mjs";
import { imageSubmit } from "./uploader-image.mjs";
import { introSubmit } from "./uploader-intro.mjs";
import { keyboardShortcuts as uploadKeys } from "./uploader-keyboard.mjs";
import { magazineSubmit } from "./uploader-magazine.mjs";
import { textUploader } from "./uploader-text.mjs";
import { submitter } from "./uploader-submitter.mjs";
import load from "./uploader-htmx.mjs";

(() => {
  "use strict";
  uploadKeys();
  advancedUploader(`advSubmit`);
  imageSubmit(`imageSubmit`);
  introSubmit(`introSubmit`);
  magazineSubmit(`magSubmit`);
  textUploader(`textSubmit`, `textUploader`);
  submitter(`demozoo-submission`, `uploader-intro-file`, `Demozoo`);
  submitter(`pouet-submission`, `uploader-intro-file`, `Pouët`);
  load();
})();
