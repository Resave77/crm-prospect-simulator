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
    <!-- Glowing Accent Circles for Ambient Glass Effect -->
    <div class="blob blob-1"></div>
    <div class="blob blob-2"></div>

    <!-- Centered Glass Card -->
    <div class="login-center">
      <form class="login-card" @submit.prevent="submit">
        <div class="header-group">
          <div class="brand-mark-glass">Y</div>
          <h2>Welcome back</h2>
          <p class="muted">Sign in to access your workspace</p>
        </div>

        <Message v-if="error" severity="error" :closable="false" class="error-msg">{{ error }}</Message>

        <div class="field-group">
          <label class="field">
            <span>Email Address</span>
            <InputText 
              v-model="email" 
              type="email" 
              placeholder="name@company.com" 
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
              required 
              placeholder="••••••••" 
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
          Restricted access for authorized Administrators &amp; Sales Executives only.
        </p>
      </form>
    </div>
  </main>
</template>

<style scoped>
/* Base Page & Ambient Lighting */
.login-page {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  position: relative;
  overflow: hidden;
  background: #092621;
}

/* Glowing Background Blobs (Memberikan efek warna tembus pandang) */
.blob {
  position: absolute;
  border-radius: 50%;
  filter: blur(80px);
  opacity: 0.6;
}

.blob-1 {
  width: 350px;
  height: 350px;
  background: #2563eb;
  top: 15%;
  left: 20%;
}

.blob-2 {
  width: 400px;
  height: 400px;
  background: #10b981;
  bottom: 15%;
  right: 20%;
}

/* Container */
.login-center {
  position: relative;
  z-index: 10;
  width: 100%;
  max-width: 440px;
  padding: 1.5rem;
}

/* Glassmorphism Card Utama */
.login-card {
  background: rgba(255, 255, 255, 0.12);
  backdrop-filter: blur(24px);
  -webkit-backdrop-filter: blur(24px);
  border: 1px solid rgba(255, 255, 255, 0.2);
  border-radius: 28px;
  padding: 2.5rem 2.25rem;
  box-shadow: 0 30px 60px -12px rgba(0, 0, 0, 0.35),
              inset 0 1px 0 rgba(255, 255, 255, 0.2);
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
  color: #ffffff;
}

/* Brand Icon & Typography */
.header-group {
  text-align: center;
}

.brand-mark-glass {
  width: 56px;
  height: 56px;
  margin: 0 auto 1.25rem;
  display: grid;
  place-items: center;
  font-size: 1.8rem;
  font-weight: 900;
  color: #ffffff;
  background: rgba(255, 255, 255, 0.15);
  border: 1px solid rgba(255, 255, 255, 0.3);
  border-radius: 16px;
  box-shadow: 0 8px 16px rgba(0, 0, 0, 0.15);
}

.header-group h2 {
  margin: 0;
  font-size: 1.85rem;
  font-weight: 800;
  letter-spacing: -0.025em;
  color: #ffffff;
}

.muted {
  color: rgba(255, 255, 255, 0.7);
  margin: 0.4rem 0 0;
  font-size: 0.9rem;
}

/* Form Fields */
.field-group {
  display: flex;
  flex-direction: column;
  gap: 1.25rem;
}

.field {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.field span {
  font-size: 0.85rem;
  font-weight: 600;
  color: rgba(255, 255, 255, 0.9);
  letter-spacing: 0.02em;
}

/* Deep Style untuk Override Komponen PrimeVue ke Tampilan Glass */
:deep(.p-inputtext),
:deep(.p-password-input) {
  background: rgba(255, 255, 255, 0.08) !important;
  border: 1px solid rgba(255, 255, 255, 0.2) !important;
  color: #ffffff !important;
  border-radius: 12px !important;
  padding: 0.75rem 1rem !important;
  transition: all 0.2s ease;
}

:deep(.p-inputtext::placeholder),
:deep(.p-password-input::placeholder) {
  color: rgba(255, 255, 255, 0.4) !important;
}

:deep(.p-inputtext:focus),
:deep(.p-password-input:focus) {
  background: rgba(255, 255, 255, 0.15) !important;
  border-color: #3b82f6 !important;
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.3) !important;
}

:deep(.p-password-toggle-icon) {
  color: rgba(255, 255, 255, 0.6) !important;
}

/* Submit Button Override */
.submit-btn {
  margin-top: 0.5rem;
  border-radius: 12px !important;
  font-weight: 700 !important;
  background: #2563eb !important;
  border: none !important;
  box-shadow: 0 10px 20px -5px rgba(37, 99, 235, 0.5) !important;
}

.privacy-copy {
  text-align: center;
  color: rgba(255, 255, 255, 0.5);
  font-size: 0.75rem;
  line-height: 1.5;
  margin: 0;
}

.error-msg {
  border-radius: 12px;
}

@media (max-width: 480px) {
  .login-card {
    padding: 2rem 1.5rem;
    border-radius: 20px;
  }
}
</style>