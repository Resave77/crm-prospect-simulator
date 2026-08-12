import { createApp } from 'vue'
import PrimeVue from 'primevue/config'
import Aura from '@primeuix/themes/aura'
import { definePreset } from '@primeuix/themes'
import ToastService from 'primevue/toastservice'
import 'primeicons/primeicons.css'
import './assets/main.css'
import App from './App.vue'
import router from './router'
import { pinia } from './stores/pinia'

const RefinedAura = definePreset(Aura, {
  semantic: {
    primary: {
      50: '#fff0f1',
      100: '#ffd9dc',
      200: '#fdb3b9',
      300: '#f98a94',
      400: '#f36070',
      500: '#e63946',
      600: '#d62839',
      700: '#c12334',
      800: '#a51e2d',
      900: '#871c28',
      950: '#5c1520'
    },
    colorScheme: {
      light: {
        surface: {
          0: '#ffffff',
          50: '#fcf9f9',
          100: '#f7f1f2',
          200: '#eee3e5',
          300: '#dccdd0',
          400: '#a39397',
          500: '#8d7d81',
          600: '#625357',
          700: '#493c40',
          800: '#352b2e',
          900: '#241c1e',
          950: '#140e10'
        },
        highlight: {
          background: '#fff0f1',
          focusBackground: '#ffd9dc',
          color: '#e63946',
          focusColor: '#d62839'
        }
      }
    },
    formField: {
      borderRadius: '8px'
    }
  }
})

createApp(App)
  .use(pinia)
  .use(PrimeVue, {
    theme: {
      preset: RefinedAura,
      options: {
        darkModeSelector: false
      }
    }
  })
  .use(router)
  .use(ToastService)
  .mount('#app')
