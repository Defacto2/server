// layout-carousel.js
//
// This is test, embedded JS script used in the layout.html template.
// It is not intended to be minified or bundled with the rest of the layout JS.
//
(() => {
  const myCarouselElement = document.querySelector("#carouselDf2Artpacks");
  if (myCarouselElement === null) {
    throw new Error("Carousel element not found");
  }
  const twoSeconds = 2000;
  // eslint-disable-next-line no-unused-vars
  const carousel = new bootstrap.Carousel(myCarouselElement, {
    interval: twoSeconds,
    touch: false,
  });
})();
