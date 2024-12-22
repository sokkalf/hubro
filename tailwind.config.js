/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
	  "./templates/**/*.gohtml",
	  "./templates/**/*.html",
	  "./static/js/**/*.js",
	  "./**/*.go",
	  "./*.go",
  ],
  theme: {
    extend: {},
  },
  plugins: [],
}
