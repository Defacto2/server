/**
 * @file layout-duplicates.mjs
 * In-page filter for the potential duplicate releasers report (/releaser/duplicates, /editor/duplicates).
 */

/**
 * Initializes the in-page tier, category, and text filter on the duplicates report page.
 */
export default function duplicatesFilter() {
  const init = () => {
    const input = document.getElementById('dup-filter');
    const clearBtn = document.getElementById('dup-filter-clear');
    const emptyClearBtn = document.getElementById('dup-empty-clear');
    const countEl = document.getElementById('dup-count');
    const emptyEl = document.getElementById('dup-empty');
    const emptyQueryEl = document.getElementById('dup-empty-query');
    const tierBtns = Array.from(document.querySelectorAll('.dup-tier-btn'));
    const kindBtns = Array.from(document.querySelectorAll('.dup-kind-btn'));
    const items = Array.from(document.querySelectorAll('.dup-item'));
    if (!input || items.length === 0) {
      return;
    }

    const total = items.length;
    let activeTier = 'all';
    let activeKind = 'all';

    for (const item of items) {
      item._text = item.textContent.toLowerCase();
    }

    const applyFilters = () => {
      const raw = input.value.trim();
      const query = raw.toLowerCase();
      const tokens = query.split(/\s+/).filter(Boolean);
      let visible = 0;

      for (const item of items) {
        const tierMatch =
          activeTier === 'all' || item.dataset.tier === activeTier;
        const kindMatch =
          activeKind === 'all' || item.dataset.kind === activeKind;
        const textMatch =
          tokens.length === 0 ||
          tokens.every((tok) => item._text.includes(tok));
        const match = tierMatch && kindMatch && textMatch;
        item.classList.toggle('d-none', !match);
        if (match) {
          visible += 1;
        }
      }

      if (clearBtn) {
        clearBtn.disabled = input.value === '';
      }
      if (countEl) {
        const isDefault =
          query === '' && activeTier === 'all' && activeKind === 'all';
        countEl.textContent = isDefault
          ? `${total} ${total === 1 ? 'cluster' : 'clusters'}`
          : `Showing ${visible} of ${total}`;
      }
      if (emptyEl) {
        emptyEl.classList.toggle('d-none', visible > 0);
        if (emptyQueryEl) {
          emptyQueryEl.textContent = raw || `${activeKind} / Tier ${activeTier}`;
        }
      }
    };

    const setTier = (tier) => {
      activeTier = tier;
      for (const btn of tierBtns) {
        const selected = (btn.dataset.tier || 'all') === tier;
        btn.classList.toggle('active', selected);
        btn.setAttribute('aria-pressed', String(selected));
      }
      applyFilters();
    };

    const setKind = (kind) => {
      activeKind = kind;
      for (const btn of kindBtns) {
        const selected = (btn.dataset.kind || 'all') === kind;
        btn.classList.toggle('active', selected);
        btn.setAttribute('aria-pressed', String(selected));
      }
      applyFilters();
    };

    const resetAll = () => {
      input.value = '';
      setTier('all');
      setKind('all');
      input.focus();
    };

    for (const btn of tierBtns) {
      btn.addEventListener('click', () => setTier(btn.dataset.tier || 'all'));
    }
    for (const btn of kindBtns) {
      btn.addEventListener('click', () => setKind(btn.dataset.kind || 'all'));
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
  };

  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', init);
    return;
  }
  init();
}
