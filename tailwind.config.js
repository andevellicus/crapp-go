// tailwind.config.js

/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    './views/**/*.templ', // Path to your templ files
    './*.go',             // Path to your Go files if they contain classes
  ],
  theme: {
    extend: {},
  },
  plugins: [],
}