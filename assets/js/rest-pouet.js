(() => {
  "use strict";

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

  const delay = 500;
  let timeout = null;

  pouet.addEventListener(`input`, parseEvent);
  pouet.addEventListener(`paste`, (change) => {
    resetEvent();
    eventFunction(change);
  });
  reset.addEventListener(`click`, resetEvent);

  function parseEvent(change) {
    console.log(`parseEvent`, change);
    clearTimeout(timeout);
    resetEvent();
    timeout = setTimeout(() => {
      eventFunction(change);
    }, delay);
  }

  function eventFunction(change) {
    const str = change.target.value;
    const mat = str.match(/\d+/g);
    if (mat === null) {
      console.log(`no numbers`);
      return;
    }
    const numbers = mat.map(Number);
    if (numbers.length === 0) {
      console.log(`no numbers`);
      return;
    }
    change.target.value = numbers[0];
    check(numbers[0]);
  }

  function resetEvent() {
    const hide = `d-none`;
    prod.classList.add(hide);
    invalid.classList.add(hide);
    title.innerText = ``;
    groups.innerText = ``;
    plats.innerText = ``;
    date.innerText = ``;
  }

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

  function releasers(data) {
    if (data === null) return ``;
    let groups = [];
    data.forEach((element) => {
      groups.push(element.name);
    });
    if (groups.length === 0) return ``;
    return "by " + groups.join(` + `);
  }

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
