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

<style scoped>
.login-page {
  min-height: 100vh;
  display: grid;
  grid-template-columns: minmax(0, 1.15fr) minmax(420px, 0.85fr);
}

.login-brand {
  position: relative;
  overflow: hidden;
  padding: clamp(3rem, 8vw, 8rem);
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  color: #fff;
  background: radial-gradient(circle at 80% 10%, #29a98f 0, transparent 35%),
              linear-gradient(145deg, #073f38, #0b6b5c 62%, #075044);
}

.login-brand::after {
  content: "";
  position: absolute;
  width: 26rem;
  height: 26rem;
  right: -11rem;
  bottom: -12rem;
  border: 1px solid rgba(255, 255, 255, 0.18);
  border-radius: 50%;
  box-shadow:
    0 0 0 4rem rgba(255, 255, 255, 0.03),
    0 0 0 8rem rgba(255, 255, 255, 0.02);
}

.login-brand :deep(.eyebrow) { color: #a6f4df; }

.login-brand h1 {
  max-width: 780px;
  margin: 0;
  font-size: clamp(2.5rem, 5vw, 5rem);
  line-height: 1.02;
  letter-spacing: -0.055em;
  font-weight: 800;
}

.brand-copy {
  max-width: 620px;
  color: #d9eee9;
  font-size: 1.1rem;
  line-height: 1.7;
}

.security-note {
  display: flex;
  gap: 0.65rem;
  align-items: center;
  color: #d9eee9;
  font-size: 0.88rem;
}

.brand-mark {
  width: 2.6rem;
  height: 2.6rem;
  display: inline-grid;
  place-items: center;
  border-radius: var(--radius-md);
  font-weight: 900;
  background: #f4c95d;
  color: #173f37;
}

.login-panel {
  padding: 2rem;
  display: grid;
  place-items: center;
  background: #fcfcfd;
}

.login-card {
  width: min(100%, 430px);
  display: grid;
  gap: 1.25rem;
}

.login-card h2 {
  margin: 0;
  font-size: 2.1rem;
  letter-spacing: -0.04em;
  font-weight: 800;
}

.privacy-copy {
  color: #8490a3;
  text-align: center;
  font-size: 0.76rem;
  line-height: 1.5;
  margin: 0;
}

@media (max-width: 900px) {
  .login-page { grid-template-columns: 1fr; }
  .login-brand { min-height: 260px; padding: 2rem; gap: 2rem; }
  .login-brand h1 { font-size: 2.3rem; }
  .brand-copy, .security-note { display: none; }
  .login-panel { padding: 2rem 1.25rem 3rem; }
}
</style>
