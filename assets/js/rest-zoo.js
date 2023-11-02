/**
 * This module handles the fetching and display of data from the Demozoo API.
 * @module rest-zoo
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
    // which are not supported by Demozoo's API.
    return `${location.protocol}//${location.host}/zoo/prod/${id}`;
  };

  const demozoo = document.getElementById(`demozooProdID`);
  const prod = document.getElementById(`demozooProd`);
  const title = document.getElementById(`demozooProdTitle`);
  const authors = document.getElementById(`demozooProdAuthors`);
  const plats = document.getElementById(`demozooProdPlat`);
  const date = document.getElementById(`demozooProdDate`);
  const invalid = document.getElementById(`demozooProdInvalid`);
  const reset = document.getElementById(`demozooProdReset`);
  const submit = document.getElementById(`demozooSubmit`);
  const hide = `d-none`;
  const errProd =  "This prod id is not valid"

  const largestID = 999999;
  const delay = 500; // milliseconds
  let timeout = null;

  demozoo.addEventListener(`input`, parseEvent);
  demozoo.addEventListener(`paste`, (change) => {
    resetEvent();
    eventFunction(change);
  });
  reset.addEventListener(`click`, resetEvent);
  submit.addEventListener(`click`, function (event) {
    document.getElementById(`demozooProdUploader`).submit();
  });

  /**
   * Parses an event and sets a timeout to execute the event function with a delay.
   * @param {any} change - The event to be parsed.
   */
  function parseEvent(change) {
    clearTimeout(timeout);
    resetEvent();
    if (change.target.value !== "") {
      prod.classList.remove(hide);
      title.innerText = `Will lookup ${change.target.value}...`;
    }
    timeout = setTimeout(() => {
      eventFunction(change);
    }, delay);
  }

  /**
   * This function is called when an event is triggered.
   * It extracts the numbers from the input string and checks if the product id is valid.
   * If the product id is invalid, it displays an error message.
   * @param {Event} change - The event object that triggered the function.
   */
  function eventFunction(change) {
    const str = change.target.value;
    if (str === "") {
      submit.disabled = true;
      return;
    }
    const mat = str.match(/\d+/g);
    if (mat === null) {
      invalid.classList.remove(hide);
      invalid.innerText = errProd;
      return;
    }
    const numbers = mat.map(Number);
    if (numbers.length === 0) {
      return;
    }
    if (numbers[0] > largestID) {
        invalid.classList.remove(hide);
        invalid.innerText = errProd;
        return;
    }
    change.target.value = numbers[0];
    check(numbers[0]);
  }

  /**
   * Resets the event by hiding the prod and invalid elements, and clearing the inner text of title, groups, plats, and date elements.
   */
  function resetEvent() {
    prod.classList.add(hide);
    invalid.classList.add(hide);
    title.innerText = ``;
    authors.innerText = ``;
    plats.innerText = ``;
    date.innerText = ``;
  }

  /**
   * Fetches data from the Demzoo API for a given production ID and updates the DOM with the result.
   * @param {number} prodID - The ID of the production to fetch from the API.
   */
  function check(prodID) {
    console.info(`Requesting the Demozoo API for production #${prodID}`);
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
        authors.innerText = releasers(result.author_nicks);
        plats.innerText = platform(result.platforms);
        if (plats.innerText === `` && result.supertype === `graphics`) {
          plats.innerText = `graphic or image`;
        }
        date.innerText = `from ` + result.release_date;
        prod.classList.remove(hide);
        const err = validate(result);
        if (err !== ``) {
          submit.disabled = true;
          invalid.classList.remove(hide);
          invalid.innerText = err;
        }
      });
  }
  /**
   * Returns a string containing the names of the releasers of a given data array.
   * @param {Array} data - An array of objects containing releaser information.
   * @returns {string} - A string containing the names of the releasers, separated by ' + '.
   */
  function releasers(authors) {
    if (authors === null) return ``;
    let groups = [];
    authors.forEach((element) => {
      if (element.releaser.is_group) {
        groups.push(element.releaser.name);
      }
    });
    if (groups.length === 0) return ``;
    return "by " + groups.join(` + `);
  }

  /**
   * Returns a string of concatenated types from an array of types.
   * @param {Array} data - An array of types.
   * @returns {string} - A string of concatenated types.
   */
  function platform(platforms) {
    if (platforms === null) return ``;
    let plats = [];
    platforms.forEach((element) => {
      plats.push(element.name);
    });
    if (plats.length === 0) return ``;
    return "for " + plats.join(` + `);
  }

  /**
   * Validates the result of a REST API call to the Demozoo API.
   * @param {Object} result - The result object returned by the API.
   * @returns {string} An error message if the result is invalid, or an empty string if the result is valid.
   */
  function validate(result) {
    if (result === null) return `result error is null`;
    switch (result.supertype) {
      case "production":
        break;
      case "graphics":
        return ``; // okay
      default:
        return `production type ${result.supertype} is not allowed`;
    }
    let plats = [];
    // list of platforms
    // https://demozoo.org/api/v1/platforms/
    result.platforms.forEach((value) => {
      switch (value.id) {
        case 1: // windows
        case 4: // msdos
        case 7: // linux
        case 10: // macos
        case 46: // js
        case 48: // java
        case 84: // freebsd
        case 94: // macos classic
          plats.push(value.name);
          break;
      }
    });
    if (plats.length === 0) return `platform is not allowed`;
    let types = [];
    // list of types
    // https://demozoo.org/api/v1/production_types/
    result.types.forEach((value) => {
      switch (value.id) {
        case 1: // demo
        case 4: // intro
        case 5: // diskmag
        case 6: // tool
        case 9: // pack
        case 13: // cracktro
        case 23: // graphics
        case 24: // ascii
        case 25: // ascii collection
        case 26: // ansi
        case 27: // exe graphics
        case 34: // video
        case 36: // photo
        case 41: // bbstro
        case 47: // magazine
        case 49: // textmag
        case 51: // artpack
          types.push(value.name);
          break;
      }
    });
    if (types.length === 0) return `production type is not allowed`;
    return ``;
  }
})();
