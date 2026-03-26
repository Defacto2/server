/**
 * @file index.js
 * This script is for the index template that is used by milestones.
 */
// Minimal content population using Bootstrap's built-in events
document.addEventListener('DOMContentLoaded', function () {
  var milestoneModal = document.getElementById('milestoneModal');
  if (milestoneModal) {
    milestoneModal.addEventListener('show.bs.modal', function (event) {
      var triggerElement = event.relatedTarget;
      var milestoneId = triggerElement.getAttribute('data-milestone-id');
      var col = milestoneId ? document.getElementById(milestoneId)?.closest('.col') : triggerElement.closest('.col');

      if (col) {
        var colClone = col.cloneNode(true);
        var buttonText = triggerElement.textContent.trim();
        document.getElementById('milestoneModalLabel').textContent = buttonText || 'Milestone Details';
        var modalBody = document.getElementById('milestoneModalBody');
        modalBody.innerHTML = '';
        modalBody.appendChild(colClone);
        var headerInClone = colClone.querySelector('.card-header');
        if (headerInClone) headerInClone.remove();
      }
    });

    milestoneModal.addEventListener('hidden.bs.modal', function () {
      document.getElementById('milestoneModalLabel').textContent = '';
      document.getElementById('milestoneModalBody').innerHTML = '';
    });
  }
});
