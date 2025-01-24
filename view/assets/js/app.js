import htmx from "../vendor/htmx/htmx.min.js";
import hljs from "../vendor/highlight/highlight.min.js";
import { removeTheme, toggleTheme, initTheme } from "./darkmode/darkmode.js";

window.hljs = hljs

function timeAgo(input) {
	const date = (input instanceof Date) ? input : new Date(input);
	const now = new Date();

	const dateAtMidnight = new Date(date.getFullYear(), date.getMonth(), date.getDate());
	const nowAtMidnight  = new Date(now.getFullYear(),  now.getMonth(),  now.getDate());

	const msPerDay = 24 * 60 * 60 * 1000;
	const diffInDays = Math.round((dateAtMidnight - nowAtMidnight) / msPerDay);

	if (diffInDays === 0) {
		return 'today';
	} else if (diffInDays === -1) {
		return 'yesterday';
	}

	const formatter = new Intl.RelativeTimeFormat('en');
	const ranges = {
		years: 3600 * 24 * 365,
		months: 3600 * 24 * 30,
		weeks: 3600 * 24 * 7,
		days: 3600 * 24,
	};

	const secondsElapsed = (date.getTime() - now.getTime()) / 1000;

	for (let key in ranges) {
		if (ranges[key] < Math.abs(secondsElapsed)) {
			const delta = secondsElapsed / ranges[key];
			return formatter.format(Math.round(delta), key);
		}
	}
	return 'today';
}

function linkIsExternal(link_element) {
   return (link_element.host !== window.location.host);
}
function highlightNewCodeBlocks() {
	document.querySelectorAll('pre code:not(.hljs)').forEach(function (el) {
		hljs.highlightElement(el);
	});
}
function boostLocalLinks() {
	document.querySelectorAll('a').forEach(function (el) {
		if (!linkIsExternal(el)) {
			el.setAttribute('data-hx-boost', 'true');
		}
	});
}

function updateWSStatus(st) {
	document.querySelector('#ws-status').innerText = st;
}

function renderPreview(data) {
	const preview = document.querySelector('#markdown-preview');
	preview.innerHTML = data.content;
	if (!data.meta.hideAuthor && data.meta.author) {
		document.querySelector('#author-block').classList.remove('hidden');
		document.querySelector('#author').innerText = data.meta.author;
		if(data.meta.date) {
			document.querySelector('#date').innerText = timeAgo(data.meta.date);
			document.querySelector('#date').setAttribute('title', data.meta.date);
		}
	} else {
		document.querySelector('#author-block').classList.add('hidden');
	}
	if (!data.meta.hideTitle && data.meta.title) {
		document.querySelector('#title').innerText = data.meta.title;
		document.querySelector('#title').classList.remove('hidden');
	} else {
		document.querySelector('#title').classList.add('hidden');
	}
	document.querySelector('#tags').innerHTML = '';
	for (tag of data.meta.tags) {
		var tagHtml = `<span class="mx-1 rounded-full bg-indigo-300 px-3 py-1 text-sm tag text-gray-800 dark:text-gray-800">` + tag + `</span>`
		document.querySelector('#tags').innerHTML += tagHtml;
	}
}

function handleWSMessage(event) {
	const data = JSON.parse(event.data);
	if (data.type === 'reload') {
		window.location.reload();
	}
	if (data.type === 'markdown') {
		renderPreview(data);
		console.log(data.meta);
	}
	if (date.type === 'filecontent') {
		document.querySelector('#editor').value = data.content;
	}
}

function initWS() {
	if (window.ws) {
		if (window.ws.readyState === WebSocket.OPEN) {
			return;
		}
	}

	const scheme = window.location.protocol === 'https:' ? 'wss' : 'ws';
	const ws = new WebSocket(scheme + '://' + window.location.host + '/admin/ws');
	window.ws = ws;

	ws.onopen = function() {
		console.log('WebSocket is open');
		updateWSStatus('🟢');
	}
	ws.onmessage = handleWSMessage;
	ws.onclose = function() {
    	console.log('WebSocket is closed. Reconnecting in 5 seconds...');
		updateWSStatus('🔴');
    	setTimeout(function() {
        	initWS(); // Attempt to reconnect
		}, 5000);
    };
	ws.onerror = function(err) {
		updateWSStatus('🔴');
		console.error('WebSocket error observed:', err);
	}
}

// Expose toggleTheme to the window for easy hooking (e.g., button onclick)
window.toggleTheme = toggleTheme;
window.removeTheme = removeTheme;
window.initTheme = initTheme;

window.HubroInit = function() {
	initTheme();
	updateWSStatus('');
	highlightNewCodeBlocks();
	boostLocalLinks();
	document.addEventListener('alpine:init', () => {
		Alpine.prefix('data-x-');
		Alpine.directive('timeago', el => {
			if (el.getAttribute("data-processed")) {
				return;
			}
			el.setAttribute("title", el.textContent);
			el.setAttribute("data-processed", "true");
			el.textContent = timeAgo(el.textContent);
		});
	});

	// Add a copy button to code blocks
	document.addEventListener('htmx:load', function() {
	  initTheme();
	  const codeBlocks = document.querySelectorAll("pre > code:not(.copy-button-added)");

	  codeBlocks.forEach((codeEl) => {
		codeEl.classList.add("copy-button-added");
		const preEl = codeEl.parentElement;

		const wrapper = document.createElement("div");
		wrapper.classList.add("relative", "group");

		preEl.parentNode.insertBefore(wrapper, preEl);
		wrapper.appendChild(preEl);

		const button = document.createElement("button");
		button.innerText = "Copy";

		button.classList.add(
		  "hidden",
		  "group-hover:inline-block",
		  "absolute",
		  "top-2",
		  "right-2",
		  "bg-gray-200",
		  "px-2",
		  "py-1",
		  "rounded",
		  "text-sm",
		  "text-gray-600",
		  "hover:text-gray-800",
		  "hover:bg-gray-300"
		);

		wrapper.appendChild(button);

		button.addEventListener("click", () => {
		  navigator.clipboard
			.writeText(codeEl.textContent)
			.then(() => {
			  button.innerText = "Copied!";
			  setTimeout(() => (button.innerText = "Copy"), 1500);
			})
			.catch((err) => {
			  console.error("Failed to copy text: ", err);
			  button.innerText = "Error";
			});
		});
	  });
	});
}

window.AdminInit = function() {
	initWS();

	document.addEventListener('htmx:load', function() {
		idx = new URLSearchParams(window.location.search).get('idx');
		file = new URLSearchParams(window.location.search).get('p');
		if (idx !== null && file !== null) {
			if (ws.readyState === WebSocket.OPEN) {
				ws.send(JSON.stringify({ type: 'load', id: file, idx: idx }));
			}
		}
	});
}

function debounce(fn, delay) {
  let timeout;
  return function(...args) {
    clearTimeout(timeout);
    timeout = setTimeout(() => {
      fn.apply(this, args);
    }, delay);
  };
}

window.sendMarkdown = debounce(function(value, id) {
	const ws = window.ws;
	const markdown = value;
	ws.send(JSON.stringify({ type: 'markdown', content: markdown, id: id }));
}, 300);

window.savePage = function(value, id) {
	const ws = window.ws;
	const idx = new URLSearchParams(window.location.search).get('idx');
	ws.send(JSON.stringify({ type: 'save', content: value, id: id, idx: idx }));
}
