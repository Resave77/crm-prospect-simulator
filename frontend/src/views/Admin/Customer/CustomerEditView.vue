<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import Button from 'primevue/button'
import InputText from 'primevue/inputtext'
import Select from 'primevue/select'
import Textarea from 'primevue/textarea'
import Message from 'primevue/message'
import { useCrmStore } from '../../../stores/crm'
import { useCustomerListStore } from '../../../stores/customerList'
import type { CustomerDetail, Contact } from '../../../types/crm'

const route = useRoute()
const router = useRouter()
const crm = useCrmStore()
const listStore = useCustomerListStore()
const error = ref('')
const detail = ref<CustomerDetail | null>(null)
const loading = ref(true)
const saving = ref(false)
const saved = ref(false)

const segmentOptions = [
  { label: 'Select segment', value: '' },
  { label: 'Key Account', value: 'Key Account' },
  { label: 'Modern Trade', value: 'Modern Trade' },
  { label: 'Food Service', value: 'Food Service' },
  { label: 'General Trade', value: 'General Trade' },
]
const categoryOptions = [
  { label: 'Select category', value: '' },
  { label: 'HORECA', value: 'HORECA' },
  { label: 'Retail', value: 'Retail' },
  { label: 'Institution', value: 'Institution' },
  { label: 'Distributor', value: 'Distributor' },
]
const regionOptions = computed(() => {
  const regs = listStore.filterOptions?.regions ?? []
  return [{ label: 'Select region', value: '' }, ...regs.map((r) => ({ label: r, value: r }))]
})
const salesOptions = computed(() => {
  const sales = listStore.filterOptions?.salesExecutives ?? []
  return [{ label: 'Select sales executive', value: '' }, ...sales.map((s) => ({ label: s.fullName, value: s.id }))]
})

const form = reactive({
  name: '',
  customerSegment: '',
  customerCategory: '',
  region: '',
  salesExecutiveId: '',
  parentCompanyName: '',
  notes: '',
  contacts: [] as Contact[],
})

function addContact() {
  form.contacts.push({ name: '', position: '', phone: '', email: '' })
}
function removeContact(index: number) {
  if (form.contacts.length > 1) form.contacts.splice(index, 1)
}

const isFormValid = computed(() =>
  form.name.trim() !== '' &&
  form.customerSegment !== '' &&
  form.customerCategory !== '' &&
  form.region !== ''
)

const hasChanges = computed(() => {
  if (!detail.value) return false
  const c = detail.value.customer
  return (
    form.name !== c.name ||
    form.customerSegment !== c.segment ||
    form.customerCategory !== c.category ||
    form.region !== c.region ||
    form.salesExecutiveId !== c.salesExecutiveId ||
    form.parentCompanyName !== c.parentCompanyName ||
    JSON.stringify(form.contacts) !== JSON.stringify(c.contacts)
  )
})

async function handleSave() {
  if (!isFormValid.value) return
  saving.value = true
  try {
    await new Promise((r) => setTimeout(r, 1000))
    saved.value = true
  } catch {
    error.value = 'Failed to update customer. Please try again.'
  } finally {
    saving.value = false
  }
}

function formatDate(dateStr: string) {
  if (!dateStr) return '—'
  return new Date(dateStr).toLocaleDateString('en-GB', { day: '2-digit', month: 'short', year: 'numeric' })
}

onMounted(async () => {
  try {
    const [custDetail] = await Promise.all([
      crm.loadAdminCustomer(String(route.params.id)),
      listStore.fetchFilterOptions().catch(() => {}),
    ])
    detail.value = custDetail
    const c = custDetail.customer
    form.name = c.name
    form.customerSegment = c.segment
    form.customerCategory = c.category
    form.region = c.region
    form.salesExecutiveId = c.salesExecutiveId
    form.parentCompanyName = custDetail.parentCompany.name
    form.contacts = c.contacts?.length ? JSON.parse(JSON.stringify(c.contacts)) : []
  } catch (e) {
    error.value = crm.errorMessage(e)
  } finally {
    loading.value = false
  }
})
</script>

