/* eslint-disable no-undef */
// uploader-htmx.mjs

export default load;

export function load() {
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
