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
        navy: {
          950: '#060A13',
          900: '#0B1120',
          850: '#0F172A',
          800: '#15213D',
          700: '#1E2F56',
          600: '#2A3F71',
        },
        purple: {
          950: '#1E1035',
          900: '#2E1065',
          800: '#4C1D95',
          700: '#6D28D9',
          600: '#7C3AED',
          500: '#8B5CF6',
          400: '#A78BFA',
          300: '#C4B5FD',
        },
      },
    },
  },
  plugins: [],
}
