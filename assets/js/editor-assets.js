// THIS FILE IS SET FOR DELETION
(() => {
  "use strict";

  //const danger = `text-danger`;
  const err = `is-invalid`;
  const ok = `is-valid`;
  const fok = `valid-feedback`;
  const ferr = `invalid-feedback`;

  const header = {
    "Content-type": "application/json; charset=UTF-8",
  };

  const saveErr = `server could not save the change`;

  // The table record id and key value, used for all fetch requests
  // It is also used to confirm the existence of the editor modal
  const id = document.getElementById(`recordID`);
  if (id == null) {
    console.info(
      `the editor modal is not open so editor assets script is not needed`
    );
    return;
  }

  // Modify the metadata, delete images asset
  document.getElementById(`edBtnImgs`).addEventListener(`click`, function () {
    if (!window.confirm("Delete the previews and thumbnail?")) {
      return;
    }
    const info = document.getElementById(`edBtnsHide`);
    const feed = document.getElementById(`edBtnsFeedback`);
    fetch("/editor/images/delete", {
      method: "POST",
      body: JSON.stringify({
        id: parseInt(id.value),
      }),
      headers: header,
    })
      .then((response) => {
        if (!response.ok) {
          throw new Error(saveErr);
        }
        info.classList.add(ok);
        feed.classList.add(fok);
        feed.textContent = `images deleted, refresh the page to see the change`;
        return response.json();
      })
      .catch((error) => {
        info.classList.add(err);
        feed.classList.add(ferr);
        feed.textContent = error.message;
      });
  });

  /// ==============
  /// TODO: below

  // Modify the assets, file artifact preview upload
  const previewUp = document.getElementById(`edUploadPreview`);
  const previewUpB = document.getElementById(`edUploadPreviewBtn`);
  const previewUpR = document.getElementById(`edUploadPreviewReset`);
  previewUp.addEventListener(`change`, function () {
    if (previewUp.value != ``) {
      previewUp.classList.remove(err);
    }
  });
  previewUpB.addEventListener(`click`, function () {
    if (previewUp.value == ``) {
      previewUp.classList.add(err);
      previewUp.classList.remove(ok);
      return;
    }
    previewUp.classList.remove(err);
    previewUp.classList.remove(ok);
    // upload here
    previewUp.classList.add(ok);
  });
  previewUpR.addEventListener(`click`, function () {
    previewUp.value = ``;
    previewUp.classList.remove(err);
    previewUp.classList.remove(ok);
  });

  // Modify the assets, file replacement upload
  const artifact = document.getElementById(`edUploadArtifact`);
  const artifactB = document.getElementById(`edUploadArtifactBtn`);
  const artifactR = document.getElementById(`edUploadArtifactReset`);
  artifact.addEventListener(`change`, function () {
    if (artifact.value != ``) {
      artifact.classList.remove(err);
    }
  });
  artifactB.addEventListener(`click`, function () {
    if (artifact.value == ``) {
      artifact.classList.add(err);
      artifact.classList.remove(ok);
      return;
    }
    artifact.classList.remove(err);
    artifact.classList.remove(ok);
    // Prompt for upload replacement
    const confirmation = window.prompt(
      `Replace ` + artifact.value + `?\nType "yes" to confirm.`
    );
    if (confirmation.toLowerCase() != `yes`) {
      return;
    }
    // upload here
    artifact.classList.add(ok);
  });
  artifactR.addEventListener(`click`, function () {
    artifact.value = ``;
    artifact.classList.remove(err);
    artifact.classList.remove(ok);
  });
})();
