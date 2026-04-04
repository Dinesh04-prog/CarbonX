/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        quant: {
          bg: '#0B0F19',       // Deepest background
          panel: '#131B2C',    // Card/Panel background
          border: '#1E293B',   // Subtle borders
          primary: '#1E66F5',  // The bright blue buttons
          accent: '#00FFA3',   // The neon green text/icons
          text: '#8B9BB4',     // Muted gray/blue text
          white: '#F8FAFC'     // Bright text
        }
      },
      fontFamily: {
        sans: ['Inter', 'system-ui', 'sans-serif'],
        mono: ['JetBrains Mono', 'Fira Code', 'monospace'],
      }
    },
  },
  plugins: [],
}