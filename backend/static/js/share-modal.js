// share-modal.js - Handles modal interactions safely with HTMX
document.addEventListener('htmx:afterSwap', function(event) {
  // Only run this if the swapped content contains a modal
  const modal = document.querySelector('.modal-container');
  if (!modal) return;
  
  // Find buttons within the modal
  const cancelButton = modal.querySelector('button[type="button"]');
  const submitButton = modal.querySelector('button[type="submit"]');
  
  // Safely add event listeners if elements exist
  if (cancelButton) {
    cancelButton.addEventListener('click', function() {
      // Clear the modal container
      const modalContainer = document.getElementById('modal-container');
      if (modalContainer) {
        modalContainer.innerHTML = '';
      }
    });
  }
  
  // Additional modal initialization can go here
  console.log('Modal initialized after HTMX swap');
});
