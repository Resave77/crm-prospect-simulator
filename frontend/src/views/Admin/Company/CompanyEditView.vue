<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import Button from 'primevue/button'
import InputText from 'primevue/inputtext'
import Textarea from 'primevue/textarea'
import Message from 'primevue/message'
import { useCustomerListStore } from '../../../stores/customerList'

const route = useRoute()
const router = useRouter()
const store = useCustomerListStore()
const error = ref('')
const loading = ref(true)
const saving = ref(false)
const saved = ref(false)

const code = computed(() => route.params.id as string)
const companyName = computed(() => {
  const match = store.allCustomers.find((c) => c.parentCode === code.value)
  return match?.parentCompanyName || code.value
})

const form = ref({
  name: '',
  address: '',
  npwpName: '',
  npwpAddress: '',
  npwpNumber: '',
  termOfPayment: '',
  notes: '',
})

const isFormValid = computed(() => form.value.name.trim() !== '')

async function handleSave() {
  if (!isFormValid.value) return
  saving.value = true
  try {
    await new Promise((r) => setTimeout(r, 1000))
    saved.value = true
  } catch {
    error.value = 'Failed to update company. Please try again.'
  } finally {
    saving.value = false
  }
}

onMounted(async () => {
  try {
    if (store.allCustomers.length === 0) await store.fetchCustomers()
    form.value.name = companyName.value
  } catch (e) { error.value = store.errorMessage(e) }
  finally { loading.value = false }
})
</script>

<template>
  <section class="admin-page">
    <div v-if="loading" class="state-box">
      <i class="pi pi-spin pi-spinner state-icon" />
      <span>Loading company data...</span>
    </div>

    <Message v-if="error && !saved" severity="error" :closable="false">{{ error }}</Message>

    <!-- SUCCESS -->
    <template v-if="saved">
      <div class="success-panel">
        <div class="success-icon"><i class="pi pi-check-circle" /></div>
        <h2>Company Updated Successfully</h2>
        <p class="muted">Changes to <strong>{{ form.name }}</strong> have been saved.</p>
        <div class="success-actions">
          <Button label="Back to Company List" icon="pi pi-list" @click="router.push('/admin/companies')" />
          <Button label="View Detail" icon="pi pi-eye" severity="secondary" outlined @click="router.push(`/admin/companies/${code}`)" />
        </div>
      </div>
    </template>

    <!-- FORM -->
    <template v-else-if="!loading">
      <header class="page-heading">
        <div class="page-title-wrapper">
          <button class="back-link" @click="router.push(`/admin/companies/${code}`)">
            <i class="pi pi-arrow-left" /> Back to Company Detail
          </button>
          <span class="eyebrow">Edit Company</span>
          <h1>{{ companyName }}</h1>
          <div class="subtitle-row">
            <code class="code-tag">{{ code }}</code>
          </div>
        </div>
        <div class="page-heading-actions">
          <Button label="Cancel" severity="secondary" text size="small" @click="router.push(`/admin/companies/${code}`)" />
          <Button label="Save Changes" icon="pi pi-check" size="small" :loading="saving" :disabled="!isFormValid || saving" @click="handleSave" />
        </div>
      </header>

      <div class="form-layout">
        <div class="form-stack">
          <!-- COMPANY INFO -->
          <div class="form-card">
            <div class="form-card-header">
              <div class="form-card-icon si-blue"><i class="pi pi-building" /></div>
              <div>
                <h3>Company Information</h3>
                <p>Basic corporate entity details.</p>
              </div>
            </div>
            <div class="form-grid">
              <div class="form-field full">
                <label>Company Name <span class="required">*</span></label>
                <InputText v-model="form.name" placeholder="Company name" />
              </div>
              <div class="form-field">
                <label>Term of Payment</label>
                <InputText v-model="form.termOfPayment" placeholder="e.g. NET 30" />
              </div>
            </div>
          </div>

          <!-- NPWP / TAX -->
          <div class="form-card">
            <div class="form-card-header">
              <div class="form-card-icon si-amber"><i class="pi pi-id-card" /></div>
              <div>
                <h3>Tax Information (NPWP)</h3>
                <p>Indonesian tax registration details.</p>
              </div>
            </div>
            <div class="form-grid">
              <div class="form-field full">
                <label>NPWP Name</label>
                <InputText v-model="form.npwpName" placeholder="Registered tax name" />
              </div>
              <div class="form-field full">
                <label>NPWP Address</label>
                <Textarea v-model="form.npwpAddress" :autoResize="true" rows="2" placeholder="Tax registered address" />
              </div>
              <div class="form-field">
                <label>NPWP Number</label>
                <InputText v-model="form.npwpNumber" placeholder="00.000.000.0-000.000" />
              </div>
            </div>
          </div>

          <!-- NOTES -->
          <div class="form-card">
            <div class="form-card-header">
              <div class="form-card-icon si-slate"><i class="pi pi-file-edit" /></div>
              <div>
                <h3>Additional Notes</h3>
                <p>Internal notes about this company.</p>
              </div>
            </div>
            <div class="form-grid">
              <div class="form-field full">
                <label>Notes</label>
                <Textarea v-model="form.notes" :autoResize="true" rows="3" placeholder="Internal notes..." />
              </div>
            </div>
          </div>
        </div>

        <!-- SIDEBAR -->
        <aside class="form-sidebar">
          <div class="sidebar-card">
            <h4>Summary</h4>
            <div class="summary-list">
              <div class="summary-row">
                <span>Company Name</span>
                <strong>{{ form.name || '—' }}</strong>
              </div>
              <div class="summary-row">
                <span>Code</span>
                <code class="code-tag code-blue">{{ code }}</code>
              </div>
              <div class="summary-row">
                <span>NPWP</span>
                <strong>{{ form.npwpNumber || '—' }}</strong>
              </div>
            </div>
          </div>
          <div class="sidebar-actions">
            <Button label="Save Changes" icon="pi pi-check" class="full-width" :loading="saving" :disabled="!isFormValid || saving" @click="handleSave" />
            <Button label="Cancel" severity="secondary" text class="full-width" @click="router.push(`/admin/companies/${code}`)" />
          </div>
        </aside>
      </div>
    </template>
  </section>
