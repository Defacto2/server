/* eslint-disable no-undef */
// uploader-htmx.mjs

export default load;

export function load() {
  // search
  document.body.addEventListener("htmx:afterRequest", function (evt) {
    console.log("htmx:afterRequest event fired");
    if (evt.detail.elt === null || evt.detail.elt.id !== "releaserSearch") {
      return;
    }
    const alert = document.getElementById("htmx-alert");
    if (typeof alert === "undefined" || alert === null) {
      return;
    }
    if (evt.detail.successful) {
      alert.setAttribute("hidden", "true");
      alert.innerText = "";
      return;
    }
    if (evt.detail.failed && evt.detail.xhr) {
      const xhr = evt.detail.xhr;
      alert.innerText = `Something on the server is not working, ${xhr.status} status: ${xhr.responseText}.`;
      alert.removeAttribute("hidden");
      return;
    }
    alert.innerText =
      "Something with the browser is not working, please refresh the page.";
    alert.removeAttribute("hidden");
  });
  const clear = document.getElementById("htmx-clear");
  if (typeof alert !== "undefined" && alert !== null) {
    clear.addEventListener("click", function () {
      const input = document.getElementById("releaserSearch");
      input.value = "";
      input.focus();
      const alert = document.getElementById("htmx-alert");
      alert.setAttribute("hidden", "true");
      const indicator = document.getElementById("indicator");
      indicator.style.opacity = 0;
      document.getElementById("search-releaser-results").innerHTML = "";
    });
  }
// uploader progress
  console.log(`htmx: load event loaded`);
  document.body.addEventListener("htmx:load", function () {
    console.log("htmx:load event fired");
    //htmx.logAll();
    htmx.on("#uploader-intro-form", "htmx:xhr:progress", function (evt) {
      console.log("htmx:xhr:progress event fired");
      htmx
        .find("#uploader-intro-progress")
        .setAttribute("value", (evt.detail.loaded / evt.detail.total) * 100);
    });
  });
}
