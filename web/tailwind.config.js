/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        'hades-dark': '#0a0e27',
        'hades-darker': '#060818',
        'hades-primary': '#3b82f6',
        'hades-secondary': '#8b5cf6',
        'hades-accent': '#10b981',
        'hades-danger': '#ef4444',
        'hades-warning': '#f59e0b',
      }
    },
  },
  plugins: [],
}
