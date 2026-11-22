/**
 * @file layout-releaser.mjs
 * In-page filter for releaser, BBS, FTP, and magazine directory listings.
 */

/**
 * Initializes the in-page quick filter on releaser directory pages.
 */
export default function releaserFilter() {
  const init = () => {
    const input = document.getElementById('releaser-filter');
    const clearBtn = document.getElementById('releaser-filter-clear');
    const emptyClearBtn = document.getElementById('releaser-empty-clear');
    const countEl = document.getElementById('releaser-count');
    const emptyEl = document.getElementById('releaser-empty');
    const emptyQueryEl = document.getElementById('releaser-empty-query');
    const items = Array.from(document.querySelectorAll('.releaser-item'));
    if (!input || items.length === 0) {
      return;
    }

    const total = items.length;
    const initialCountLabel = countEl ? countEl.textContent : `${total} items`;

    for (const item of items) {
      item._text = item.textContent.toLowerCase();
    }

    const applyFilter = () => {
      const raw = input.value.trim();
      const query = raw.toLowerCase();
      const tokens = query.split(/\s+/).filter(Boolean);
      let visible = 0;

      for (const item of items) {
        const match =
          tokens.length === 0 || tokens.every((tok) => item._text.includes(tok));
        item.classList.toggle('d-none', !match);
        if (match) {
          visible += 1;
        }
      }

      if (clearBtn) {
        clearBtn.disabled = input.value === '';
      }
      if (countEl) {
        countEl.textContent =
          query === ''
            ? initialCountLabel
            : `Showing ${visible} of ${total}`;
      }
      if (emptyEl) {
        emptyEl.classList.toggle('d-none', query === '' || visible > 0);
        if (emptyQueryEl) {
          emptyQueryEl.textContent = raw;
        }
      }
    };

    const resetFilter = () => {
      input.value = '';
      applyFilter();
      input.focus();
    };

    input.addEventListener('input', applyFilter);
    input.addEventListener('keydown', (e) => {
      if (e.key === 'Escape' && input.value !== '') {
        e.preventDefault();
        resetFilter();
      }
    });
    if (clearBtn) {
      clearBtn.addEventListener('click', resetFilter);
    }
    if (emptyClearBtn) {
      emptyClearBtn.addEventListener('click', resetFilter);
    }
    window.addEventListener('pageshow', applyFilter);
  };

  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', init);
    return;
  }
  init();
}
