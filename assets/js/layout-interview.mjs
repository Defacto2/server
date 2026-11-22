/**
 * @file layout-interview.mjs
 * In-page filter and releaser selector for the /interview directory.
 */

/**
 * Initializes the in-page quick filter and group filtering on interview pages.
 */
export default function interviewFilter() {
  const init = () => {
    const input = document.getElementById('interview-filter');
    const clearBtn = document.getElementById('interview-filter-clear');
    const emptyClearBtn = document.getElementById('interview-empty-clear');
    const countEl = document.getElementById('interview-count');
    const emptyEl = document.getElementById('interview-empty');
    const emptyQueryEl = document.getElementById('interview-empty-query');
    const groupBtns = Array.from(
      document.querySelectorAll('.interview-group-btn')
    );
    const items = Array.from(document.querySelectorAll('.interview-item'));
    if (!input || items.length === 0) {
      return;
    }

    const total = items.length;
    let activeGroup = '';

    const applyFilters = () => {
      const raw = input.value.trim();
      const query = raw.toLowerCase();
      let visible = 0;

      for (const item of items) {
        const groupMatch =
          activeGroup === '' || item.dataset.group === activeGroup;
        const textMatch =
          query === '' || (item.dataset.search || '').includes(query);
        const match = groupMatch && textMatch;
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
          query === '' && activeGroup === ''
            ? `${total} ${total === 1 ? 'interview' : 'interviews'}`
            : `Showing ${visible} of ${total}`;
      }
      if (emptyEl) {
        emptyEl.classList.toggle('d-none', visible > 0);
        if (emptyQueryEl) {
          emptyQueryEl.textContent = raw || activeGroup;
        }
      }
    };

    const setGroup = (group) => {
      activeGroup = group;
      for (const btn of groupBtns) {
        const isSelected = (btn.dataset.group || '') === group;
        btn.classList.toggle('active', isSelected);
        btn.setAttribute('aria-pressed', String(isSelected));
      }
      applyFilters();
    };

    const resetAll = () => {
      input.value = '';
      setGroup('');
      input.focus();
    };

    for (const btn of groupBtns) {
      btn.addEventListener('click', () => setGroup(btn.dataset.group || ''));
    }

    input.addEventListener('input', applyFilters);
    input.addEventListener('keydown', (e) => {
      if (e.key === 'Escape' && input.value !== '') {
        e.preventDefault();
        input.value = '';
        applyFilters();
      }
    });
    if (clearBtn) {
      clearBtn.addEventListener('click', () => {
        input.value = '';
        applyFilters();
        input.focus();
      });
    }
    if (emptyClearBtn) {
      emptyClearBtn.addEventListener('click', resetAll);
    }
    window.addEventListener('pageshow', applyFilters);
    applyFilters();
  };

  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', init);
    return;
  }
  init();
}
