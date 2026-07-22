<script setup lang="ts">
import { ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import Button from 'primevue/button'
import InputText from 'primevue/inputtext'
import Message from 'primevue/message'
import Password from 'primevue/password'
import { useAuthStore } from '../../stores/auth'

const auth = useAuthStore()
const route = useRoute()
const router = useRouter()
const email = ref('')
const password = ref('')
const error = ref('')

async function submit() {
  error.value = ''
  try {
    const user = await auth.login(email.value, password.value)
    const intended = typeof route.query.redirect === 'string' && route.query.redirect.startsWith('/')
      ? route.query.redirect
      : null
    const roleHome = user.role === 'ADMINISTRATOR' ? '/admin/dashboard' : '/sales/dashboard'
    await router.replace(intended ?? roleHome)
  } catch (caught) {
    error.value = auth.errorMessage(caught)
  }
}
</script>

<template>
  <main class="login-page">
    <section class="login-brand" aria-label="Product introduction">
      <div class="brand-mark">Y</div>
      <div>
        <p class="eyebrow">Enterprise field sales</p>
        <h1>Turn every qualified place into a lasting customer relationship.</h1>
        <p class="brand-copy">Prospect discovery, accountable field visits, and customer attendance in one controlled workflow.</p>
      </div>
      <div class="security-note"><i class="pi pi-shield" aria-hidden="true" /> Secure role-based workspace</div>
    </section>

    <section class="login-panel">
      <form class="login-card" @submit.prevent="submit">
        <div>
          <p class="eyebrow">Yummy CRM</p>
          <h2>Welcome back</h2>
          <p class="muted">Sign in with your assigned company account.</p>
        </div>

        <Message v-if="error" severity="error" :closable="false">{{ error }}</Message>

        <label class="field">
          <span>Email address</span>
          <InputText v-model="email" type="email" autocomplete="username" placeholder="name@yummy.test" required fluid />
        </label>
        <label class="field">
          <span>Password</span>
          <Password v-model="password" autocomplete="current-password" :feedback="false" toggle-mask required fluid />
        </label>

        <Button type="submit" label="Sign in" icon="pi pi-arrow-right" icon-pos="right" :loading="auth.loading" fluid />
        <p class="privacy-copy">Access is restricted to authorized Administrators and Sales Executives.</p>
      </form>
    </section>
  </main>
</template>
