import htmx from "../vendor/htmx/htmx.min.js";
import hljs from "../vendor/highlight/highlight.min.js";

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

window.HubroInit = function() {
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
}
