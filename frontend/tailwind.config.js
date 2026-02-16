/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{vue,js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        bg: "#181825",
        surface: "#1e1e2e",
        overlay: "#313244",
        text: "#cdd6f4",
        subtext: "#a6adc8",
        accent: "#cba6f7",
        warn: "#f9e2af",
      }
    },
  },
  plugins: [],
}