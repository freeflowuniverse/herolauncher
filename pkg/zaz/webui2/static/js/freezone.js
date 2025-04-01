/**
 * Freezone Manager JavaScript
 * Main application JS file for the Freezone Manager UI
 */

// DOM ready handler
document.addEventListener('DOMContentLoaded', function() {
  console.log('Freezone Manager UI initialized');
  
  // Initialize all interactive elements
  initDropdowns();
  initTables();
  initSearch();
});

/**
 * Initialize dropdown menus
 */
function initDropdowns() {
  const dropdowns = document.querySelectorAll('.dropdown-trigger');
  
  dropdowns.forEach(dropdown => {
    dropdown.addEventListener('click', function(e) {
      e.preventDefault();
      const target = this.getAttribute('data-target');
      const dropdownContent = document.getElementById(target);
      
      if (dropdownContent) {
        dropdownContent.classList.toggle('active');
      }
    });
  });
  
  // Close dropdowns when clicking outside
  document.addEventListener('click', function(e) {
    if (!e.target.matches('.dropdown-trigger')) {
      const dropdownContents = document.querySelectorAll('.dropdown-content');
      dropdownContents.forEach(content => {
        if (content.classList.contains('active')) {
          content.classList.remove('active');
        }
      });
    }
  });
}

/**
 * Initialize interactive tables
 */
function initTables() {
  const tables = document.querySelectorAll('table.sortable');
  
  tables.forEach(table => {
    const headers = table.querySelectorAll('th');
    
    headers.forEach(header => {
      if (header.getAttribute('data-sortable') !== 'false') {
        header.addEventListener('click', function() {
          const sortDirection = this.getAttribute('data-sort-direction') === 'asc' ? 'desc' : 'asc';
          
          // Update all headers
          headers.forEach(h => {
            h.removeAttribute('data-sort-direction');
            h.classList.remove('sort-asc', 'sort-desc');
          });
          
          // Update current header
          this.setAttribute('data-sort-direction', sortDirection);
          this.classList.add(sortDirection === 'asc' ? 'sort-asc' : 'sort-desc');
          
          // You would implement actual sorting here
          console.log(`Sorting by ${this.textContent} in ${sortDirection} order`);
        });
      }
    });
  });
}

/**
 * Initialize search functionality
 */
function initSearch() {
  const searchInputs = document.querySelectorAll('.search-input');
  
  searchInputs.forEach(input => {
    input.addEventListener('keypress', function(e) {
      if (e.key === 'Enter') {
        e.preventDefault();
        const searchTerm = this.value.trim();
        
        if (searchTerm) {
          console.log(`Searching for: ${searchTerm}`);
          // In a real app, you would handle the search here
          window.location.href = `/search?q=${encodeURIComponent(searchTerm)}`;
        }
      }
    });
  });
}
