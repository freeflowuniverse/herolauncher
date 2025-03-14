// Admin Dashboard JavaScript

document.addEventListener('DOMContentLoaded', function() {
  // Highlight active navigation links
  highlightActiveLinks();
  
  // Setup UI toggles
  setupUIToggles();
  
  // Setup logging functionality
  setupLogging();
});

// Highlight the current active navigation link
function highlightActiveLinks() {
  const currentPath = window.location.pathname;
  const navLinks = document.querySelectorAll('nav a');
  
  navLinks.forEach(link => {
    if (link.getAttribute('href') === currentPath) {
      link.classList.add('active');
    }
  });
}

// Setup UI toggle functionality
function setupUIToggles() {
  // Toggle sidebar on mobile
  const menuToggle = document.querySelector('.menu-toggle');
  const sidebar = document.querySelector('.sidebar');
  
  if (menuToggle && sidebar) {
    menuToggle.addEventListener('click', function() {
      sidebar.classList.toggle('open');
    });
  }
  
  // Toggle log panel
  const logToggle = document.querySelector('.log-toggle');
  const logPanel = document.querySelector('.log-panel');
  
  if (logToggle && logPanel) {
    logToggle.addEventListener('click', function() {
      logPanel.classList.toggle('open');
    });
  }
}

// Setup logging functionality
function setupLogging() {
  // Log panel functionality
  function appendToLog(message, type = 'info') {
    const logContent = document.querySelector('.log-content');
    if (logContent) {
      const timestamp = new Date().toISOString();
      const logEntry = document.createElement('div');
      logEntry.className = `log-entry log-${type}`;
      logEntry.textContent = `[${timestamp}] ${message}`;
      logContent.appendChild(logEntry);
      logContent.scrollTop = logContent.scrollHeight;
    }
  }
  
  // Expose log function globally
  window.adminLog = appendToLog;
  
  // Initialize with a log message
  appendToLog('Admin dashboard initialized', 'info');
}
