/**
 * @module layout-anim
 * Swaps static thumbnails with animated WebP previews on hover/focus for ANSImation cards.
 */
export default function animPreviewHover() {
  const cards = document.querySelectorAll('.card[data-anim-src]');
  if (cards.length === 0) return;

  const hoverDelayMs = 150;

  for (const card of cards) {
    const animSrc = card.dataset.animSrc;
    const img = card.querySelector('img.card-img-bottom, img.card-img-top');
    if (!animSrc || !img) continue;

    const picture = img.closest('picture');
    const source = picture ? picture.querySelector('source') : null;
    const badge = card.querySelector('.anim-badge');
    const origSrc = img.getAttribute('src') ?? '';
    const origSrcset = source ? (source.getAttribute('srcset') ?? '') : '';

    let timer = null;
    let playing = false;

    const play = () => {
      if (playing) return;
      playing = true;
      const rect = img.getBoundingClientRect();
      if (rect.height > 0) {
        img.style.height = `${rect.height}px`;
        img.style.objectFit = 'contain';
        img.style.backgroundColor = '#000';
      }
      if (source) source.setAttribute('srcset', animSrc);
      img.setAttribute('src', animSrc);
      if (badge) badge.classList.add('d-none');
    };

    const stop = () => {
      if (timer !== null) {
        clearTimeout(timer);
        timer = null;
      }
      if (!playing) return;
      playing = false;
      if (source) source.setAttribute('srcset', origSrcset);
      img.setAttribute('src', origSrc);
      img.style.height = '';
      img.style.objectFit = '';
      img.style.backgroundColor = '';
      if (badge) badge.classList.remove('d-none');
    };

    card.addEventListener('mouseenter', () => {
      if (timer !== null) clearTimeout(timer);
      timer = setTimeout(play, hoverDelayMs);
    });

    card.addEventListener('mouseleave', stop);
    card.addEventListener('focusin', play);
    card.addEventListener('focusout', (evt) => {
      if (!card.contains(evt.relatedTarget)) {
        stop();
      }
    });
  }
}
