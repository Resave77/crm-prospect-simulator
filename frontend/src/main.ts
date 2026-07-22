import { createApp } from 'vue'
import PrimeVue from 'primevue/config'
import Aura from '@primeuix/themes/aura'
import 'primeicons/primeicons.css'
import './assets/main.css'
import App from './App.vue'
import router from './router'
import { pinia } from './stores/pinia'

createApp(App)
  .use(pinia)
  .use(PrimeVue, { theme: { preset: Aura } })
  .use(router)
  .mount('#app')
