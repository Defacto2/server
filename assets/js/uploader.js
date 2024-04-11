// uploader.js

import { advancedUploader } from "./uploader-advanced.mjs";
import { imageSubmit } from "./uploader-image.mjs";
import {
  submit as introSubmit,
  progress as introProgress,
} from "./uploader-intro.mjs";
import { keyboardShortcuts as uploadKeys } from "./uploader-keyboard.mjs";
import { magazineSubmit } from "./uploader-magazine.mjs";
import { textUploader } from "./uploader-text.mjs";
import { submitter } from "./uploader-submitter.mjs";

(() => {
  "use strict";
  uploadKeys();
  advancedUploader(`advSubmit`);
  imageSubmit(`imageSubmit`);
  introSubmit(`uploader-intro-submit`);
  introProgress();
  magazineSubmit(`magSubmit`);
  textUploader(`textSubmit`, `textUploader`);
  submitter(`demozoo-submission`, `Demozoo`);
  submitter(`pouet-submission`, `Pouët`);
})();
