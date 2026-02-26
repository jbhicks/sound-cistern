(function() {
    'use strict';

    const OFFLINE_CLASS = 'offline';
    const INDICATOR_SELECTOR = '#offline-indicator';

    function updateOnlineStatus() {
        const isOnline = navigator.onLine;
        const indicator = document.querySelector(INDICATOR_SELECTOR);

        if (isOnline) {
            document.body.classList.remove(OFFLINE_CLASS);
            if (indicator) {
                indicator.setAttribute('hidden', '');
                indicator.style.display = 'none';
            }
            console.log('[Offline] Status: Online');
        } else {
            document.body.classList.add(OFFLINE_CLASS);
            if (indicator) {
                indicator.removeAttribute('hidden');
                indicator.style.display = '';
            }
            console.log('[Offline] Status: Offline');
        }
    }

    function handleServiceWorkerUpdate(registration) {
        registration.addEventListener('updatefound', () => {
            const newWorker = registration.installing;

            newWorker.addEventListener('statechange', () => {
                if (newWorker.state === 'installed' && navigator.serviceWorker.controller) {
                    console.log('[Offline] New content available, refresh to update');
                } else if (newWorker.state === 'installed') {
                    console.log('[Offline] Content cached for offline use');
                }
            });
        });
    }

    function registerServiceWorker() {
        if (!navigator.serviceWorker) {
            console.log('[Offline] Service Worker not supported');
            return;
        }

        navigator.serviceWorker.register('/sw.js')
            .then(registration => {
                console.log('[Offline] Service Worker registered:', registration.scope);

                handleServiceWorkerUpdate(registration);

                setInterval(() => {
                    registration.update().catch(err => {
                        console.log('[Offline] SW update check failed:', err.message);
                    });
                }, 60 * 60 * 1000);
            })
            .catch(error => {
                console.error('[Offline] Service Worker registration failed:', error);
            });
    }

    function init() {
        updateOnlineStatus();

        window.addEventListener('online', updateOnlineStatus);
        window.addEventListener('offline', updateOnlineStatus);

        if (document.readyState === 'complete') {
            registerServiceWorker();
        } else {
            window.addEventListener('load', registerServiceWorker);
        }
    }

    init();
})();
