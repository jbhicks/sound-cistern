/**
 * Theme switching functionality for Pico.css
 * Handles dark/light/auto theme switching with localStorage persistence
 * Includes ARIA accessibility attributes for screen readers
 */

function setTheme(theme) {
  // Try to get the header icon first, then the general one if it exists.
  let themeIcon = document.getElementById('theme-icon-header');
  if (!themeIcon) {
    themeIcon = document.getElementById('theme-icon'); 
  }

  // Update aria-pressed state for accessibility
  const themeButtons = document.querySelectorAll('.theme-toggle-btn');
  
  if (theme === 'auto') {
    localStorage.removeItem('picoPreferredColorScheme');
    document.documentElement.removeAttribute('data-theme');
    
    // Update aria-pressed state - auto is not pressed
    themeButtons.forEach(btn => {
      btn.setAttribute('aria-pressed', 'false');
      btn.setAttribute('aria-label', 'Toggle between light and dark mode (currently auto)');
    });
    
    if (themeIcon) { // Check if themeIcon was found
      themeIcon.innerHTML = `<svg fill="none" stroke="currentColor" viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg" aria-hidden="true" width="20" height="20">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 17.25v1.007a3 3 0 01-.879 2.122L7.5 21h9l-.621-.621A3 3 0 0115 18.257V17.25m6-12V15a2.25 2.25 0 01-2.25 2.25H5.25A2.25 2.25 0 013 15V5.25A2.25 2.25 0 015.25 3h13.5A2.25 2.25 0 0121 5.25z"></path>
      </svg>`; // Auto icon (desktop monitor)
    }
  } else {
    localStorage.setItem('picoPreferredColorScheme', theme);
    document.documentElement.setAttribute('data-theme', theme);
    
    // Update aria-pressed state - the active theme is "pressed"
    themeButtons.forEach(btn => {
      btn.setAttribute('aria-pressed', 'true');
    });
    
    if (themeIcon) { // Check if themeIcon was found
      if (theme === 'dark') {
        themeIcon.innerHTML = `<svg fill="none" stroke="currentColor" viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg" aria-hidden="true" width="20" height="20">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M20.354 15.354A9 9 0 018.646 3.646 9.003 9.003 0 0012 21a9.003 9.003 0 008.354-5.646z"></path>
        </svg>`; // Dark icon (moon)
        themeButtons.forEach(btn => {
          btn.setAttribute('aria-label', 'Toggle to light mode (currently dark)');
        });
      } else { // light
        themeIcon.innerHTML = `<svg fill="none" stroke="currentColor" viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg" aria-hidden="true" width="20" height="20">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 3v1m0 16v1m9-9h-1M4 12H3m15.364 6.364l-.707-.707M6.343 6.343l-.707-.707m12.728 0l-.707.707M6.343 17.657l-.707.707M16 12a4 4 0 11-8 0 4 4 0 018 0z"></path>
        </svg>`; // Light icon (sun)
        themeButtons.forEach(btn => {
          btn.setAttribute('aria-label', 'Toggle to dark mode (currently light)');
        });
      }
    }
  }
}

window.toggleTheme = function() { // Explicitly global
  const currentTheme = localStorage.getItem('picoPreferredColorScheme');
  
  if (!currentTheme || currentTheme === 'auto') {
    setTheme('dark');
  } else if (currentTheme === 'dark') {
    setTheme('light');
  } else {
    setTheme('dark');
  }
}

// Initialize theme on page load
document.addEventListener('DOMContentLoaded', function() {
  const savedTheme = localStorage.getItem('picoPreferredColorScheme');
  // Query all buttons that might toggle the theme, not just the one in the header initially.
  const themeToggleButtons = document.querySelectorAll('.theme-toggle-btn'); 

  let initialTheme = savedTheme;
  if (!savedTheme) {
    initialTheme = 'dark'; // Default to dark
  }
  
  setTheme(initialTheme); 
  
  // Ensure all theme toggle buttons have proper accessibility attributes
  themeToggleButtons.forEach(button => {
    // Ensure button is keyboard accessible
    if (!button.hasAttribute('tabindex')) {
      button.setAttribute('tabindex', '0');
    }
    
    // Add keyboard support (Enter and Space)
    button.addEventListener('keydown', function(event) {
      if (event.key === 'Enter' || event.key === ' ') {
        event.preventDefault();
        toggleTheme();
      }
    });
  });
});

// Pico.css Modal Toggle Function
window.toggleModal = function(event) {
  if (event) event.preventDefault();
  const modalId = event.currentTarget.dataset.target;
  const modal = document.getElementById(modalId);
  if (modal) {
    if (modal.open) {
      modal.close();
    } else {
      // Before showing, ensure HTMX has a chance to load content if it hasn't already.
      // This is a simple approach; more robust would be to listen for htmx:afterSwap on the modal content div.
      // For now, we assume the hx-get on the button will trigger and populate before or as this runs.
      modal.showModal();
    }
  }
};

// Switch between modals
function switchModal(event, currentModalId, targetModalId) {
  event.preventDefault(); // Prevent default link navigation
  const currentModal = document.getElementById(currentModalId);
  const targetModal = document.getElementById(targetModalId);

  if (currentModal && currentModal.open) { // Only close if open
    currentModal.open = false; 
  }
  if (targetModal) {
    targetModal.open = true; // Open the target modal
    // HTMX attributes on the link (event.currentTarget) will handle content loading.
  }
}
