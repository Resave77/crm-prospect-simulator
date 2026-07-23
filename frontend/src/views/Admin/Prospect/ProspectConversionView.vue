<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import Button from 'primevue/button'
import Checkbox from 'primevue/checkbox'
import InputNumber from 'primevue/inputnumber'
import InputText from 'primevue/inputtext'
import Message from 'primevue/message'
import RadioButton from 'primevue/radiobutton'
import Select from 'primevue/select'
import Tag from 'primevue/tag'
import Textarea from 'primevue/textarea'
import { convertProspect, getConversionForm } from '../../../api/crm'
import { useCrmStore } from '../../../stores/crm'
import type { Address, Contact, ConversionFormData, ConversionInput, ParentCompany, PeriodAssignment } from '../../../types/crm'

const route = useRoute()
const router = useRouter()
const crm = useCrmStore()
const data = ref<ConversionFormData | null>(null)
const error = ref('')
const submitted = ref(false)
const saving = ref(false)

const blankAddress = (): Address => ({ mode: '', province: '', district: '', subDistrict: '', village: '', latitude: null, longitude: null, previewAddress: '' })
const blankContact = (phone = ''): Contact => ({ name: '', position: '', phone, email: '' })
const blankAssignment = (): PeriodAssignment => ({ ownerId: '', ownerName: '', startMonth: new Date().getMonth() + 1, startYear: new Date().getFullYear(), end: 'UNTIL_NOW' })

const form = reactive<ConversionInput>({
  customerName: '', customerSegment: '', customerCategory: '', parentMethod: '', existingParentCompanyId: null,
  parentCompanyName: '', sameAsSiteAddress: false, siteAddress: blankAddress(), companyAddress: blankAddress(),
  siteContacts: [], companyContacts: [], ppn: '', idTkuNumber: '', nik: '', companyNpwpName: '',
  companyNpwpAddress: '', companyNpwpNumber: '', shipmentCost: '', invoiceType: '', bankAccount: '',
  termOfPayment: '', billToSource: '', shipToSource: '', billingAddressPreview: '', shippingAddressPreview: '',
  salesExecutiveId: '', salesAssignments: [], kamAssignments: [],
})

const parentMethods = [
  { label: 'Manual Entry', value: 'MANUAL_ENTRY' },
  { label: 'Company Name Matches Customer Name', value: 'MATCH_CUSTOMER_NAME' },
  { label: 'Existing Company', value: 'EXISTING_COMPANY' },
]
const addressModes = [{ label: 'Search by Gmaps (simulated)', value: 'GMAPS_SIMULATION' }, { label: 'Manual', value: 'MANUAL' }]
const documentSources = ['Company', 'Site', 'Other Delivery']
const months = Array.from({ length: 12 }, (_, index) => ({ label: new Date(2026, index).toLocaleString('en', { month: 'long' }), value: index + 1 }))
const years = Array.from({ length: 8 }, (_, index) => new Date().getFullYear() - 1 + index)

const selectedParent = computed(() => data.value?.parentCompanies.find((item) => item.id === form.existingParentCompanyId) ?? null)
const isExistingParent = computed(() => form.parentMethod === 'EXISTING_COMPANY')
const parentCodePreview = computed(() => selectedParent.value?.parentCode ?? data.value?.parentCodePreview ?? 'Generated on save')
const customerCodePreview = computed(() => data.value?.customerCodePreview ?? 'Generated on save')

function addressPreview(source: string) {
  if (source === 'Company') return selectedParent.value?.address.previewAddress || form.companyAddress.previewAddress || 'Company address not completed'
  if (source === 'Site') return form.siteAddress.previewAddress || 'Site address not completed'
  if (source === 'Other Delivery') return 'Other Delivery — optional address not configured in this simulation'
  return 'Select a source to preview its address'
}

const billPreview = computed(() => addressPreview(form.billToSource))
const shipPreview = computed(() => addressPreview(form.shipToSource))

