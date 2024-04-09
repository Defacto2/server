/**
 * This function fetches data from Pouët's API and updates the DOM with the production, user votes results.
 * @returns {void}
 */
(() => {
  "use strict";

  /**
   * Returns the URL for the given production ID.
   * @param {string} id - The ID of the production to fetch.
   * @returns {string} The URL for the given production ID.
   */
  const url = (id) => {
    // This URL is to avoid CORS errors,
    // which are not supported by Pouët's API.
    return `${location.protocol}//${location.host}/pouet/vote/${id}`;
  };

  const element = document.getElementById(`pouetVoteID`);
  const row = document.getElementById(`pouetRow`);
  const stars = document.getElementById(`pouetStars`);
  const votes = document.getElementById(`pouetVotes`);
  if (element === null || row === null || stars === null || votes === null)
    return;

  const prodID = element.innerHTML.trim();
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
      const v = result.votes_up + result.votes_down + result.votes_meh;
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
