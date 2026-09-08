<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import Button from 'primevue/button'
import InputText from 'primevue/inputtext'
import Select from 'primevue/select'
import Textarea from 'primevue/textarea'
import Message from 'primevue/message'
import type { Contact } from '../../../types/crm'
import { useCustomerListStore } from '../../../stores/customerList'
import { getSalesExecutives } from '../../../api/crm'
import { listCategories, listSegments, type MasterDataCategory, type MasterDataSegment } from '../../../api/masterData'

const router = useRouter()
const store = useCustomerListStore()
const error = ref('')
const saved = ref(false)
const saving = ref(false)

const segments = ref<MasterDataSegment[]>([])
const categories = ref<MasterDataCategory[]>([])
const salesExecutives = ref<{ id: string; fullName: string }[]>([])
const optionsLoading = ref(false)
const optionsError = ref('')
const segmentOptions = computed(() => [{ label: 'Select segment', value: '' }, ...segments.value.map((s) => ({ label: s.name, value: s.name }))])
const categoryOptions = computed(() => {
  const segment = segments.value.find((item) => item.name === form.customerSegment)
  const available = segment ? categories.value.filter((item) => item.segmentId === segment.id) : categories.value
  return [{ label: 'Select category', value: '' }, ...available.map((c) => ({ label: c.name, value: c.name }))]
})
const regionOptions = computed(() => {
  const regs = store.filterOptions?.regions ?? []
  return [{ label: 'Select region', value: '' }, ...regs.map((r) => ({ label: r, value: r }))]
})
const salesOptions = computed(() => {
  return [{ label: 'Select sales executive', value: '' }, ...salesExecutives.value.map((s) => ({ label: s.fullName, value: s.id }))]
})

const blankContact = (): Contact => ({ name: '', position: '', phone: '', email: '' })

const form = reactive({
  name: '',
  customerSegment: '',
  customerCategory: '',
  region: '',
  salesExecutiveId: '',
  parentCompanyName: '',
  parentCode: '',
  address: '',
  province: '',
  district: '',
  subDistrict: '',
  village: '',
  notes: '',
  contacts: [blankContact()] as Contact[],
})

