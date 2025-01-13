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

  // Update any UI elements (icons, etc.)
  updateBodyClass(newTheme);
  updateThemeIcons(newTheme);
}

/**
 * Update any icons/UI based on the new theme.
 * (Replace this with whatever DOM logic you need.)
 * @param {string} theme - The new theme, "dark" or "light".
 */
function updateThemeIcons(theme) {
  if (theme === "dark") {
    document.getElementById("icon-moon").classList.remove("bw");
    document.getElementById("icon-sun").classList.add("bw");
  } else {
    document.getElementById("icon-moon").classList.add("bw");
    document.getElementById("icon-sun").classList.remove("bw");
  }
}

function updateBodyClass(theme) {
	body = document.querySelector("body");
	html = document.querySelector("html");
	body.classList.remove("dark");
	body.classList.remove("light");
	body.classList.add(theme);
	html.removeAttribute("data-theme");
	html.setAttribute("data-theme", theme);
}

/**
 * Remove the user’s saved theme from local storage. If the user’s theme
 * was different from the system theme, flip the stylesheet rules back.
 */
export function removeTheme() {
  const storedTheme = getStoredTheme();
  if (storedTheme) {
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
