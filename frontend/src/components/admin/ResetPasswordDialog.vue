<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import Button from 'primevue/button'
import Dialog from 'primevue/dialog'
import InputText from 'primevue/inputtext'
import Message from 'primevue/message'
import SelectButton from 'primevue/selectbutton'
import Tag from 'primevue/tag'
import { useToast } from 'primevue/usetoast'
import { useAdminStore } from '../../stores/admin'
import { useAuthStore } from '../../stores/auth'
import type { AdminResetPasswordMode, AdminResetPasswordResult, AdminUserListItem } from '../../types/admin'
import { adminRoleLabel, adminRoleSeverity, adminTemporaryPasswordError } from '../../utils/admin'

const props = defineProps<{
  visible: boolean
  user: AdminUserListItem | null
}>()

const emit = defineEmits<{ (e: 'update:visible', value: boolean): void }>()

const store = useAdminStore()
const auth = useAuthStore()
const toast = useToast()

const mode = ref<AdminResetPasswordMode>('AUTO')
const tempPassword = ref('')
const confirmPassword = ref('')
const showPasswords = ref(false)
const showResultPassword = ref(false)
const error = ref('')
const result = ref<AdminResetPasswordResult | null>(null)
const submitting = ref(false)

const modeOptions = [
  { label: 'Generate automatically', value: 'AUTO' },
  { label: 'Enter manually', value: 'MANUAL' },
]

const isSelf = computed(() => props.user?.id === auth.user?.id)

const initials = computed(() => {
  const name = props.user?.fullName?.trim() ?? ''
  if (!name) return '?'
  return name.split(/\s+/).slice(0, 2).map((part) => part[0]?.toUpperCase() ?? '').join('')
})

const fieldErrors = computed(() => {
  if (mode.value !== 'MANUAL') return {}
  const passwordError = adminTemporaryPasswordError(tempPassword.value)
  if (passwordError) return { tempPassword: passwordError }
  if (!confirmPassword.value) return { confirmPassword: 'Confirm the temporary password by entering it again.' }
  if (confirmPassword.value !== tempPassword.value) return { confirmPassword: 'The passwords do not match.' }
  return {}
})

const canSubmit = computed(() => !submitting.value && (mode.value === 'AUTO' || Object.keys(fieldErrors.value).length === 0))

function clearAll() {
  mode.value = 'AUTO'
  tempPassword.value = ''
  confirmPassword.value = ''
  showPasswords.value = false
  showResultPassword.value = false
  error.value = ''
  result.value = null
  submitting.value = false
}

function updateVisible(value: boolean) {
  emit('update:visible', value)
}

watch(() => props.visible, (visible) => { if (visible) clearAll() })

async function submit() {
  if (!canSubmit.value || !props.user) return
  submitting.value = true
  error.value = ''
  try {
    const payload = mode.value === 'MANUAL'
      ? { mode: 'MANUAL' as const, temporaryPassword: tempPassword.value }
      : { mode: 'AUTO' as const }
    const res = await store.resetPassword(props.user.id, payload)
    if (!res) return
    result.value = res
    tempPassword.value = ''
    confirmPassword.value = ''
  } catch (e) {
    error.value = store.errorMessage(e)
  } finally {
    submitting.value = false
  }
}

async function copyPassword() {
  if (!result.value) return
  try {
    await navigator.clipboard.writeText(result.value.temporaryPassword)
    toast.add({ severity: 'success', summary: 'Temporary password copied', life: 3000 })
  } catch {
    toast.add({ severity: 'error', summary: 'Copy failed', detail: 'Copy the temporary password manually from the field below.', life: 5000 })
  }
}

function close() {
  emit('update:visible', false)
}
</script>

