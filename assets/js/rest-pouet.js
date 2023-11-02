/**
 * This module handles the fetching and display of data from the Pouët API.
 * @module rest-pouet
 */

(() => {
  "use strict";

  /**
   * Returns the URL for a given production ID.
   * @param {number} id - The ID of the production.
   * @returns {string} - The URL for the production.
   */
  const url = (id) => {
    // This URL is to avoid CORS errors,
    // which are not supported by Pouet's API.
    return `${location.protocol}//${location.host}/pouet/prod/${id}`;
  };

  const pouet = document.getElementById(`pouetProdsID`);
  const prod = document.getElementById(`pouetProd`);
  const title = document.getElementById(`pouetProdTitle`);
  const groups = document.getElementById(`pouetProdGroups`);
  const plats = document.getElementById(`pouetProdPlat`);
  const date = document.getElementById(`pouetProdDate`);
  const invalid = document.getElementById(`pouetProdInvalid`);
  const reset = document.getElementById(`pouetProdReset`);

  const delay = 500; // milliseconds
  let timeout = null;

  pouet.addEventListener(`input`, parseEvent);
  pouet.addEventListener(`paste`, (change) => {
    resetEvent();
    eventFunction(change);
  });
  reset.addEventListener(`click`, resetEvent);

  /**
   * Parses an event and sets a timeout to execute the event function with a delay.
   * @param {any} change - The event to be parsed.
   */
  function parseEvent(change) {
    clearTimeout(timeout);
    resetEvent();
    timeout = setTimeout(() => {
      eventFunction(change);
    }, delay);
  }

  /**
   * This function handles the event when the value of an input field changes.
   * It extracts the first number from the input value and passes it to the check function.
   * If the input value does not contain any numbers, or the extracted number is invalid, it displays an error message.
   *
   * @param {Event} change - The change event object.
   */
  function eventFunction(change) {
    const str = change.target.value;
    if (str === "") {
      return;
    }
    const mat = str.match(/\d+/g);
    if (mat === null) {
      invalid.classList.remove(`d-none`);
      invalid.innerText = "This prod id is invalid";
      return;
    }
    const numbers = mat.map(Number);
    if (numbers.length === 0) {
      return;
    }
    change.target.value = numbers[0];
    check(numbers[0]);
  }

  /**
   * Resets the event by hiding the prod and invalid elements, and clearing the inner text of title, groups, plats, and date elements.
   */
  function resetEvent() {
    const hide = `d-none`;
    prod.classList.add(hide);
    invalid.classList.add(hide);
    title.innerText = ``;
    groups.innerText = ``;
    plats.innerText = ``;
    date.innerText = ``;
  }

  /**
   * Fetches data from the Pouët API for a given production ID and updates the DOM with the result.
   * @param {number} prodID - The ID of the production to fetch from the API.
   */
  function check(prodID) {
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
          throw new Error(
            `${error}: ${response.statusText} ${response.status}`
          );
        }
        return response.json();
      })
      .then((result) => {
        title.innerText = result.title;
        groups.innerText = releasers(result.groups);
        plats.innerText = `on ` + result.platform + typers(result.types);
        date.innerText = `from ` + result.release_date;
        prod.classList.remove(`d-none`);
        if (result.valid !== true) {
          invalid.classList.remove(`d-none`);
          invalid.innerText = "This prod is not valid";
        }
      });
  }

  /**
   * Returns a string containing the names of the releasers of a given data array.
   * @param {Array} data - An array of objects containing releaser information.
   * @returns {string} - A string containing the names of the releasers, separated by ' + '.
   */
  function releasers(data) {
    if (data === null) return ``;
    let groups = [];
    data.forEach((element) => {
      groups.push(element.name);
    });
    if (groups.length === 0) return ``;
    return "by " + groups.join(` + `);
  }

  /**
   * Returns a string of concatenated types from an array of types.
   * @param {Array} data - An array of types.
   * @returns {string} - A string of concatenated types.
   */
  function typers(data) {
    if (data === null) return ``;
    let types = [];
    data.forEach((element) => {
      types.push(element);
    });
    if (types.length === 0) return ``;
    return " " + types.join(` + `);
  }
})();
