/**
 * @file index.js
 * This script is for the index template that is used by milestones.
 */
// Minimal content population using Bootstrap's built-in events
document.addEventListener('DOMContentLoaded', () => {
  const milestoneModal = document.getElementById('milestoneModal');
  if (!milestoneModal) return;

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
    const colClone = col.cloneNode(true);
    colClone.querySelector('.card-header')?.remove();
    colClone.querySelectorAll('[id]').forEach((el) => el.removeAttribute('id'));

    modalLabel.textContent = triggerElement.textContent.trim() || 'Milestone Details';
    modalBody.replaceChildren(colClone);
  });

  milestoneModal.addEventListener('hidden.bs.modal', () => {
    modalLabel.textContent = '';
    modalBody.replaceChildren();
  });
});
