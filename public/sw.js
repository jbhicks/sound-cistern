const CACHE_NAME = 'sound-cistern-v4';
const MAX_TRACKS = 100;

const STATIC_ASSETS = [
  '/',
  '/css/pico.min.css',
  '/css/custom.css',
  '/css/stream.css',
  '/js/theme.js',
  '/js/htmx.min.js',
  '/js/htmx-enhancements.js',
  '/favicon.svg',
  '/manifest.json'
];

const STATIC_EXTENSIONS = [
  '.css',
  '.js',
  '.svg',
  '.png',
  '.jpg',
  '.jpeg',
  '.gif',
  '.ico',
  '.woff',
  '.woff2',
  '.ttf',
  '.eot'
];

const AUTH_ENDPOINTS = [
  '/api/auth',
  '/api/oauth',
  '/api/users/auth'
];

function isStaticAsset(url) {
  const pathname = new URL(url).pathname;
  return STATIC_EXTENSIONS.some(ext => pathname.endsWith(ext));
}

function isApiCall(url) {
  const pathname = new URL(url).pathname;
  return pathname.startsWith('/api/');
}

function isAuthEndpoint(url) {
  const pathname = new URL(url).pathname;
  return AUTH_ENDPOINTS.some(ep => pathname.startsWith(ep));
}

function isTrackData(url) {
  const pathname = new URL(url).pathname;
  return pathname.includes('/tracks') || pathname.includes('/stream');
}

async function limitCacheSize(cacheName, maxItems) {
  const cache = await caches.open(cacheName);
  const keys = await cache.keys();
  
  if (keys.length > maxItems) {
    const itemsToDelete = keys.slice(0, keys.length - maxItems);
    for (const request of itemsToDelete) {
      await cache.delete(request);
    }
  }
}

self.addEventListener('install', (event) => {
  event.waitUntil(
    caches.open(CACHE_NAME)
      .then((cache) => {
        return cache.addAll(STATIC_ASSETS);
      })
      .then(() => {
        return self.skipWaiting();
      })
      .catch((error) => {
        console.error('Service Worker install failed:', error);
      })
  );
});

self.addEventListener('activate', (event) => {
  event.waitUntil(
    caches.keys()
      .then((cacheNames) => {
        return Promise.all(
          cacheNames
            .filter((name) => name !== CACHE_NAME && name.startsWith('sound-cistern-'))
            .map((name) => {
              return caches.delete(name);
            })
        );
      })
      .then(() => {
        return self.clients.claim();
      })
  );
});

self.addEventListener('fetch', (event) => {
  const { request } = event;
  const url = request.url;

  if (request.method !== 'GET') {
    return;
  }

  if (isAuthEndpoint(url)) {
    return;
  }

  if (isStaticAsset(url)) {
    event.respondWith(cacheFirst(request));
    return;
  }

  if (isApiCall(url)) {
    event.respondWith(networkFirstWithCache(request));
    return;
  }

  event.respondWith(networkFirstWithCache(request));
});

async function cacheFirst(request) {
  const cachedResponse = await caches.match(request);
  
  if (cachedResponse) {
    return cachedResponse;
  }

  try {
    const networkResponse = await fetch(request);
    
    if (networkResponse.ok) {
      const cache = await caches.open(CACHE_NAME);
      cache.put(request, networkResponse.clone());
    }
    
    return networkResponse;
  } catch (error) {
    console.error('Cache-first fetch failed:', error);
    
    return new Response('Offline - content not available', {
      status: 503,
      statusText: 'Service Unavailable',
      headers: { 'Content-Type': 'text/plain' }
    });
  }
}

async function networkFirstWithCache(request) {
  try {
    const networkResponse = await fetch(request);
    
    if (networkResponse.ok) {
      const cache = await caches.open(CACHE_NAME);
      
      if (isTrackData(request.url)) {
        await limitCacheSize(CACHE_NAME, MAX_TRACKS);
      }
      
      cache.put(request, networkResponse.clone());
    }
    
    return networkResponse;
  } catch (error) {
    const cachedResponse = await caches.match(request);
    
    if (cachedResponse) {
      return cachedResponse;
    }

    if (request.headers.get('Accept')?.includes('text/html')) {
      const cachedRoot = await caches.match('/');
      if (cachedRoot) {
        return cachedRoot;
      }
    }

    return new Response(JSON.stringify({ 
      error: 'Offline', 
      message: 'Content not available while offline' 
    }), {
      status: 503,
      statusText: 'Service Unavailable',
      headers: { 'Content-Type': 'application/json' }
    });
  }
}
