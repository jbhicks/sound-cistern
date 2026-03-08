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
    });
    
    // Toast notification system
    function showErrorNotification(message) {
        // Remove existing toasts first
        const existing = document.querySelector('.toast-notification.error');
        if (existing) existing.remove();
        
        const toast = document.createElement('div');
        toast.className = 'toast-notification error';
        toast.textContent = message;
        toast.setAttribute('role', 'alert');
        toast.setAttribute('aria-live', 'assertive');
        
        Object.assign(toast.style, {
            position: 'fixed',
            bottom: '20px',
            right: '20px',
            padding: '12px 20px',
            background: '#dc3545',
            color: '#fff',
            borderRadius: '6px',
            boxShadow: '0 4px 12px rgba(0,0,0,0.3)',
            zIndex: '9999',
            maxWidth: '300px',
            animation: 'slideIn 0.3s ease'
        });
        
        document.body.appendChild(toast);
        
        setTimeout(() => {
            toast.style.animation = 'slideOut 0.3s ease';
            setTimeout(() => toast.remove(), 300);
        }, 4000);
    }
    
    // Success notification
    function showSuccessNotification(message) {
        const existing = document.querySelector('.toast-notification.success');
        if (existing) existing.remove();
        
        const toast = document.createElement('div');
        toast.className = 'toast-notification success';
        toast.textContent = message;
        toast.setAttribute('role', 'status');
        toast.setAttribute('aria-live', 'polite');
        
        Object.assign(toast.style, {
            position: 'fixed',
            bottom: '20px',
            right: '20px',
            padding: '12px 20px',
            background: '#28a745',
            color: '#fff',
            borderRadius: '6px',
            boxShadow: '0 4px 12px rgba(0,0,0,0.3)',
            zIndex: '9999',
            maxWidth: '300px',
            animation: 'slideIn 0.3s ease'
        });
        
        document.body.appendChild(toast);
        
        setTimeout(() => {
            toast.style.animation = 'slideOut 0.3s ease';
            setTimeout(() => toast.remove(), 300);
        }, 3000);
    }
    
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
            const originalText = trigger.getAttribute('data-original-text');
            if (originalText) {
                trigger.innerHTML = originalText;
                trigger.removeAttribute('data-original-text');
            }
        }
        
        // Handle error responses
        if (evt.detail.failed) {
            const status = evt.detail.xhr.status;
            let message = 'An error occurred.';
            
            if (status === 401) {
                message = 'Please log in to continue.';
                // Optionally redirect to login
                // window.location.href = '/login';
            } else if (status === 0) {
                // Network error - could be offline
                message = 'Network error. Please check your connection.';
                setTimeout(() => {
                    showErrorNotification('Retrying...');
                }, 2000);
            } else if (status === 403) {
                message = 'You do not have permission to perform this action.';
            } else if (status === 404) {
                message = 'The requested resource was not found.';
            } else if (status >= 500) {
                message = 'Server error. Please try again later.';
            }
            
            showErrorNotification(message);
        }
    });
    
    // Announce content updates for accessibility
    function announceContentUpdate(target) {
        const announcement = document.createElement('div');
        announcement.setAttribute('aria-live', 'polite');
        announcement.setAttribute('aria-atomic', 'true');
        announcement.className = 'sr-only';
        
        const itemCount = target.querySelectorAll('.track-card, li').length;
        if (itemCount > 0) {
            announcement.textContent = 'Content updated. ' + itemCount + ' items displayed.';
        } else {
            announcement.textContent = 'Content updated.';
        }
        
        document.body.appendChild(announcement);
        
        // Remove after announcement
        setTimeout(() => {
            announcement.remove();
        }, 1000);
    }
    
    // Toggle favorite functionality
    function toggleFavorite(button) {
        const trackId = button.dataset.trackId;
        const isFavorite = button.classList.contains('favorited');
        
        // Optimistic update
        button.classList.toggle('favorited');
        button.setAttribute('aria-pressed', !isFavorite);
        
        // Send request
        fetch('/api/favorites/toggle', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({ track_id: trackId })
        })
        .then(response => {
            if (!response.ok) {
                // Revert on error
                button.classList.toggle('favorited');
                button.setAttribute('aria-pressed', isFavorite);
                showErrorNotification('Failed to update favorite');
            } else {
                showSuccessNotification(isFavorite ? 'Removed from favorites' : 'Added to favorites');
            }
        })
        .catch(() => {
            // Revert on error
            button.classList.toggle('favorited');
            button.setAttribute('aria-pressed', isFavorite);
            showErrorNotification('Failed to update favorite');
        });
    }
    
    // Re-attach event listeners to dynamically added content
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
});
