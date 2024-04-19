// uploader.js

import { advancedUploader } from "./uploader-advanced.mjs";
import { keyboardShortcuts as uploadKeys } from "./uploader-keyboard.mjs";
import { submitter } from "./uploader-submitter.mjs";
import { submit as imageSubmit } from "./uploader-image.mjs";
import { submit as introSubmit } from "./uploader-intro.mjs";
import { submit as magazineSubmit } from "./uploader-magazine.mjs";
import { submit as textSubmit } from "./uploader-text.mjs";
import { progress } from "./uploader.mjs";

(() => {
  "use strict";
  uploadKeys();

  submitter(`demozoo-submission`, `Demozoo`);
  submitter(`pouet-submission`, `Pouët`);

  imageSubmit(`uploader-image-submit`);
  progress(`uploader-image-form`, `uploader-image-progress`);

  introSubmit(`uploader-intro-submit`);
  progress(`uploader-intro-form`, `uploader-intro-progress`);

  magazineSubmit(`uploader-magazine-submit`);
  progress(`uploader-magazine-form`, `uploader-magazine-progress`);

  textSubmit(`uploader-text-submit`);
  progress(`uploader-text-form`, `uploader-text-progress`);

  advancedUploader(`advSubmit`);
})();
