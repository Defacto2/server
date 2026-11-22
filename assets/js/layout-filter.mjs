/**
 * @file layout-filter.mjs
 * In-page quick filter for artifact cards on paginated listing pages.
 */

/**
 * Initializes the in-page quick filter on artifact listing pages.
 */
export default function artifactFilter() {
  const init = () => {
    const input = document.getElementById('artifact-filter');
    const clearBtn = document.getElementById('artifact-filter-clear');
    const emptyClearBtn = document.getElementById('artifact-empty-clear');
    const countEl = document.getElementById('artifact-count');
    const emptyEl = document.getElementById('artifact-empty');
    const emptyQueryEl = document.getElementById('artifact-empty-query');
    const unkHeader = document.getElementById('artifact-unknown-years-header');
    const items = Array.from(document.querySelectorAll('.artifact-item'));
    if (!input || items.length < 3) {
      return;
    }

    const total = items.length;

    for (const item of items) {
      item._text = item.textContent.toLowerCase();
    }

    const applyFilter = () => {
      const raw = input.value.trim();
      const query = raw.toLowerCase();
      const tokens = query.split(/\s+/).filter(Boolean);
      let visible = 0;
      let unkVisible = 0;

      for (const item of items) {
        const match =
          tokens.length === 0 || tokens.every((tok) => item._text.includes(tok));
        item.classList.toggle('d-none', !match);
        if (match) {
          visible += 1;
          if (
            unkHeader &&
            unkHeader.compareDocumentPosition(item) &
              Node.DOCUMENT_POSITION_FOLLOWING
          ) {
            unkVisible += 1;
          }
        }
      }

      if (unkHeader) {
        unkHeader.classList.toggle('d-none', unkVisible === 0);
      }

      if (clearBtn) {
        clearBtn.disabled = input.value === '';
      }
      if (countEl) {
        countEl.textContent =
          query === ''
            ? `${total} ${total === 1 ? 'artifact' : 'artifacts'}`
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
