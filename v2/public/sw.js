// Self-destruct: clear all caches and unregister this SW.
// The app runs via a live dev tunnel; a caching SW causes stale-chunk
// errors that break React. PWA caching is disabled until a proper
// versioned build pipeline is in place.

self.addEventListener('install', () => {
  self.skipWaiting();
});

self.addEventListener('activate', (event) => {
  event.waitUntil(
    caches.keys()
      .then((keys) => Promise.all(keys.map((key) => caches.delete(key))))
      .then(() => self.registration.unregister())
      .then(() => self.clients.matchAll())
      .then((clients) => clients.forEach((client) => client.navigate(client.url)))
  );
});
