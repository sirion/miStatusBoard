const cacheName = "v0.2.0";

const cacheURIs = [
	"/index.html",
	"js/main.js",
	"js/board.js",
	"js/ui/dialog.js",
	"js/utils/domutils.js",
	"css/main.css"
];


const cacheOpened = caches.open(cacheName);

function onActivate() {
	return caches.keys().then(keys => {
		return Promise.all(keys.map(key => caches.delete(key)));
	}).then(() => {
		return self.clients.claim();
	});
}

function onInstall(/* e */) {
	self.skipWaiting();

	return cacheOpened.then(cache => {
		return cache.addAll(cacheURIs);
	});
}

function updateCache(request) {
	console.log("[Service Worker] Caching resource " + request.url);
	return fetch(request).then(response => {
		const cacheResponse = response.clone();
		cacheOpened.then(cache => {
			cache.put(request, cacheResponse);
		});
		return response;
	});
}

async function fromCache(request) {
	const cache = await cacheOpened;
	return cache.match(request);
}

async function cachedRequest(request) {
	let response = await fromCache(request);
	if (!response) {
		response = updateCache(request);
	}
	return response;
}

self.addEventListener("install", e => {
	console.log("[Service Worker] Install");
	e.waitUntil(onInstall(e));
});

self.addEventListener("activate", e => {
	e.waitUntil(onActivate());
});

self.addEventListener("fetch", e => {
	const isAPI = e.request.url.startsWith(self.origin + "/api");
	if (!isAPI) {
		e.respondWith(cachedRequest(e.request));
	} else if (e.request.method === "GET") {
		e.respondWith(fetch(request));
	}
});
