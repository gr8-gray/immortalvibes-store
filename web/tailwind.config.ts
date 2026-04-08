import type { Config } from 'tailwindcss';

export default {
  content: ['./src/**/*.{html,js,svelte,ts}'],
  theme: {
    extend: {
      colors: {
        void: '#030308',
        space: '#08080f',
        navy: '#0A0E2A',
        lunar: '#F0EDE6',
        gold: '#C8922A',
        'earth-blue': '#4FC3F7'
      },
      fontFamily: {
        display: ['"Cormorant Garamond"', 'serif'],
        body: ['Inter', 'sans-serif']
      },
      letterSpacing: {
        hero: '0.6em'
      },
      backdropBlur: {
        nav: '12px'
      }
    }
  },
  plugins: []
} satisfies Config;
