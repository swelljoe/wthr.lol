const CACHE_NAME = 'wthr-v1';
const ASSETS = [
    '/',
    '/static/css/style.css',
    '/static/icons/icon-192.png',
    '/static/icons/icon-512.png',
    '/static/manifest.json'
];

self.addEventListener('install', (event) => {
    event.waitUntil(
        caches.open(CACHE_NAME)
            .then((cache) => cache.addAll(ASSETS))
    );
});

self.addEventListener('fetch', (event) => {
    event.respondWith(
        caches.match(event.request)
            .then((response) => response || fetch(event.request))
    );
});
