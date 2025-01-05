import htmx from "../vendor/htmx/htmx.min.js";
import hljs from "../vendor/highlight/highlight.min.js";

window.hljs = hljs

function timeAgo(input) {
	const date = (input instanceof Date) ? input : new Date(input);
	const formatter = new Intl.RelativeTimeFormat('en');
	const ranges = {
		years: 3600 * 24 * 365,
		months: 3600 * 24 * 30,
		weeks: 3600 * 24 * 7,
		days: 3600 * 24,
		hours: 3600,
		minutes: 60,
		seconds: 1
	};
	const secondsElapsed = (date.getTime() - Date.now()) / 1000;
	for (let key in ranges) {
		if (ranges[key] < Math.abs(secondsElapsed)) {
			const delta = secondsElapsed / ranges[key];
			return formatter.format(Math.round(delta), key);
		}
	}
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

window.HubroInit = function() {
	highlightNewCodeBlocks();
	boostLocalLinks();
	document.addEventListener('alpine:init', () => {
		Alpine.directive('timeago', el => {
			if (el.getAttribute("data-processed")) {
				return;
			}
			el.setAttribute("title", el.textContent);
			el.setAttribute("data-processed", "true");
			el.textContent = timeAgo(el.textContent);
		});
	});
}
