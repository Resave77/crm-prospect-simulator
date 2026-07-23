import { createApp } from 'vue'
import PrimeVue from 'primevue/config'
import Aura from '@primeuix/themes/aura'
import { definePreset } from '@primeuix/themes'
import 'primeicons/primeicons.css'
import './assets/main.css'
import App from './App.vue'
import router from './router'
import { pinia } from './stores/pinia'

const RefinedAura = definePreset(Aura, {
  semantic: {
    primary: {
      50: '#eef5ff',
      100: '#e0edff',
      200: '#c8dbff',
      300: '#a3c2ff',
      400: '#749fff',
      500: '#3b82f6',
      600: '#2563eb',
      700: '#1d4fd8',
      800: '#1e40af',
      900: '#1e3a8a',
      950: '#172554'
    },
    colorScheme: {
      light: {
        surface: {
          0: '#ffffff',
          50: '#f8fafc',
          100: '#f1f5f9',
          200: '#e2e8f0',
          300: '#d5dde9',
          400: '#9aa5b7',
          500: '#8490a3',
          600: '#52615d',
          700: '#374151',
          800: '#1f2937',
          900: '#111827',
          950: '#030712'
        },
        highlight: {
          background: '#eef5ff',
          focusBackground: '#e0edff',
          color: '#2563eb',
          focusColor: '#1d4fd8'
        }
      }
    },
    form: {
      border: {
        radius: '8px'
      }
    },
    list: {
      border: {
        radius: '8px'
      }
    },
    menu: {
      border: {
        radius: '8px'
      }
    },
    button: {
      border: {
        radius: '8px'
      }
    },
    tag: {
      border: {
        radius: '8px'
      }
    },
    badge: {
      border: {
        radius: '8px'
      }
    },
    input: {
      border: {
        radius: '8px'
      }
    }
  }
})

createApp(App)
  .use(pinia)
  .use(PrimeVue, { theme: { preset: RefinedAura } })
  .use(router)
  .mount('#app')