onMounted(async () => {
  try {
    data.value = await getConversionForm(String(route.params.id))
    const prospect = data.value.prospect.prospect
    form.customerName = prospect.placeName
    form.customerCategory = data.value.options.categories.includes(prospect.placeCategory) ? prospect.placeCategory : ''
    form.siteAddress = { mode: 'GMAPS_SIMULATION', province: '', district: '', subDistrict: '', village: '', latitude: prospect.latitude, longitude: prospect.longitude, previewAddress: prospect.formattedAddress }
    form.siteContacts = prospect.phoneNumber ? [blankContact(prospect.phoneNumber)] : []
    form.salesExecutiveId = prospect.assignedSalesExecutiveId
  } catch (caught) {
    error.value = crm.errorMessage(caught)
  }
})

watch(() => form.customerName, (name) => {
  if (form.parentMethod === 'MATCH_CUSTOMER_NAME') form.parentCompanyName = name
})

watch(() => form.parentMethod, (method) => {
  form.existingParentCompanyId = null
  form.sameAsSiteAddress = false
  if (method === 'MATCH_CUSTOMER_NAME') form.parentCompanyName = form.customerName
  if (method === 'EXISTING_COMPANY') form.parentCompanyName = ''
})

watch(() => form.existingParentCompanyId, () => {
  const parent = selectedParent.value
  if (!parent) return
  form.parentCompanyName = parent.name
  form.companyAddress = structuredClone(parent.address)
  form.companyContacts = structuredClone(parent.contacts ?? [])
  form.companyNpwpName = parent.npwpName
  form.companyNpwpAddress = parent.npwpAddress
  form.companyNpwpNumber = parent.npwpNumber
  form.termOfPayment = parent.termOfPayment
  form.kamAssignments = structuredClone(parent.kamAssignments ?? [])
})

watch(() => form.sameAsSiteAddress, (same) => {
  if (same && !isExistingParent.value) form.companyAddress = structuredClone(form.siteAddress)
})

watch(() => form.siteAddress, (address) => {
  if (form.sameAsSiteAddress && !isExistingParent.value) form.companyAddress = structuredClone(address)
}, { deep: true })

function applySuggestion(target: 'site' | 'company', suggestion: Address) {
  if (target === 'site') form.siteAddress = structuredClone(suggestion)
  else form.companyAddress = structuredClone(suggestion)
}

function addContact(target: 'site' | 'company') {
  ;(target === 'site' ? form.siteContacts : form.companyContacts).push(blankContact())
}

function addAssignment(target: 'sales' | 'kam') {
  ;(target === 'sales' ? form.salesAssignments : form.kamAssignments).push(blankAssignment())
}

function salesName(id: string) {
  return data.value?.salesExecutives.find((item) => item.id === id)?.fullName ?? ''
}

function coreInvalid() {
  return !form.customerName.trim() || !form.customerSegment || !form.customerCategory || !form.parentMethod ||
    (isExistingParent.value ? !form.existingParentCompanyId : !form.parentCompanyName.trim()) ||
    !form.siteAddress.mode || !form.siteAddress.province || !form.siteAddress.district || !form.siteAddress.subDistrict ||
    !form.siteAddress.village || !form.siteAddress.previewAddress.trim() || !form.salesExecutiveId
}

async function submit() {
  submitted.value = true
  error.value = ''
  if (coreInvalid()) {
    error.value = 'Complete all required Customer Site, Parent Company, core address, and Sales Executive fields.'
    window.scrollTo({ top: 0, behavior: 'smooth' })
    return
  }
  form.billingAddressPreview = form.billToSource ? billPreview.value : ''
  form.shippingAddressPreview = form.shipToSource ? shipPreview.value : ''
  saving.value = true
  try {
    await convertProspect(String(route.params.id), form)
    await router.push({ path: '/admin/customers', query: { converted: '1' } })
  } catch (caught) {
    error.value = crm.errorMessage(caught)
    window.scrollTo({ top: 0, behavior: 'smooth' })
  } finally {
    saving.value = false
  }
}
</script>

