/**
 * @file layout-scener.mjs
 * In-page filter for scener and role directory pages (/scener, /artist, /coder, /musician, /writer).
 */

/**
 * Initializes the in-page quick filter on scener directory pages.
 */
export default function scenerFilter() {
  const init = () => {
    const input = document.getElementById('scener-filter');
    const clearBtn = document.getElementById('scener-filter-clear');
    const emptyClearBtn = document.getElementById('scener-empty-clear');
    const countEl = document.getElementById('scener-count');
    const emptyEl = document.getElementById('scener-empty');
    const emptyQueryEl = document.getElementById('scener-empty-query');
    const groups = Array.from(document.querySelectorAll('.scener-group'));
    const jumps = new Map(
      Array.from(document.querySelectorAll('.scener-jump')).map((el) => [
        el.dataset.groupId,
        el,
      ])
    );
    if (!input || groups.length === 0) {
      return;
    }

    const items = Array.from(document.querySelectorAll('.scener-item'));
    const total = items.length;

    const applyFilter = () => {
      const raw = input.value.trim();
      const query = raw.toLowerCase();
      let visible = 0;

      for (const group of groups) {
        const groupItems = group.querySelectorAll('.scener-item');
        let groupVisible = 0;
        for (const item of groupItems) {
          const match =
            query === '' || (item.dataset.scener || '').includes(query);
          item.classList.toggle('d-none', !match);
          if (match) {
            groupVisible += 1;
          }
        }
        visible += groupVisible;
        const groupCountEl = group.querySelector('.scener-group-count');
        if (groupCountEl) {
          groupCountEl.textContent = String(groupVisible);
        }
        group.classList.toggle('d-none', groupVisible === 0);

        const jump = jumps.get(group.id);
        if (jump) {
          jump.classList.toggle('disabled', groupVisible === 0);
          if (groupVisible === 0) {
            jump.setAttribute('aria-disabled', 'true');
          } else {
            jump.removeAttribute('aria-disabled');
          }
        }
      }

      if (clearBtn) {
        clearBtn.disabled = input.value === '';
      }
      if (countEl) {
        countEl.textContent =
          query === ''
            ? `${total} ${total === 1 ? 'scener' : 'sceners'}`
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
