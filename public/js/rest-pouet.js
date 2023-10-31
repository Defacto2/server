(() => {
  "use strict";

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

  const delay = 500;
  let timeout = null;

  demozoo.addEventListener(`input`, parseEvent);
  demozoo.addEventListener(`paste`, (change) => {
    resetEvent();
    eventFunction(change);
  });
  reset.addEventListener(`click`, resetEvent);

  function parseEvent(change) {
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
    authors.innerText = ``;
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
        authors.innerText = releasers(result.author_nicks);
        plats.innerText = platform(result.platforms);
        if (plats.innerText === `` && result.supertype === `graphics`) {
          plats.innerText = `graphic or image`;
        }
        date.innerText = `from ` + result.release_date;
        prod.classList.remove(`d-none`);
        const err = validate(result);
        if (err !== ``) {
          invalid.classList.remove(`d-none`);
          invalid.innerText = err;
        }
      });
  }

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

  function platform(platforms) {
    if (platforms === null) return ``;
    let plats = [];
    platforms.forEach((element) => {
      plats.push(element.name);
    });
    if (plats.length === 0) return ``;
    return "for " + plats.join(` + `);
  }

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
