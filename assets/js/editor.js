(() => {
  "use strict";

  const danger = `text-danger`;
  const err = `is-invalid`;
  const ok = `is-valid`;
  const fok = `valid-feedback`;
  const ferr = `invalid-feedback`;
  const hide = `d-none`;
  const header = {
    "Content-type": "application/json; charset=UTF-8",
  };

  // The table record id and key value, used for all fetch requests
  // It is also used to confirm the existence of the editor modal
  const id = document.getElementById(`recordID`);
  if (id == null) {
    console.info(
      `the editor modal is not open so the editor script is not needed`
    );
    return;
  }

  // Modify the file metadata, File artifact is online and public
  const online = document.getElementById(`recordOnline`);
  const onlineL = document.getElementById(`recordOnlineLabel`);
  if (online.checked != true) {
    onlineL.classList.add(danger);
  }
  online.addEventListener(`change`, function (event) {
    let path = `/editor/online/false`;
    if (online.checked == true) {
      path = `/editor/online/true`;
    }
    fetch(path, {
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
        if (online.checked == true) {
          onlineL.classList.remove(danger);
        } else {
          onlineL.classList.add(danger);
        }
        return response.json();
      })
      .catch((error) => {
        console.log(error.message);
      });
  });

  // Modify the file metadata, Title
  const recTitle = document.getElementById(`recordTitle`);
  recTitle.addEventListener(`input`, function (event) {
    const infoErr = document.getElementById(`recordTitleErr`);
    const ogLabel = document.getElementById(`recordTitleOG`);
    const ogText = document.getElementById(`recordTitleOGValue`).textContent;
    if (recTitle.value != ogText && recTitle.value.length > 0) {
      ogLabel.classList.remove(hide);
    } else {
      ogLabel.classList.add(hide);
    }
    recTitle.classList.remove(err);
    infoErr.classList.add(hide);
    fetch("/editor/title", {
      method: "POST",
      body: JSON.stringify({
        id: parseInt(id.value),
        value: recTitle.value,
      }),
      headers: header,
    })
      .then((response) => {
        if (!response.ok) {
          throw new Error(saveErr);
        }
        infoErr.classList.add(hide);
        recTitle.classList.remove(err);
        recTitle.classList.add(ok);
        return response.json();
      })
      .catch((error) => {
        infoErr.classList.remove(hide);
        recTitle.classList.add(err);
        console.log(error.message);
      });
  });

  // Modify the file metadata, Year, month, day of release, save button
  const recYMDSave = document.getElementById(`recordYMDSave`);
  recYMDSave.addEventListener(`click`, function (event) {
    year.classList.remove(ok);
    month.classList.remove(ok);
    day.classList.remove(ok);
    fetch("/editor/ymd", {
      method: "POST",
      body: JSON.stringify({
        id: parseInt(id.value),
        year: parseInt(year.value),
        month: parseInt(month.value),
        day: parseInt(day.value),
      }),
      headers: header,
    })
      .then((response) => {
        if (!response.ok) {
          throw new Error(saveErr);
        }
        year.classList.add(ok);
        month.classList.add(ok);
        day.classList.add(ok);
        recYMDSave.classList.remove(danger);
        return response.json();
      })
      .catch((error) => {
        recYMDSave.classList.add(danger);
        console.log(error.message);
      });
  });

  // Modify the file metadata, Year, month, day of release, Use last modification button
  document
    .getElementById(`recordLMBtn`)
    .addEventListener(`click`, function (event) {
      year.classList.remove(err);
      month.classList.remove(err);
      day.classList.remove(err);
      year.classList.remove(ok);
      month.classList.remove(ok);
      day.classList.remove(ok);
      const split = document.getElementById(`recordLastMod`).value.split(`-`);
      if (split.length != 3) {
        console.error(`invalid last modified date provided by server`);
        return;
      }
      year.value = split[0];
      month.value = split[1];
      day.value = split[2];
    });

  // Modify the file metadata, Year, month, day of release, reset button
  document
    .getElementById(`recordYMDReset`)
    .addEventListener(`click`, function (event) {
      const ogy = document.getElementById(`recordOgY`).value;
      const ogm = document.getElementById(`recordOgM`).value;
      const ogd = document.getElementById(`recordOgD`).value;
      year.value = ogy;
      month.value = ogm;
      day.value = ogd;
      year.classList.remove(err);
      month.classList.remove(err);
      day.classList.remove(err);
      year.classList.remove(ok);
      month.classList.remove(ok);
      day.classList.remove(ok);
      recYMDSave.disabled = false;
    });

  // Modify the file metadata, Year, month, day of release
  const year = document.getElementById(`recordYear`);
  const month = document.getElementById(`recordMonth`);
  const day = document.getElementById(`recordDay`);
  year.addEventListener(`input`, function (event) {
    if (year.value >= 1980 && year.value <= 2023) {
      year.classList.remove(err);
      recYMDSave.disabled = false;
      return;
    }
    // year can only be empty when month and day are empty
    if (year.value == `` && month.value == `` && day.value == ``) {
      year.classList.remove(err);
      month.classList.remove(err);
      day.classList.remove(err);
      recYMDSave.disabled = false;
      return;
    }
    year.classList.add(err);
    recYMDSave.disabled = true;
  });
  month.addEventListener(`input`, function (event) {
    if (month.value >= 1 && month.value <= 12) {
      month.classList.remove(err);
      recYMDSave.disabled = false;
      return;
    }
    if (year.value == `` && month.value == `` && day.value == ``) {
      year.classList.remove(err);
      month.classList.remove(err);
      day.classList.remove(err);
      recYMDSave.disabled = false;
      return;
    }
    // month can only be empty when day is empty
    if (month.value == `` && day.value == ``) {
      month.classList.remove(err);
      day.classList.remove(err);
      recYMDSave.disabled = false;
      return;
    }
    month.classList.add(err);
    recYMDSave.disabled = true;
  });
  day.addEventListener(`input`, function (event) {
    if (day.value >= 1 && day.value <= 31) {
      day.classList.remove(err);
      recYMDSave.disabled = false;
      return;
    }
    if (year.value == `` && month.value == `` && day.value == ``) {
      year.classList.remove(err);
      month.classList.remove(err);
      day.classList.remove(err);
      recYMDSave.disabled = false;
      return;
    }
    if (month.value == `` && day.value == ``) {
      month.classList.remove(err);
      day.classList.remove(err);
      recYMDSave.disabled = false;
      return;
    }
    if (day.value == ``) {
      day.classList.remove(err);
      recYMDSave.disabled = false;
      return;
    }
    day.classList.add(err);
    recYMDSave.disabled = true;
  });

  // releasers
  const releasers = document.getElementById(`recordReleasers`);
  const releasersMax = document.getElementById(`recordReleasersMax`);
  releasers.addEventListener(`input`, function (event) {
    // enforce text input
    releasers.value = releasers.value.toUpperCase();
    releasers.value = releasers.value.replace(/[^a-zA-Z0-9-+& ]/g, "");
    if (releasers.value == ``) {
      releasers.classList.add(err);
      return;
    }
    releasers.classList.remove(err);
    // enforce max releasers
    const count = releasers.value.split(`+`).length;
    const maximum = 2;
    if (count > maximum) {
      releasers.classList.add(err);
      releasersMax.classList.add(dang);
    } else {
      releasers.classList.remove(err);
      releasersMax.classList.remove(dang);
    }

    //     	// hyphen to underscore
    // re := regexp.MustCompile(`\-`)
    // s = re.ReplaceAllString(s, "_")
    // // multiple groups get separated with asterisk
    // re = regexp.MustCompile(`\, `)
    // s = re.ReplaceAllString(s, "*")
    // // any & characters need replacement due to HTML escaping
    // re = regexp.MustCompile(` \& `)
    // s = re.ReplaceAllString(s, " ampersand ")
    // // numbers receive a leading hyphen
    // re = regexp.MustCompile(` ([0-9])`)
    // s = re.ReplaceAllString(s, "-$1")
    // // delete all other characters
    // const deleteAllExcept = `[^A-Za-z0-9 \-\+\.\_\*]`
    // re = regexp.MustCompile(deleteAllExcept)
    // s = re.ReplaceAllString(s, "")
    // // trim whitespace and replace any space separators with hyphens
    // s = strings.TrimSpace(strings.ToLower(s))
    // re = regexp.MustCompile(` `)
    // s = re.ReplaceAllString(s, "-")
  });

  // people
  function parseName(name) {
    let str = name;
    str = str.toLowerCase();
    str = str.replace(/[^A-Za-z0-9À-ÖØ-öø-ÿ\s,]/g, "");
    return str;
  }
  const artist = document.getElementById(`recordArtist`);
  const coder = document.getElementById(`recordCoder`);
  const music = document.getElementById(`recordMusic`);
  const writer = document.getElementById(`recordWriter`);
  artist.addEventListener(`input`, function (event) {
    artist.value = parseName(artist.value);
  });
  coder.addEventListener(`input`, function (event) {
    coder.value = parseName(coder.value);
  });
  music.addEventListener(`input`, function (event) {
    music.value = parseName(music.value);
  });
  writer.addEventListener(`input`, function (event) {
    writer.value = parseName(writer.value);
  });

  // demozoo copy and paste
  const dz = document.getElementById(`recordDemozoo`);
  dz.addEventListener(`paste`, function (event) {
    // delay execution to allow the paste action to complete
    setTimeout(() => {
      try {
        const urlObj = new URL(`${dz.value}`);
        if (urlObj.hostname != `demozoo.org`) {
          dz.classList.add(err);
          return;
        }
      } catch (error) {
        // do nothing, incase a partial URL was pasted
        return;
      }
      try {
        // https://demozoo.org/productions/332978/
        // https://demozoo.org/graphics/332980/
        const urlObj = new URL(`${dz.value}`);
        const pathname = urlObj.pathname;
        const split = pathname.split(`/`).filter(Boolean);
        if (split.length != 2) {
          dz.classList.add(err);
          return;
        }
        const type = split[0];
        switch (type) {
          case `productions`:
            dz.value = split[1];
            break;
          case `graphics`:
            dz.value = split[1];
            break;
          default:
            dz.classList.add(err);
            return;
        }
        dz.classList.remove(err);
      } catch (error) {
        // if a URL was pasted, but it's not a prod.php URL
        dz.classList.add(err);
      }
    }, 0);
  });
  // demozoo input
  dz.addEventListener(`input`, function (event) {
    let id = Number(dz.value); // remove leading zeros
    if (isNaN(id)) {
      dz.classList.add(err);
      return;
    }
    if (id < 0 || id > 999999) {
      dz.classList.add(err);
      return;
    }
    if (id == 0) {
      dz.value = ``;
    }
    dz.classList.remove(err);
  });
  // pouet copy and paste
  const pouet = document.getElementById(`recordPouet`);
  pouet.addEventListener(`paste`, function (event) {
    // delay execution to allow the paste action to complete
    setTimeout(() => {
      try {
        const urlObj = new URL(`${pouet.value}`);
        console.log(urlObj.hostname);
        if (urlObj.hostname != `www.pouet.net`) {
          pouet.classList.add(err);
          return;
        }
      } catch (error) {
        // do nothing, incase a partial URL was pasted
        return;
      }
      try {
        // https://www.pouet.net/prod.php?which=123
        const urlObj = new URL(`${pouet.value}`);
        const pathname = urlObj.pathname;
        if (pathname != `/prod.php`) {
          pouet.classList.add(err);
          return;
        }
        const params = new URLSearchParams(urlObj.search);
        const prod = params.get("which");
        pouet.value = prod;
        pouet.classList.remove(err);
      } catch (error) {
        // if a URL was pasted, but it's not a prod.php URL
        pouet.classList.add(err);
      }
    }, 0);
  });
  // pouet input
  pouet.addEventListener(`input`, function (event) {
    let id = Number(pouet.value); // remove leading zeros
    if (isNaN(id)) {
      pouet.classList.add(err);
      return;
    }
    if (id < 0 || id > 199999) {
      pouet.classList.add(err);
      return;
    }
    if (id == 0) {
      pouet.value = ``;
    }
    pouet.classList.remove(err);
  });
  // 16colors
  const sixteen = document.getElementById(`record16colors`);
  sixteen.addEventListener(`paste`, function (event) {
    // delay execution to allow the paste action to complete
    setTimeout(() => {
      try {
        const urlObj = new URL(`${sixteen.value}`);
        if (urlObj.hostname != `16colo.rs`) {
          sixteen.classList.add(err);
          return;
        }
        const pathname = urlObj.pathname;
        sixteen.value = pathname;
        sixteen.classList.remove(err);
      } catch (error) {
        // do nothing, incase a partial URL was pasted
      }
    }, 0);
  });
  sixteen.addEventListener(`input`, function (event) {
    if (sixteen.value == ``) {
      sixteen.classList.remove(err);
      return;
    }
  });
  // github
  const gh = document.getElementById(`recordGitHub`);
  gh.addEventListener(`paste`, function (event) {
    // delay execution to allow the paste action to complete
    setTimeout(() => {
      try {
        const urlObj = new URL(`${gh.value}`);
        if (urlObj.hostname != `github.com`) {
          gh.classList.add(err);
          return;
        }
        const pathname = urlObj.pathname;
        gh.value = pathname;
        gh.classList.remove(err);
      } catch (error) {
        // do nothing, incase a partial URL was pasted
      }
    }, 0);
  });
  gh.addEventListener(`input`, function (event) {
    if (gh.value == ``) {
      gh.classList.remove(err);
      return;
    }
  });
  // youtube
  const yt = document.getElementById(`recordYouTube`);
  yt.addEventListener(`paste`, function (event) {
    // delay execution to allow the paste action to complete
    setTimeout(() => {
      try {
        const urlObj = new URL(`${yt.value}`);
        if (
          urlObj.hostname != `youtube.com` &&
          urlObj.hostname != `www.youtube.com`
        ) {
          yt.classList.add(err);
          return;
        }
        const params = new URLSearchParams(urlObj.search);
        const videoId = params.get("v");
        yt.value = videoId;
        gh.classList.remove(err);
      } catch (error) {
        // do nothing, incase an ID was pasted
      }
    }, 0);
  });
  yt.addEventListener(`input`, function (event) {
    setTimeout(() => {
      if (yt.value == ``) {
        yt.classList.remove(err);
        return;
      }
      const re = new RegExp(`^[a-zA-Z0-9_-]{11}$`);
      if (re.test(yt.value)) {
        yt.classList.remove(err);
        return;
      }
      yt.classList.add(err);
    }, 0);
  });

  // filename, support any characters except empty and all whitespace
  const filename = document.getElementById(`recordFilename`);
  filename.addEventListener(`input`, function () {
    filename.value = filename.value.trimStart();
    if (filename.value == ``) {
      filename.classList.add(err);
      return;
    }
    filename.classList.remove(err);
  });

  // Modify the file metadata, Platform and Tag
  const platform = document.getElementById(`recordPlatform`);
  const tag = document.getElementById(`recordTag`);

  document.getElementById(`recTagsReset`).addEventListener(`click`, function () {
    const ogp = document.getElementById(`recOSOg`).value;
    const ogt = document.getElementById(`recTagOg`).value;
    platform.value = ogp;
    tag.value = ogt;
  });



  const releaserL = document.getElementById(`recordReleasersLabel`);
  const titleL = document.getElementById(`recordTitleLabel`);

  // special handler for magazine tag
  tag.addEventListener(`change`, function (event) {
    setTimeout(() => {
      if (tag.value == `magazine`) {
        magazineTag();
        return;
      }
      titleTag();
    }, 0);
  });
  function magazineTag() {
    releaserL.textContent = `Magazine`;
    titleL.textContent = `Issue`;
  }
  function titleTag() {
    releaserL.textContent = `Releasers`;
    titleL.textContent = `Title`;
  }

  // platform and tag shortcut buttons
  document
    .getElementById(`recordDosText`)
    .addEventListener(`click`, function (event) {
      platform.value = `text`;
      tag.value = ``;
      titleTag();
    });
  document
    .getElementById(`recordAmigaText`)
    .addEventListener(`click`, function (event) {
      platform.value = `textamiga`;
      tag.value = ``;
      titleTag();
    });
  document
    .getElementById(`recordProof`)
    .addEventListener(`click`, function (event) {
      platform.value = `image`;
      tag.value = `releaseproof`;
      titleTag();
    });
  document
    .getElementById(`recordDostro`)
    .addEventListener(`click`, function (event) {
      platform.value = `dos`;
      tag.value = `releaseadvert`;
      titleTag();
    });
  document
    .getElementById(`recordWintro`)
    .addEventListener(`click`, function (event) {
      platform.value = `windows`;
      tag.value = `releaseadvert`;
      titleTag();
    });
  document
    .getElementById(`recordBBStro`)
    .addEventListener(`click`, function (event) {
      platform.value = `dos`;
      tag.value = `bbs`;
      titleTag();
    });
  document
    .getElementById(`recordBBSAnsi`)
    .addEventListener(`click`, function (event) {
      platform.value = `ansi`;
      tag.value = `bbs`;
      titleTag();
    });
  document
    .getElementById(`recordTextMag`)
    .addEventListener(`click`, function (event) {
      platform.value = `text`;
      tag.value = `magazine`;
      magazineTag();
    });
  document
    .getElementById(`recordDosMag`)
    .addEventListener(`click`, function (event) {
      platform.value = `dos`;
      tag.value = `magazine`;
      magazineTag();
    });

  const reset = document.getElementById(`recordReset`);
  // reset all input elements
  reset.addEventListener(`click`, function (event) {
    // delay execution to allow the reset action to complete
    setTimeout(() => {
      if (online.checked != true) {
        onlineL.classList.add(dang);
      } else {
        onlineL.classList.remove(dang);
      }
      releasersMax.classList.remove(dang);
      if (tag.value == `magazine`) {
        magazineTag();
      } else {
        titleTag();
      }
      const inputs = document.querySelectorAll(`input`);
      inputs.forEach((input) => {
        input.classList.remove(err);
      }, 0);
    });
  });
})();