<template>
  <Dialog
    :visible="visible"
    header="Reset Password"
    modal
    :style="{ width: '520px' }"
    :breakpoints="{ '575px': '96vw' }"
    :closable="!submitting"
    :close-on-escape="!submitting"
    :dismissable-mask="!submitting"
    @update:visible="updateVisible"
    @hide="clearAll"
  >
    <!-- SUCCESS -->
    <template v-if="result">
      <div class="reset-banner">
        <i class="pi pi-check-circle" />
        <div>
          <strong>Password reset successful</strong>
          <span>Use the temporary password below to sign in.</span>
        </div>
      </div>

      <div class="identity-strip">
        <div class="avatar">{{ initials }}</div>
        <div class="identity-meta">
          <strong>{{ user?.fullName }}</strong>
          <span><code>{{ user?.employeeId }}</code> · {{ user?.email }}</span>
        </div>
        <Tag :value="adminRoleLabel(user?.role ?? '')" :severity="adminRoleSeverity(user?.role ?? '')" />
      </div>

      <div class="field-block">
        <label class="field-label">Temporary Password</label>
        <div class="password-field">
          <InputText :type="showResultPassword ? 'text' : 'password'" :model-value="result.temporaryPassword" readonly fluid />
          <Button :icon="showResultPassword ? 'pi pi-eye-slash' : 'pi pi-eye'" severity="secondary" text rounded :title="showResultPassword ? 'Hide' : 'Show'" @click="showResultPassword = !showResultPassword" />
        </div>
        <small class="field-hint">Shown only once and not stored anywhere. Copy it now or close the dialog.</small>
      </div>

      <div class="result-rows">
        <div class="result-row">
          <span class="result-label">Must change password</span>
          <Tag :value="result.mustChangePassword ? 'Yes' : 'No'" :severity="result.mustChangePassword ? 'warn' : 'secondary'" :icon="result.mustChangePassword ? 'pi pi-key' : ''" />
        </div>
        <div class="result-row">
          <span class="result-label">Active sessions revoked</span>
          <span class="result-value">{{ result.sessionsRevoked }}</span>
        </div>
      </div>

      <Message v-if="isSelf" severity="warn" :closable="false">
        You reset your own password. Your current session may be invalidated.
      </Message>
    </template>

    <!-- FORM -->
    <template v-else>
      <div class="identity-strip">
        <div class="avatar">{{ initials }}</div>
        <div class="identity-meta">
          <strong>{{ user?.fullName }}</strong>
          <span><code>{{ user?.employeeId }}</code> · {{ user?.email }}</span>
        </div>
        <Tag :value="adminRoleLabel(user?.role ?? '')" :severity="adminRoleSeverity(user?.role ?? '')" />
      </div>

      <Message severity="info" :closable="false">
        This replaces the user's current password and revokes all active sessions. The temporary password will be shown only once.
      </Message>

      <Message v-if="isSelf" severity="warn" :closable="false">
        You are resetting your own password. Your current session will be revoked after this reset.
      </Message>

      <Message v-if="user?.status === 'INACTIVE'" severity="info" :closable="false">
        Resetting the password does not activate this account.
      </Message>

      <div class="field-block">
        <label class="field-label">Reset Mode</label>
        <SelectButton v-model="mode" :options="modeOptions" option-label="label" option-value="value" :allow-empty="false" />
        <small class="field-hint">
          {{ mode === 'AUTO'
            ? 'A secure temporary password will be generated for the user.'
            : 'Enter the temporary password you want to set for the user.' }}
        </small>
      </div>

      <template v-if="mode === 'MANUAL'">
        <div class="field-block">
          <label class="field-label">Temporary Password</label>
          <div class="password-field">
            <InputText v-model="tempPassword" :type="showPasswords ? 'text' : 'password'" autocomplete="new-password" placeholder="Temporary password" fluid />
            <Button :icon="showPasswords ? 'pi pi-eye-slash' : 'pi pi-eye'" severity="secondary" text rounded :title="showPasswords ? 'Hide' : 'Show'" @click="showPasswords = !showPasswords" />
          </div>
          <small v-if="fieldErrors.tempPassword" class="field-error">{{ fieldErrors.tempPassword }}</small>
        </div>

        <div class="field-block">
          <label class="field-label">Confirm Temporary Password</label>
          <div class="password-field">
            <InputText v-model="confirmPassword" :type="showPasswords ? 'text' : 'password'" autocomplete="new-password" placeholder="Re-enter temporary password" fluid />
          </div>
          <small v-if="fieldErrors.confirmPassword" class="field-error">{{ fieldErrors.confirmPassword }}</small>
        </div>
      </template>

      <Message v-if="error" severity="error" :closable="false">{{ error }}</Message>
    </template>

    <template #footer>
      <template v-if="result">
        <Button label="Close" icon="pi pi-times" severity="secondary" text @click="close" />
        <Button label="Copy Password" icon="pi pi-copy" @click="copyPassword" />
      </template>
      <template v-else>
        <Button label="Cancel" severity="secondary" text :disabled="submitting" @click="close" />
        <Button label="Reset Password" icon="pi pi-key" :loading="submitting" :disabled="!canSubmit" @click="submit" />
      </template>
    </template>
  </Dialog>
