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
      50: '#fff1f2',
      100: '#ffd9dd',
      200: '#ffc7cd',
      300: '#f49ba5',
      400: '#e47781',
      500: '#df5a66',
      600: '#d14350',
      700: '#bb3342',
      800: '#a92c39',
      900: '#922833',
      950: '#591923'
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
          background: '#fff1f2',
          focusBackground: '#ffd9dd',
          color: '#d14350',
          focusColor: '#bb3342'
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
