(() => {
  "use strict";

  const replace = `Are you sure you want to replace `;
  const prmpt = `yes`;
  const dang = `text-danger`;
  const err = `is-invalid`;
  const ok = `is-valid`;

  // The table record id and key value, used for all fetch requests
  // It is also used to confirm the existence of the editor modal
  const id = document.getElementById(`recordID`);
  if (id == null) {
    console.info(`the editor modal is not open so this script is not needed`);
    return;
  }

  // Modify the file metadata, delete readme asset
  const edmeBtn = document.getElementById(`edMeBtn`);
  edmeBtn.addEventListener(`click`, function (event) {
    fetch("/editor/readme/delete", {
      method: "POST",
      body: JSON.stringify({
        id: parseInt(id.value),
      }),
      headers: {
        "Content-type": "application/json; charset=UTF-8",
      },
    })
      .then((response) => {
        if (!response.ok) {
          throw new Error("server could not save the change");
        }
        return response.json();
      })
      .catch((error) => {
        console.log(error);
      });
  })

    // Modify the file metadata, delete images asset
    const edImgBtn = document.getElementById(`edImgBtn`);
    edImgBtn.addEventListener(`click`, function (event) {
      alert(`delete images`)
      fetch("/editor/images/delete", {
        method: "POST",
        body: JSON.stringify({
          id: parseInt(id.value),
        }),
        headers: {
          "Content-type": "application/json; charset=UTF-8",
        },
      })
        .then((response) => {
          if (!response.ok) {
            throw new Error("server could not save the change");
          }
          return response.json();
        })
        .catch((error) => {
          console.log(error);
        });
    })

  // Modify the file assets, readme in archive
  const readmeValue = document.getElementById(`recordReadme`);
  const readmeList = document.getElementById(`recordReadmeList`);
  readmeValue.addEventListener(`input`, function (event) {
    readmeValue.classList.remove(err);
    readmeValue.classList.remove(ok);
    if (readmeValue.value == ``) {
      return;
    }
    // automatic upload
    let exists = Array.from(readmeList.options).some(
      (option) => option.value === readmeValue.value
    );
    if (!exists) {
      readmeValue.classList.add(err);
      return;
    }

    fetch("/editor/readme/copy", {
      method: "POST",
      body: JSON.stringify({
        id: parseInt(id.value),
        target: readmeValue.value,
      }),
      headers: {
        "Content-type": "application/json; charset=UTF-8",
      },
    })
      .then((response) => {
        if (!response.ok) {
          throw new Error("server could not save the change");
        }
        readme.classList.add(ok);
        return response.json();
      })
      .catch((error) => {
        readmeErr.textContent = error.message;
        readmeHide.classList.add(err);
      });

    readmeValue.classList.add(ok);
  });

  // Modify the file assets, record readme hide/show
  const readme = document.getElementById(`recordHideReadme`);
  const readmeL = document.getElementById(`recordHideReadmeLabel`);
  const readmeName = document.getElementById(`recordReadme`);
  const readmeHide = document.getElementById(`recordHideReadme`);
  const readmeErr = document.getElementById(`recordHideReadmeErr`);
  if (readme.checked == true) {
    readmeL.classList.add(dang);
    readmeName.disabled = true;
  }
  readme.addEventListener(`change`, function (event) {
    if (readme.checked == true) {
      readmeL.classList.add(dang);
      readmeName.disabled = true;
    } else {
      readmeL.classList.remove(dang);
      readmeName.disabled = false;
    }
    fetch("/editor/readme", {
      method: "POST",
      body: JSON.stringify({
        id: parseInt(id.value),
        readme: readme.checked,
      }),
      headers: {
        "Content-type": "application/json; charset=UTF-8",
      },
    })
      .then((response) => {
        if (!response.ok) {
          throw new Error("server could not save the change");
        }
        readme.classList.add(ok);
        return response.json();
      })
      .catch((error) => {
        readmeErr.textContent = error.message;
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
      return;
    }
    previewValue.classList.add(ok);
  });

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
      replace + artifact.value + `?\nType "` + prmpt + `" to confirm.`
    );
    if (confirmation.toLowerCase() != prmpt) {
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
    onlineL.classList.add(dang);
  }
  online.addEventListener(`change`, function (event) {
    if (online.checked == true) {
      onlineL.classList.remove(dang);
    } else {
      onlineL.classList.add(dang);
    }
  });
})();
