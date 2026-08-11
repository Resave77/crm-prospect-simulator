<script setup lang="ts">
import { computed, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useToast } from 'primevue/usetoast'
import Button from 'primevue/button'
import Message from 'primevue/message'
import Password from 'primevue/password'
import Toast from 'primevue/toast'
import { useAuthStore } from '../../stores/auth'
import { validateChangePassword } from '../../utils/changePasswordValidation'

const auth = useAuthStore()
const router = useRouter()
const toast = useToast()

const form = reactive({
  currentPassword: '',
  newPassword: '',
  confirmPassword: '',
})
const error = ref('')

const validation = computed(() => validateChangePassword(form))
const canSubmit = computed(() => validation.value.valid && !auth.changingPassword)

function clearPasswords() {
  form.currentPassword = ''
  form.newPassword = ''
  form.confirmPassword = ''
}

async function submit() {
  if (!canSubmit.value) return

  error.value = ''
  try {
    await auth.changePassword({
      currentPassword: form.currentPassword,
      newPassword: form.newPassword,
      confirmPassword: form.confirmPassword,
    })
    clearPasswords()
    toast.add({
      severity: 'success',
      summary: 'Password changed',
      detail: 'Password changed successfully. Please sign in again.',
      life: 3500,
    })
    await router.replace('/login?passwordChanged=1')
  } catch (caught) {
    error.value = auth.errorMessage(caught)
  }
}
</script>

<template>
  <main class="change-password-page">
    <Toast />

    <div class="change-password-center">
      <form class="change-password-card" @submit.prevent="submit">
        <div class="header-group">
          <div class="brand-mark">
            <img src="/yummy-logo.png" alt="Yummy Dairy" />
          </div>

          <div>
            <h2>Change Temporary Password</h2>
            <p class="muted">
              For security, replace your temporary password before continuing to the application.
            </p>
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
            <span>Current Password</span>

            <Password
              v-model="form.currentPassword"
              :feedback="false"
              toggle-mask
              autocomplete="current-password"
              required
              placeholder="Enter current password"
              fluid
            />
          </label>

          <label class="field">
            <span>New Password</span>

            <Password
              v-model="form.newPassword"
              :feedback="false"
              toggle-mask
              autocomplete="new-password"
              required
              placeholder="Enter new password"
              fluid
            />
          </label>

          <label class="field">
            <span>Confirm New Password</span>

            <Password
              v-model="form.confirmPassword"
              :feedback="false"
              toggle-mask
              autocomplete="new-password"
              required
              placeholder="Confirm new password"
              fluid
            />
          </label>
        </div>

        <ul class="requirements" aria-label="Password requirements">
          <li
            v-for="requirement in validation.requirements"
            :key="requirement.key"
            :class="{ valid: requirement.valid }"
          >
            <i :class="requirement.valid ? 'pi pi-check-circle' : 'pi pi-circle'" aria-hidden="true" />
            <span>{{ requirement.label }}</span>
          </li>
        </ul>

        <Button
          type="submit"
          label="Change password"
          icon="pi pi-lock"
          :disabled="!canSubmit"
          :loading="auth.changingPassword"
          size="large"
          fluid
          class="submit-btn"
        />
      </form>
    </div>
  </main>
</template>

<style scoped>
.change-password-page {
  min-height: 100dvh;
  display: grid;
  place-items: center;
  padding: 1.5rem;
  background: #f1f5f9;
}

.change-password-center {
  width: 100%;
  max-width: 460px;
}

.change-password-card {
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
  width: 72px;
  height: 52px;
  flex-shrink: 0;
  display: grid;
  place-items: center;
  padding: 0.35rem;
  border-radius: 14px;
  background: #ffffff;
  border: 1px solid #e2e8f0;
  box-shadow: 0 8px 18px rgba(15, 23, 42, 0.08);
}

.brand-mark img {
  width: 100%;
  height: 100%;
  object-fit: contain;
}

.header-group h2 {
  margin: 0;
  color: #0f172a;
  font-size: 1.45rem;
  font-weight: 800;
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

:deep(.p-inputtext:hover),
:deep(.p-password-input:hover) {
  border-color: #94a3b8;
}

:deep(.p-inputtext:focus),
:deep(.p-password-input:focus) {
  border-color: #d14350;
  box-shadow: 0 0 0 3px rgba(209, 67, 80, 0.12);
}

:deep(.p-password) {
  width: 100%;
}

:deep(.p-password-toggle-mask-icon),
:deep(.p-password-toggle-icon) {
  color: #64748b;
}

.requirements {
  display: grid;
  gap: 0.45rem;
  margin: 0;
  padding: 0;
  list-style: none;
}

.requirements li {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  color: #64748b;
  font-size: 0.82rem;
  line-height: 1.35;
}

.requirements li.valid {
  color: #166534;
}

.requirements i {
  font-size: 0.9rem;
}

.submit-btn {
  margin-top: 0.25rem;
  border-radius: 10px !important;
  background: #d14350 !important;
  border-color: #d14350 !important;
  font-weight: 700 !important;
  box-shadow: 0 8px 18px rgba(209, 67, 80, 0.18) !important;
}

.submit-btn:hover {
  background: #bb3342 !important;
  border-color: #bb3342 !important;
}

.error-msg {
  border-radius: 10px;
}

@media (max-width: 480px) {
  .change-password-page {
    padding: 1rem;
  }

  .change-password-card {
    padding: 1.6rem;
    border-radius: 16px;
  }

  .header-group {
    align-items: flex-start;
  }

  .brand-mark {
    width: 66px;
    height: 46px;
    border-radius: 12px;
  }

  .header-group h2 {
    font-size: 1.3rem;
  }
}
</style>