<template>
  <section class="admin-page">
    <!-- LOADING -->
    <div v-if="loading" class="state-box">
      <i class="pi pi-spin pi-spinner state-icon" />
      <span>Loading customer data...</span>
    </div>

    <!-- ERROR -->
    <Message v-if="error && !detail" severity="error" :closable="false">{{ error }}</Message>

    <!-- SAVED SUCCESS -->
    <template v-if="saved">
      <div class="success-panel">
        <div class="success-icon">
          <i class="pi pi-check-circle" />
        </div>
        <h2>Customer Updated Successfully</h2>
        <p class="muted">Changes to <strong>{{ form.name }}</strong> have been saved.</p>
        <div class="success-actions">
          <Button label="Back to Customer List" icon="pi pi-list" @click="router.push('/admin/customers')" />
          <Button label="View Detail" icon="pi pi-eye" severity="secondary" outlined @click="router.push(`/admin/customers/${route.params.id}`)" />
        </div>
      </div>
    </template>

    <!-- EDIT FORM -->
    <template v-else-if="detail">
      <Message v-if="error" severity="error">{{ error }}</Message>

      <!-- PAGE HEADER -->
      <Button icon="pi pi-arrow-left" severity="secondary" text rounded @click="router.push(`/admin/customers/${route.params.id}`)" title="Back" />
      <header class="page-heading">
        <div class="page-title-wrapper">
          <span class="eyebrow">Edit Customer</span>
          <h1>{{ detail.customer.name }}</h1>
          <div class="subtitle-row">
            <code class="code-tag code-blue">{{ detail.customer.customerCode }}</code>
            <span class="muted">&mdash; Last updated {{ formatDate(detail.customer.updatedAt) }}</span>
          </div>
        </div>
        <div class="page-heading-actions">
          <Button label="Cancel" severity="secondary" text size="small" @click="router.push(`/admin/customers/${route.params.id}`)" />
          <Button label="Save Changes" icon="pi pi-check" size="small" :loading="saving" :disabled="!isFormValid || saving || !hasChanges" @click="handleSave" />
        </div>
      </header>

      <div class="form-layout">
        <div class="form-stack">
          <!-- CUSTOMER SITE INFO -->
          <div class="form-card">
            <div class="form-card-header">
              <div class="form-card-icon si-blue"><i class="pi pi-map-marker" /></div>
              <div>
                <h3>Customer Site Information</h3>
                <p>Basic details about the customer site location.</p>
              </div>
            </div>
            <div class="form-grid">
              <div class="form-field full">
                <label>Customer Site Name <span class="required">*</span></label>
                <InputText v-model="form.name" placeholder="Customer site name" />
              </div>
              <div class="form-field">
                <label>Segment <span class="required">*</span></label>
                <Select v-model="form.customerSegment" :options="segmentOptions" optionLabel="label" optionValue="value" />
              </div>
              <div class="form-field">
                <label>Category <span class="required">*</span></label>
                <Select v-model="form.customerCategory" :options="categoryOptions" optionLabel="label" optionValue="value" />
              </div>
              <div class="form-field">
                <label>Region <span class="required">*</span></label>
                <Select v-model="form.region" :options="regionOptions" optionLabel="label" optionValue="value" />
              </div>
              <div class="form-field">
                <label>Sales Executive</label>
                <Select v-model="form.salesExecutiveId" :options="salesOptions" optionLabel="label" optionValue="value" />
              </div>
            </div>
          </div>

          <!-- PARENT COMPANY -->
          <div class="form-card">
            <div class="form-card-header">
              <div class="form-card-icon si-violet"><i class="pi pi-building" /></div>
              <div>
                <h3>Parent Company</h3>
                <p>Corporate entity this site belongs to.</p>
              </div>
            </div>
            <div class="form-grid">
              <div class="form-field full">
                <label>Company Name</label>
                <InputText v-model="form.parentCompanyName" placeholder="Parent company name" />
              </div>
            </div>
          </div>

          <!-- CONTACTS -->
          <div class="form-card">
            <div class="form-card-header">
              <div class="form-card-icon si-amber"><i class="pi pi-users" /></div>
              <div>
                <h3>Site Contacts</h3>
                <p>People to contact at this customer site.</p>
              </div>
            </div>
            <div v-if="form.contacts.length" class="contacts-list">
              <div v-for="(contact, idx) in form.contacts" :key="idx" class="contact-row">
                <div class="contact-row-header">
                  <span class="contact-label">Contact {{ idx + 1 }}</span>
                  <Button icon="pi pi-times" text rounded size="small" class="act-delete" @click="removeContact(idx)" />
                </div>
                <div class="form-grid">
                  <div class="form-field">
                    <label>Name</label>
                    <InputText v-model="contact.name" placeholder="Contact name" />
                  </div>
                  <div class="form-field">
                    <label>Position</label>
                    <InputText v-model="contact.position" placeholder="e.g. Owner, Manager" />
                  </div>
                  <div class="form-field">
                    <label>Phone</label>
                    <InputText v-model="contact.phone" placeholder="Phone number" />
                  </div>
                  <div class="form-field">
                    <label>Email</label>
                    <InputText v-model="contact.email" placeholder="Email address" />
                  </div>
                </div>
              </div>
            </div>
            <div v-else class="empty-contacts">
              <i class="pi pi-users" />
              <span>No contacts registered for this site.</span>
            </div>
            <Button label="Add Contact" icon="pi pi-plus" severity="secondary" text size="small" class="add-contact-btn" @click="addContact" />
          </div>

          <!-- NOTES -->
          <div class="form-card">
            <div class="form-card-header">
              <div class="form-card-icon si-slate"><i class="pi pi-file-edit" /></div>
              <div>
                <h3>Additional Notes</h3>
                <p>Internal notes about this customer site.</p>
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
            <h4>Change Summary</h4>
            <div class="summary-list">
              <div class="summary-row">
                <span>Customer Name</span>
                <strong>{{ form.name || '—' }}</strong>
              </div>
              <div class="summary-row">
                <span>Segment</span>
                <strong>{{ form.customerSegment || '—' }}</strong>
              </div>
              <div class="summary-row">
                <span>Category</span>
                <strong>{{ form.customerCategory || '—' }}</strong>
              </div>
              <div class="summary-row">
                <span>Region</span>
                <strong>{{ form.region || '—' }}</strong>
              </div>
              <div class="summary-row">
                <span>Company</span>
                <strong>{{ form.parentCompanyName || '—' }}</strong>
              </div>
              <div class="summary-row">
                <span>Contacts</span>
                <strong>{{ form.contacts.length }}</strong>
              </div>
            </div>
          </div>
          <div class="sidebar-card" :class="hasChanges ? 'changes-card' : 'no-changes-card'">
            <i :class="hasChanges ? 'pi pi-exclamation-circle' : 'pi pi-check-circle'" />
            <p v-if="hasChanges">You have unsaved changes. Click "Save Changes" to apply.</p>
            <p v-else>No changes detected. Edit the form to make updates.</p>
          </div>
          <div class="sidebar-actions">
            <Button label="Save Changes" icon="pi pi-check" class="full-width" :loading="saving" :disabled="!isFormValid || saving || !hasChanges" @click="handleSave" />
            <Button label="Cancel" severity="secondary" text class="full-width" @click="router.push(`/admin/customers/${route.params.id}`)" />
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
.subtitle-row {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  margin-top: 0.1rem;
}
.page-heading-actions {
  display: flex;
  gap: 0.5rem;
  align-items: center;
  padding-top: 0.15rem;
}

