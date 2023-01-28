import Board from "./board.js";

main();

async function main() {
	const sw = await registerServiceWorker();
	const board = new Board("/api/");
	board.render(document.querySelector(".overview"));
}


async function registerServiceWorker() {
	try {
		const reg = await navigator.serviceWorker.register("/sw.js");
		const serviceWorker = reg.installing ?? reg.waiting ?? reg.active;
		if (serviceWorker) {
			serviceWorker.addEventListener("statechange", e => {
				console.log("[SW State] " + e.target.state);
			});

			serviceWorker.addEventListener("error", e => {
				console.error("[SW] Error:" +  e.message);
			});

			return serviceWorker;
		}

	} catch (ex) {
		console.error("Error registering service worker: " + ex.message);
	}

	return null;
}
