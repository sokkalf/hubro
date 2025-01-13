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

// Expose toggleTheme to the window for easy hooking (e.g., button onclick)
window.toggleTheme = toggleTheme;
window.removeTheme = removeTheme;

window.HubroInit = function() {
	initTheme();
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
