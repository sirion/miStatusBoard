import Dialog from "./ui/dialog.js";
import { d, clear } from "./utils/domutils.js";

export default class Board {

    constructor(apiUrl) {
        this.apiUrl = apiUrl;

        this._promiseConfig = this.request("config");
        this._promiseData = this.request("readAll");
    }

    async render(domRoot) {
        this.domRoot = domRoot;

        this.config = await this._promiseConfig;
        this.data = await this._promiseData;
        this.updateTiles();

        // Refresh every minute;
        setInterval(async () => {
            this._promiseData = this.request("readAll");
            this.data = await this._promiseData;
            this.updateTiles();
        }, 60000);
    }


    async request(url) {
        const response = await fetch(this.apiUrl + url);
        return response.json();
    }

    updateTiles() {
        document.querySelector("title").textContent = this.config.title;
        document.querySelector(".mainTitle").textContent = this.config.title;

        for (const group of this.config.groups) {
            this.updateTile(group);
        }
        setTimeout(() => {
            this.zoomTileTitles();
        }, 0);

    }

    rt = new Intl.RelativeTimeFormat('en', { style: 'narrow' });
    rtMs = {
        year: 24 * 60 * 60 * 1000 * 365,
        month: 24 * 60 * 60 * 1000 * 365 / 12,
        day: 24 * 60 * 60 * 1000,
        hour: 60 * 60 * 1000,
        minute: 60 * 1000
    };

    formatRelativeTime(date, referenceDate = new Date()) {
        const delta = date - referenceDate;

        for (const step in this.rtMs) {
            if (Math.abs(delta) > this.rtMs[step]) {
                return this.rt.format(Math.round(100 * delta / this.rtMs[step]) / 100, step);
            }
        }
        return this.rt.format(Math.round(delta / 1000), "second");
    }

    zoomTileTitles() {
        const tiles = this.domRoot.querySelectorAll(".tile");
        for (let i = 0; i < tiles.length; i++) {
            const tile = tiles[i];
            const title = tile.querySelector(".title");
            this.zoom(title, tile);
        }
    }

    async zoom(element, reference, zoomTarget = 0.9) {
        if (reference === true) {
            // Special case: reference is the viewport
            reference = {
                getBoundingClientRect() {
                    return {
                        height: window.innerHeight,
                        width: window.innerWidth
                    }
                }
            }
        }

        const el = element.getBoundingClientRect();
        const re = reference.getBoundingClientRect();
        const rScale = /scale\(.?\)/;

        const widthRatio = el.width / re.width;
        const heightRatio = el.height / re.height;

        const ratio = (Math.abs(1 - widthRatio) > Math.abs(1 - heightRatio)) ? heightRatio : widthRatio;
        const scale = `scale(${1 / ratio * zoomTarget})`;

        if (rScale.test(element.style.transform)) {
            element.style.transform.replace(rScale, scale);
        } else {
            element.style.transform += scale;
        }
    }

    updateTile(group) {
        const activeGrid = document.querySelector(".overview .active");
        const inactiveGrid = document.querySelector(".overview .inactive");

        const tile = this.groupTile(group)



        tile.classList.toggle("inactive", group.inactive);
        (group.inactive ? inactiveGrid : activeGrid).append(tile);

        const statusDots = tile.querySelector(".dots");
        clear(statusDots);

        const status = {
            "grey": 0,
            "green": 0,
            "yellow": 0,
            "red": 0
        };

        for (const endpoint of group.endpoints) {
            const url = new URL(endpoint.url, group.url || undefined);

            let st = this.data?.[url.toString()]?.status;
            if (!st) {
                // Data not available yet
                st = "grey";
            }
            status[st]++;

            const statusDot = d({
                classes: ["dot", `status_${st}`],
                attributes: {
                    title: `${endpoint.name} status: ${st}`
                }
            });
            statusDots.append(statusDot)
        }

        if (group.inactive) {
            tile.classList.add(`status_grey`);
        } else if (status.red > status.green) {
            tile.classList.add(`status_red`);
        } else if (status.red > 0) {
            tile.classList.add(`status_yellow`);
        } else {
            tile.classList.add(`status_green`);
        }
    }

