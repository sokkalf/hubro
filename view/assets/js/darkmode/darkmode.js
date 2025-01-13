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
export function toggleTheme() {
  const currentTheme = getActiveTheme();
  const newTheme = currentTheme === "light" ? "dark" : "light";

  // Save the new theme
  localStorage.setItem("theme", newTheme);

  // Flip all prefers-color-scheme rules
  switchThemeRules();

  // Update any UI elements (icons, etc.)
  updateBodyClass(newTheme);
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

function updateBodyClass(theme) {
	body = document.querySelector("body");
	body.classList.remove("dark");
	body.classList.remove("light");
	body.classList.add(theme);
}

/**
 * Remove the user’s saved theme from local storage. If the user’s theme
 * was different from the system theme, flip the stylesheet rules back.
 */
export function removeTheme() {
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
 * Initial page load: If there’s a stored theme and it doesn’t match
 * the system theme, flip the rules so the user’s theme stays consistent.
 */
export function initTheme() {
	// Update icons/UI to reflect whichever theme is active
	updateBodyClass(getActiveTheme());
	updateThemeIcons(getActiveTheme());
	checkBox = document.getElementById("dark-mode-toggle");
	checkBox.checked = getActiveTheme() === "dark";
}
