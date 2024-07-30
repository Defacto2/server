/**
 * @file editor-assets.js
 * This script is the entry point for the artifact editor assets page.
 */
import { progress } from "./uploader.mjs";

(() => {
  "use strict";

  // New file download progress bar.
  progress(`artifact-editor-dl-form`, `artifact-editor-dl-progress`);

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
    afterRequest(
      event,
      `artifact-editor-dl-form`,
      `artifact-editor-dl-up`,
      `artifact-editor-dl-feedback`
    );
  });

  /**
   * The htmx event listener for the artifact editor upload a new file download form.
   * @param {Event} event - The htmx event.
   * @param {string} formId - The form id.
   * @param {string} inputName - The input name.
   * @param {string} feedbackName - The feedback name.
   * @returns {void}
   **/
  function afterRequest(event, formId, inputName, feedbackName) {
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
    artifact.classList.remove(`is-invalid`);
    artifact.classList.remove(`is-valid`);
  });

  //alert(`editor assets script is running`);

  //const danger = `text-danger`;
  // const err = `is-invalid`;
  // const ok = `is-valid`;
  // const fok = `valid-feedback`;
  // const ferr = `invalid-feedback`;

  // const header = {
  //   "Content-type": "application/json; charset=UTF-8",
  // };

  // const saveErr = `server could not save the change`;

  // The table record id and key value, used for all fetch requests
  // It is also used to confirm the existence of the editor modal
  // const id = document.getElementById(`recordID`);
  // if (id == null) {
  //   console.info(
  //     `the editor modal is not open so editor assets script is not needed`
  //   );
  //   return;
  // }

  // // Modify the metadata, delete images asset
  // document
  //   .getElementById(`asset-editor-delete-images`)
  //   .addEventListener(`click`, function () {
  //     if (!window.confirm("Delete the previews and thumbnail?")) {
  //       return;
  //     }
  //     const info = document.getElementById(`asset-editor-hidden`);
  //     const feed = document.getElementById(`asset-editor-feedback`);
  //     fetch("/editor/images/delete", {
  //       method: "POST",
  //       body: JSON.stringify({
  //         id: parseInt(id.value),
  //       }),
  //       headers: header,
  //     })
  //       .then((response) => {
  //         if (!response.ok) {
  //           throw new Error(saveErr);
  //         }
  //         info.classList.add(ok);
  //         feed.classList.add(fok);
  //         feed.textContent = `images deleted, refresh the page to see the change`;
  //         return response.json();
  //       })
  //       .catch((error) => {
  //         info.classList.add(err);
  //         feed.classList.add(ferr);
  //         feed.textContent = error.message;
  //       });
  //   });

  // /// ==============
  // /// TODO: below

  // // Modify the assets, file artifact preview upload
  // const previewUp = document.getElementById(`asset-editor-preview`);
  // const previewUpB = document.getElementById(`edUploadPreviewBtn`);
  // const previewUpR = document.getElementById(`edUploadPreviewReset`);
  // previewUp.addEventListener(`change`, function () {
  //   if (previewUp.value != ``) {
  //     previewUp.classList.remove(err);
  //   }
  // });
  // previewUpB.addEventListener(`click`, function () {
  //   if (previewUp.value == ``) {
  //     previewUp.classList.add(err);
  //     previewUp.classList.remove(ok);
  //     return;
  //   }
  //   previewUp.classList.remove(err);
  //   previewUp.classList.remove(ok);
  //   // upload here
  //   previewUp.classList.add(ok);
  // });
  // previewUpR.addEventListener(`click`, function () {
  //   previewUp.value = ``;
  //   previewUp.classList.remove(err);
  //   previewUp.classList.remove(ok);
  // });

  // Modify the assets, file replacement upload
  // console.log(`file replacement upload`);
  // const artifact = document.getElementById(`artifact-editor-dl-up`);
  // const artifactB = document.getElementById(`asset-editor-dl-submit`);
  // const artifactR = document.getElementById(`asset-editor-dl-reset`);
  // artifact.addEventListener(`change`, function () {
  //   if (artifact.value != ``) {
  //     artifact.classList.remove(err);
  //   }
  // });
  // artifactB.addEventListener(`click`, function () {
  //   if (artifact.value == ``) {
  //     artifact.classList.add(err);
  //     artifact.classList.remove(ok);
  //     return;
  //   }
  //   artifact.classList.remove(err);
  //   artifact.classList.remove(ok);
  //   // Prompt for upload replacement
  //   const confirmation = window.prompt(
  //     `Replace ` + artifact.value + `?\nType "yes" to confirm.`
  //   );
  //   if (confirmation.toLowerCase() != `yes`) {
  //     return;
  //   }
  //   // upload here
  //   artifact.classList.add(ok);
  // });
  // artifactR.addEventListener(`click`, function () {
  //   artifact.value = ``;
  //   artifact.classList.remove(err);
  //   artifact.classList.remove(ok);
  // });
})();
