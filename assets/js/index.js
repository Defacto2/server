/**
 * @file index.js
 * This script is for the index template that is used by milestones.
 */
document.addEventListener('DOMContentLoaded', () => {
  const milestoneModal = document.getElementById('milestoneModal');
  if (milestoneModal) {
    const modalLabel = document.getElementById('milestoneModalLabel');
    const modalBody = document.getElementById('milestoneModalBody');

    milestoneModal.addEventListener('show.bs.modal', (event) => {
      const triggerElement = event.relatedTarget;
      if (!triggerElement) return;

      const milestoneId = triggerElement.getAttribute('data-milestone-id');
      const col = milestoneId
        ? document.getElementById(milestoneId)?.closest('.col')
        : triggerElement.closest('.col');

      if (!col) return;
      const dateText = col.querySelector('.milestone-date')?.textContent.trim();
      const colClone = col.cloneNode(true);
      colClone.classList.remove('d-none');
      colClone.querySelector('.card-header')?.remove();
      colClone.querySelectorAll('[id]').forEach((el) => el.removeAttribute('id'));

      modalLabel.textContent = dateText || triggerElement.textContent.trim() || 'Milestone Details';
      modalBody.replaceChildren(colClone);
    });

    milestoneModal.addEventListener('hidden.bs.modal', () => {
      modalLabel.textContent = '';
      modalBody.replaceChildren();
    });
  }

  const grid = document.getElementById('milestone-grid');
  const decadeButtons = document.querySelectorAll('[data-decade-filter]');
  const highlightsToggle = document.getElementById('milestone-highlights-only');
  const countEl = document.getElementById('milestone-count');
  if (!grid || decadeButtons.length === 0) return;

  const items = Array.from(grid.querySelectorAll('.milestone-item'));
  const total = items.length;
  let activeDecade = 'all';

  const applyFilter = () => {
    const onlyHighlights = Boolean(highlightsToggle?.checked);
    let visible = 0;

    items.forEach((item) => {
      const decadeMatch = activeDecade === 'all' || item.dataset.decade === activeDecade;
      const highlightMatch = !onlyHighlights || item.dataset.highlight === 'true';
      const show = decadeMatch && highlightMatch;
      item.classList.toggle('d-none', !show);
      if (show) visible++;
    });

    if (countEl) {
      countEl.textContent = `${visible} of ${total} milestones`;
    }
  };

  decadeButtons.forEach((btn) => {
    btn.addEventListener('click', () => {
      activeDecade = btn.dataset.decadeFilter || 'all';
      decadeButtons.forEach((b) => {
        const active = b === btn;
        b.classList.toggle('btn-primary', active);
        b.classList.toggle('btn-outline-secondary', !active);
        b.setAttribute('aria-pressed', active ? 'true' : 'false');
      });
      applyFilter();
    });
  });

  if (highlightsToggle) {
    highlightsToggle.addEventListener('change', applyFilter);
  }
});
