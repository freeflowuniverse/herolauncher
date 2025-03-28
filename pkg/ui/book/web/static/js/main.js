// Initialize Unpoly
up.log.enable();

// Configure Unpoly
up.fragment.config.mainTargets = ['.content-area', '.right-sidebar'];

// Add a transition effect when loading content
up.transition.config.duration = 300;

// Update active link in sidebar and initialize right sidebar functionality
document.addEventListener('up:fragment:inserted', function(event) {
  // Get current path
  const path = window.location.pathname;
  
  // Find all sidebar links
  const links = document.querySelectorAll('.left-sidebar a');
  
  // Remove active class from all links
  links.forEach(link => {
    link.classList.remove('active');
    
    // Add active class to current link
    if (link.getAttribute('href') === path) {
      link.classList.add('active');
    }
  });
  
  // Initialize the right sidebar highlighting
  // This runs regardless of which fragment was updated
  initRightSidebar();
});

// Function to initialize the right sidebar functionality
function initRightSidebar() {
  // Add smooth scrolling to anchor links
  document.querySelectorAll('.right-sidebar a.anchor-link').forEach(anchor => {
    anchor.addEventListener('click', function(e) {
      e.preventDefault();
      const targetId = this.getAttribute('href');
      const targetElement = document.querySelector(targetId);
      
      if (targetElement) {
        // Smooth scroll to element
        targetElement.scrollIntoView({ behavior: 'smooth' });
        
        // Update URL without page reload
        history.pushState(null, null, targetId);
        
        // Highlight the clicked anchor in the sidebar
        highlightAnchor(targetId.substring(1)); // Remove the # from the ID
      }
    });
  });
  
  // Set up intersection observer to highlight current section while scrolling
  setupScrollSpy();
}

// Function to highlight the current anchor in the right sidebar
function highlightAnchor(id) {
  // Remove active class from all anchor links
  document.querySelectorAll('.right-sidebar a.anchor-link').forEach(anchor => {
    anchor.classList.remove('active');
  });
  
  // Add active class to the current anchor link
  const currentAnchor = document.querySelector(`.right-sidebar a.anchor-link[href="#${id}"]`);
  if (currentAnchor) {
    currentAnchor.classList.add('active');
  }
}

// Function to set up scroll spy using Intersection Observer
function setupScrollSpy() {
  const headings = document.querySelectorAll('.markdown-content h1, .markdown-content h2, .markdown-content h3, .markdown-content h4, .markdown-content h5, .markdown-content h6');
  
  if (headings.length === 0) return;
  
  const observerOptions = {
    root: null, // Use the viewport
    rootMargin: '0px 0px -80% 0px', // Consider element in view when it's 20% from the top
    threshold: 0
  };
  
  const observer = new IntersectionObserver((entries) => {
    entries.forEach(entry => {
      if (entry.isIntersecting) {
        // Get the ID of the heading that is currently visible
        const id = entry.target.getAttribute('id');
        if (id) {
          // Highlight the corresponding anchor in the sidebar
          highlightAnchor(id);
        }
      }
    });
  }, observerOptions);
  
  // Observe all headings
  headings.forEach(heading => {
    if (heading.hasAttribute('id')) {
      observer.observe(heading);
    }
  });
}

// Initialize right sidebar when the page loads
document.addEventListener('DOMContentLoaded', function() {
  initRightSidebar();
  initBackToTop();
});

// Function to initialize the back-to-top button
function initBackToTop() {
  const backToTopButton = document.querySelector('.back-to-top');
  
  if (!backToTopButton) return;
  
  // Show/hide the button based on scroll position
  window.addEventListener('scroll', function() {
    if (window.scrollY > 300) {
      backToTopButton.classList.add('visible');
    } else {
      backToTopButton.classList.remove('visible');
    }
  });
  
  // Scroll to top when clicked
  backToTopButton.addEventListener('click', function(e) {
    e.preventDefault();
    window.scrollTo({
      top: 0,
      behavior: 'smooth'
    });
  });
};