<template>
  <section class="admin-page conversion-page">
    <Button label="Back to review" icon="pi pi-arrow-left" severity="secondary" text @click="router.push(`/admin/prospects/${route.params.id}/review`)" />
    <div class="page-heading"><div><p class="eyebrow">Points 12–14</p><h1>Convert to Customer Existing</h1><p class="muted">Review Google snapshot data, complete required master data, and create the Customer Site atomically.</p></div><Tag value="WON — eligible" severity="success" /></div>
    <Message v-if="error" severity="error" :closable="false">{{ error }}</Message>
    <div v-if="!data && !error" class="empty-state">Preparing conversion form…</div>

    <form v-if="data" class="conversion-layout" @submit.prevent="submit">
      <div class="conversion-main">
        <section class="form-section">
          <div class="section-heading"><span>01</span><div><h2>Customer Information</h2><p>Customer Site identity and Parent Company relationship.</p></div></div>
          <div class="dual-cards">
            <article class="form-card"><div class="card-label"><strong>Customer Site</strong><Tag value="Required" severity="danger" /></div>
              <label class="field"><span>Customer Name / Outlet / Branch / Store *</span><InputText v-model="form.customerName" :invalid="submitted && !form.customerName.trim()" fluid /></label>
              <div class="two-fields"><label class="field"><span>Customer Segment *</span><Select v-model="form.customerSegment" :options="data.options.segments" placeholder="Select segment" :invalid="submitted && !form.customerSegment" fluid /></label><label class="field"><span>Customer Category *</span><Select v-model="form.customerCategory" :options="data.options.categories" placeholder="Select category" :invalid="submitted && !form.customerCategory" fluid /></label></div>
              <div class="code-preview"><label class="field"><span>Parent Code Preview</span><InputText :model-value="parentCodePreview" disabled fluid /></label><label class="field"><span>Customer Code Preview</span><InputText :model-value="customerCodePreview" disabled fluid /></label><small>Simulation-only codes are safely generated by the backend on save.</small></div>
            </article>
            <article class="form-card"><div class="card-label"><strong>Parent Company</strong><Tag value="Required" severity="danger" /></div>
              <div class="radio-stack"><label v-for="method in parentMethods" :key="method.value"><RadioButton v-model="form.parentMethod" name="parentMethod" :value="method.value" /><span>{{ method.label }}</span></label></div>
              <label v-if="isExistingParent" class="field"><span>Search Existing Company *</span><Select v-model="form.existingParentCompanyId" :options="data.parentCompanies" option-label="name" option-value="id" filter placeholder="Search name or code" fluid><template #option="slot"><div><strong>{{ slot.option.name }}</strong><small class="block-muted">{{ slot.option.parentCode }}</small></div></template></Select></label>
              <label v-else class="field"><span>Customer Company / Parent *</span><InputText v-model="form.parentCompanyName" :disabled="form.parentMethod === 'MATCH_CUSTOMER_NAME'" :invalid="submitted && !!form.parentMethod && !form.parentCompanyName.trim()" fluid /></label>
              <div v-if="selectedParent" class="master-preview"><span>Locked master preview</span><strong>{{ selectedParent.parentCode }} · {{ selectedParent.name }}</strong><p>{{ selectedParent.address.previewAddress }}</p></div>
            </article>
          </div>
        </section>

        <section class="form-section">
          <div class="section-heading"><span>02</span><div><h2>Address Information</h2><p>Google snapshot prefill with explicit Site and Company scopes.</p></div></div>
          <div class="dual-cards">
            <article class="form-card"><div class="card-label"><strong>Site Address</strong><Tag value="Required" severity="danger" /></div>
              <label class="field"><span>Input Method *</span><Select v-model="form.siteAddress.mode" :options="addressModes" option-label="label" option-value="value" fluid /></label>
              <label v-if="form.siteAddress.mode === 'GMAPS_SIMULATION'" class="field"><span>Local Gmaps Suggestions</span><Select :options="data.options.addressSuggestions" option-label="previewAddress" placeholder="Choose local suggestion" fluid @change="applySuggestion('site', $event.value)" /></label>
              <div class="address-grid"><label class="field"><span>Province *</span><InputText v-model="form.siteAddress.province" fluid /></label><label class="field"><span>District *</span><InputText v-model="form.siteAddress.district" fluid /></label><label class="field"><span>Sub-District *</span><InputText v-model="form.siteAddress.subDistrict" fluid /></label><label class="field"><span>Village *</span><InputText v-model="form.siteAddress.village" fluid /></label><label class="field"><span>Latitude <em>Optional</em></span><InputNumber v-model="form.siteAddress.latitude" :min-fraction-digits="4" :max-fraction-digits="7" fluid /></label><label class="field"><span>Longitude <em>Optional</em></span><InputNumber v-model="form.siteAddress.longitude" :min-fraction-digits="4" :max-fraction-digits="7" fluid /></label></div>
              <label class="field"><span>Preview Address *</span><Textarea v-model="form.siteAddress.previewAddress" rows="3" fluid /></label>
            </article>
            <article class="form-card" :class="{ 'locked-card': isExistingParent }"><div class="card-label"><strong>Company Address</strong><Tag :value="isExistingParent ? 'Master locked' : 'Optional'" severity="secondary" /></div>
              <label v-if="!isExistingParent" class="check-row"><Checkbox v-model="form.sameAsSiteAddress" binary /><span>Same as Site Address</span></label>
              <label class="field"><span>Input Method</span><Select v-model="form.companyAddress.mode" :options="addressModes" option-label="label" option-value="value" :disabled="isExistingParent || form.sameAsSiteAddress" fluid /></label>
              <label v-if="!isExistingParent && !form.sameAsSiteAddress && form.companyAddress.mode === 'GMAPS_SIMULATION'" class="field"><span>Local Gmaps Suggestions</span><Select :options="data.options.addressSuggestions" option-label="previewAddress" placeholder="Choose local suggestion" fluid @change="applySuggestion('company', $event.value)" /></label>
              <div class="address-grid"><label class="field"><span>Province</span><InputText v-model="form.companyAddress.province" :disabled="isExistingParent || form.sameAsSiteAddress" fluid /></label><label class="field"><span>District</span><InputText v-model="form.companyAddress.district" :disabled="isExistingParent || form.sameAsSiteAddress" fluid /></label><label class="field"><span>Sub-District</span><InputText v-model="form.companyAddress.subDistrict" :disabled="isExistingParent || form.sameAsSiteAddress" fluid /></label><label class="field"><span>Village</span><InputText v-model="form.companyAddress.village" :disabled="isExistingParent || form.sameAsSiteAddress" fluid /></label><label class="field"><span>Latitude <em>Optional</em></span><InputNumber v-model="form.companyAddress.latitude" :disabled="isExistingParent || form.sameAsSiteAddress" :max-fraction-digits="7" fluid /></label><label class="field"><span>Longitude <em>Optional</em></span><InputNumber v-model="form.companyAddress.longitude" :disabled="isExistingParent || form.sameAsSiteAddress" :max-fraction-digits="7" fluid /></label></div>
              <label class="field"><span>Preview Address</span><Textarea v-model="form.companyAddress.previewAddress" :disabled="isExistingParent || form.sameAsSiteAddress" rows="3" fluid /></label>
            </article>
          </div>
        </section>

        <section class="form-section">
          <div class="section-heading"><span>03</span><div><h2>Contact Information</h2><p>Optional repeatable contacts. A public Google phone is not assumed to be a named PIC.</p></div></div>
          <div class="dual-cards"><article class="form-card"><div class="card-label"><strong>Customer Site Contacts</strong><Tag value="Optional" severity="secondary" /></div><div v-for="(contact, index) in form.siteContacts" :key="index" class="repeat-card"><div class="repeat-heading"><span>Site Contact {{ index + 1 }}</span><Button icon="pi pi-trash" severity="danger" text rounded type="button" @click="form.siteContacts.splice(index, 1)" /></div><div class="two-fields"><label class="field"><span>Contact Name</span><InputText v-model="contact.name" fluid /></label><label class="field"><span>Position</span><InputText v-model="contact.position" fluid /></label><label class="field"><span>Phone Number</span><InputText v-model="contact.phone" fluid /></label><label class="field"><span>Email Address</span><InputText v-model="contact.email" type="email" fluid /></label></div></div><Button label="Add Another Site Contact" icon="pi pi-plus" severity="secondary" outlined type="button" @click="addContact('site')" /></article>
            <article class="form-card"><div class="card-label"><strong>Company Contacts</strong><Tag :value="isExistingParent ? 'Master locked' : 'Optional'" severity="secondary" /></div><div v-for="(contact, index) in form.companyContacts" :key="index" class="repeat-card"><div class="repeat-heading"><span>Company Contact {{ index + 1 }}</span><Button v-if="!isExistingParent" icon="pi pi-trash" severity="danger" text rounded type="button" @click="form.companyContacts.splice(index, 1)" /></div><div class="two-fields"><label class="field"><span>Contact Name</span><InputText v-model="contact.name" :disabled="isExistingParent" fluid /></label><label class="field"><span>Position</span><InputText v-model="contact.position" :disabled="isExistingParent" fluid /></label><label class="field"><span>Phone Number</span><InputText v-model="contact.phone" :disabled="isExistingParent" fluid /></label><label class="field"><span>Email Address</span><InputText v-model="contact.email" :disabled="isExistingParent" type="email" fluid /></label></div></div><Button v-if="!isExistingParent" label="Add Another Company Contact" icon="pi pi-plus" severity="secondary" outlined type="button" @click="addContact('company')" /></article></div>
        </section>

        <section class="form-section"><div class="section-heading"><span>04</span><div><h2>Tax Information</h2><p>Optional data; never sourced from Google Places.</p></div></div><div class="dual-cards"><article class="form-card"><div class="card-label"><strong>Site Tax and Identity</strong><Tag value="Optional" severity="secondary" /></div><label class="field"><span>PPN</span><Select v-model="form.ppn" :options="['PKP', 'Non-PKP']" show-clear fluid /></label><label class="field"><span>ID TKU Number</span><InputText v-model="form.idTkuNumber" fluid /></label><label class="field"><span>NIK</span><InputText v-model="form.nik" fluid /></label></article><article class="form-card"><div class="card-label"><strong>Company Tax</strong><Tag :value="isExistingParent ? 'Master synced' : 'Optional'" severity="secondary" /></div><label class="field"><span>Company NPWP Name</span><InputText v-model="form.companyNpwpName" :disabled="isExistingParent" fluid /></label><label class="field"><span>Company NPWP Address</span><Textarea v-model="form.companyNpwpAddress" :disabled="isExistingParent" rows="2" fluid /></label><label class="field"><span>Company NPWP Number</span><InputText v-model="form.companyNpwpNumber" :disabled="isExistingParent" fluid /></label></article></div></section>

        <section class="form-section"><div class="section-heading"><span>05</span><div><h2>Other Master Data</h2><p>Local options simulate ERP master selections.</p></div></div><div class="dual-cards"><article class="form-card"><div class="card-label"><strong>Customer Site</strong><Tag value="Optional" severity="secondary" /></div><label class="field"><span>Shipment Cost</span><Select v-model="form.shipmentCost" :options="data.options.shipmentCosts" show-clear fluid /></label><label class="field"><span>Invoice Type</span><Select v-model="form.invoiceType" :options="data.options.invoiceTypes" show-clear fluid /></label><label class="field"><span>Bank Account</span><InputText v-model="form.bankAccount" placeholder="Simulation reference" fluid /></label></article><article class="form-card"><div class="card-label"><strong>Company</strong><Tag :value="isExistingParent ? 'Master synced' : 'Optional'" severity="secondary" /></div><label class="field"><span>Term of Payment</span><Select v-model="form.termOfPayment" :options="data.options.termsOfPayment" :disabled="isExistingParent" show-clear fluid /></label></article></div></section>

        <section class="form-section"><div class="section-heading"><span>06</span><div><h2>Billing and Shipment Configuration</h2><p>Optional Document Header preview.</p></div></div><article class="form-card"><div class="document-header"><div><span>Seller Identity</span><strong>{{ data.sellerIdentity }}</strong></div><div><span>Customer ID</span><strong>{{ customerCodePreview }}</strong></div></div><div class="two-fields"><label class="field"><span>Bill To Source</span><Select v-model="form.billToSource" :options="documentSources" show-clear fluid /></label><label class="field"><span>Ship To Source</span><Select v-model="form.shipToSource" :options="documentSources" show-clear fluid /></label></div><div class="dual-preview"><div><span>Billing address preview</span><p>{{ billPreview }}</p></div><div><span>Shipment address preview</span><p>{{ shipPreview }}</p></div></div></article></section>

        <section class="form-section"><div class="section-heading"><span>07</span><div><h2>Sales Assignment and Company KAM</h2><p>Sales users come from active application users; KAM remains ERP/master data only.</p></div></div><div class="dual-cards"><article class="form-card"><div class="card-label"><strong>Sales Assignment</strong><Tag value="Primary required" severity="warn" /></div><label class="field"><span>Sales Executive *</span><Select v-model="form.salesExecutiveId" :options="data.salesExecutives" option-label="fullName" option-value="id" filter fluid /></label><div v-for="(assignment, index) in form.salesAssignments" :key="index" class="repeat-card"><div class="repeat-heading"><span>Additional Sales Assignment {{ index + 1 }}</span><Button icon="pi pi-trash" severity="danger" text rounded type="button" @click="form.salesAssignments.splice(index, 1)" /></div><label class="field"><span>Sales Executive</span><Select v-model="assignment.ownerId" :options="data.salesExecutives" option-label="fullName" option-value="id" @change="assignment.ownerName = salesName($event.value)" fluid /></label><div class="period-grid"><Select v-model="assignment.startMonth" :options="months" option-label="label" option-value="value" /><Select v-model="assignment.startYear" :options="years" /><InputText v-model="assignment.end" placeholder="UNTIL_NOW or YYYY-MM" /></div></div><Button label="Add Another Sales Assignment" icon="pi pi-plus" severity="secondary" outlined type="button" @click="addAssignment('sales')" /></article><article class="form-card"><div class="card-label"><strong>Company KAM</strong><Tag value="Optional master data" severity="secondary" /></div><div v-for="(assignment, index) in form.kamAssignments" :key="index" class="repeat-card"><div class="repeat-heading"><span>KAM Assignment {{ index + 1 }}</span><Button v-if="!isExistingParent" icon="pi pi-trash" severity="danger" text rounded type="button" @click="form.kamAssignments.splice(index, 1)" /></div><label class="field"><span>Key Account Manager</span><Select v-model="assignment.ownerName" :options="data.options.kams" :disabled="isExistingParent" @change="assignment.ownerId = $event.value" fluid /></label><div class="period-grid"><Select v-model="assignment.startMonth" :options="months" option-label="label" option-value="value" :disabled="isExistingParent" /><Select v-model="assignment.startYear" :options="years" :disabled="isExistingParent" /><InputText v-model="assignment.end" :disabled="isExistingParent" placeholder="UNTIL_NOW or YYYY-MM" /></div></div><Button v-if="!isExistingParent" label="Add Another KAM Assignment" icon="pi pi-plus" severity="secondary" outlined type="button" @click="addAssignment('kam')" /></article></div></section>

        <div class="conversion-actions"><Button label="Cancel" severity="secondary" outlined type="button" @click="router.push(`/admin/prospects/${route.params.id}/review`)" /><Button label="Convert to Customer Existing" icon="pi pi-check" type="submit" :loading="saving" /></div>
      </div>

      <aside class="scope-panel"><p class="eyebrow">Field Scope & Requirements</p><h3>Conversion checklist</h3><div><Tag value="Required" severity="danger" /><p>Name, segment, category, company method, core Site address, active Sales Executive.</p></div><div><Tag value="Autofill + review" severity="info" /><p>Place name, formatted address, coordinates, category suggestion, phone, assigned Sales Executive.</p></div><div><Tag value="System generated" severity="success" /><p>Locked Parent Code and Customer Code previews. Final values are returned after save.</p></div><div><Tag value="Optional" severity="secondary" /><p>Contacts, tax, ERP master simulation, document header, additional Sales and KAM periods.</p></div><div class="scope-warning"><i class="pi pi-info-circle" /><p>No live Google/ERP call is made. KAM is not an application login role.</p></div></aside>
    </form>
  </section>
</template>
