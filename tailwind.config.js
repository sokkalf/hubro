/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
	  "./web/templates/**/*.gohtml",
	  "./web/templates/**/*.html",
	  "./web/layouts/**/*.gohtml",
	  "./web/layouts/**/*.html",
	  "./web/static/js/**/*.js",
	  "./**/*.go",
	  "./*.go",
  ],
  theme: {
    extend: {},
  },
  plugins: [],
}
