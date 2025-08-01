/**
 * @file editor-artifact.js
 * This script is the entry point for the artifact editor page.
 */
import {
  date as validateDate,
  releaser as validateReleaser,
  repository as validateGitHub,
  color16 as validate16color,
  youtube as validateYouTube,
  number as validateNumber,
} from "./artifact-validate.mjs";
import { clipValue, formatPaste, getElmById, titleize } from "./helper.mjs";

(() => {
  "use strict";

  /**
   * Footer buttons to toggle the editor modals
   */
  function activeBtn(elms) {
    return function () {
      elms.forEach((e) => {
        e.disabled = true;
        e.classList.remove("btn-outline-primary");
        e.classList.add("btn-light");
      });
    };
  }
  function inactiveBtn(elms) {
    return function () {
      elms.forEach((e) => {
        e.disabled = false;
        e.classList.remove("btn-light");
        e.classList.add("btn-outline-primary");
      });
    };
  }
  const artifactEditor = document.getElementById("artifact-editor-modal");
  const artifactEditors = document.getElementsByName(
    "artifact-editor-dataeditor",
  );
  artifactEditor.addEventListener("shown.bs.modal", () => {
    activeBtn(artifactEditors)();
  });
  artifactEditor.addEventListener("hidden.bs.modal", () => {
    inactiveBtn(artifactEditors)();
  });
  const assetEditor = document.getElementById("asset-editor-modal");
  const assetEditors = document.getElementsByName("artifact-editor-fileeditor");
  assetEditor.addEventListener("shown.bs.modal", () => {
    activeBtn(assetEditors)();
  });
  assetEditor.addEventListener("hidden.bs.modal", () => {
    inactiveBtn(assetEditors)();
  });
  const emulateEditor = document.getElementById("emulate-editor-modal");
  const emulateEditors = document.getElementsByName(
    "artifact-editor-emueditor",
  );
  emulateEditor.addEventListener("shown.bs.modal", () => {
    activeBtn(emulateEditors)();
  });
  emulateEditor.addEventListener("hidden.bs.modal", () => {
    inactiveBtn(emulateEditors)();
  });

  const erp = document.getElementById("emulate-run-program");
  if (erp !== null) {
    const egp = document.getElementById("emulate-guess-program");
    if (egp === null) {
      throw new Error("The guess program input is missing.");
    }
    erp.addEventListener("input", () => {
      erp.value = erp.value.toUpperCase().replace(/\s{2,}/g, " ");
      const val = erp.value;
      if (val !== "" && val !== " ") {
        egp.disabled = true;
        return;
      }
      egp.disabled = false;
    });
  }

  const aekv = getElmById(`artifact-dataeditor-key-value`);
  if (aekv === null) {
    throw new Error("The key value is missing.");
  }
  const aekl = getElmById(`artifact-dataeditor-key-label`);
  if (aekl === null) {
    throw new Error("The key label is missing.");
  }
  aekl.addEventListener(`click`, () =>
    clipValue(`artifact-dataeditor-key-value`),
  );

  const afkv = getElmById(`artifact-fileeditor-key-value`);
  if (afkv === null) {
    throw new Error("The key value is missing.");
  }
  const afkl = getElmById(`artifact-fileeditor-key-label`);
  if (aekl === null) {
    throw new Error("The key label is missing.");
  }
  afkl.addEventListener(`click`, () =>
    clipValue(`artifact-fileeditor-key-value`),
  );

  const udid = getElmById(`artifact-dataeditor-unique-id-value`);
  if (udid === null) {
    throw new Error("The unique id value is missing.");
  }
  const udidl = getElmById(`artifact-dataeditor-unique-id-label`);
  if (udidl === null) {
    throw new Error("The unique id label is missing.");
  }
  udidl.addEventListener(`click`, () =>
    clipValue(`artifact-dataeditor-unique-id-value`),
  );

  const ufid = getElmById(`artifact-fileeditor-unique-id-value`);
  if (ufid === null) {
    throw new Error("The unique id value is missing.");
  }
  const ufidl = getElmById(`artifact-fileeditor-unique-id-label`);
  if (ufidl === null) {
    throw new Error("The unique id label is missing.");
  }
  ufidl.addEventListener(`click`, () =>
    clipValue(`artifact-fileeditor-unique-id-value`),
  );

  const locv = getElmById(`artifact-editor-location-value`);
  if (locv === null) {
    throw new Error("The location value is missing.");
  }
  const locvl = getElmById(`artifact-editor-location-label`);
  if (locvl === null) {
    throw new Error("The location label is missing.");
  }
  locvl.addEventListener(`click`, () =>
    clipValue(`artifact-editor-location-value`),
  );

  const tmploc = getElmById(`artifact-editor-templocation`);
  if (tmploc !== null && tmploc !== undefined) {
    const tmplocl = getElmById(`artifact-editor-templocation-label`);
    if (tmplocl !== null) {
      tmplocl.addEventListener(`click`, () =>
        clipValue(`artifact-editor-templocation`),
      );
    }
  }

  const osl = document.getElementById("artifact-editor-os-label");
  if (osl === null) {
    throw new Error("The operating system label is missing.");
  }
  const osv = document.getElementById("artifact-editor-operating-system");
  if (osv === null) {
    throw new Error("The operating system input is missing.");
  }
  osv.addEventListener("input", newOSLabel);
  const tagv = document.getElementById("artifact-editor-category");
  if (tagv === null) {
    throw new Error("The category input is missing.");
  }
  tagv.addEventListener("input", newTagLabel);
  newOSLabel();
  newTagLabel();
  /**
   * New operating system label based on the selected option in the dropdown.
   */
  function newOSLabel() {
    const index = osv.selectedIndex;
    if (index == 0) {
      osv.classList.remove("is-valid");
      osv.classList.add("is-invalid");
    }
  }
  /**
   * New tag or category label based on the selected option in the dropdown.
   */
  function newTagLabel() {
    const index = tagv.selectedIndex;
    if (index == 0) {
      tagv.classList.remove("is-valid");
      tagv.classList.add("is-invalid");
    }
  }

  const presetTags = document.getElementsByName("prereset-classifications");
  if (presetTags.length === 0) {
    throw new Error("The preset classifications are missing.");
  }
  for (let i = 0; i < presetTags.length; i++) {
    presetTag(i);
  }
  /**
   * Undo the classification for a given element.
   *
   * @param {number} i - The index of the element in the classifications array.
   */
  function presetTag(i) {
    const elm = presetTags[i];
    const os = elm.getAttribute("data-preset-os");
    if (os === null) {
      throw new Error("data-preset-os attribute is required for ${elm.id}.");
    }
    const cat = elm.getAttribute("data-preset-tag");
    if (cat === null) {
      throw new Error("data-preset-tag attribute is required for ${elm.id}.");
    }
    elm.addEventListener("click", (e) => {
      e.preventDefault();
      osv.value = os;
      osv.classList.remove("is-invalid");
      tagv.value = cat;
      tagv.classList.remove("is-invalid");
      newOSLabel();
      newTagLabel();
    });
  }

  const filename = document.getElementById("artifact-editor-filename");
  if (filename === null) {
    throw new Error("The filename input is missing.");
  }
  filename.addEventListener("input", (e) => {
    e.target.classList.remove("is-valid");
    e.target.classList.remove("is-invalid");
    if (e.target.value.trim().length === 0) {
      e.target.classList.add("is-invalid");
    }
  });
  const fnr = document.getElementById("artifact-editor-filename-reset");
  if (fnr === null) {
    throw new Error("The filename reset is missing.");
  }
  const fnUndo = document.getElementsByName("artifact-editor-filename-undo");
  if (fnUndo.length === 0) {
    throw new Error("The filename resetter is missing.");
  }
  fnr.addEventListener("click", () => {
    filename.classList.remove("is-valid");
    if (fnUndo.length === 0) {
      throw new Error("The filename resetter is missing.");
    }
    filename.value = fnUndo[0].value;
    filename.classList.add("is-valid");
    filename.classList.remove("is-invalid");
    if (filename.value.trim().length === 0) {
      filename.classList.add("is-invalid");
    }
  });

  const rel1 = document.getElementById("artifact-editor-releaser-1");
  if (rel1 === null) {
    throw new Error("The releaser 1 input is missing.");
  }
  rel1.addEventListener("input", (e) => validateReleaser(e.target));

  const rel2 = document.getElementById("artifact-editor-releaser-2");
  if (rel2 === null) {
    throw new Error("The releaser 2 input is missing.");
  }
  rel2.addEventListener("input", (e) => validateReleaser(e.target));

  const relUndo = document.getElementById("artifact-editor-releaser-undo");
  if (relUndo === null) {
    throw new Error("The releasers reset is missing.");
  }
  relUndo.addEventListener("click", undoRels);
  function undoRels() {
    const revert1 = rel1.getAttribute("data-reset-rel1");
    if (revert1 === null) {
      throw new Error(
        "data-reset-rel1 attribute is required for artifact-editor-releaser-1.",
      );
    }
    rel1.value = revert1;
    validateReleaser(rel1);
    const revert2 = rel2.getAttribute("data-reset-rel2");
    if (revert2 === null) {
      throw new Error(
        "data-reset-rel2 attribute is required for artifact-editor-releaser-2.",
      );
    }
    rel2.value = revert2;
    validateReleaser(rel2);
  }

  const title = document.getElementById("artifact-editor-title");
  if (title === null) {
    throw new Error("The title input is missing.");
  }
  title.addEventListener("paste", formatPaste);
  title.addEventListener("input", (e) => {
    e.target.classList.remove("is-valid");
  });
  const titleUndo = document.getElementById("artifact-editor-title-undo");
  if (titleUndo === null) {
    throw new Error("The title reset is missing.");
  }
  const titleU = document.getElementsByName("artifact-editor-titleundo");
  if (titleU.length === 0) {
    throw new Error("The title resetter is missing.");
  }
  titleUndo.addEventListener("click", () => {
    title.classList.remove("is-valid");
    if (titleU.length === 0) {
      throw new Error("The title resetter is missing.");
    }
    title.value = titleU[0].value;
    title.classList.add("is-valid");
  });
  const titleizeBtn = document.getElementById("artifact-editor-titleize");
  if (titleizeBtn.length === 0) {
    throw new Error("The titleize button is missing.");
  }
  titleizeBtn.addEventListener("click", () => {
    title.value = titleize(title.value);
    let event = new Event("keyup");
    title.dispatchEvent(event);
  });
  const titleDelete = document.getElementById("artifact-editor-title-delete");
  if (titleDelete.length === 0) {
    throw new Error("The title delete button is missing.");
  }
  titleDelete.addEventListener("click", () => {
    title.value = "";
    let event = new Event("keyup");
    title.dispatchEvent(event);
  });

  const ct = document.getElementById("artifact-editor-credit-text");
  if (ct === null) {
    throw new Error("The creator text input is missing.");
  }
  ct.addEventListener("input", (e) => {
    e.target.classList.remove("is-valid");
  });
  const ci = document.getElementById("artifact-editor-credit-ill");
  if (ci === null) {
    throw new Error("The creator illustrator input is missing.");
  }
  ci.addEventListener("input", (e) => {
    e.target.classList.remove("is-valid");
  });
  const cp = document.getElementById("artifact-editor-credit-prog");
  if (cp === null) {
    throw new Error("The creator programmer input is missing.");
  }
  cp.addEventListener("input", (e) => {
    e.target.classList.remove("is-valid");
  });
  const ca = document.getElementById("artifact-editor-credit-audio");
  if (ca === null) {
    throw new Error("The creator audio input is missing.");
  }
  ca.addEventListener("input", (e) => {
    e.target.classList.remove("is-valid");
  });
  const cUndos = document.getElementById("artifact-editor-credits-undo");
  if (cUndos === null) {
    throw new Error("The creator resetter is missing.");
  }
  const cUndo = document.getElementById("artifact-editor-credit-undo");
  if (cUndo === null) {
    throw new Error("The creator reset is missing.");
  }
  cUndo.addEventListener("click", () => {
    if (cUndos.length === 0) {
      throw new Error("The creator resetter is missing.");
    }
    const creators = cUndos.value.split(";");
    if (creators.length != 4) {
      throw new Error("The creator resetter values are invalid.");
    }
    const text = creators[0];
    const ill = creators[1];
    const prog = creators[2];
    const audio = creators[3];
    ct.value = text;
    ci.value = ill;
    cp.value = prog;
    ca.value = audio;
  });

  const vt = document.getElementById("artifact-editor-virustotal");
  if (vt === null) {
    throw new Error("The virustotal input is missing.");
  }
  vt.addEventListener("input", (e) => {
    e.target.classList.remove("is-valid", "is-invalid");
    const value = e.target.value.trim();
    if (value.length != 0) {
      if (!value.startsWith("https://www.virustotal.com/")) {
        e.target.classList.add("is-invalid");
      }
    }
  });

  const year = document.getElementById("artifact-editor-year");
  if (year === null) {
    throw new Error("The year input is missing.");
  }
  year.addEventListener("input", () => {
    const val = parseInt(year.value, 10);
    if (val >= 79 && val <= 99) {
      year.value = 1900 + val;
    }
    validateDate(year, month, day, unknownDate);
  });
  const month = document.getElementById("artifact-editor-month");
  if (month === null) {
    throw new Error("The month input is missing.");
  }
  month.addEventListener("input", () => {
    validateDate(year, month, day, unknownDate);
  });
  const day = document.getElementById("artifact-editor-day");
  if (day === null) {
    throw new Error("The day input is missing.");
  }
  day.addEventListener("input", () => {
    validateDate(year, month, day, unknownDate);
  });

  let unknownDate = false;
  if (year.value == 0 && month.value == 0 && day.value == 0) {
    unknownDate = true;
  }
  const dateReset = document.getElementById("artifact-editor-date-reset");
  if (dateReset === null) {
    throw new Error("The date reset is missing.");
  }
  const dateResetter = document.getElementById("artifact-editor-date-resetter");
  if (dateResetter === null) {
    throw new Error("The date resetter is missing.");
  }
  dateReset.addEventListener("click", () => {
    year.classList.remove("is-invalid", "is-valid");
    month.classList.remove("is-invalid", "is-valid");
    day.classList.remove("is-invalid", "is-valid");
    const value = dateResetter.value;
    const values = value.split("-");
    if (values.length != 3) {
      throw new Error("The date resetter values are invalid.");
    }
    year.value = values[0];
    month.value = values[1];
    day.value = values[2];
  });

  const cmmt = document.getElementById("artifact-editor-comment");
  if (cmmt === null) {
    throw new Error("The comment input is missing.");
  }
  cmmt.addEventListener("input", (e) => {
    e.target.classList.remove("is-valid");
    const unsetDateOfRelease =
      year.value == 0 && month.value == 0 && day.value == 0;
    if (unsetDateOfRelease == false) {
      return;
    }
    const mmddyyDatePattern =
      /(0[1-9]|1[0-2])\/(0[1-9]|[12][0-9]|3[01])\/(\d{2})/;
    let match = mmddyyDatePattern.exec(e.target.value);
    if (!match) {
      const mmddyyDatePatternDash =
        /(0[1-9]|1[0-2])-(0[1-9]|[12][0-9]|3[01])-(\d{2})/;
      match = mmddyyDatePatternDash.exec(e.target.value);
    }
    if (match) {
      const mm = match[1];
      const md = match[2];
      const my = match[3];
      const val = parseInt(my, 10);
      year.value = 2000 + val;
      if (val >= 79 && val <= 99) {
        year.value = 1900 + val;
      }
      month.value = mm;
      day.value = md;
      const submitValues = document.getElementById(
        "artifact-editor-date-update",
      );
      if (submitValues !== null) {
        submitValues.click();
      }
    }
  });
  const cmmtReset = document.getElementById("artifact-editor-comment-undo");
  if (cmmtReset === null) {
    throw new Error("The comment reset is missing.");
  }
  const cmmtResetter = document.getElementById(
    "artifact-editor-comment-resetter",
  );
  if (cmmtResetter === null) {
    throw new Error("The comment resetter is missing.");
  }
  cmmtReset.addEventListener("click", () => {
    cmmt.classList.remove("is-valid");
    cmmt.value = cmmtResetter.value;
  });

  const dateLastMod = document.getElementById("artifact-editor-date-lastmod");
  if (dateLastMod === null) {
    // do nothing as the date last mod input is optional
  } else {
    const dateLastModder = document.getElementById(
      "artifact-editor-date-lastmodder",
    );
    if (dateLastModder === null) {
      throw new Error("The date last modder input is missing.");
    }
    dateLastMod.addEventListener("click", () => {
      year.classList.remove("is-invalid", "is-valid");
      month.classList.remove("is-invalid", "is-valid");
      day.classList.remove("is-invalid", "is-valid");
      const value = dateLastModder.value;
      const values = value.split("-");
      if (values.length != 3) {
        throw new Error("The date last modder values are invalid.");
      }
      year.value = values[0];
      month.value = values[1];
      day.value = values[2];
    });
  }

  const linksReset = document.getElementById("artifact-editor-links-reset");
  if (linksReset === null) {
    throw new Error("The links reset is missing.");
  }
  const youtube = document.getElementById("artifact-editor-youtube");
  const youtubeReset = document.getElementById("artifact-editor-youtube-reset");
  if (youtube === null || youtubeReset === null) {
    throw new Error("A YouTube input is missing.");
  }
  const demozoo = document.getElementById("artifact-editor-demozoo");
  const demozooReset = document.getElementById("artifact-editor-demozoo-reset");
  if (demozoo === null || demozooReset === null) {
    throw new Error("A Demozoo input is missing.");
  }
  const pouet = document.getElementById("artifact-editor-pouet");
  const pouetReset = document.getElementById("artifact-editor-pouet-reset");
  if (pouet === null || pouetReset === null) {
    throw new Error("A Pouet input is missing.");
  }
  const colors16 = document.getElementById("artifact-editor-16colors");
  const colors16Reset = document.getElementById(
    "artifact-editor-16colors-reset",
  );
  if (colors16 === null || colors16Reset === null) {
    throw new Error("A 16colors input is missing.");
  }
  const github = document.getElementById("artifact-editor-github");
  const githubReset = document.getElementById("artifact-editor-github-reset");
  if (github === null || githubReset === null) {
    throw new Error("A GitHub input is missing.");
  }
  const relations = document.getElementById("artifact-editor-relations");
  const relationsReset = document.getElementById(
    "artifact-editor-relations-reset",
  );
  if (relations === null || relationsReset === null) {
    throw new Error("A relations input is missing.");
  }
  const websites = document.getElementById("artifact-editor-websites");
  const websitesReset = document.getElementById(
    "artifact-editor-websites-reset",
  );
  if (websites === null || websitesReset === null) {
    throw new Error("A websites input is missing.");
  }
  // on paste event for websites remove any http:// or https:// protcols
  websites.addEventListener("paste", () => {
    setTimeout(() => {
      websites.value = websites.value.replace(/https?:\/\//, "");
    }, 0);
  });

  linksReset.addEventListener("click", () => {
    youtube.classList.remove("is-invalid", "is-valid");
    demozoo.classList.remove("is-invalid", "is-valid");
    pouet.classList.remove("is-invalid", "is-valid");
    colors16.classList.remove("is-invalid", "is-valid");
    github.classList.remove("is-invalid", "is-valid");
    relations.classList.remove("is-invalid", "is-valid");
    websites.classList.remove("is-invalid", "is-valid");
    youtube.value = youtubeReset.value;
    demozoo.value = demozooReset.value;
    pouet.value = pouetReset.value;
    colors16.value = colors16Reset.value;
    github.value = githubReset.value;
    relations.value = relationsReset.value;
    websites.value = websitesReset.value;
  });
  const demozooSanity = 450000,
    pouetSanity = 200000;

  // on paste event for websites remove the watch url: https://www.youtube.com/watch?v=
  youtube.addEventListener("paste", () => {
    setTimeout(() => {
      youtube.value = youtube.value.replace(
        /https?:\/\/www\.youtube\.com\/watch\?v=/,
        "",
      );
    }, 0);
  });
  youtube.addEventListener("input", (e) => validateYouTube(e.target));
  demozoo.addEventListener("input", (e) =>
    validateNumber(e.target, demozooSanity),
  );
  pouet.addEventListener("input", (e) => validateNumber(e.target, pouetSanity));
  // on paste event for websites remove the https://16colo.rs/ URL
  colors16.addEventListener("paste", () => {
    setTimeout(() => {
      colors16.value = colors16.value.replace(/https?:\/\/16colo\.rs\//, "");
    }, 0);
  });
  colors16.addEventListener("input", (e) => validate16color(e.target));
  // on paste event for github remove the https://github.com/ URL
  github.addEventListener("paste", () => {
    setTimeout(() => {
      github.value = github.value.replace(/https?:\/\/github\.com\//, "");
    }, 0);
  });
  github.addEventListener("input", (e) => validateGitHub(e.target));
  // relations and websites are optional
})();
