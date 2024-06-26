// THIS FILE IS SET FOR DELETION
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
    console.info(
      `the editor modal is not open so the editor archive script is not needed`
    );
    return;
  }

  // Modify the assets, readme in archive
  const readmeCP = document.getElementById(`edCopyMe`);
  if (readmeCP == null) {
    console.info(
      `this file artifact is a single-file so the editor archive script is not needed`
    );
    return;
  }

  readmeCP.addEventListener(`input`, function () {
    readmeCP.classList.remove(err);
    readmeCP.classList.remove(ok);
    if (readmeCP.value == ``) {
      return;
    }

    const list = document.getElementById(`edCopyMeList`);
    const exists = Array.from(list.options).some(
      (option) => option.value === readmeCP.value
    );
    if (!exists) {
      readmeCP.classList.add(err);
      document.getElementById(`edCopyMeErr`).textContent = `unknown filename`;
      list.classList.remove(err);
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
  // Modify the assets, readme in archive reset
  document
    .getElementById(`edCopyMeReset`)
    .addEventListener(`click`, function () {
      readmeCP.value = ``;
      readmeCP.classList.remove(err);
      readmeCP.classList.remove(ok);
      readmeHide.classList.remove(err);
      readmeHide.classList.remove(ok);
      document.getElementById(`edCopyMeList`).classList.remove(err);
    });

  // Modify the assets, record readme hide/show
  const readmeHide = document.getElementById(`edHideMe`);
  if (readmeHide.checked == true) {
    document.getElementById(`edHideMeLabel`).classList.add(danger);
    readmeCP.disabled = true;
  }
  readmeHide.addEventListener(`change`, function () {
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

  // Modify the assets, preview in archive
  const previewValue = document.getElementById(`edCopyPreview`);
  const previewList = document.getElementById(`edCopyPreviewList`);
  const previewB = document.getElementById(`edCopyPreviewBtn`);
  previewValue.addEventListener(`input`, function () {
    previewValue.classList.remove(err);
    previewValue.classList.remove(ok);
    if (previewValue.value == ``) {
      return;
    }
  });
  previewB.addEventListener(`click`, function () {
    const exists = Array.from(previewList.options).some(
      (option) => option.value === previewValue.value
    );
    if (!exists) {
      previewValue.classList.add(err);
      document.getElementById(`edCopyPreviewErr`).textContent =
        `unknown filename`;
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
        }
        previewValue.classList.add(ok);
        return response.json();
      })
      .catch((error) => {
        console.log(error);
        document.getElementById(`edCopyPreviewErr`).textContent = error.message;
        previewList.classList.add(err);
        //list.classList.add(err);
        return;
      });
  });
  // Modify the assets, preview in archive reset
  document
    .getElementById(`edCopyPreviewReset`)
    .addEventListener(`click`, function () {
      previewValue.value = ``;
      previewValue.classList.remove(err);
      previewValue.classList.remove(ok);
      document.getElementById(`edCopyPreviewErr`).textContent = ``;
      previewList.classList.remove(err);
    });

  // Modify the assets, ansilove preview in archive
  const ansiloveValue = document.getElementById(`edAnsiLove`);
  const ansiloveList = document.getElementById(`edAnsiLoveList`);
  const ansiloveB = document.getElementById(`edAnsiLoveBtn`);
  ansiloveValue.addEventListener(`input`, function () {
    ansiloveValue.classList.remove(err);
    ansiloveValue.classList.remove(ok);
    if (ansiloveValue.value == ``) {
      return;
    }
  });
  ansiloveB.addEventListener(`click`, function () {
    const exists = Array.from(ansiloveList.options).some(
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
  // Modify the assets, ansilove preview in archive reset
  document
    .getElementById(`edAnsiLoveReset`)
    .addEventListener(`click`, function () {
      ansiloveValue.value = ``;
      ansiloveValue.classList.remove(err);
      ansiloveValue.classList.remove(ok);
      document.getElementById(`edAnsiLoveErr`).textContent = ``;
      ansiloveList.classList.remove(err);
    });

  // Modify the metadata, delete readme asset
  const readmeDel = document.getElementById(`asset-editor-delete-text`);
  readmeDel.disabled = false;
  readmeDel.addEventListener(`click`, function () {
    if (!window.confirm("Delete the readme or textfile?")) {
      return;
    }
    const info = document.getElementById(`asset-editor-hidden`);
    const feed = document.getElementById(`asset-editor-feedback`);
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
})();
