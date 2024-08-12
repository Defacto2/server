/**
 * @file editor-assets.js
 * This script is the entry point for the artifact editor assets page.
 */
import { progress } from "./uploader.mjs";

(() => {
  "use strict";

  // New file download progress bar.
  progress(`artifact-editor-dl-form`, `artifact-editor-dl-progress`);
  progress(`artifact-editor-preview-form`, `artifact-editor-preview-progress`);

  const previewReset = document.getElementById(`artifact-editor-preview-reset`);
  if (previewReset == null) {
    console.error(`the reset preview button is missing`);
    return;
  }
  const previewInput = document.getElementById(
    `artifact-editor-replace-preview`
  );
  if (previewInput == null) {
    console.error(`the form preview input is missing`);
    return;
  }

  const reset = document.getElementById(`artifact-editor-dl-reset`);
  if (reset == null) {
    console.error(`the reset button is missing`);
    return;
  }
  const artifact = document.getElementById(`artifact-editor-dl-up`);
  if (artifact == null) {
    console.error(`the artifact file input is missing`);
    return;
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

  // Automatically open the editor modals based on the URL hash.
  const dataModal = new bootstrap.Modal(dataEditor);
  const assetModal = new bootstrap.Modal(assetEditor);
  const emulateModal = new bootstrap.Modal(jsdosEditor);
  const parsedUrl = new URL(window.location.href);
  switch (parsedUrl.hash) {
    case `#data-editor`:
      dataModal.show();
      break;
    case `#file-editor`:
      assetModal.show();
      break;
    case `#emulate-editor`:
      emulateModal.show();
      break;
    default:
  }

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
      "artifact-editor-image-pixelate",
      "artifact-editor-preview-feedback"
    );
    afterLinkRequest(
      event,
      "artifact-editor-link-delete",
      "artifact-editor-link-feedback"
    );
  });

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

  function afterLinkRequest(event, inputId, feedbackId) {
    if (event.detail.elt === null) return;
    if (event.detail.elt.id !== `${inputId}`) return;
    const feedback = document.getElementById(feedbackId);
    if (feedback === null) {
      throw new Error(
        `The htmx successful feedback element ${feedbackId} is null`
      );
    }
    const errClass = "text-danger";
    const xhr = event.detail.xhr;
    if (event.detail.successful) {
      feedback.innerText = `${xhr.responseText}`;
      feedback.classList.remove(errClass);
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
    artifact.value = ``;
    artifact.classList.remove(`is-invalid`, `is-valid`);
  });

  previewReset.addEventListener(`click`, function () {
    previewInput.value = ``;
    previewInput.classList.remove(`is-invalid`, `is-valid`);
  });
})();
