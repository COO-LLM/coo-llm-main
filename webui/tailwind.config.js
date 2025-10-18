/** @type {import('tailwindcss').Config} */
module.exports = {
  darkMode: 'class', // Enable dark mode with class strategy
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        // COO-LLM brand colors from docs (Crimson Red theme)
        coo: {
          crimson: '#DC143C',   // Primary color from docs
          dark: '#C41235',      // Dark variant
          darker: '#B31030',    // Darker variant
          light: '#E41642',     // Light variant
        },
      },
    },
  },
  plugins: [],
}