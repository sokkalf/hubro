/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
	  "./view/templates/**/*.gohtml",
	  "./view/templates/**/*.html",
	  "./view/layouts/**/*.gohtml",
	  "./view/layouts/**/*.html",
	  "./view/static/js/**/*.js",
	  "./**/*.go",
	  "./*.go",
  ],
  theme: {
    extend: {},
  },
  plugins: [],
}
