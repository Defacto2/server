/**
 * jsdos.js
 * JS-DOS 6.22 configuration and initialization.
 */
const canvas = document.getElementById("jsdos6");
const ctx = canvas.getContext("2d");
ctx.font = "18px serif";
ctx.fillText("Not loading?", 10, 50);
ctx.fillText("Try the Console log for errors.", 10, 70);
const stopButton = document.getElementById("jsdosStop");
stopButton.addEventListener("click", () => {
  try {
    ci.exit();
    stopButton.disabled = true;
  } catch (error) {
    console.log("ci.exit() error: ", error);
  }
});
document.getElementById(`jsdosFullscreen`).addEventListener("click", () => {
  canvas.requestFullscreen();
});
document
  .getElementById(`jsdosScreenshot`)
  .addEventListener("click", screenshot);
function screenshot() {
  console.log("screenshot: for canvas of {{$filename}}");
  let dataURL = canvas.toDataURL("image/png");
  let a = document.createElement("a");
  a.href = dataURL;
  a.download = "{{$filename}}.png";
  document.body.appendChild(a);
  a.click();
  document.body.removeChild(a);
}
function startBinary(options) {
  Dos(canvas, options).ready((fs, main) => {
    // create a dosbox configuration file within the virtual file system.
    fs.createFile("dosbox.conf", "{{$config}}");
    // fetch the artifact file download.
    fetch("/d/{{$download}}")
      .then((response) => response.arrayBuffer())
      .then((arrayBuffer) => {
        // recreate the artifact file download within the virtual file system.
        fs.createFile("{{$filename}}", arrayBuffer);
        // start the dosbox with the configuration and the artifact file.
        main(["-conf", "dosbox.conf", "-c", "{{$filename}}"]).then((ci) => {
          window.ci = ci;
          console.log(
            `width: ${ci.dos.canvas.width}, height: ${ci.dos.canvas.width}`
          );
        });
      });
  });
}
function startZip(options) {
  Dos(canvas, options).ready((fs, main) => {
    // create a dosbox configuration file within the virtual file system.
    fs.createFile("dosbox.conf", "{{$config}}");
    console.log("{{$config}}");
    // fetch the artifact file download and if it is a zip file, extract it to the virtual file system.
    fs.extract("/d/{{$download}}").then(() => {
      // start the dosbox with the configuration and the extracted artifact file.
      main(["-conf", "dosbox.conf", "-c", "{{$runProgram}}"]).then((ci) => {
        window.ci = ci;
        console.log(
          `width: ${ci.dos.canvas.width}, height: ${ci.dos.canvas.width}`
        );
      });
    });
  });
}
const jsdos = document.getElementById("jsdosRunLink");
jsdos.addEventListener("click", function () {
  this.style.pointerEvents = "none";
  this.textContent = "Running";
  DosBoxConfig = {
    wdosboxUrl: "/js/wdosbox.js",
    cycles: "auto", // int value, "max" or "auto"
    autolock: false,
  };
  {
    {
      $runJS | safeJS;
    }
  }
});
const jsQuit = document.getElementById("jsdosCloser");
jsQuit.addEventListener("click", () => {
  location.reload();
});
