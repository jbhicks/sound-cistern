/**
 * Global HTMX Event Handlers and Enhancements
 * Provides consistent error handling, loading states, and accessibility
 */

document.addEventListener('DOMContentLoaded', function() {
    // Global HTMX configuration
    htmx.config.defaultSwapStyle = 'innerHTML';
    htmx.config.defaultSwapDelay = 0;
    htmx.config.defaultSettleDelay = 0;
    
    // Global error handler for HTMX requests
    document.body.addEventListener('htmx:error', function(evt) {
        console.error('HTMX Error:', evt.detail);
        
        const target = evt.detail.target;
        const errorMessage = evt.detail.error || 'An error occurred. Please try again.';
        
        // Create error notification
        showErrorNotification(errorMessage);
        
        // Remove loading states from buttons
        document.querySelectorAll('.htmx-request').forEach(el => {
            el.classList.remove('htmx-request');
        });
    });
    
    // Handle network errors
    document.body.addEventListener('htmx:sendError', function(evt) {
        console.error('HTMX Send Error:', evt.detail);
        showErrorNotification('Network error. Please check your connection and try again.');
    });
    
    // Handle HTTP errors (4xx, 5xx)
    document.body.addEventListener('htmx:responseError', function(evt) {
        const xhr = evt.detail.xhr;
        const status = xhr.status;
        let message = 'An error occurred. Please try again.';
        
        if (status === 401) {
            // Try to refresh the token first before redirecting to login
            fetch('/api/auth/refresh', {
                method: 'POST',
                credentials: 'include'
            }).then(refreshResp => {
                if (refreshResp.ok) {
                    // Token refreshed successfully, reload the page to retry the original request
                    window.location.reload();
                } else {
                    // Refresh failed, redirect to login
                    message = 'Your session has expired. Please sign in again.';
                    setTimeout(() => {
                        window.location.href = '/signin';
                    }, 2000);
                }
            }).catch(() => {
                // Network error during refresh, redirect to login
                message = 'Your session has expired. Please sign in again.';
                setTimeout(() => {
                    window.location.href = '/signin';
                }, 2000);
            });
        } else if (status === 403) {
            message = 'You do not have permission to perform this action.';
        } else if (status === 404) {
            message = 'The requested resource was not found.';
        } else if (status >= 500) {
            message = 'Server error. Please try again later.';
        }
        
        showErrorNotification(message);
    });
    
    // Focus management after HTMX swaps
    document.body.addEventListener('htmx:afterSwap', function(evt) {
        // Announce content changes to screen readers
        announceContentUpdate(evt.detail.target);
        
        // Re-attach event listeners to new content
        reattachEventListeners(evt.detail.target);
    });
    
    // Handle beforeRequest to set loading states
    document.body.addEventListener('htmx:beforeRequest', function(evt) {
        const trigger = evt.detail.elt;
        
        // Add loading state to buttons
        if (trigger.tagName === 'BUTTON' || trigger.getAttribute('role') === 'button') {
            trigger.setAttribute('data-original-text', trigger.innerHTML);
            trigger.disabled = true;
            trigger.setAttribute('aria-busy', 'true');
        }
    });
    
    // Handle afterRequest to remove loading states
    document.body.addEventListener('htmx:afterRequest', function(evt) {
        const trigger = evt.detail.elt;
        
        // Remove loading state from buttons
        if (trigger.tagName === 'BUTTON' || trigger.getAttribute('role') === 'button') {
            trigger.disabled = false;
            trigger.removeAttribute('aria-busy');
        }
        
        // Update filter results announcement
        if (evt.detail.target.id === 'track-container') {
            updateResultsAnnouncement(evt.detail.target);
        }
    });
    
    // Keyboard shortcuts
    document.addEventListener('keydown', function(evt) {
        // Escape key clears filters
        if (evt.key === 'Escape') {
            const filterForm = document.getElementById('filter-form');
            if (filterForm && document.activeElement.closest('.filter-section')) {
                evt.preventDefault();
                clearFilters();
            }
        }
    });
});

