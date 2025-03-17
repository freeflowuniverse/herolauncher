// Admin Dashboard JavaScript - Documentation Style

document.addEventListener('DOMContentLoaded', function() {
  // Highlight active navigation links
  highlightActiveLinks();
  
  // Setup UI toggles
  setupUIToggles();
  
  // Setup logging functionality
  setupLogging();
  
  // Setup search functionality
  setupSearch();
});

// Highlight the current active navigation links
function highlightActiveLinks() {
  const currentPath = window.location.pathname;
  
  // Handle top navigation links
  const navLinks = document.querySelectorAll('.nav-link');
  navLinks.forEach(link => {
    link.classList.remove('active');
    const href = link.getAttribute('href');
    
    // Check if current path starts with the nav link path
    // This allows section links to be highlighted when on sub-pages
    if (currentPath === href || 
        (href !== '/admin' && currentPath.startsWith(href))) {
      link.classList.add('active');
    }
  });
  
  // Handle sidebar links
  const sidebarLinks = document.querySelectorAll('.doc-link');
  sidebarLinks.forEach(link => {
    link.classList.remove('active');
    if (link.getAttribute('href') === currentPath) {
      link.classList.add('active');
      
      // Also highlight parent section if needed
      const parentSection = link.closest('.sidebar-section');
      if (parentSection) {
        parentSection.classList.add('active-section');
      }
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
  
  // Setup Docusaurus-style collapsible menu
  setupTreeviewMenu();
}

// Setup sidebar navigation
function setupTreeviewMenu() {
  // Set active sidebar links based on current URL
  setActiveSidebarLinks();
  
  // Setup collapsible sections
  setupCollapsibleSections();
}

// Set active sidebar links based on current URL
function setActiveSidebarLinks() {
  const currentPath = window.location.pathname;
  
  // Find all sidebar links
  const sidebarLinks = document.querySelectorAll('.sidebar-link');
  
  // Remove any existing active classes
  sidebarLinks.forEach(link => {
    link.classList.remove('active');
  });
  
  // Find and mark active links
  let activeFound = false;
  sidebarLinks.forEach(link => {
    const linkPath = link.getAttribute('href');
    
    // Check if the current path matches or starts with the link path
    // For exact matches or if it's a parent path
    if (currentPath === linkPath || 
        (linkPath !== '/admin' && currentPath.startsWith(linkPath))) {
      // Mark this link as active
      link.classList.add('active');
      activeFound = true;
      
      // Expand the parent section if this link is inside a collapsible section
      const parentSection = link.closest('.sidebar-content-section')?.parentElement;
      if (parentSection && parentSection.classList.contains('collapsible')) {
        parentSection.classList.remove('collapsed');
      }
    }
  });
}

// Setup collapsible sections
function setupCollapsibleSections() {
  // Find all toggle headings
  const toggleHeadings = document.querySelectorAll('.sidebar-heading.toggle');
  
  // Set all sections as collapsed by default
  document.querySelectorAll('.sidebar-section.collapsible').forEach(section => {
    section.classList.add('collapsed');
  });
  
  toggleHeadings.forEach(heading => {
    // Add click event to toggle section
    heading.addEventListener('click', function() {
      const section = this.parentElement;
      section.classList.toggle('collapsed');
    });
  });
  
  // Open the section that contains the active link
  const activeLink = document.querySelector('.sidebar-link.active');
  if (activeLink) {
    const parentSection = activeLink.closest('.sidebar-section.collapsible');
    if (parentSection) {
      parentSection.classList.remove('collapsed');
    }
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

// Setup search functionality
function setupSearch() {
  const searchInput = document.querySelector('.search-box input');
  if (searchInput) {
    searchInput.addEventListener('keyup', function(e) {
      if (e.key === 'Enter') {
        performSearch(this.value);
      }
    });
  }
}

// Perform search
function performSearch(query) {
  if (!query.trim()) return;
  
  // Log the search query
  window.adminLog(`Searching for: ${query}`, 'info');
  
  // In a real application, this would send an AJAX request to search the docs
  // For now, just simulate a search by redirecting to a search results page
  // window.location.href = `/admin/search?q=${encodeURIComponent(query)}`;
  
  // For demo purposes, show a message in the console
  console.log(`Search query: ${query}`);
}
