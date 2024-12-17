/**
 * @file editor-assets.js
 * This script is the entry point for the artifact editor assets page.
 */
import {
  checkDuplicate,
  checkErrors,
  checkSize,
  progress,
  resetInput,
} from "./uploader.mjs";
import { getElmById } from "./helper.mjs";

(() => {
  "use strict";

  // New file download progress bar.
  progress(`artifact-editor-dl-form`, `artifact-editor-dl-progress`);
  progress(`artifact-editor-preview-form`, `artifact-editor-preview-progress`);

  const previewReset = getElmById(`artifact-editor-preview-reset`);
  const previewInput = getElmById(`artifact-editor-replace-preview`);
  const previewSubmit = getElmById(`artifact-editor-preview-submit`);
  const alert = getElmById(`artifact-editor-dl-alert`);
  const reset = getElmById(`artifact-editor-dl-reset`);
  const lastMod = getElmById(`artifact-editor-last-modified`);
  const results = getElmById("artifact-editor-dl-results");
  const fileInput = getElmById(`artifact-editor-dl-up`);
  fileInput.addEventListener("change", checkFile);

  async function checkFile() {
    resetInput(fileInput, alert, results);
    const file1 = this.files[0];
    let errors = [checkSize(file1)];
    checkErrors(errors, alert, fileInput, results);
    checkDuplicate(file1, alert, fileInput, results);

    const lastModified = file1.lastModified,
      currentTime = new Date().getTime(),
      oneHourMs = 60 * 60 * 1000;
    const underOneHour = currentTime - lastModified < oneHourMs;
    if (!underOneHour) {
      lastMod.value = lastModified;
    }
  }

  const dataEditor = document.getElementById("artifact-editor-modal");
  if (dataEditor == null) {
    console.error(`the data editor modal is missing`);
    return;
  }
  const assetEditor = document.getElementById("asset-editor-modal");
  if (assetEditor == null) {
    console.error(`the asset editor modal is missing`);
    return;
  }
  const jsdosEditor = document.getElementById("emulate-editor-modal");
  if (jsdosEditor == null) {
    console.error(`the emulate editor modal is missing`);
    return;
  }
  const deleteEditor = document.getElementById("artifact-delete-forever-modal");
  if (deleteEditor == null) {
    console.error(`the delete editor modal is missing`);
    return;
  }

  // Automatically open the editor modals based on the URL hash.
  const dataModal = new bootstrap.Modal(dataEditor);
  const assetModal = new bootstrap.Modal(assetEditor);
  const emulateModal = new bootstrap.Modal(jsdosEditor);
  const deleteModal = new bootstrap.Modal(deleteEditor);
  const parsedUrl = new URL(window.location.href);
  switch (parsedUrl.hash) {
    case `#data-editor`:
      dataModal.show();
      history.replaceState(null, "", window.location.pathname);
      break;
    case `#file-editor`:
      assetModal.show();
      history.replaceState(null, "", window.location.pathname);
      break;
    case `#emulate-editor`:
      emulateModal.show();
      history.replaceState(null, "", window.location.pathname);
      break;
    default:
    // note, the #runapp hash is used by js-dos
  }

  // Keyboard shortcuts for the editor modals.
  document.addEventListener("keydown", function (event) {
    if (!event.altKey || !event.shiftKey) {
      return;
    }
    const refresher = "Enter";
    const dataEditor = "F9";
    const assetEditor = "F10";
    const emulateEditor = "F11";
    const deleteEditor = "Delete";
    const closeEditors = "F12";
    switch (event.key) {
      case refresher:
        event.preventDefault();
        location.reload();
        break;
      case dataEditor:
        event.preventDefault();
        assetModal.hide();
        emulateModal.hide();
        deleteModal.hide();
        dataModal.show();
        break;
      case assetEditor:
        event.preventDefault();
        dataModal.hide();
        emulateModal.hide();
        deleteModal.hide();
        assetModal.show();
        break;
      case emulateEditor:
        event.preventDefault();
        dataModal.hide();
        assetModal.hide();
        deleteModal.hide();
        emulateModal.show();
        break;
      case deleteEditor:
        event.preventDefault();
        dataModal.hide();
        assetModal.hide();
        emulateModal.hide();
        deleteModal.show();
        break;
      case closeEditors:
        event.preventDefault();
        dataModal.hide();
        assetModal.hide();
        emulateModal.hide();
        deleteModal.hide();
        break;
    }
  });

  // New file download form event listener.
  document.body.addEventListener("htmx:afterRequest", function (event) {
    afterFormRequest(
      event,
      `artifact-editor-dl-form`,
      `artifact-editor-dl-up`,
      `artifact-editor-dl-feedback`
    );
    afterFormRequest(
      event,
      `artifact-editor-preview-form`,
      `artifact-editor-replace-preview`,
      `artifact-editor-preview-feedback`
    );
    afterDeleteRequest(
      event,
      "artifact-editor-image-delete",
      "artifact-editor-image-feedback"
    );
    afterDeleteRequest(
      event,
      "artifact-editor-imagepreview-delete",
      "artifact-editor-image-feedback"
    );
    afterDeleteRequest(
      event,
      "artifact-editor-imagethumb-delete",
      "artifact-editor-image-feedback"
    );
    afterDeleteRequest(
      event,
      "artifact-editor-image-pixelate",
      "artifact-editor-preview-feedback"
    );
    afterLinkRequest(
      event,
      "artifact-editor-link-delete",
      "artifact-editor-link-feedback"
    );
    afterLinkRequest(
      event,
      "artifact-editor-comp-previewcopy",
      "artifact-editor-comp-feedback"
    );
    afterLinkRequest(
      event,
      "artifact-editor-comp-previewtext",
      "artifact-editor-comp-feedback"
    );
    afterLinkRequest(
      event,
      "artifact-editor-comp-textcopy",
      "artifact-editor-comp-feedback"
    );
    afterLinkRequest(
      event,
      "artifact-editor-comp-dizcopy",
      "artifact-editor-comp-feedback"
    );
  });

  /**
   * After link request event listener.
   * @param {Event} event - The htmx event.
   * @param {string} inputId - The inputId is the id of the input element that triggered the request,
   * or the name of the input element that triggered the request.
   * @param {string} feedbackId - The feedback name.
   * @returns {void}
   **/
  function afterLinkRequest(event, inputId, feedbackId) {
    if (event.detail.elt === null) return;
    if (
      event.detail.elt.id !== `${inputId}` &&
      event.detail.elt.name !== inputId
    )
      return;
    const feedback = document.getElementById(feedbackId);
    if (feedback === null) {
      throw new Error(
        `The htmx successful feedback element ${feedbackId} is null`
      );
    }
    const errClass = "text-danger";
    const xhr = event.detail.xhr;
    const statusFound = xhr.status === 200 || xhr.status === 302;
    if ((event.detail && event.detail.successful) || statusFound) {
      feedback.innerText = `${xhr.responseText}`;
      feedback.classList.remove(errClass);
      return;
    }
    if (
      event.detail &&
      event.detail.failed !== undefined &&
      event.detail.failed &&
      event.detail.xhr
    ) {
      feedback.classList.add(errClass);
      feedback.innerText =
        `Something on the server is not working, ` +
        `${xhr.status} status: ${xhr.responseText}.`;
      return;
    }
    feedback.classList.add(errClass);
    feedback.innerText =
      "Something with the browser is not working," +
      " please try again or refresh the page.";
  }

  function afterDeleteRequest(event, inputId, feedbackId) {
    if (event.detail.elt === null) return;
    if (event.detail.elt.id !== `${inputId}`) return;
    const feedback = document.getElementById(feedbackId);
    if (feedback === null) {
      throw new Error(
        `The htmx successful feedback element ${feedbackId} is null`
      );
    }
    const errClass = "text-danger";
    const okClass = "text-success";
    const xhr = event.detail.xhr;
    if (event.detail.successful) {
      feedback.innerText = `The delete request was successful, about to refresh the page.`;
      feedback.classList.remove(errClass);
      feedback.classList.add(okClass);
      setTimeout(() => {
        location.reload();
      }, 500);
      return;
    }
    if (event.detail.failed && event.detail.xhr) {
      feedback.classList.add(errClass);
      feedback.innerText =
        `Something on the server is not working, ` +
        `${xhr.status} status: ${xhr.responseText}.`;
      return;
    }
    feedback.classList.add(errClass);
    feedback.innerText =
      "Something with the browser is not working," +
      " please try again or refresh the page.";
  }

  /**
   * The htmx event listener for the artifact editor upload a new file download form.
   * @param {Event} event - The htmx event.
   * @param {string} formId - The form id.
   * @param {string} inputName - The input name.
   * @param {string} feedbackName - The feedback name.
   * @returns {void}
   **/
  function afterFormRequest(event, formId, inputName, feedbackName) {
    if (event.detail.elt === null) return;
    if (event.detail.elt.id !== `${formId}`) return;
    const input = document.getElementById(inputName);
    if (input === null) {
      throw new Error(`The htmx successful input element ${inputName} is null`);
    }
    const feedback = document.getElementById(feedbackName);
    if (feedback === null) {
      throw new Error(
        `The htmx successful feedback element ${feedbackName} is null`
      );
    }
    if (event.detail.successful) {
      return successful(event, input, feedback);
    }
    if (event.detail.failed && event.detail.xhr) {
      return errorXhr(event, input, feedback);
    }
    errorBrowser(input, feedback);
  }

  function successful(event, input, feedback) {
    const xhr = event.detail.xhr;
    feedback.innerText = `${xhr.responseText}`;
    feedback.classList.remove("invalid-feedback");
    feedback.classList.add("valid-feedback");
    input.classList.remove("is-invalid");
    input.classList.add("is-valid");
    setTimeout(() => {
      location.reload();
    }, 500);
  }

  function errorXhr(event, input, feedback) {
    const xhr = event.detail.xhr;
    feedback.innerText = `Something on the server is not working, ${xhr.status} status: ${xhr.responseText}.`;
    feedback.classList.remove("valid-feedback");
    feedback.classList.add("invalid-feedback");
    input.classList.remove("is-valid");
    input.classList.add("is-invalid");
  }

  function errorBrowser(input, feedback) {
    input.classList.remove("is-valid");
    input.classList.add("is-invalid");
    feedback.innerText =
      "Something with the browser is not working, please try again or refresh the page.";
    feedback.classList.remove("d-none");
  }

  // New file download form reset button.
  reset.addEventListener(`click`, function () {
    alert.innerText = "";
    alert.classList.add("d-none");
    fileInput.value = ``;
    fileInput.classList.remove(`is-invalid`, `is-valid`);
  });

  previewReset.addEventListener(`click`, function () {
    previewInput.value = ``;
    previewInput.classList.remove(`is-invalid`, `is-valid`);
  });

  // Automatically submit the preview form when a file is selected.
  previewInput.addEventListener("change", function (evt) {
    if (evt.target.value.trim() === ``) {
      return;
    }
    console.log(`Submitting the image or photo preview form`);
    previewSubmit.click();
  });
})();
