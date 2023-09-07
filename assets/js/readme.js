(() => {
  "use strict";
  const hide = `d-none`;
  const preLatin1 = document.getElementById("readmeLatin1");
  const pre437 = document.getElementById("readmeCP437");
  document.getElementById("topazFont").addEventListener("click", function () {
    preLatin1.classList.remove(hide);
    pre437.classList.add(hide);
    console.log(`a`)
  });
  document.getElementById("vgaFont").addEventListener("click", function () {
    preLatin1.classList.add(hide);
    pre437.classList.remove(hide);
    console.log(`b`)
  });
})();