/**
 * Show error notification toast
 * @param {string} message - Error message to display
 * @param {string} type - Type of notification: 'error', 'success', 'warning', 'info'
 */
function showErrorNotification(message, type = 'error') {
    // Remove existing notifications of same type
    const existing = document.querySelector(`.notification-${type}`);
    if (existing) {
        existing.remove();
    }
    
    const notification = document.createElement('div');
    notification.className = `error-notification notification-${type}`;
    notification.setAttribute('role', 'alert');
    notification.setAttribute('aria-live', 'assertive');
    
    const icons = {
        error: '<circle cx="12" cy="12" r="10"></circle><line x1="15" y1="9" x2="9" y2="15"></line><line x1="9" y1="9" x2="15" y2="15"></line>',
        success: '<path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"></path><polyline points="22 4 12 14.01 9 11.01"></polyline>',
        warning: '<path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z"></path><line x1="12" y1="9" x2="12" y2="13"></line><line x1="12" y1="17" x2="12.01" y2="17"></line>',
        info: '<circle cx="12" cy="12" r="10"></circle><line x1="12" y1="16" x2="12" y2="12"></line><line x1="12" y1="8" x2="12.01" y2="8"></line>'
    };
    
    notification.innerHTML = `
        <div class="error-notification-content">
            <span class="error-icon" aria-hidden="true">
                <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    ${icons[type] || icons.info}
                </svg>
            </span>
            <span class="error-message">${message}</span>
            <button class="error-close" onclick="this.parentElement.parentElement.remove()" aria-label="Dismiss">
                <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <line x1="18" y1="6" x2="6" y2="18"></line>
                    <line x1="6" y1="6" x2="18" y2="18"></line>
                </svg>
            </button>
        </div>
    `;
    `;
    
    document.body.appendChild(notification);
    
    // Auto-remove after 5 seconds
    setTimeout(() => {
        if (notification.parentElement) {
            notification.remove();
        }
    }, 5000);
}

/**
 * Announce content updates to screen readers
 * @param {HTMLElement} target - Element that was updated
 */
function announceContentUpdate(target) {
    // Check if there's an aria-live region
    const liveRegion = target.closest('[aria-live="polite"]') || 
                       target.querySelector('[aria-live="polite"]');
    
    if (!liveRegion) {
        // Create temporary live region
        const announcement = document.createElement('div');
        announcement.setAttribute('role', 'status');
        announcement.setAttribute('aria-live', 'polite');
        announcement.setAttribute('aria-atomic', 'true');
        announcement.className = 'sr-only';
        
        const itemCount = target.querySelectorAll('.track-card, li').length;
        if (itemCount > 0) {
            announcement.textContent = `Content updated. ${itemCount} items displayed.`;
        } else {
            announcement.textContent = 'Content updated.';
        }
        
        document.body.appendChild(announcement);
        
        // Remove after announcement
        setTimeout(() => {
            announcement.remove();
        }, 1000);
    }
}

/**
 * Update results announcement for screen readers
 * @param {HTMLElement} container - Track container
 */
function updateResultsAnnouncement(container) {
    const trackCount = container.querySelectorAll('.track-card').length;
    const announcement = document.createElement('div');
    announcement.setAttribute('role', 'status');
    announcement.setAttribute('aria-live', 'polite');
    announcement.className = 'sr-only';
    announcement.textContent = `Found ${trackCount} track${trackCount !== 1 ? 's' : ''}`;
    
    document.body.appendChild(announcement);
    
    setTimeout(() => {
        announcement.remove();
    }, 1000);
}

/**
 * Re-attach event listeners to dynamically added content
 * @param {HTMLElement} container - Container with new content
 */
function reattachEventListeners(container) {
    // Re-attach favorite button listeners
    const favoriteButtons = container.querySelectorAll('.favorite-btn');
    favoriteButtons.forEach(btn => {
        btn.onclick = function(event) {
            event.preventDefault();
            toggleFavorite(this);
        };
    });
    
    // Ensure proper keyboard accessibility
    container.querySelectorAll('button:not([tabindex])').forEach(btn => {
        btn.setAttribute('tabindex', '0');
    });
}
