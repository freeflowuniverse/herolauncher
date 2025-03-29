// Initialize Unpoly
up.log.enable();

// Configure Unpoly
up.fragment.config.mainTargets = ['.mail-list', '.mail-content'];

// Add a transition effect when loading content
up.transition.config.duration = 300;

// Update active link in sidebar and handle mail list selection
document.addEventListener('up:fragment:inserted', function(event) {
  // Get current path
  const path = window.location.pathname;
  
  // Find all sidebar links
  const sidebarLinks = document.querySelectorAll('.mail-sidebar a');
  
  // Remove active class from all sidebar links
  sidebarLinks.forEach(link => {
    link.classList.remove('active');
    
    // Add active class to current sidebar link
    if (link.getAttribute('href') === path || 
        (path.startsWith('/mail/') && link.getAttribute('href').startsWith('/mailbox/') && 
         path.includes(link.getAttribute('href').replace('/mailbox/', '')))) {
      link.classList.add('active');
    }
  });
  
  // Handle mail list selection
  const mailLinks = document.querySelectorAll('.mail-preview');
  
  // Remove active class from all mail links
  mailLinks.forEach(link => {
    link.classList.remove('selected');
    
    // Add selected class to current mail link
    if (link.getAttribute('href') === path) {
      link.classList.add('selected');
      // Mark as read
      link.classList.remove('unread');
    }
  });
});

// Initialize when the page loads
document.addEventListener('DOMContentLoaded', function() {
  // Get current path
  const path = window.location.pathname;
  
  // Find all sidebar links
  const sidebarLinks = document.querySelectorAll('.mail-sidebar a');
  
  // Set active class on the current sidebar link
  sidebarLinks.forEach(link => {
    if (link.getAttribute('href') === path || 
        (path.startsWith('/mail/') && link.getAttribute('href').startsWith('/mailbox/') && 
         path.includes(link.getAttribute('href').replace('/mailbox/', '')))) {
      link.classList.add('active');
    }
  });
  
  // Handle mail list selection
  const mailLinks = document.querySelectorAll('.mail-preview');
  
  // Set selected class on the current mail link
  mailLinks.forEach(link => {
    if (link.getAttribute('href') === path) {
      link.classList.add('selected');
    }
  });
});