function addContact() {
  form.contacts.push(blankContact())
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

async function handleSubmit() {
  if (!isFormValid.value) return
  saving.value = true
  try {
    await new Promise((r) => setTimeout(r, 1200))
    saved.value = true
  } catch (e) {
    error.value = 'Failed to save customer. Please try again.'
  } finally {
    saving.value = false
  }
}

onMounted(async () => {
  optionsLoading.value = true
  try {
    const [segmentData, categoryData, salesData] = await Promise.all([
      listSegments({ status: 'ACTIVE' }),
      listCategories({ status: 'ACTIVE' }),
      getSalesExecutives(),
      store.fetchFilterOptions(),
    ])
    segments.value = segmentData
    categories.value = categoryData
    salesExecutives.value = salesData
  } catch {
    optionsError.value = 'Unable to load master data and sales executives. Please try again.'
  } finally {
    optionsLoading.value = false
  }
})

watch(() => form.customerSegment, (segmentName) => {
  const segment = segments.value.find((item) => item.name === segmentName)
  if (segment && form.customerCategory && !categories.value.some((item) => item.segmentId === segment.id && item.name === form.customerCategory)) {
    form.customerCategory = ''
  }
})
</script>

<template>
  <section class="admin-page bg-surface min-h-full">
    <!-- SUCCESS STATE -->
    <template v-if="saved">
      <div class="success-panel">
        <div class="success-icon">
          <i class="pi pi-check-circle" />
        </div>
        <h2>Customer Created Successfully</h2>
        <p class="muted">The new customer site <strong>{{ form.name }}</strong> has been added to the system.</p>
        <div class="success-actions">
          <Button label="View Customer List" icon="pi pi-list" @click="router.push('/admin/customers')" />
          <Button label="Add Another" icon="pi pi-plus" severity="secondary" outlined @click="saved = false; form.name = ''; form.customerSegment = ''; form.customerCategory = ''; form.region = ''; form.salesExecutiveId = ''; form.parentCompanyName = ''; form.parentCode = ''; form.address = ''; form.province = ''; form.district = ''; form.subDistrict = ''; form.village = ''; form.notes = ''; form.contacts = [blankContact()]" />
        </div>
      </div>
    </template>

    <!-- FORM -->
    <template v-else>
      <!-- PAGE HEADER -->
      <header class="sticky top-0 z-40 border-b border-[#e2e8f0] bg-white/95 px-[16px] py-[8px] shadow-[0px_2px_8px_0px_rgba(15,23,42,0.04)] backdrop-blur sm:px-[24px] lg:px-[24px] mb-4 -mx-6 -mt-6">
        <div class="mx-auto flex flex-wrap items-center justify-between gap-[12px] max-w-[1200px]">
          <div class="min-w-0 flex items-center gap-[12px]">
            <button
              type="button"
              class="flex shrink-0 items-center gap-[6px] font-semibold text-[#64748b] transition-colors hover:text-[#dc2626]"
              @click="router.push('/admin/customers')"
            >
              <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="size-[14px]"><path d="m12 19-7-7 7-7"></path><path d="M19 12H5"></path></svg> Back to Customer
            </button>
            <div class="hidden h-[28px] w-px bg-[#e2e8f0] sm:block"></div>
            <div class="min-w-0 flex items-center gap-[8px]">
              <h5 class="truncate font-['Inter'] text-[15px] font-bold leading-[20px] text-[#1e293b]">Create New Customer</h5>
              <span class="hidden truncate text-[11px] leading-[15px] text-[#94a3b8] sm:block">Customer List &gt; Customer Site &gt; Create</span>
            </div>
          </div>
          <section class="flex items-center gap-[8px]">
            <button
              type="button"
              class="inline-flex items-center justify-center border font-semibold tracking-normal shadow-[0_1px_2px_rgba(15,23,42,0.08)] transition-all duration-150 ease-out [-webkit-tap-highlight-color:transparent] hover:-translate-y-px active:translate-y-0 active:scale-[0.99] active:shadow-[0_1px_1px_rgba(15,23,42,0.08)] focus:outline-none focus:ring-1 focus:ring-offset-0 disabled:translate-y-0 disabled:scale-100 disabled:cursor-not-allowed disabled:opacity-50 disabled:shadow-none border-slate-200 bg-white text-slate-700 shadow-[0_1px_3px_rgba(15,23,42,0.06)] hover:border-slate-300 hover:bg-slate-50 hover:text-slate-900 hover:shadow-[0_2px_6px_rgba(15,23,42,0.08)] focus:ring-slate-100 gap-1.5 rounded-md px-2.5 py-1.5 text-[11.5px]"
              @click="router.push('/admin/customers')"
            >
              Cancel
            </button>
            <button
              type="button"
              class="inline-flex items-center justify-center border font-semibold tracking-normal shadow-[0_1px_2px_rgba(15,23,42,0.08)] transition-all duration-150 ease-out [-webkit-tap-highlight-color:transparent] hover:-translate-y-px active:translate-y-0 active:scale-[0.99] active:shadow-[0_1px_1px_rgba(15,23,42,0.08)] focus:outline-none focus:ring-1 focus:ring-offset-0 disabled:translate-y-0 disabled:scale-100 disabled:cursor-not-allowed disabled:opacity-50 disabled:shadow-none border-primary-600 bg-primary-600 text-white shadow-[0_3px_8px_rgba(220,38,38,0.14)] hover:border-primary-700 hover:bg-primary-700 hover:shadow-[0_4px_10px_rgba(220,38,38,0.16)] focus:ring-primary-100 gap-1.5 rounded-md px-2.5 py-1.5 text-[11.5px]"
              :disabled="!isFormValid || saving"
              @click="handleSubmit"
            >
              <i class="pi text-xs shrink-0" :class="saving ? 'pi-spin pi-spinner' : 'pi-send'"></i> Create Customer
            </button>
          </section>
        </div>
      </header>

      <Message v-if="error" severity="error">{{ error }}</Message>

      <div class="form-layout">
        <!-- LEFT COLUMN: FORM -->
        <div class="form-stack">
          <section class="customer-information-shell">
            <div class="customer-information-heading">
              <div>
                <h2>Customer Information</h2>
                <p>Define the customer site and its parent-company reference.</p>
              </div>
              <span class="information-scope-badge">Site + Company</span>
            </div>
            <div class="customer-information-grid">
          <!-- CUSTOMER SITE INFO -->
          <div class="form-card">
            <div class="form-card-header">
              <div class="form-card-icon si-blue"><i class="pi pi-map-marker" /></div>
              <div>
                <h3>Site Identity</h3>
                <p>Basic details about the customer site location.</p>
              </div>
            </div>
            <div class="form-grid">
              <div class="form-field full">
                <label>Customer Site Name <span class="required">*</span></label>
                <InputText v-model="form.name" placeholder="e.g. Yummy Cabang Jakarta Selatan" />
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
                <label>Company Name <span class="field-note">(reference)</span></label>
                <InputText v-model="form.parentCompanyName" placeholder="e.g. PT Yummy Food Indonesia" />
              </div>
              <div class="form-field">
                <label>Company Code <span class="field-note">(reference)</span></label>
                <InputText v-model="form.parentCode" placeholder="Auto-generated if empty" />
              </div>
              <div class="company-link-note">
                <i class="pi pi-info-circle" />
                <span>Select or create a Parent Company separately. Company records are not created when saving this site.</span>
                <Button label="Create Parent Company" severity="secondary" outlined size="small" @click="router.push('/admin/customers/add/company')" />
              </div>
            </div>
          </div>
            </div>
          </section>

          <!-- ADDRESS -->
          <div class="form-card">
            <div class="form-card-header">
              <div class="form-card-icon si-emerald"><i class="pi pi-map" /></div>
              <div>
                <div class="section-title-line"><h3>Site Address</h3><span class="address-scope-badge site-address-badge">Site</span></div>
                <p>Address for customer transactions and site-level documents.</p>
              </div>
            </div>
            <div class="form-grid">
              <div class="form-field full">
                <div class="address-mode-label"><label>Search by Gmaps / Manual <span class="required">*</span></label><span>Manual</span></div>
                <Textarea v-model="form.address" :autoResize="true" rows="2" placeholder="Search or type customer address" />
                <small class="address-helper">Manual entry is available. Add a Google Maps API key to enable autocomplete.</small>
              </div>
              <div class="location-detail-label">Location Detail</div>
              <div class="form-field">
                <label>Province</label>
                <InputText v-model="form.province" placeholder="Province" />
              </div>
              <div class="form-field">
                <label>District</label>
                <InputText v-model="form.district" placeholder="District" />
              </div>
              <div class="form-field">
                <label>Sub-District</label>
                <InputText v-model="form.subDistrict" placeholder="Sub-district" />
              </div>
              <div class="form-field">
                <label>Village</label>
                <InputText v-model="form.village" placeholder="Village" />
              </div>
            </div>
          </div>

          <!-- CONTACTS -->
          <div class="form-card">
            <div class="form-card-header">
              <div class="form-card-icon si-amber"><i class="pi pi-users" /></div>
              <div>
                <div class="section-title-line"><h3>Contact Information</h3><span class="shared-scope-badge">Site + Company</span><span class="optional-badge">Optional</span></div>
                <p>Keep site contacts and company contacts separate so operational and billing communication does not get mixed.</p>
              </div>
            </div>
            <div class="contact-scope-heading"><strong>Customer Site Contacts</strong><span class="site-scope-badge">Site</span></div>
            <p class="contact-scope-copy">Use this list for outlet, branch, or site-level contacts.</p>
            <div class="contacts-list">
              <div v-for="(contact, idx) in form.contacts" :key="idx" class="contact-row">
                <div class="contact-row-header">
                  <span class="contact-label">Contact {{ idx + 1 }}</span>
                  <Button v-if="form.contacts.length > 1" icon="pi pi-times" text rounded size="small" class="act-delete" @click="removeContact(idx)" />
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
            <Button label="Add Contact" icon="pi pi-plus" severity="secondary" text size="small" class="add-contact-btn" @click="addContact" />
            <div class="company-contacts-note">
              <div class="company-contact-heading"><strong>Company Contacts</strong><span class="company-scope-badge">Company</span></div>
              <p>Use a dedicated Parent Company record for billing, tax, or general company contacts.</p>
              <Button label="Open Parent Company Flow" severity="secondary" outlined size="small" @click="router.push('/admin/customers/add/company')" />
            </div>
          </div>

          <!-- NOTES -->
          <div class="form-card">
            <div class="form-card-header">
              <div class="form-card-icon si-slate"><i class="pi pi-file-edit" /></div>
              <div>
                <h3>Additional Notes</h3>
                <p>Any extra information about this customer site.</p>
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
          <div class="sidebar-card scope-card">
            <h4>Submission Summary</h4>
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
          <div class="sidebar-card tip-card">
            <i class="pi pi-info-circle" />
            <p>Customer codes will be automatically generated upon save in the format <code>PC-XXXXXX-SXXX</code>.</p>
          </div>
          <div class="sidebar-card field-scope-card">
            <h4>Field Scope &amp; Requirements</h4>
            <p class="scope-description">Scope badges show where each field belongs. The check icon only shows completion.</p>
            <div class="scope-legend">
              <span class="scope-badge site">Site</span>
              <span class="scope-badge company">Company</span>
              <span class="scope-badge shared">Shared / Preview</span>
            </div>
            <div class="scope-list">
              <div class="scope-group">
                <div class="scope-group-heading"><span>SITE</span><span class="scope-badge site">Site</span></div>
                <div class="scope-checklist">
                  <div class="scope-check-row"><span class="scope-check" :class="{ complete: form.name.trim() }"><i v-if="form.name.trim()" class="pi pi-check" /></span><span>Customer Name / Outlet / Branch / Store</span></div>
                  <div class="scope-check-row"><span class="scope-check" :class="{ complete: form.parentCompanyName.trim() }"><i v-if="form.parentCompanyName.trim()" class="pi pi-check" /></span><span>Customer Company / Parent</span></div>
                  <div class="scope-check-row"><span class="scope-check" :class="{ complete: form.customerSegment }"><i v-if="form.customerSegment" class="pi pi-check" /></span><span>Customer Segment</span></div>
                  <div class="scope-check-row"><span class="scope-check" :class="{ complete: form.customerCategory }"><i v-if="form.customerCategory" class="pi pi-check" /></span><span>Customer Category</span></div>
                  <div class="scope-check-row"><span class="scope-check" :class="{ complete: form.address }"><i v-if="form.address" class="pi pi-check" /></span><span>Customer / Outlet / Branch Address</span></div>
                </div>
              </div>
              <div class="scope-row"><span class="scope-badge company">Reference</span><span>Parent company name and code are reference fields; create companies separately.</span></div>
            </div>
          </div>
          <div class="sidebar-actions">
            <Button label="Create Customer" icon="pi pi-check" class="full-width" :loading="saving" :disabled="!isFormValid || saving" @click="handleSubmit" />
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
.add-contact-btn {
  margin-top: 0.5rem;
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
.tip-card code {
  background: rgba(230, 57, 70, 0.1);
  padding: 0.1rem 0.3rem;
  border-radius: 3px;
  font-size: 0.72rem;
  font-weight: 600;
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

.act-delete {
  color: #dc2626 !important;
}
.act-delete:hover {
  background: #fef2f2 !important;
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
.back-add-button{display:inline-flex;margin:0 12px 0 0;padding:0 10px 0 0;border-right:1px solid #e2e8f0;border-radius:0;background:transparent;color:#64748b;font-size:12px}.back-add-button .p-button-icon{font-size:13px}.page-heading{position:sticky;top:0;z-index:20;align-items:center;min-height:56px;padding:8px 24px;border-bottom:1px solid #e2e8f0;background:rgba(255,255,255,.96);box-shadow:0 2px 8px rgba(15,23,42,.04);backdrop-filter:blur(10px)}.page-title-wrapper{display:flex;align-items:center;gap:10px}.page-title-wrapper .eyebrow{display:none}.page-title-wrapper h1{font-size:17px;font-weight:700;line-height:22px}.page-title-wrapper .muted{font-size:11px;color:#94a3b8}.page-heading-actions :deep(.p-button){height:32px;border-radius:8px;font-size:12px}.form-layout{width:min(100%,1200px);margin:0 auto;grid-template-columns:minmax(0,1fr) 320px;gap:24px;padding:32px 24px}.form-stack{gap:16px}.form-card,.sidebar-card{border:1px solid #e2e8f0;border-radius:18px;background:#fff;box-shadow:0 1px 2px rgba(15,23,42,.04)}.form-card{padding:20px}.form-card-header{gap:12px;margin-bottom:16px;padding-bottom:14px;border-bottom-color:#eef2f7}.form-card-icon{width:36px;height:36px;border-radius:10px;font-size:14px}.form-card-header h3{font-size:15px}.form-card-header p{font-size:12px;color:#64748b}.form-grid{gap:16px 20px}.form-field{gap:6px}.form-field label{text-transform:none;letter-spacing:0;color:#334155;font-size:12px;font-weight:600}.form-field :deep(.p-inputtext),.form-field :deep(.p-select),.form-field :deep(.p-textarea){min-height:40px;border:1px solid #d9e2ec;border-radius:10px;background:#fff;font-size:12px;box-shadow:0 1px 2px rgba(15,23,42,.03)}.form-field :deep(.p-inputtext),.form-field :deep(.p-textarea){padding:0 12px}.form-field :deep(.p-textarea){padding-top:10px}.form-field :deep(.p-select-label){padding:0 12px;font-size:12px}.form-field :deep(.p-inputtext:focus),.form-field :deep(.p-select:focus),.form-field :deep(.p-textarea:focus){border-color:#dc2626;box-shadow:0 0 0 3px rgba(220,38,38,.12)}.form-sidebar{top:72px;gap:16px}.sidebar-card{padding:16px}.sidebar-card h4{margin-bottom:12px;font-size:12px}.summary-list{gap:10px}.summary-row span,.summary-row strong{font-size:12px}.summary-row span{color:#64748b}.summary-row strong{color:#0f172a}.tip-card{border-color:#bfdbfe;background:#eff6ff}.tip-card i{color:#2563eb;font-size:14px}.tip-card p{color:#1d4ed8;font-size:12px}.tip-card code{background:#dbeafe;font-size:11px}.sidebar-actions :deep(.p-button){min-height:40px;border-radius:10px;font-size:12px}.contact-row{border-color:#eef2f7;border-radius:12px;padding:12px;background:#f8fafc}.scope-card{padding:16px}.scope-list{display:grid;gap:10px}.scope-row{display:grid;grid-template-columns:auto 1fr;align-items:start;gap:8px;color:#64748b;font-size:11px;line-height:1.45}.scope-badge{display:inline-flex;align-items:center;height:22px;padding:0 8px;border:1px solid #bfdbfe;border-radius:999px;background:#eff6ff;color:#1d4ed8;font-size:10px;font-weight:700}.scope-badge.company{border-color:#fecdd3;background:#fff1f2;color:#be123c}.scope-badge.required-scope{border-color:#fde68a;background:#fffbeb;color:#a16207}@media(max-width:1024px){.form-layout{grid-template-columns:1fr;padding:24px 20px}.form-sidebar{position:static;order:-1}}@media(max-width:768px){.admin-page{padding:0}.page-heading{padding:10px 16px}.page-title-wrapper .muted{display:none}.form-layout{padding:20px 16px}.form-sidebar{order:0}}@media(max-width:560px){.page-heading{align-items:flex-start;padding:10px 12px}.page-heading-actions{width:100%}.page-heading-actions :deep(.p-button){flex:1}.form-layout{padding:16px 12px}.form-card{padding:16px;border-radius:14px}.form-grid{grid-template-columns:1fr;gap:14px}.form-field.full{grid-column:1}}
.field-note{color:#94a3b8;font-size:10px;font-weight:500;text-transform:none;letter-spacing:0}.company-link-note{grid-column:1/-1;display:grid;grid-template-columns:auto 1fr auto;align-items:center;gap:8px;padding:10px 12px;border:1px solid #dbeafe;border-radius:10px;background:#eff6ff;color:#1d4ed8;font-size:11px;line-height:1.4}.company-link-note>i{font-size:13px}.company-link-note :deep(.p-button){height:32px;border-radius:8px;font-size:11px;white-space:nowrap}@media(max-width:560px){.company-link-note{grid-template-columns:auto 1fr}.company-link-note :deep(.p-button){grid-column:1/-1;width:100%}}
.customer-information-shell{border:1px solid #e2e8f0;border-radius:18px;background:#fff;box-shadow:0 1px 2px rgba(15,23,42,.04);overflow:hidden}.customer-information-heading{display:flex;align-items:flex-start;justify-content:space-between;gap:12px;padding:16px 20px;border-bottom:1px solid #eef2f7}.customer-information-heading h2{margin:0;color:#0f172a;font-size:16px;line-height:22px}.customer-information-heading p{margin:3px 0 0;color:#64748b;font-size:12px}.information-scope-badge{display:inline-flex;align-items:center;height:24px;padding:0 9px;border:1px solid #bfdbfe;border-radius:999px;background:#eff6ff;color:#1d4ed8;font-size:10px;font-weight:700;white-space:nowrap}.customer-information-grid{display:grid;grid-template-columns:1fr 1fr;gap:16px;padding:16px}.customer-information-grid>.form-card{box-shadow:none;border-color:#eef2f7;padding:16px}.customer-information-grid>.form-card .form-card-header{margin-bottom:14px;padding-bottom:12px}.form-sidebar .field-scope-card{order:-3}.form-sidebar .tip-card{order:-2}.form-sidebar .scope-card{order:-1}@media(max-width:1024px){.customer-information-grid{grid-template-columns:1fr}}@media(max-width:560px){.customer-information-heading{padding:14px 16px;flex-direction:column}.customer-information-grid{padding:12px;gap:12px}}
.scope-description{margin:-4px 0 10px;color:#64748b;font-size:12px;line-height:1.5}.scope-legend{display:flex;flex-wrap:wrap;gap:8px;margin-bottom:14px}.scope-badge.shared{border-color:#ddd6fe;background:#f5f3ff;color:#6d28d9}.scope-group{padding:14px 12px;border:1px solid #dbe3ee;border-radius:16px;background:#f8fafc}.scope-group-heading{display:flex;align-items:center;justify-content:space-between;margin:0 2px 10px;color:#64748b;font-size:11px;letter-spacing:.08em}.scope-checklist{display:grid;gap:8px}.scope-check-row{display:flex;align-items:center;gap:10px;min-height:42px;padding:10px 12px;border:1px solid #dce5f0;border-radius:14px;background:#fff;color:#334155;font-size:12px;line-height:1.35}.scope-check{display:grid;place-items:center;width:20px;height:20px;flex:none;border:1px solid #cbd5e1;border-radius:50%;color:#fff;font-size:10px}.scope-check.complete{border-color:#22c55e;background:#22c55e}.scope-row{margin-top:12px}
.site-address-badge{border-color:#bbf7d0;background:#f0fdf4;color:#15803d}.location-detail-label{grid-column:1/-1;margin-top:2px;color:#64748b;font-size:11px;font-weight:700;letter-spacing:.08em;text-transform:uppercase}.site-address-badge + p{margin-top:4px}
.section-title-line{display:flex;align-items:center;flex-wrap:wrap;gap:8px}.section-title-line h3{margin:0}.shared-scope-badge,.optional-badge,.site-scope-badge,.company-scope-badge{display:inline-flex;align-items:center;height:22px;padding:0 8px;border:1px solid #ddd6fe;border-radius:999px;background:#f5f3ff;color:#6d28d9;font-size:10px;font-weight:700;white-space:nowrap}.optional-badge{border-color:#dbe3ee;background:#f8fafc;color:#64748b;font-weight:600}.site-scope-badge{border-color:#bbf7d0;background:#f0fdf4;color:#15803d}.company-scope-badge{border-color:#bfdbfe;background:#eff6ff;color:#1d4ed8}.contact-scope-heading,.company-contact-heading{display:flex;align-items:center;gap:8px;margin:0 0 4px}.contact-scope-heading strong,.company-contact-heading strong{color:#0f172a;font-size:13px}.contact-scope-copy,.company-contacts-note p{margin:0;color:#64748b;font-size:11px;line-height:1.45}.company-contacts-note{display:grid;gap:8px;margin-top:16px;padding:14px;border:1px solid #dbeafe;border-radius:14px;background:#eff6ff}.company-contacts-note :deep(.p-button){width:max-content;min-height:32px;border-radius:8px;font-size:11px}@media(max-width:560px){.company-contacts-note :deep(.p-button){width:100%}}
.address-mode-label{display:flex;align-items:center;justify-content:space-between;gap:10px}.address-mode-label>span{display:inline-flex;align-items:center;height:26px;padding:0 9px;border:1px solid #dbe3ee;border-radius:999px;background:#f8fafc;color:#64748b;font-size:10px;font-weight:600}.address-helper{display:block;color:#64748b;font-size:11px;line-height:1.4}@media(max-width:560px){.address-mode-label{align-items:flex-start;flex-direction:column;gap:6px}.address-mode-label>span{align-self:flex-end}}
</style>
