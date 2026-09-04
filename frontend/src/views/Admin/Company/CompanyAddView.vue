<script setup lang="ts">
import { computed, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import Button from 'primevue/button'
import InputText from 'primevue/inputtext'
import Textarea from 'primevue/textarea'
import Message from 'primevue/message'

const router = useRouter()
const error = ref('')
const saved = ref(false)
const saving = ref(false)

const form = reactive({
  name: '',
  termOfPayment: '',
  npwpName: '',
  npwpAddress: '',
  npwpNumber: '',
  address: '',
  notes: '',
})

const isFormValid = computed(() => form.name.trim() !== '')

async function handleSubmit() {
  if (!isFormValid.value) return
  saving.value = true
  try {
    await new Promise((r) => setTimeout(r, 1200))
    saved.value = true
  } catch {
    error.value = 'Failed to save company. Please try again.'
  } finally {
    saving.value = false
  }
}

function resetForm() {
  form.name = ''
  form.termOfPayment = ''
  form.npwpName = ''
  form.npwpAddress = ''
  form.npwpNumber = ''
  form.address = ''
  form.notes = ''
}
</script>

<template>
  <section class="admin-page">
    <!-- SUCCESS STATE -->
    <template v-if="saved">
      <div class="success-panel">
        <div class="success-icon">
          <i class="pi pi-check-circle" />
        </div>
        <h2>Company Created Successfully</h2>
        <p class="muted">The new company <strong>{{ form.name }}</strong> has been added to the system.</p>
        <div class="success-actions">
          <Button label="View Customers" icon="pi pi-list" @click="router.push('/admin/customers')" />
          <Button label="Add Another" icon="pi pi-plus" severity="secondary" outlined @click="saved = false; resetForm()" />
        </div>
      </div>
    </template>

    <!-- FORM -->
    <template v-else>
      <!-- PAGE HEADER -->
      <header class="page-heading">
        <div class="page-title-wrapper">
          <Button label="Back to Customer" icon="pi pi-arrow-left" text class="back-add-button" @click="router.push('/admin/customers')" />
          <span class="eyebrow">New Company</span>
          <h1>Add Company</h1>
          <p class="muted">Register a new corporate company entity into the CRM system.</p>
        </div>
        <div class="page-heading-actions">
          <Button label="Cancel" severity="secondary" text size="small" @click="router.push('/admin/customers')" />
          <Button label="Save Company" icon="pi pi-check" size="small" :loading="saving" :disabled="!isFormValid || saving" @click="handleSubmit" />
        </div>
      </header>

      <Message v-if="error" severity="error">{{ error }}</Message>

      <div class="form-layout">
        <!-- LEFT COLUMN: FORM -->
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
                <InputText v-model="form.name" placeholder="e.g. PT Yummy Food Indonesia" />
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

          <!-- ADDRESS -->
          <div class="form-card">
            <div class="form-card-header">
              <div class="form-card-icon si-emerald"><i class="pi pi-map" /></div>
              <div>
                <div class="section-title-line"><h3>Company Address</h3><span class="address-scope-badge">Company</span></div>
                <p>Legal and billing address used for company and tax records.</p>
              </div>
            </div>
            <div class="form-grid">
              <div class="form-field full">
                <div class="address-mode-label"><label>Search by Gmaps / Manual <span class="required">*</span></label><span>Manual</span></div>
                <Textarea v-model="form.address" :autoResize="true" rows="2" placeholder="Search or type company address" />
                <small class="address-helper">Manual entry is available. Add a Google Maps API key to enable autocomplete.</small>
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
                <Textarea v-model="form.notes" :autoResize="true" rows="3" placeholder="Internal notes, special instructions, etc." />
              </div>
            </div>
          </div>
        </div>

        <!-- RIGHT COLUMN: SIDEBAR -->
        <aside class="form-sidebar">
          <div class="sidebar-card">
            <h4>Submission Summary</h4>
            <div class="summary-list">
              <div class="summary-row">
                <span>Company Name</span>
                <strong>{{ form.name || '—' }}</strong>
              </div>
              <div class="summary-row">
                <span>Term of Payment</span>
                <strong>{{ form.termOfPayment || '—' }}</strong>
              </div>
              <div class="summary-row">
                <span>NPWP</span>
                <strong>{{ form.npwpNumber || '—' }}</strong>
              </div>
            </div>
          </div>
          <div class="sidebar-card tip-card">
            <i class="pi pi-info-circle" />
            <p>Company codes will be auto-generated upon save. Sites can be linked to this company afterward.</p>
          </div>
          <div class="sidebar-actions">
            <Button label="Save Company" icon="pi pi-check" class="full-width" :loading="saving" :disabled="!isFormValid || saving" @click="handleSubmit" />
            <Button label="Cancel" severity="secondary" text class="full-width" @click="router.push('/admin/customers')" />
          </div>
        </aside>
      </div>
    </template>
  </section>
</template>

<style scoped>
.admin-page {
  display: flex;
  flex-direction: column;
  gap: 1.25rem;
  padding: 1.75rem 2rem;
  min-height: 100vh;
}

/* ── PAGE HEADER ──────────────────────────────────────────────────── */
.page-heading {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  flex-wrap: wrap;
  gap: 1rem;
}
.page-title-wrapper {
  display: flex;
  flex-direction: column;
  gap: 0.15rem;
}
.page-title-wrapper .eyebrow {
  font-size: 0.68rem;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.1em;
  color: var(--brand-green-light, #0b7766);
  margin-top: 0.5rem;
}
.page-title-wrapper h1 {
  font-size: 1.65rem;
  font-weight: 800;
  color: var(--text-primary);
  margin: 0.2rem 0 0.15rem;
  letter-spacing: -0.03em;
}
.page-title-wrapper .muted {
  font-size: 0.85rem;
  color: var(--text-muted);
}
.page-heading-actions {
  display: flex;
  gap: 0.5rem;
  align-items: center;
  padding-top: 0.15rem;
}

/* ── FORM LAYOUT ──────────────────────────────────────────────────── */
.form-layout {
  display: grid;
  grid-template-columns: 1fr 320px;
  gap: 1.5rem;
  align-items: start;
}
.form-stack {
  display: flex;
  flex-direction: column;
  gap: 1.25rem;
}

/* ── FORM CARDS ────────────────────────────────────────────────────── */
.form-card {
  background: var(--surface-card);
  border: 1px solid var(--border-light);
  border-radius: var(--radius-lg);
  padding: 1.5rem;
  box-shadow: var(--shadow-xs);
}
.form-card-header {
  display: flex;
  align-items: flex-start;
  gap: 0.85rem;
  margin-bottom: 1.25rem;
  padding-bottom: 1rem;
  border-bottom: 1px solid #f0f3f7;
}
.form-card-icon {
  width: 40px;
  height: 40px;
  border-radius: var(--radius-md);
  display: grid;
  place-content: center;
  font-size: 1rem;
  flex-shrink: 0;
}
.si-blue { background: #fff0f1; color: #e63946; }
.si-emerald { background: #ecfdf5; color: #059669; }
.si-amber { background: #fffbeb; color: #d97706; }
.si-slate { background: #f1f5f9; color: #64748b; }

.form-card-header h3 {
  margin: 0;
  font-size: 0.95rem;
  font-weight: 700;
  color: var(--text-primary);
}
.form-card-header p {
  margin: 0.15rem 0 0;
  font-size: 0.78rem;
  color: var(--text-muted);
}

/* ── FORM GRID ─────────────────────────────────────────────────────── */
.form-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 1rem;
}
.form-field {
  display: flex;
  flex-direction: column;
  gap: 0.3rem;
}
.form-field.full {
  grid-column: 1 / -1;
}
.form-field label {
  font-size: 0.72rem;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: var(--text-muted);
}
.required {
  color: #dc2626;
}

/* ── SIDEBAR ───────────────────────────────────────────────────────── */
.form-sidebar {
  display: flex;
  flex-direction: column;
  gap: 1rem;
  position: sticky;
  top: 1.5rem;
}
.sidebar-card {
  background: var(--surface-card);
  border: 1px solid var(--border-light);
  border-radius: var(--radius-lg);
  padding: 1.15rem 1.25rem;
  box-shadow: var(--shadow-xs);
}
.sidebar-card h4 {
  margin: 0 0 0.85rem;
  font-size: 0.8rem;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: var(--text-muted);
}
.summary-list {
  display: flex;
  flex-direction: column;
  gap: 0.6rem;
}
.summary-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 0.5rem;
}
.summary-row span {
  font-size: 0.78rem;
  color: var(--text-muted);
}
.summary-row strong {
  font-size: 0.82rem;
  font-weight: 600;
  color: var(--text-primary);
  text-align: right;
}
.tip-card {
  display: flex;
  gap: 0.6rem;
  align-items: flex-start;
  background: #fff0f1;
  border-color: #f4b3ba;
}
.tip-card i {
  color: #e63946;
  margin-top: 0.1rem;
  font-size: 0.95rem;
  flex-shrink: 0;
}
.tip-card p {
  margin: 0;
  font-size: 0.78rem;
  color: #d62839;
  line-height: 1.5;
}
.sidebar-actions {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}
.full-width {
  width: 100%;
}

/* ── SUCCESS PANEL ─────────────────────────────────────────────────── */
.success-panel {
  min-height: 60vh;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 0.6rem;
  text-align: center;
  padding: 2rem;
}
.success-icon {
  font-size: 3rem;
  color: #059669;
  margin-bottom: 0.5rem;
}
.success-panel h2 {
  margin: 0;
  font-size: 1.5rem;
  font-weight: 800;
  color: var(--text-primary);
}
.success-panel .muted {
  max-width: 400px;
  font-size: 0.9rem;
  color: var(--text-muted);
}
.success-actions {
  display: flex;
  gap: 0.75rem;
  margin-top: 1rem;
}

/* ── RESPONSIVE ────────────────────────────────────────────────────── */
@media (max-width: 1024px) {
  .form-layout {
    grid-template-columns: 1fr;
  }
  .form-sidebar {
    position: static;
    order: -1;
  }
}
@media (max-width: 768px) {
  .admin-page { padding: 1.25rem 1rem; }
  .page-heading { flex-direction: column; }
  .form-grid { grid-template-columns: 1fr; }
  .form-field.full { grid-column: 1; }
  .success-actions { flex-direction: column; width: 100%; }
}
.back-add-button{display:inline-flex;margin:0 12px 8px 0;padding:0 12px 0 0;border-right:1px solid #e2e8f0;border-radius:0;background:#fff0f1;color:#64748b;font-size:12px}.back-add-button .p-button-icon{font-size:13px}.page-heading{align-items:center;min-height:48px;padding:0 32px;border-bottom:1px solid #e2e8f0}.page-title-wrapper{display:flex;align-items:center;gap:8px}.page-title-wrapper .eyebrow{display:none}.page-title-wrapper h1{font-size:17px;font-weight:700}.page-title-wrapper .muted{font-size:11px;color:#94a3b8}.page-heading-actions :deep(.p-button){height:32px;border-radius:8px;font-size:12px}
.address-scope-badge{display:inline-flex;align-items:center;height:22px;margin:3px 0 0;padding:0 8px;border:1px solid #bfdbfe;border-radius:999px;background:#eff6ff;color:#1d4ed8;font-size:10px;font-weight:700}.form-card-header:has(.address-scope-badge){align-items:flex-start}.form-card-header:has(.address-scope-badge) p{margin-top:4px}.address-mode-label{display:flex;align-items:center;justify-content:space-between;gap:10px}.address-mode-label>span{display:inline-flex;align-items:center;height:26px;padding:0 9px;border:1px solid #dbe3ee;border-radius:999px;background:#f8fafc;color:#64748b;font-size:10px;font-weight:600}.address-helper{display:block;color:#64748b;font-size:11px;line-height:1.4}@media(max-width:560px){.address-mode-label{align-items:flex-start;flex-direction:column;gap:6px}.address-mode-label>span{align-self:flex-end}}
</style>
