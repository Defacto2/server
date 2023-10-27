/**
 * Immediately invoked function expression that sets up the functionality for the readme text page.
 * @function
 */
(() => {
  "use strict";
  const hide = `d-none`;
  const preLatin1 = document.getElementById("readmeLatin1");
  const pre437 = document.getElementById("readmeCP437");
  const copyBtn = document.getElementById(`copyReadme`);
  const topaz = document.getElementById(`topazFont`);
  const vga = document.getElementById(`vgaFont`);

  /**
   * Converts a file size in bytes to a human-readable format.
   *
   * @param {number} size - The file size in bytes.
   * @returns {string} A human-readable string representation of the file size.
   */
  function humanFilesize(size = 0) {
    const three = 3,
      round = 100,
      kB = 1000,
      MB = Math.pow(kB, 2),
      GB = Math.pow(kB, three);
    if (size > GB)
      return `${(Math.round((size * round) / GB) / round).toFixed(2)} GB`;
    if (size > MB)
      return `${(Math.round((size * round) / MB) / round).toFixed(1)} MB`;
    if (size > kB)
      return `${(Math.round((size * round) / kB) / round).toFixed()} kB`;
    return `${Math.round(size).toFixed()} bytes`;
  }

  /**
   * Copies the text content of an HTML element to the clipboard.
   * @async
   * @function clipText
   * @param {string} [id=""] - The ID of the HTML element to copy the text from.
   * @throws {Error} Throws an error if the specified element is missing.
   * @returns {Promise<void>} A Promise that resolves when the text has been copied to the clipboard.
   */
  async function clipText(id = ``) {
    const element = document.getElementById(id);
    if (element === null) throw Error(`select text element "${id}" is missing`);
    element.focus(); // select the element to avoid NotAllowedError: Clipboard write is not allowed in this context
    await navigator.clipboard.writeText(`${element.textContent}`).then(
      function () {
        console.log(
          `Copied ${humanFilesize(element.textContent.length)} to the clipboard`
        );
        const button = document.getElementById(`copyReadme`),
          oneSecond = 1000;
        if (button === null) return;
        const save = button.textContent;
        button.textContent = `✓ Copied`;
        window.setTimeout(() => {
          button.textContent = `${save}`;
        }, oneSecond);
      },
      function (err) {
        console.error(`could not save any text to the clipboard: ${err}`);
      }
    );
  }

  /**
   * Event listener for the Topaz font radio button.
   * @function
   */
  topaz.addEventListener("click", function () {
    preLatin1.classList.remove(hide);
    pre437.classList.add(hide);
  });

  /**
   * Event listener for the VGA font radio button.
   * @function
   */
  vga.addEventListener("click", function () {
    preLatin1.classList.add(hide);
    pre437.classList.remove(hide);
  });

  if (typeof navigator.clipboard === `undefined`)
    copyBtn.classList.add(hide);
  else
    /**
     * Event listener for the copy button.
     * @function
     */
    copyBtn.addEventListener(`click`, () => {
      if (topaz.checked)
        clipText(`readmeLatin1`);
      else if (vga.checked)
        clipText(`readmeCP437`);
    });
})();
