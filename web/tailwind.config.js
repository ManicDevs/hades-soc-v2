/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
  ],
  darkMode: 'class',
  theme: {
    extend: {
      colors: {
        'hades-dark': 'var(--bg-primary)',
        'hades-darker': 'var(--bg-secondary)',
        'hades-primary': 'var(--accent-color)',
        'hades-secondary': '#8b5cf6',
        'hades-accent': '#10b981',
        'hades-danger': 'var(--danger-color)',
        'hades-warning': 'var(--warning-color)',
        'hades-card': 'var(--card-bg)',
        'hades-input': 'var(--input-bg)',
        'theme-border': 'var(--border-color)',
        'theme-text': 'var(--text-primary)',
        'theme-text-secondary': 'var(--text-secondary)',
      },
      backgroundColor: {
        'theme-primary': 'var(--bg-primary)',
        'theme-secondary': 'var(--bg-secondary)',
        'theme-tertiary': 'var(--bg-tertiary)',
      },
      textColor: {
        'theme-primary': 'var(--text-primary)',
        'theme-secondary': 'var(--text-secondary)',
      },
      borderColor: {
        'theme-border': 'var(--border-color)',
      },
      animation: {
        shimmer: 'shimmer 2s infinite linear',
      },
      keyframes: {
        shimmer: {
          '0%': { backgroundPosition: '-200% 0' },
          '100%': { backgroundPosition: '200% 0' },
        },
      },
    },
  },
  plugins: [],
}