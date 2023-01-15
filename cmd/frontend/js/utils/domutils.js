const c = document.createElement.bind(document);

function d(options) {
	const type = options.type || "div";

	let namespace = null;
	if (!options.namespace) {
		if (type.toLowerCase() === "svg") {
			namespace = "http://www.w3.org/2000/svg";
		}
	} else {
		namespace = options.namespace;
	}

	let el;
	if (namespace) {
		el = document.createElementNS(namespace, type);
	} else {
		el = document.createElement(type);
	}


	if (options.id) {
		el.id = options.id;
	}

	if (options.classes) {
		const classes = Array.isArray(options.classes) ? options.classes : [ options.classes ];
		classes.forEach(cl => el.classList.add(cl));
	}

	if (options.textContent) {
		el.textContent = options.textContent;
	}

	if (options.attributes) {
		Object.keys(options.attributes).forEach(name => {
			if (options.attributes[name] !== undefined) {
				el.setAttribute(name, options.attributes[name]);
			}
		});
	}

	if (options.style) {
		Object.assign(el.style, options.style);
	}

	if (options.events) {
		Object.keys(options.events).forEach(name => {
			el.addEventListener(name, options.events[name]);
		});
	}

	if (options.content || options.children) {
		options.content = options.content || [];
		options.children = options.children || [];

		const content = Array.isArray(options.content) ? options.content : [ options.content ];
		const children = Array.isArray(options.children) ? options.children : [ options.children ];

		const subElements = content.concat(children).map(c => {
			if (typeof c === "string") {
				return document.createTextNode(c)
			} else if (c instanceof Node) {
				// Allow using Nodes in children-array
				return c;
			}

			if (namespace) {
				c = Object.assign({ namespace: namespace }, c);
			}
			return d(c)
		})

		el.append(...subElements);
	}

	return el;
}

/**
 * Clears als children from the given DOM Node
 * 
 * @param {Node} - The DOM Node to be cleared
 * @returns {void}
 */
function clear(domNode) {
	while (domNode.children.length > 0) {
		domNode.removeChild(domNode.children[0]);
	}
}

export { c, d, clear };