.code-tag {
  display: inline-block;
  font-family: 'SF Mono', 'Fira Code', 'Consolas', monospace;
  font-size: 0.78rem;
  font-weight: 600;
  padding: 0.15rem 0.5rem;
  border-radius: 4px;
  background: #f1f5f9;
  color: var(--text-secondary);
}
.code-blue { background: #fff0f1; color: #e63946; }

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
.si-violet { background: #fff5f5; color: #ef4e5d; }
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
.form-field.full { grid-column: 1 / -1; }
.form-field label {
  font-size: 0.72rem;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: var(--text-muted);
}
.required { color: #dc2626; }

/* ── CONTACTS ──────────────────────────────────────────────────────── */
.contacts-list {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}
.contact-row {
  background: #f8fafc;
  border: 1px solid #eef1f5;
  border-radius: var(--radius-md);
  padding: 1rem 1.1rem;
}
.contact-row-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 0.75rem;
}
.contact-label {
  font-size: 0.75rem;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.04em;
  color: var(--brand-blue);
}
.add-contact-btn { margin-top: 0.5rem; }
.empty-contacts {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 1rem;
  background: #f8fafc;
  border-radius: var(--radius-md);
  color: var(--text-muted);
  font-size: 0.85rem;
}
.empty-contacts i { color: var(--text-faint); }
.act-delete { color: #dc2626 !important; }
.act-delete:hover { background: #fef2f2 !important; }

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
.changes-card {
  display: flex;
  gap: 0.6rem;
  align-items: flex-start;
  background: #fffbeb;
  border-color: #fde68a;
}
.changes-card i {
  color: #d97706;
  margin-top: 0.1rem;
  font-size: 0.95rem;
  flex-shrink: 0;
}
.changes-card p {
  margin: 0;
  font-size: 0.78rem;
  color: #92400e;
  line-height: 1.5;
}
.no-changes-card {
  display: flex;
  gap: 0.6rem;
  align-items: flex-start;
  background: #ecfdf5;
  border-color: #a7f3d0;
}
.no-changes-card i {
  color: #059669;
  margin-top: 0.1rem;
  font-size: 0.95rem;
  flex-shrink: 0;
}
.no-changes-card p {
  margin: 0;
  font-size: 0.78rem;
  color: #065f46;
  line-height: 1.5;
}
.sidebar-actions {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}
.full-width { width: 100%; }

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

/* ── STATE BOX ─────────────────────────────────────────────────────── */
.state-box {
  min-height: 300px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  color: var(--text-muted);
}
.state-icon {
  font-size: 1.75rem;
  color: var(--brand-blue);
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
</style>
