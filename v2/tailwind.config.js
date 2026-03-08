/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        surface: {
          50:  '#fafafa',
          100: '#f4f4f5',
          200: '#e4e4e7',
          300: '#d4d4d8',
          400: '#a1a1aa',
          500: '#71717a',
          600: '#52525b',
          700: '#3f3f46',
          800: '#27272a',
          900: '#18181b',
          950: '#0a0a0b',
        },
        accent: {
          DEFAULT: '#8b5cf6',
          hover:   '#7c3aed',
          light:   '#a78bfa',
          dark:    '#6d28d9',
        },
        vapor: {
          pink:   '#ff6b9d',
          cyan:   '#00d4ff',
          purple: '#8b5cf6',
          dark:   '#1a0a2e',
        },
      },
      animation: {
        'gradient':    'gradient 8s ease infinite',
        'float':       'float 6s ease-in-out infinite',
        'pulse-slow':  'pulse 4s cubic-bezier(0.4, 0, 0.6, 1) infinite',
        'glow':        'glow 2s ease-in-out infinite alternate',
        'bar':         'bar 0.8s ease-in-out infinite alternate',
      },
      keyframes: {
        gradient: {
          '0%, 100%': { backgroundPosition: '0% 50%' },
          '50%':      { backgroundPosition: '100% 50%' },
        },
        float: {
          '0%, 100%': { transform: 'translateY(0)' },
          '50%':      { transform: 'translateY(-8px)' },
        },
        glow: {
          '0%':   { boxShadow: '0 0 5px rgba(139, 92, 246, 0.4)' },
          '100%': { boxShadow: '0 0 20px rgba(139, 92, 246, 0.7)' },
        },
        bar: {
          '0%':   { transform: 'scaleY(0.4)' },
          '100%': { transform: 'scaleY(1)' },
        },
      },
    },
  },
  plugins: [],
}
