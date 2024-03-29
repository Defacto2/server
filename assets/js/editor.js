(() => {
  "use strict";

  const dang = `text-danger`;
  const err = `is-invalid`;
  const ok = `is-valid`;
  // const fok = `valid-feedback`;
  // const ferr = `invalid-feedback`;
  const hide = `d-none`;
  const header = {
    "Content-type": "application/json; charset=UTF-8",
  };
  const saveErr = `server could not save the change`;

  // The table record id and key value, used for all fetch requests
  // It is also used to confirm the existence of the editor modal
  const id = document.getElementById(`recordID`);
  if (id == null) {
    console.info(
      `the editor modal is not open so the editor script is not needed`
    );
    return;
  }

  // Modify the metadata, Artifact is online and public
  const elm0 = document.getElementById(`recordOnline`);
  const label0 = document.getElementById(`recordOnlineLabel`);
  if (elm0 == null) {
    console.info(`the online checkbox is not present`);
  } else if (label0 == null) {
    console.info(`the online checkbox label is not present`);
  } else {
    if (elm0.checked != true) {
      label0.classList.add(dang);
    }
    elm0.addEventListener(`change`, () => {
      let path = `/editor/online/false`;
      if (elm0.checked == true) {
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
          if (elm0.checked == true) {
            label0.classList.remove(dang);
          } else {
            label0.classList.add(dang);
          }
          return response.json();
        })
        .catch((error) => {
          elm0.checked = !elm0.checked;
          console.info(
            `the artifact online status could not be saved: ${error.message}`
          );
        });
    });
  }

  // Modify the metadata, Title
  const elm1 = document.getElementById(`recordTitle`);
  if (elm1 == null) {
    console.info(`the record title is not present`);
  } else {
    elm1.addEventListener(`input`, () => {
      const infoErr = document.getElementById(`recordTitleErr`);
      if (infoErr == null) {
        console.info(`the record title error is not present`);
        return;
      }
      const label = document.getElementById(`recordTitleOG`);
      if (label == null) {
        console.info(`the record title original label is not present`);
        return;
      }
      const text = document.getElementById(`recordTitleOGValue`).textContent;
      if (text == null) {
        console.info(`the record title original value is not present`);
        return;
      }
      if (elm1.value != text && elm1.value.length > 0) {
        label.classList.remove(hide);
      } else {
        label.classList.add(hide);
      }
      elm1.classList.remove(err);
      infoErr.classList.add(hide);
      fetch("/editor/title", {
        method: "POST",
        body: JSON.stringify({
          id: parseInt(id.value),
          value: elm1.value,
        }),
        headers: header,
      })
        .then((response) => {
          if (!response.ok) {
            throw new Error(saveErr);
          }
          infoErr.classList.add(hide);
          elm1.classList.remove(err);
          elm1.classList.add(ok);
          return response.json();
        })
        .catch((error) => {
          infoErr.classList.remove(hide);
          elm1.classList.add(err);
          elm1.value = text;
          console.info(`the title could not be saved: ${error.message}`);
        });
    });
  }

  // Modify the Releasers and enforce a maximum of 2
  const elm2 = document.getElementById(`recordReleasers`);
  if (elm2 == null) {
    console.info(`the record releasers is not present`);
  } else {
    elm2.addEventListener(`input`, function () {
      // enforce text input
      elm2.value = elm2.value.toUpperCase();
      elm2.value = elm2.value.replace(/[^a-zA-Z0-9-+& ]/g, "");
      if (elm2.value == ``) {
        elm2.classList.add(err);
        return;
      }
      elm2.classList.remove(err);
      // enforce max releasers
      const max = document.getElementById(`recordReleasersMax`);
      if (max == null) {
        console.info(`the record releasers max is not present`);
        return;
      }
      const count = elm2.value.split(`+`).length;
      const maximum = 2;
      if (count > maximum) {
        elm2.classList.add(err);
        max.classList.add(dang);
        return;
      }
      elm2.classList.remove(err);
      max.classList.remove(dang);
      fetch("/editor/releasers", {
        method: "POST",
        body: JSON.stringify({
          id: parseInt(id.value),
          value: elm2.value,
        }),
        headers: header,
      })
        .then((response) => {
          if (!response.ok) {
            throw new Error(saveErr);
          }
          // infoErr.classList.add(hide);
          // elm1.classList.remove(err);
          // elm1.classList.add(ok);
          return response.json();
        })
        .catch((error) => {
          // infoErr.classList.remove(hide);
          // elm1.classList.add(err);
          // elm1.value = text;
          console.info(`the title could not be saved: ${error.message}`);
        });
    });
  }

  // Modify the metadata, use last modification button
  const elm3 = document.getElementById(`recordLMBtn`);
  if (elm3 == null) {
    console.info(`the last modification button is not present`);
  } else {
    elm3.addEventListener(`click`, function () {
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
  }

  // Modify the metadata, Year, month, day of release, save button
  const saveYMD = document.getElementById(`recordYMDSave`);
  if (saveYMD == null) {
    console.info(`the record ymd save button is not present`);
  } else {
    saveYMD.addEventListener(`click`, function () {
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
          saveYMD.classList.remove(dang);
          return response.json();
        })
        .catch((error) => {
          saveYMD.classList.add(dang);
          console.info(`the ymd could not be saved: ${error.message}`);
        });
    });
  }

  // Modify the metadata, Year, month, day of release, reset button
  const elm4 = document.getElementById(`recordYMDReset`);
  if (elm4 == null || saveYMD == null) {
    console.info(`the record ymd reset button is not present or usable`);
  } else {
    elm4.addEventListener(`click`, function () {
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
      saveYMD.disabled = false;
    });
  }

  // Modify the metadata, Year, month, day of release
  const year = document.getElementById(`recordYear`);
  if (year == null || saveYMD == null) {
    console.info(`the record year is not present or usable`);
  } else {
    year.addEventListener(`input`, function () {
      if (year.value >= 1980 && year.value <= 2023) {
        year.classList.remove(err);
        saveYMD.disabled = false;
        return;
      }
      if (year == null || month == null || day == null) {
        console.info(`the record year, month, or day is not present`);
        return;
      }
      // year can only be empty when month and day are empty
      if (year.value == `` && month.value == `` && day.value == ``) {
        year.classList.remove(err);
        month.classList.remove(err);
        day.classList.remove(err);
        saveYMD.disabled = false;
        return;
      }
      year.classList.add(err);
      saveYMD.disabled = true;
    });
  }
  const month = document.getElementById(`recordMonth`);
  if (month == null || saveYMD == null) {
    console.info(`the record month is not present or usable`);
  } else {
    month.addEventListener(`input`, function () {
      if (month.value >= 1 && month.value <= 12) {
        month.classList.remove(err);
        saveYMD.disabled = false;
        return;
      }
      if (year == null || month == null || day == null) {
        console.info(`the record year, month, or day is not present`);
        return;
      }
      if (year.value == `` && month.value == `` && day.value == ``) {
        year.classList.remove(err);
        month.classList.remove(err);
        day.classList.remove(err);
        saveYMD.disabled = false;
        return;
      }
      // month can only be empty when day is empty
      if (month.value == `` && day.value == ``) {
        month.classList.remove(err);
        if (day != null) {
          day.classList.remove(err);
        }
        saveYMD.disabled = false;
        return;
      }
      month.classList.add(err);
      saveYMD.disabled = true;
    });
  }
  const day = document.getElementById(`recordDay`);
  if (day == null || saveYMD == null) {
    console.info(`the record day is not present or usable`);
  } else {
    day.addEventListener(`input`, function () {
      if (day.value >= 1 && day.value <= 31) {
        day.classList.remove(err);
        saveYMD.disabled = false;
        return;
      }
      if (year == null || month == null || day == null) {
        console.info(`the record year, month, or day is not present`);
        return;
      }
      if (year.value == `` && month.value == `` && day.value == ``) {
        year.classList.remove(err);
        month.classList.remove(err);
        day.classList.remove(err);
        saveYMD.disabled = false;
        return;
      }
      if (month.value == `` && day.value == ``) {
        month.classList.remove(err);
        day.classList.remove(err);
        saveYMD.disabled = false;
        return;
      }
      if (day.value == ``) {
        day.classList.remove(err);
        saveYMD.disabled = false;
        return;
      }
      day.classList.add(err);
      saveYMD.disabled = true;
    });
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
  artist.addEventListener(`input`, function () {
    artist.value = parseName(artist.value);
  });
  coder.addEventListener(`input`, function () {
    coder.value = parseName(coder.value);
  });
  music.addEventListener(`input`, function () {
    music.value = parseName(music.value);
  });
  writer.addEventListener(`input`, function () {
    writer.value = parseName(writer.value);
  });

  // demozoo copy and paste
  const dz = document.getElementById(`recordDemozoo`);
  dz.addEventListener(`paste`, function () {
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
  dz.addEventListener(`input`, function () {
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
  pouet.addEventListener(`paste`, function () {
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
  pouet.addEventListener(`input`, function () {
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
  sixteen.addEventListener(`paste`, function () {
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
  sixteen.addEventListener(`input`, function () {
    if (sixteen.value == ``) {
      sixteen.classList.remove(err);
      return;
    }
  });
  // github
  const gh = document.getElementById(`recordGitHub`);
  gh.addEventListener(`paste`, function () {
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
  gh.addEventListener(`input`, function () {
    if (gh.value == ``) {
      gh.classList.remove(err);
      return;
    }
  });
  // youtube
  const yt = document.getElementById(`recordYouTube`);
  yt.addEventListener(`paste`, function () {
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
  yt.addEventListener(`input`, function () {
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

  // Modify the metadata, Platform
  const platform = document.getElementById(`recordPlatform`);
  platform.addEventListener(`change`, function (event) {
    platform.classList.remove(err);
    const value = event.target.value;
    if (value.length == 0) {
      platform.classList.add(err);
      return;
    }
    platformChange(value);
  });

  function platformChange(value) {
    fetch("/editor/platform", {
      method: "POST",
      body: JSON.stringify({
        id: parseInt(id.value),
        value: value,
      }),
      headers: header,
    })
      .then((response) => {
        if (!response.ok) {
          throw new Error(saveErr);
        }
        platform.classList.remove(err);
        platform.classList.add(ok);
        return response.json();
      })
      .catch((error) => {
        platform.classList.remove(ok);
        platform.classList.add(dang);
        console.log(error.message);
      });
    platformTagInfo(value, tag.value);
  }

  // Modify the metadata, Tag
  const tag = document.getElementById(`recordTag`);
  tag.addEventListener(`change`, function (event) {
    const value = event.target.value;
    tagChange(value);
  });

  function tagChange(value) {
    tag.classList.remove(err);
    platformTagInfo(platform.value, value);
    tagInfo(value);
    if (value.length == 0) {
      tag.classList.add(err);
      tag.value = ``; // incase a hyperlink was clicked
      return;
    }
    fetch("/editor/tag", {
      method: "POST",
      body: JSON.stringify({
        id: parseInt(id.value),
        value: value,
      }),
      headers: header,
    })
      .then((response) => {
        if (!response.ok) {
          throw new Error(saveErr);
        }
        tag.classList.remove(err);
        tag.classList.add(ok);
        return response.json();
      })
      .catch((error) => {
        tag.classList.remove(ok);
        tag.classList.add(dang);
        document.getElementById(`tagInfo`).textContent = ``;
        console.log(error.message);
      });
  }

  function platformTagInfo(platform, tag) {
    fetch("/editor/platform+tag", {
      method: "POST",
      body: JSON.stringify({
        platform: platform,
        tag: tag,
      }),
      headers: header,
    })
      .then((response) => response.text())
      .then((text) => {
        document.getElementById(`platformTagInfo`).textContent = text;
      });
  }

  function tagInfo(tag) {
    fetch("/editor/tag/info", {
      method: "POST",
      body: JSON.stringify({
        tag: tag,
      }),
      headers: header,
    })
      .then((response) => response.text())
      .then((text) => {
        document.getElementById(`tagInfo`).textContent = text;
      });
  }

  // Modify the metadata, Reset Platform and Tag
  document
    .getElementById(`recTagsReset`)
    .addEventListener(`click`, function () {
      const ogp = document.getElementById(`recOSOg`).value;
      const ogt = document.getElementById(`recTagOg`).value;
      platform.value = ogp;
      tag.value = ogt;
      platform.classList.remove(err);
      tag.classList.remove(err);
      if (platform.value.length == 0) {
        platform.classList.add(err);
      }
      if (tag.value.length == 0) {
        tag.classList.add(err);
      }
      platformChange(ogp);
      tagChange(ogt);
    });

  const releaserL = document.getElementById(`recordReleasersLabel`);
  const titleL = document.getElementById(`recordTitleLabel`);

  // special handler for magazine tag
  tag.addEventListener(`change`, function () {
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
    .addEventListener(`click`, function () {
      platformChange(`text`);
      tagChange(``);
      titleTag();
      // platform.value = `text`;
      // platform.classList.remove(err);
      // tag.value = ``;
      // tag.classList.add(err);
      // platformTagInfo(`text`);
      // platformChange(`text`);
      // tagChange(``);
      // tagInfo(``);
    });
  document
    .getElementById(`recordAmigaText`)
    .addEventListener(`click`, function () {
      platform.value = `textamiga`;
      platform.classList.remove(err);
      tag.value = ``;
      tag.classList.add(err);
      titleTag();
    });
  document.getElementById(`recordProof`).addEventListener(`click`, function () {
    platform.value = `image`;
    platform.classList.remove(err);
    tag.value = `releaseproof`;
    tag.classList.remove(err);
    titleTag();
  });
  document
    .getElementById(`recordDostro`)
    .addEventListener(`click`, function () {
      platform.value = `dos`;
      platform.classList.remove(err);
      tag.value = `releaseadvert`;
      tag.classList.remove(err);
      titleTag();
    });
  document
    .getElementById(`recordWintro`)
    .addEventListener(`click`, function () {
      platform.value = `windows`;
      platform.classList.remove(err);
      tag.value = `releaseadvert`;
      tag.classList.remove(err);
      titleTag();
    });
  document
    .getElementById(`recordBBStro`)
    .addEventListener(`click`, function () {
      platform.value = `dos`;
      platform.classList.remove(err);
      tag.value = `bbs`;
      tag.classList.remove(err);
      titleTag();
    });
  document
    .getElementById(`recordBBSAnsi`)
    .addEventListener(`click`, function () {
      platform.value = `ansi`;
      platform.classList.remove(err);
      tag.value = `bbs`;
      tag.classList.remove(err);
      titleTag();
    });
  document
    .getElementById(`recordTextMag`)
    .addEventListener(`click`, function () {
      platform.value = `text`;
      tag.value = `magazine`;
      magazineTag();
    });
  document
    .getElementById(`recordDosMag`)
    .addEventListener(`click`, function () {
      platform.value = `dos`;
      platform.classList.remove(err);
      tag.value = `magazine`;
      tag.classList.remove(err);
      magazineTag();
    });

  // TODO: {{brief (index . "platform") (index . "section")}}
  // create an fetch response to get the platform and section as text

  const reset = document.getElementById(`recordReset`);
  if (reset == null) {
    console.info(`the editor reset button is not present`);
  } else {
    // reset all input elements
    reset.addEventListener(`click`, function () {
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
  }
})();
