(() => {
  "use strict";

  const danger = `text-danger`;
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
    console.info(`the editor modal is not open so this script is not needed`);
    return;
  }

  // Modify the file metadata, delete readme asset
  document
    .getElementById(`edBtnRead`)
    .addEventListener(`click`, function (event) {
      if (!window.confirm("Delete the readme or textfile?")) {
        return;
      }
      const info = document.getElementById(`edBtnsHide`);
      const feed = document.getElementById(`edBtnsFeedback`);
      fetch("/editor/readme/delete", {
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
          feed.textContent = `readme or textfile deleted, refresh the page to see the change`;
          return response.json();
        })
        .catch((error) => {
          info.classList.add(err);
          feed.classList.add(ferr);
          feed.textContent = error.message;
          console.log(error);
        });
    });

  // Modify the file metadata, delete images asset
  document
    .getElementById(`edBtnImgs`)
    .addEventListener(`click`, function (event) {
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

  // Modify the file assets, readme in archive
  const readmeCP = document.getElementById(`edCopyMe`);
  readmeCP.addEventListener(`input`, function (event) {
    readmeCP.classList.remove(err);
    readmeCP.classList.remove(ok);
    if (readmeCP.value == ``) {
      return;
    }

    const list = document.getElementById(`edCopyMeList`);
    let exists = Array.from(list.options).some(
      (option) => option.value === readmeCP.value
    );
    if (!exists) {
      readmeCP.classList.add(err);
      return;
    }

    fetch("/editor/readme/copy", {
      method: "POST",
      body: JSON.stringify({
        id: parseInt(id.value),
        target: readmeCP.value,
      }),
      headers: header,
    })
      .then((response) => {
        if (!response.ok) {
          throw new Error(saveErr);
        }
        readmeCP.classList.add(ok);
        return response.json();
      })
      .catch((error) => {
        document.getElementById(`edCopyMeErr`).textContent = error.message;
        list.classList.add(err);
        return;
      });
  });
  // Modify the file assets, readme in archive reset
  document
    .getElementById(`edCopyMeReset`)
    .addEventListener(`click`, function (event) {
      readmeCP.value = ``;
      readmeCP.classList.remove(err);
      readmeCP.classList.remove(ok);
      readmeHide.classList.remove(err);
      readmeHide.classList.remove(ok);
    });

  // Modify the file assets, record readme hide/show
  const readmeHide = document.getElementById(`edHideMe`);
  if (readmeHide.checked == true) {
    document.getElementById(`edHideMeLabel`).classList.add(danger);
    readmeCP.disabled = true;
  }
  readmeHide.addEventListener(`change`, function (event) {
    const label = document.getElementById(`edHideMeLabel`);
    if (readmeHide.checked == true) {
      label.classList.add(danger);
      readmeCP.disabled = true;
    } else {
      label.classList.remove(danger);
      readmeCP.disabled = false;
    }
    fetch("/editor/readme/hide", {
      method: "POST",
      body: JSON.stringify({
        id: parseInt(id.value),
        readme: readmeHide.checked,
      }),
      headers: header,
    })
      .then((response) => {
        if (!response.ok) {
          throw new Error(saveErr);
        }
        readmeHide.classList.add(ok);
        return response.json();
      })
      .catch((error) => {
        document.getElementById(`edHideMeErr`).textContent = error.message;
        readmeHide.classList.add(err);
      });
  });

  // Modify the file assets, preview in archive
  const previewValue = document.getElementById(`recordPreview`);
  const previewList = document.getElementById(`recordPreviewList`);
  const previewB = document.getElementById(`recordPreviewBtn`);
  previewValue.addEventListener(`input`, function (event) {
    previewValue.classList.remove(err);
    previewValue.classList.remove(ok);
    if (previewValue.value == ``) {
      return;
    }
  });
  previewB.addEventListener(`click`, function (event) {
    let exists = Array.from(previewList.options).some(
      (option) => option.value === previewValue.value
    );
    if (!exists) {
      previewValue.classList.add(err);
      document.getElementById(`edCopyImgsErr`).textContent = `unknown filename`;
      previewList.classList.remove(err);
      return;
    }
    fetch("/editor/images/copy", {
      method: "POST",
      body: JSON.stringify({
        id: parseInt(id.value),
        target: previewValue.value,
      }),
      headers: header,
    })
      .then((response) => {
        console.log(response);
        if (!response.ok) {
          console.log(`not ok`);
          throw new Error(saveErr);
          return;
        }
        previewValue.classList.add(ok);
        return response.json();
      })
      .catch((error) => {
        console.log(error);
        document.getElementById(`edCopyImgsErr`).textContent = error.message;
        previewList.classList.add(err);
        //list.classList.add(err);
        return;
      });
  });
  // Modify the file assets, preview in archive reset
  document
    .getElementById(`edCopyImgsReset`)
    .addEventListener(`click`, function () {
      previewValue.value = ``;
      previewValue.classList.remove(err);
      previewValue.classList.remove(ok);
      document.getElementById(`edCopyImgsErr`).textContent = ``;
      previewList.classList.remove(err);
    });

  // Modify the file assets, ansilove preview in archive
  const ansiloveValue = document.getElementById(`edAnsiLove`);
  const ansiloveList = document.getElementById(`edAnsiLoveList`);
  const ansiloveB = document.getElementById(`edAnsiLoveBtn`);
  ansiloveValue.addEventListener(`input`, function (event) {
    ansiloveValue.classList.remove(err);
    ansiloveValue.classList.remove(ok);
    if (ansiloveValue.value == ``) {
      return;
    }
  });
  ansiloveB.addEventListener(`click`, function (event) {
    let exists = Array.from(ansiloveList.options).some(
      (option) => option.value === ansiloveValue.value
    );
    if (!exists) {
      ansiloveValue.classList.add(err);
      document.getElementById(`edAnsiLoveErr`).textContent = `unknown filename`;
      ansiloveList.classList.remove(err);
      return;
    }
    fetch("/editor/ansilove/copy", {
      method: "POST",
      body: JSON.stringify({
        id: parseInt(id.value),
        target: ansiloveValue.value,
      }),
      headers: header,
    })
      .then((response) => {
        console.log(response);
        if (!response.ok) {
          console.log(`not ok`);
          throw new Error(saveErr);
          return;
        }
        ansiloveValue.classList.add(ok);
        return response.json();
      })
      .catch((error) => {
        console.log(error);
        document.getElementById(`edAnsiLoveErr`).textContent = error.message;
        ansiloveList.classList.add(err);
        //list.classList.add(err);
        return;
      });
  });
  // Modify the file assets, ansilove preview in archive reset
  document
    .getElementById(`edAnsiLoveReset`)
    .addEventListener(`click`, function () {
      ansiloveValue.value = ``;
      ansiloveValue.classList.remove(err);
      ansiloveValue.classList.remove(ok);
      document.getElementById(`edAnsiLoveErr`).textContent = ``;
      ansiloveList.classList.remove(err);
    });

  /// ==============
  /// TODO: below

  // Modify the file assets, file artifact preview upload
  const previewUp = document.getElementById(`recordPreviewUp`);
  const previewUpB = document.getElementById(`recordPreviewUpBtn`);
  const previewUpR = document.getElementById(`recordPreviewUpReset`);
  previewUp.addEventListener(`change`, function (event) {
    if (previewUp.value != ``) {
      previewUp.classList.remove(err);
    }
  });
  previewUpB.addEventListener(`click`, function (event) {
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
  previewUpR.addEventListener(`click`, function (event) {
    previewUp.value = ``;
    previewUp.classList.remove(err);
    previewUp.classList.remove(ok);
  });

  // Modify the file assets, file artifact replacement upload
  const artifact = document.getElementById(`recordArtifact`);
  const artifactB = document.getElementById(`recordArtifactBtn`);
  const artifactR = document.getElementById(`recordArtifactReset`);
  artifact.addEventListener(`change`, function (event) {
    if (artifact.value != ``) {
      artifact.classList.remove(err);
    }
  });
  artifactB.addEventListener(`click`, function (event) {
    if (artifact.value == ``) {
      artifact.classList.add(err);
      artifact.classList.remove(ok);
      return;
    }
    artifact.classList.remove(err);
    artifact.classList.remove(ok);
    // Prompt for upload replacement
    let confirmation = window.prompt(
      `Replace ` + artifact.value + `?\nType "yes" to confirm.`
    );
    if (confirmation.toLowerCase() != `yes`) {
      return;
    }
    // upload here
    artifact.classList.add(ok);
  });
  artifactR.addEventListener(`click`, function (event) {
    artifact.value = ``;
    artifact.classList.remove(err);
    artifact.classList.remove(ok);
  });

  // Modify the file metadata, online and public
  const online = document.getElementById(`recordOnline`);
  const onlineL = document.getElementById(`recordOnlineLabel`);
  if (online.checked != true) {
    onlineL.classList.add(danger);
  }
  online.addEventListener(`change`, function (event) {
    if (online.checked == true) {
      onlineL.classList.remove(danger);
    } else {
      onlineL.classList.add(danger);
    }
  });
})();