</template>

<style scoped>
.reset-banner {
  display: flex;
  align-items: flex-start;
  gap: 0.75rem;
  background: #ecfdf5;
  border: 1px solid #a7f3d0;
  border-radius: var(--radius-md, 10px);
  padding: 1rem;
  margin-bottom: 1.1rem;
}
.reset-banner i {
  font-size: 1.25rem;
  color: #059669;
}
.reset-banner strong {
  display: block;
  font-size: 0.9rem;
  color: #065f46;
}
.reset-banner span {
  font-size: 0.78rem;
  color: #047857;
}

.identity-strip {
  display: flex;
  align-items: center;
  gap: 0.7rem;
  padding: 0.85rem 0.9rem;
  border: 1px solid var(--border-light, #e5e9f0);
  border-radius: var(--radius-md, 10px);
  background: var(--surface-subtle, #f8fafc);
  margin-bottom: 1.1rem;
}
.avatar {
  width: 38px;
  height: 38px;
  flex-shrink: 0;
  border-radius: 50%;
  background: var(--brand-blue, #2563eb);
  color: #fff;
  display: grid;
  place-content: center;
  font-size: 0.8rem;
  font-weight: 700;
}
.identity-meta {
  display: flex;
  flex-direction: column;
  gap: 0.1rem;
  min-width: 0;
  flex: 1;
}
.identity-meta strong {
  font-size: 0.9rem;
  color: var(--text-primary, #1e293b);
}
.identity-meta span {
  font-size: 0.78rem;
  color: var(--text-muted, #64748b);
  word-break: break-all;
}
.identity-meta code {
  font-family: 'SF Mono', 'Fira Code', 'Consolas', monospace;
  background: #f1f5f9;
  padding: 0.1rem 0.4rem;
  border-radius: 4px;
}

.field-block {
  display: flex;
  flex-direction: column;
  gap: 0.45rem;
  margin-bottom: 1rem;
}
.field-label {
  font-size: 0.78rem;
  font-weight: 700;
  color: var(--text-primary, #1e293b);
}
.field-hint {
  font-size: 0.74rem;
  color: var(--text-muted, #64748b);
}
.field-error {
  font-size: 0.74rem;
  color: #dc2626;
}

.password-field {
  display: flex;
  align-items: center;
  gap: 0.4rem;
}

.result-rows {
  display: flex;
  flex-direction: column;
  gap: 0.65rem;
  padding: 0.9rem 1rem;
  border: 1px solid var(--border-light, #e5e9f0);
  border-radius: var(--radius-md, 10px);
  margin-bottom: 1.1rem;
}
.result-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.result-label {
  font-size: 0.75rem;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: var(--text-muted, #64748b);
}
.result-value {
  font-size: 0.85rem;
  font-weight: 700;
  color: var(--text-primary, #1e293b);
}
</style>