</template>

<style scoped>
.admin-page { display: flex; flex-direction: column; gap: 1.25rem; padding: 1.75rem 2rem; min-height: 100vh; }
.page-heading { display: flex; justify-content: space-between; align-items: flex-start; flex-wrap: wrap; gap: 1rem; }
.page-title-wrapper { display: flex; flex-direction: column; gap: 0.15rem; }
.page-title-wrapper .eyebrow { font-size: 0.68rem; font-weight: 700; text-transform: uppercase; letter-spacing: 0.1em; color: var(--brand-green-light, #0b7766); margin-top: 0.5rem; }
.page-title-wrapper h1 { font-size: 1.65rem; font-weight: 800; color: var(--text-primary); margin: 0.2rem 0 0.15rem; letter-spacing: -0.03em; }
.subtitle-row { display: flex; align-items: center; gap: 0.5rem; margin-top: 0.1rem; }
.page-heading-actions { display: flex; gap: 0.5rem; align-items: center; padding-top: 0.15rem; }
.back-link { display: inline-flex; align-items: center; gap: 0.35rem; color: var(--brand-blue); font-size: 0.8rem; font-weight: 600; cursor: pointer; background: none; border: none; padding: 0; font: inherit; transition: opacity 0.15s; }
.back-link:hover { opacity: 0.8; }
.code-tag { display: inline-block; font-family: 'SF Mono', 'Fira Code', 'Consolas', monospace; font-size: 0.78rem; font-weight: 600; padding: 0.15rem 0.5rem; border-radius: 4px; background: #f1f5f9; color: var(--text-secondary); }
.code-blue { background: #eff6ff; color: #2563eb; }

.form-layout { display: grid; grid-template-columns: 1fr 320px; gap: 1.5rem; align-items: start; }
.form-stack { display: flex; flex-direction: column; gap: 1.25rem; }
.form-card { background: var(--surface-card); border: 1px solid var(--border-light); border-radius: var(--radius-lg); padding: 1.5rem; box-shadow: var(--shadow-xs); }
.form-card-header { display: flex; align-items: flex-start; gap: 0.85rem; margin-bottom: 1.25rem; padding-bottom: 1rem; border-bottom: 1px solid #f0f3f7; }
.form-card-icon { width: 40px; height: 40px; border-radius: var(--radius-md); display: grid; place-content: center; font-size: 1rem; flex-shrink: 0; }
.si-blue { background: #eff6ff; color: #2563eb; }
.si-amber { background: #fffbeb; color: #d97706; }
.si-slate { background: #f1f5f9; color: #64748b; }
.form-card-header h3 { margin: 0; font-size: 0.95rem; font-weight: 700; color: var(--text-primary); }
.form-card-header p { margin: 0.15rem 0 0; font-size: 0.78rem; color: var(--text-muted); }
.form-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 1rem; }
.form-field { display: flex; flex-direction: column; gap: 0.3rem; }
.form-field.full { grid-column: 1 / -1; }
.form-field label { font-size: 0.72rem; font-weight: 700; text-transform: uppercase; letter-spacing: 0.05em; color: var(--text-muted); }
.required { color: #dc2626; }

.form-sidebar { display: flex; flex-direction: column; gap: 1rem; position: sticky; top: 1.5rem; }
.sidebar-card { background: var(--surface-card); border: 1px solid var(--border-light); border-radius: var(--radius-lg); padding: 1.15rem 1.25rem; box-shadow: var(--shadow-xs); }
.sidebar-card h4 { margin: 0 0 0.85rem; font-size: 0.8rem; font-weight: 700; text-transform: uppercase; letter-spacing: 0.05em; color: var(--text-muted); }
.summary-list { display: flex; flex-direction: column; gap: 0.6rem; }
.summary-row { display: flex; justify-content: space-between; align-items: center; gap: 0.5rem; }
.summary-row span { font-size: 0.78rem; color: var(--text-muted); }
.summary-row strong { font-size: 0.82rem; font-weight: 600; color: var(--text-primary); text-align: right; }
.sidebar-actions { display: flex; flex-direction: column; gap: 0.5rem; }
.full-width { width: 100%; }

.success-panel { min-height: 60vh; display: flex; flex-direction: column; align-items: center; justify-content: center; gap: 0.6rem; text-align: center; padding: 2rem; }
.success-icon { font-size: 3rem; color: #059669; margin-bottom: 0.5rem; }
.success-panel h2 { margin: 0; font-size: 1.5rem; font-weight: 800; color: var(--text-primary); }
.success-panel .muted { max-width: 400px; font-size: 0.9rem; color: var(--text-muted); }
.success-actions { display: flex; gap: 0.75rem; margin-top: 1rem; }

.state-box { min-height: 300px; display: flex; flex-direction: column; align-items: center; justify-content: center; gap: 0.5rem; color: var(--text-muted); }
.state-icon { font-size: 1.75rem; color: var(--brand-blue); }

@media (max-width: 1024px) { .form-layout { grid-template-columns: 1fr; } .form-sidebar { position: static; order: -1; } }
@media (max-width: 768px) { .admin-page { padding: 1.25rem 1rem; } .page-heading { flex-direction: column; } .form-grid { grid-template-columns: 1fr; } .form-field.full { grid-column: 1; } .success-actions { flex-direction: column; width: 100%; } }
</style>
