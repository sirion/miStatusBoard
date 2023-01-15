import AmsStatus from "./amsstatus.js";

main();

async function main() {
    const amsStatus = new AmsStatus("/api/");
    amsStatus.render(document.querySelector(".overview"));
}



document.querySelector("#btnRefreshAll")?.addEventListener("click", async () => {
	document.body.classList.toggle("busy", true);
	const response = await fetch("/api/refreshAll");
	const data = await response.json();
	document.body.classList.toggle("busy", false);
	updateTiles(data);
});
