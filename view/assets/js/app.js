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

/**
 * Get the user's system-level preference (dark or light).
 * @returns {string} "dark" or "light"
 */
function getSystemTheme() {
  return window.matchMedia("(prefers-color-scheme: dark)").matches
    ? "dark"
    : "light";
}

/**
 * Return any user-saved theme from localStorage, or null if none is saved.
 * @returns {string|null} "dark", "light", or null
 */
function getStoredTheme() {
  return localStorage.getItem("theme");
}

/**
 * Determine the currently active theme:
 *   - If a user-saved theme is present, return that;
 *   - Otherwise, return the system preference.
 * @returns {string} "dark" or "light"
 */
function getActiveTheme() {
  const storedTheme = getStoredTheme();
  return storedTheme ? storedTheme : getSystemTheme();
}

/**
 * Toggle the current theme. If the active theme is "light",
 * change it to "dark", and vice versa. Save the new value to localStorage.
 * Then, flip media queries and update icons.
 */
function toggleTheme() {
  const currentTheme = getActiveTheme();
  const newTheme = currentTheme === "light" ? "dark" : "light";

  // Save the new theme
  localStorage.setItem("theme", newTheme);

  // Flip all prefers-color-scheme rules
  switchThemeRules();

  // Update any UI elements (icons, etc.)
  updateThemeIcons(newTheme);
}

/**
 * Switch the prefers-color-scheme references in all loaded stylesheets.
 * This effectively flips "light" to "dark" or "dark" to "light".
 */
function switchThemeRules() {
  for (let sheetIndex = 0; sheetIndex < document.styleSheets.length; sheetIndex++) {
    const styleSheet = document.styleSheets[sheetIndex];

    try {
      for (
        let ruleIndex = 0;
        ruleIndex < styleSheet.cssRules.length;
        ruleIndex++
      ) {
        const rule = styleSheet.cssRules[ruleIndex];
        // Check if the rule is a @media rule that includes "prefers-color-scheme"
        if (rule?.media && rule.media.mediaText.includes("prefers-color-scheme")) {
          const oldMedia = rule.media.mediaText;
          let newMedia = oldMedia;

          if (oldMedia.includes("light")) {
            newMedia = oldMedia.replace("light", "dark");
          } else if (oldMedia.includes("dark")) {
            newMedia = oldMedia.replace("dark", "light");
          }

          rule.media.deleteMedium(oldMedia);
          rule.media.appendMedium(newMedia);
        }
      }
    } catch (e) {
      console.warn(
        `Stylesheet ${styleSheet.href} threw an error while toggling theme: `,
        e
      );
    }
  }
}

/**
 * Update any icons/UI based on the new theme.
 * (Replace this with whatever DOM logic you need.)
 * @param {string} theme - The new theme, "dark" or "light".
 */
function updateThemeIcons(theme) {
  if (theme === "dark") {
    document.getElementById("icon-moon").classList.remove("hidden");
    document.getElementById("icon-sun").classList.add("hidden");
  } else {
    document.getElementById("icon-moon").classList.add("hidden");
    document.getElementById("icon-sun").classList.remove("hidden");
  }
}

/**
 * Remove the user’s saved theme from local storage. If the user’s theme
 * was different from the system theme, flip the stylesheet rules back.
 */
function removeTheme() {
  const storedTheme = getStoredTheme();
  if (storedTheme) {
    const systemTheme = getSystemTheme();
    if (systemTheme !== storedTheme) {
      switchThemeRules();
    }
    localStorage.removeItem("theme");
  }
}

/**
 * Listen for system-level theme changes. If the user has manually chosen
 * a theme (i.e., localStorage has one), we flip the rules each time the
 * system changes, effectively overriding the system preference.
 */
window
  .matchMedia("(prefers-color-scheme: dark)")
  .addEventListener("change", (event) => {
    if (getStoredTheme()) {
      // Switch the stylesheet rules again to stay in user-chosen mode
      switchThemeRules();
    }
  });


// Expose toggleTheme to the window for easy hooking (e.g., button onclick)
window.toggleTheme = toggleTheme;
window.removeTheme = removeTheme;

window.HubroInit = function() {
	/**
	 * Initial page load: If there’s a stored theme and it doesn’t match
	 * the system theme, flip the rules so the user’s theme stays consistent.
	 */
	const storedTheme = getStoredTheme();
	if (storedTheme && storedTheme !== getSystemTheme()) {
		switchThemeRules();
	}
	// Update icons/UI to reflect whichever theme is active
	updateThemeIcons(getActiveTheme());
	checkBox = document.getElementById("dark-mode-toggle");
	checkBox.checked = getActiveTheme() === "dark";

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