    groupTile(group) {
        const groupId = group.name.replaceAll(/[^a-z0-9_]/ig, "_");
        let tile = document.querySelector(`#${groupId}`);
        if (!tile) {
            tile = document.createElement("div");
            tile.id = groupId;
            tile.classList.add("tile");

            const title = document.createElement("span");
            title.classList.add("title");
            title.textContent = group.name;
            tile.append(title);

            // const btn = document.createElement("span");
            // btn.classList.add("more");
            // btn.textContent = "[show more]";
            // tile.append(btn);

            const statusDots = document.createElement("div");
            statusDots.classList.add("dots");
            tile.append(statusDots);


            tile.addEventListener("click", e => {
                this.displayGroupDetails(group);
            });
        }
        return tile;
    }

    async displayGroupDetails(group) {
        const sortedEndpoints = group.endpoints.sort((e1, e2) => {
            if (e1.inactive === e2.inactive) {
                return 0;
            } else if (e1.inactive) {
                return 1
            }
            return -1;
        })

        const rows = sortedEndpoints.map(endpoint => {
            const url = new URL(endpoint.url, group.url || undefined);
            const e = this.data[url.toString()];
            return {
                type: "tr",
                classes: ["status_" + e.status],
                children: [{
                    type: "td",
                    textContent: endpoint.name
                }, {
                    type: "td",
                    textContent: e.code
                }, {
                    type: "td",
                    textContent: endpoint.inactive ? "inactive" : this.formatRelativeTime(new Date(e.updated))
                }, {
                    type: "td",
                    children: [{
                        type: "button",
                        textContent: "body",
                        style: {
                            display: (!e.body || endpoint.inactive) ? "none" : undefined
                        },
                        events: {
                            click: event => {
                                event.preventDefault();

                                let body = e.body ? atob(e.body) : "";
                                const shortened = body.length > 28;

                                let content;
                                if (e.contentType.startsWith("application/json")) {
                                    try {
                                        body = JSON.stringify(JSON.parse(body), null, 2);
                                    } catch (ex) {
                                        // Ignore
                                    }
                                }



                                if (e.contentType.startsWith("text/html")) {
                                    content = d({
                                        type: "iframe",
                                        style: {
                                            "min-width": "30vw",
                                            "max-width": "85vw"
                                        }
                                    });
                                    content.addEventListener("load", () => {
                                        content.contentWindow.document.write(body);
                                    });
                                } else {
                                    content = d({
                                        textContent: body,
                                        style: {
                                            "white-space": "break-spaces",
                                            "min-width": "30vw",
                                            "max-width": "85vw"
                                        }
                                    });
                                }


                                const dialog = Dialog.create("", [content]);
                                dialog.type = "none";
                                dialog.resize = false;
                                dialog.blocklayerCloses = true;
                                dialog.classList.add("endpointBody")
                                dialog.showModal();
                                this.zoom(dialog, true);
                            }
                        }
                    }]
                }]
            };

        });


        const content = [d({
            children: [{
                type: "table",
                children: [{
                    type: "tr",
                    children: [{
                        type: "th",
                        textContent: "Endpoint"
                    }, {
                        type: "th",
                        textContent: "Code"
                    }, {
                        type: "th",
                        textContent: "Time"
                    }, {
                        type: "th",
                        textContent: ""
                    }]
                }].concat(rows)
            }]
        })];

        const title = `Group Status for "${group.name}"`;

        const dialog = Dialog.create(title, content);
        dialog.type = "none";
        dialog.blocklayerCloses = true;
        dialog.classList.add("endpointsDetails")
        dialog.showModal();

        this.zoom(dialog, true);

        return dialog.closed;
    }

}
