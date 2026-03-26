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

      // Check if this is a milestone link with data-milestone-id attribute
      var milestoneId = triggerElement.getAttribute('data-milestone-id');

      if (milestoneId) {
        // This is a milestone link - find the anchor with matching ID
        var anchor = document.getElementById(milestoneId);
        if (anchor) {
          var col = anchor.closest('.col');
          if (col) {
            var colClone = col.cloneNode(true);
            // title - use button text
            var buttonText = triggerElement.textContent.trim();
            document.getElementById('milestoneModalLabel').textContent =
              buttonText || 'Milestone Details';
            // content
            var modalBody = document.getElementById('milestoneModalBody');
            modalBody.innerHTML = '';
            modalBody.appendChild(colClone);
            // cleanup duplicate header
            var headerInClone = colClone.querySelector('.card-header');
            if (headerInClone) {
              headerInClone.remove();
            }
            return; // Found and processed the milestone
          }
        }
      } else {
        // This is a direct milestone button click - use original logic
        var col = triggerElement.closest('.col');
        if (col) {
          var colClone = col.cloneNode(true);
          // title - use button text
          var buttonText = triggerElement.textContent.trim();
          document.getElementById('milestoneModalLabel').textContent =
            buttonText || 'Milestone Details';
          // content
          var modalBody = document.getElementById('milestoneModalBody');
          modalBody.innerHTML = '';
          modalBody.appendChild(colClone);
          // cleanup duplicate header
          var headerInClone = colClone.querySelector('.card-header');
          if (headerInClone) {
            headerInClone.remove();
          }
        }
      }
    });
    // clear content when modal closes
    milestoneModal.addEventListener('hidden.bs.modal', function () {
      document.getElementById('milestoneModalLabel').textContent = '';
      document.getElementById('milestoneModalBody').innerHTML = '';
    });
  }
});
