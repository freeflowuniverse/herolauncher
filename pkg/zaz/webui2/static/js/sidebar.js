// sidebar.js - Handles sidebar navigation behavior
document.addEventListener('DOMContentLoaded', function() {
  // Get all sidebar details elements
  const sidebarDetails = document.querySelectorAll('#freezone-menu details');
  const currentPath = window.location.pathname;
  
  // Function to close all details elements
  function closeAllDetails() {
    sidebarDetails.forEach(detail => {
      detail.removeAttribute('open');
    });
  }
  
  // Close all details initially
  closeAllDetails();
  
  // Find and open the correct section based on current path
  let sectionToOpen = null;
  
  // First try to find exact path match
  const matchingLink = document.querySelector(`#freezone-menu a[data-path="${currentPath}"]`);
  if (matchingLink) {
    sectionToOpen = matchingLink.closest('details');
  } else {
    // If no exact match, try to find section by checking if path starts with section path
    // This handles nested routes
    sidebarDetails.forEach(detail => {
      const section = detail.getAttribute('data-section');
      const sectionLinks = detail.querySelectorAll('a');
      
      sectionLinks.forEach(link => {
        const linkPath = link.getAttribute('data-path');
        if (currentPath.startsWith(linkPath) && linkPath !== '/') {
          sectionToOpen = detail;
        }
      });
    });
    
    // Default to dashboard for home page
    if (currentPath === '/' || !sectionToOpen) {
      sectionToOpen = document.querySelector('#freezone-menu details[data-section="dashboard"]');
    }
  }
  
  // Open the matching section
  if (sectionToOpen) {
    sectionToOpen.setAttribute('open', '');
  }
  
  // Add click event listeners to all summary elements
  sidebarDetails.forEach(detail => {
    const summary = detail.querySelector('summary');
    
    summary.addEventListener('click', function(e) {
      // Prevent the default toggle behavior
      e.preventDefault();
      
      // If this detail is already open, close it and return
      if (detail.hasAttribute('open')) {
        detail.removeAttribute('open');
        return;
      }
      
      // Close all details
      closeAllDetails();
      
      // Open this detail
      detail.setAttribute('open', '');
    });
  });
});
