// layout.js
(() => {
  "use strict";
  const tooltipTriggerList = document.querySelectorAll(
    '[data-bs-toggle="tooltip"]'
  );
  if (tooltipTriggerList === null) {
    throw new Error("Tooltip trigger list not found");
  }
  // eslint-disable-next-line no-unused-vars
  const tooltipList = [...tooltipTriggerList].map(
    (tooltipTriggerEl) => new bootstrap.Tooltip(tooltipTriggerEl)
  );
})();
