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
    <div class="login-center">
      <form class="login-card" @submit.prevent="submit">
        <div class="header-group">
          <div class="brand-mark">Y</div>

          <div>
            <h2>Welcome back</h2>
            <p class="muted">Sign in to access your workspace</p>
          </div>
        </div>

        <Message
          v-if="error"
          severity="error"
          :closable="false"
          class="error-msg"
        >
          {{ error }}
        </Message>

        <div class="field-group">
          <label class="field">
            <span>Email address</span>

            <InputText
              v-model="email"
              type="email"
              placeholder="name@company.com"
              autocomplete="email"
              required
              fluid
            />
          </label>

          <label class="field">
            <span>Password</span>

            <Password
              v-model="password"
              :feedback="false"
              toggle-mask
              autocomplete="current-password"
              required
              placeholder="Enter your password"
              fluid
            />
          </label>
        </div>

        <Button
          type="submit"
          label="Sign in"
          icon="pi pi-arrow-right"
          icon-pos="right"
          :loading="auth.loading"
          size="large"
          fluid
          class="submit-btn"
        />

        <p class="privacy-copy">
          Restricted access for authorized Administrators and Sales Executives.
        </p>
      </form>
    </div>
  </main>
</template>

<style scoped>
.login-page {
  min-height: 100dvh;
  display: grid;
  place-items: center;
  padding: 1.5rem;
  background: #f1f5f9;
}

.login-center {
  width: 100%;
  max-width: 420px;
}

.login-card {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
  padding: 2.25rem;
  background: #ffffff;
  border: 1px solid #e2e8f0;
  border-radius: 20px;
  box-shadow:
    0 20px 45px rgba(15, 23, 42, 0.08),
    0 2px 8px rgba(15, 23, 42, 0.04);
}

.header-group {
  display: flex;
  align-items: center;
  gap: 1rem;
}

.brand-mark {
  width: 52px;
  height: 52px;
  flex-shrink: 0;
  display: grid;
  place-items: center;
  border-radius: 14px;
  background: #2563eb;
  color: #ffffff;
  font-size: 1.45rem;
  font-weight: 800;
  box-shadow: 0 8px 18px rgba(37, 99, 235, 0.2);
}

.header-group h2 {
  margin: 0;
  color: #0f172a;
  font-size: 1.55rem;
  font-weight: 800;
  letter-spacing: -0.025em;
}

.muted {
  margin: 0.3rem 0 0;
  color: #64748b;
  font-size: 0.875rem;
  line-height: 1.5;
}

.field-group {
  display: flex;
  flex-direction: column;
  gap: 1.15rem;
}

.field {
  display: flex;
  flex-direction: column;
  gap: 0.45rem;
}

.field > span {
  color: #334155;
  font-size: 0.82rem;
  font-weight: 700;
}

:deep(.p-inputtext),
:deep(.p-password-input) {
  width: 100%;
  padding: 0.78rem 0.9rem;
  background: #ffffff;
  border: 1px solid #cbd5e1;
  border-radius: 10px;
  color: #0f172a;
  font-size: 0.9rem;
  box-shadow: none;
  transition:
    border-color 150ms ease,
    box-shadow 150ms ease;
}

:deep(.p-inputtext::placeholder),
:deep(.p-password-input::placeholder) {
  color: #94a3b8;
}

:deep(.p-inputtext:hover),
:deep(.p-password-input:hover) {
  border-color: #94a3b8;
}

:deep(.p-inputtext:focus),
:deep(.p-password-input:focus) {
  border-color: #2563eb;
  box-shadow: 0 0 0 3px rgba(37, 99, 235, 0.12);
}

:deep(.p-password) {
  width: 100%;
}

:deep(.p-password-toggle-mask-icon),
:deep(.p-password-toggle-icon) {
  color: #64748b;
}

.submit-btn {
  margin-top: 0.25rem;
  border-radius: 10px !important;
  background: #2563eb !important;
  border-color: #2563eb !important;
  font-weight: 700 !important;
  box-shadow: 0 8px 18px rgba(37, 99, 235, 0.18) !important;
}

.submit-btn:hover {
  background: #1d4ed8 !important;
  border-color: #1d4ed8 !important;
}

.privacy-copy {
  margin: 0;
  color: #94a3b8;
  font-size: 0.72rem;
  line-height: 1.5;
  text-align: center;
}

.error-msg {
  border-radius: 10px;
}

@media (max-width: 480px) {
  .login-page {
    align-items: center;
    padding: 1rem;
  }

  .login-card {
    padding: 1.6rem;
    border-radius: 16px;
  }

  .header-group {
    align-items: flex-start;
  }

  .brand-mark {
    width: 46px;
    height: 46px;
    border-radius: 12px;
    font-size: 1.25rem;
  }

  .header-group h2 {
    font-size: 1.35rem;
  }
}
</style>