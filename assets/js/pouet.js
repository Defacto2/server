(() => {
  "use strict";

  const url = (id) => {
    // This URL is to avoid CORS errors,
    // which are not supported by Pouët's API.
    return `${location.protocol}//${location.host}/pouet/vote/${id}`;
  };

  let prodElm = document.getElementById(`pouetProdID`);
  let row = document.getElementById(`pouetRow`);
  let stars = document.getElementById(`pouetStars`);
  let votes = document.getElementById(`pouetVotes`);
  if (prodElm === null || row === null || stars === null || votes === null)
    return;

  let prodID = prodElm.innerHTML.trim();
  console.info(`Requesting the Pouët API for production #${prodID}`);
  fetch(url(prodID), {
    method: `GET`,
    headers: {
      "Content-Type": `application/json charset=UTF-8`,
    },
  })
    .then((response) => {
      if (!response.ok) {
        const error = `A network error occurred requesting API`;
        throw new Error(`${error}: ${response.statusText} ${response.status}`);
      }
      return response.json();
    })
    .then((result) => {
      let v = result.votes_up + result.votes_down + result.votes_meh;
      if (v === 0) {
        row.classList.add(`d-none`);
        return;
      }
      let s = `${result.stars} star`;
      if (result.stars !== 1) s += `s`;
      stars.innerHTML = s;

      s = `${v} vote`;
      if (v !== 1) s += `s`;
      votes.innerHTML = s;

      row.classList.remove(`d-none`);
    })
    .catch((err) => {
      const error = `An error occurred requesting API`;
      console.error(`${error}: ${err}`);
    });
})();
