<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, reactive, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import Button from 'primevue/button'
import Checkbox from 'primevue/checkbox'
import Dialog from 'primevue/dialog'
import InputNumber from 'primevue/inputnumber'
import InputText from 'primevue/inputtext'
import Message from 'primevue/message'
import RadioButton from 'primevue/radiobutton'
import Select from 'primevue/select'
import Tag from 'primevue/tag'
import Textarea from 'primevue/textarea'
import { convertProspect, getConversionForm, searchParentCompanies } from '../../../api/crm'
import { useCrmStore } from '../../../stores/crm'
import type { Address, Contact, ConversionFormData, ConversionInput, ParentCompany, PeriodAssignment } from '../../../types/crm'

const route = useRoute()
const router = useRouter()
const crm = useCrmStore()
const data = ref<ConversionFormData | null>(null)
const error = ref('')
const submitted = ref(false)
const saving = ref(false)
const companySuggestions = ref<ParentCompany[]>([])
const showConfirmDialog = ref(false)
let searchTimeout: ReturnType<typeof setTimeout> | null = null

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

const sectionStatus = computed(() => [
  {
    id: 'sec-01', label: 'Customer Info', required: true,
    done: !!form.customerName.trim() && !!form.customerSegment && !!form.customerCategory && !!form.parentMethod && (isExistingParent.value ? !!form.existingParentCompanyId : !!form.parentCompanyName.trim()),
  },
  {
    id: 'sec-02', label: 'Address', required: true,
    done: !!form.siteAddress.province && !!form.siteAddress.district && !!form.siteAddress.subDistrict && !!form.siteAddress.village && !!form.siteAddress.previewAddress.trim(),
  },
  { id: 'sec-03', label: 'Contacts', required: false, done: true },
  { id: 'sec-04', label: 'Tax', required: false, done: true },
  { id: 'sec-05', label: 'Master Data', required: false, done: true },
  { id: 'sec-06', label: 'Billing', required: false, done: true },
  {
    id: 'sec-07', label: 'Sales', required: true,
    done: !!form.salesExecutiveId,
  },
])

const requiredSectionsDone = computed(() => sectionStatus.value.filter((s) => s.required).every((s) => s.done))
const progressPercent = computed(() => {
  const required = sectionStatus.value.filter((s) => s.required)
  return Math.round((required.filter((s) => s.done).length / required.length) * 100)
})

const googleDataIssues = computed(() => {
  if (!data.value) return []
  const p = data.value.prospect.prospect
  const issues: string[] = []
  if (!p.phoneNumber) issues.push('No phone number stored')
  if (p.latitude == null || p.longitude == null) issues.push('No GPS coordinates stored')
  if (!p.websiteUrl) issues.push('No website URL stored')
  if (!p.formattedAddress) issues.push('No formatted address stored')
  return issues
})

function scrollToSection(id: string) {
  document.getElementById(id)?.scrollIntoView({ behavior: 'smooth', block: 'start' })
}

function parseAddressFromFormatted(formatted: string): { province: string; district: string; subDistrict: string; village: string } {
  const parts = formatted.split(',').map((s) => s.trim()).filter(Boolean)
  const result = { province: '', district: '', subDistrict: '', village: '' }
  if (parts.length >= 4) {
    result.village = parts[parts.length - 4]
    result.subDistrict = parts[parts.length - 3]
    result.district = parts[parts.length - 2]
    result.province = parts[parts.length - 1]
  } else if (parts.length === 3) {
    result.subDistrict = parts[0]
    result.district = parts[1]
    result.province = parts[2]
  } else if (parts.length === 2) {
    result.district = parts[0]
    result.province = parts[1]
  }
  return result
}

function autoParseAddress() {
  const source = form.siteAddress.previewAddress || data.value?.prospect.prospect.formattedAddress || ''
  if (!source) return
  const parsed = parseAddressFromFormatted(source)
  if (parsed.province) form.siteAddress.province = parsed.province
  if (parsed.district) form.siteAddress.district = parsed.district
  if (parsed.subDistrict) form.siteAddress.subDistrict = parsed.subDistrict
  if (parsed.village) form.siteAddress.village = parsed.village
}

function suggestCompanies(name: string) {
  if (searchTimeout) clearTimeout(searchTimeout)
  if (!name || name.length < 2 || isExistingParent.value) {
    companySuggestions.value = []
    return
  }
  searchTimeout = setTimeout(async () => {
    try {
      const results = await searchParentCompanies(name)
      companySuggestions.value = results.filter((c) => c.name.toLowerCase() !== name.toLowerCase())
    } catch {
      companySuggestions.value = []
    }
  }, 300)
}

function selectCompanySuggestion(company: ParentCompany) {
  form.existingParentCompanyId = company.id
  form.parentMethod = 'EXISTING_COMPANY'
  companySuggestions.value = []
}

onMounted(async () => {
  try {
    data.value = await getConversionForm(String(route.params.id))
    const prospect = data.value.prospect.prospect
    form.customerName = prospect.placeName
    form.customerCategory = data.value.options.categories.includes(prospect.placeCategory) ? prospect.placeCategory : ''
    form.siteAddress = { mode: 'GMAPS_SIMULATION', province: '', district: '', subDistrict: '', village: '', latitude: prospect.latitude, longitude: prospect.longitude, previewAddress: prospect.formattedAddress }
    form.siteContacts = prospect.phoneNumber ? [blankContact(prospect.phoneNumber)] : []
    form.salesExecutiveId = prospect.assignedSalesExecutiveId
    autoParseAddress()
  } catch (caught) {
    error.value = crm.errorMessage(caught)
  }
})

onBeforeUnmount(() => { if (searchTimeout) clearTimeout(searchTimeout) })

watch(() => form.customerName, (name) => {
  if (form.parentMethod === 'MATCH_CUSTOMER_NAME') form.parentCompanyName = name
})

watch(() => form.parentMethod, (method) => {
  form.existingParentCompanyId = null
  form.sameAsSiteAddress = false
  if (method === 'MATCH_CUSTOMER_NAME') form.parentCompanyName = form.customerName
  if (method === 'EXISTING_COMPANY') form.parentCompanyName = ''
  companySuggestions.value = []
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

watch(() => form.parentCompanyName, (name) => suggestCompanies(name))

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

function confirmConvert() {
  submitted.value = true
  error.value = ''
  if (coreInvalid()) {
    error.value = 'Complete all required Customer Site, Parent Company, core address, and Sales Executive fields.'
    window.scrollTo({ top: 0, behavior: 'smooth' })
    return
  }
  form.billingAddressPreview = form.billToSource ? billPreview.value : ''
  form.shippingAddressPreview = form.shipToSource ? shipPreview.value : ''
  showConfirmDialog.value = true
}

async function executeConvert() {
  showConfirmDialog.value = false
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
    <div class="page-heading"><div><p class="eyebrow"></p><h1>Convert to Customer Existing</h1><p class="muted">Review the Google snapshot below, complete required master data, and create the Customer Site atomically.</p></div><Tag value="WON — eligible" severity="success" /></div>
    <Message v-if="error" severity="error" :closable="false">{{ error }}</Message>
    <div v-if="!data && !error" class="empty-state">Preparing conversion form…</div>

    <template v-if="data">
      <article class="snapshot-card">
        <div class="snapshot-header">
          <div><p class="eyebrow">Google Snapshot</p><h2>{{ data.prospect.prospect.placeName }}</h2></div>
          <Tag value="Source data — read only" severity="secondary" />
        </div>
        <div class="snapshot-grid">
          <div><span>Place Name</span><strong>{{ data.prospect.prospect.placeName }}</strong><small class="source-badge source-google">Google</small></div>
          <div><span>Category</span><strong>{{ data.prospect.prospect.placeCategory }}</strong><small class="source-badge source-google">Google</small></div>
          <div><span>Industry Group</span><strong>{{ data.prospect.prospect.industryGroup }}</strong><small class="source-badge source-crm">CRM Mapping</small></div>
          <div class="wide"><span>Formatted Address</span><strong>{{ data.prospect.prospect.formattedAddress }}</strong><small class="source-badge source-google">Google</small></div>
          <div><span>Phone</span><strong>{{ data.prospect.prospect.phoneNumber || 'Not available' }}</strong><small class="source-badge source-google">Google</small></div>
          <div><span>Website</span><strong><a v-if="data.prospect.prospect.websiteUrl" :href="data.prospect.prospect.websiteUrl" target="_blank" rel="noopener">{{ data.prospect.prospect.websiteUrl }}</a><span v-else>Not available</span></strong><small class="source-badge source-google">Google</small></div>
          <div><span>Coordinates</span><strong v-if="data.prospect.prospect.latitude != null">{{ data.prospect.prospect.latitude.toFixed(6) }}, {{ data.prospect.prospect.longitude?.toFixed(6) }}</strong><strong v-else>Not available</strong><small class="source-badge source-google">Google</small></div>
          <div><span>Google Place ID</span><strong class="mono">{{ data.prospect.prospect.googlePlaceId }}</strong><small class="source-badge source-google">Google</small></div>
          <div><span>Sales Executive</span><strong>{{ data.prospect.prospect.assignedSalesExecutive }}</strong><small class="source-badge source-crm">CRM Mapping</small></div>
          <div><span>Visit Notes</span><strong>{{ data.prospect.prospect.visitNotes || '—' }}</strong><small class="source-badge source-crm">CRM Mapping</small></div>
          <div><span>Follow-up Notes</span><strong>{{ data.prospect.prospect.followUpNotes || '—' }}</strong><small class="source-badge source-crm">CRM Mapping</small></div>
        </div>
      </article>

      <Message v-if="googleDataIssues.length" severity="warn" :closable="false">
        <strong>Google snapshot incomplete:</strong> {{ googleDataIssues.join(' · ') }}. Verify address fields manually.
      </Message>

      <div class="progress-rail">
        <div class="progress-track"><div class="progress-fill" :style="{ width: progressPercent + '%' }" /></div>
        <div class="progress-labels">
          <button v-for="s in sectionStatus" :key="s.id" type="button" class="progress-label" :class="{ done: s.done, active: false }" @click="scrollToSection(s.id)">
            <i :class="s.done ? 'pi pi-check-circle' : s.required ? 'pi pi-exclamation-circle' : 'pi pi-circle'" />
            <span>{{ s.label }}</span>
          </button>
        </div>
      </div>

      <form class="conversion-layout" @submit.prevent="confirmConvert">
        <div class="conversion-main">
          <section id="sec-01" class="form-section">
            <div class="section-heading"><span>01</span><div><h2>Customer Information</h2><p>Customer Site identity and Parent Company relationship.</p></div></div>
            <div class="dual-cards">
              <article class="form-card"><div class="card-label"><strong>Customer Site</strong><Tag value="Required" severity="danger" /></div>
                <label class="field"><span>Customer Name / Outlet / Branch / Store * <small class="source-badge source-google">Google</small></span><InputText v-model="form.customerName" :invalid="submitted && !form.customerName.trim()" fluid /></label>
                <div class="two-fields"><label class="field"><span>Customer Segment * <small class="source-badge source-manual">Manual</small></span><Select v-model="form.customerSegment" :options="data.options.segments" placeholder="Select segment" :invalid="submitted && !form.customerSegment" fluid /></label><label class="field"><span>Customer Category * <small class="source-badge source-google">Google</small></span><Select v-model="form.customerCategory" :options="data.options.categories" placeholder="Select category" :invalid="submitted && !form.customerCategory" fluid /></label></div>
                <div class="code-preview"><label class="field"><span>Parent Code Preview <small class="source-badge source-system">System Generated</small></span><InputText :model-value="parentCodePreview" disabled fluid /></label><label class="field"><span>Customer Code Preview <small class="source-badge source-system">System Generated</small></span><InputText :model-value="customerCodePreview" disabled fluid /></label><small>Simulation-only codes are safely generated by the backend on save.</small></div>
              </article>
              <article class="form-card"><div class="card-label"><strong>Parent Company</strong><Tag value="Required" severity="danger" /></div>
                <div class="radio-stack" :class="{ 'field-invalid': submitted && !form.parentMethod }"><label v-for="method in parentMethods" :key="method.value"><RadioButton v-model="form.parentMethod" name="parentMethod" :value="method.value" /><span>{{ method.label }}</span></label></div>
                <label v-if="isExistingParent" class="field"><span>Search Existing Company * <small class="source-badge source-system">System Generated</small></span><Select v-model="form.existingParentCompanyId" :options="data.parentCompanies" option-label="name" option-value="id" :invalid="submitted && isExistingParent && !form.existingParentCompanyId" filter placeholder="Search name or code" fluid><template #option="slot"><div><strong>{{ slot.option.name }}</strong><small class="block-muted">{{ slot.option.parentCode }}</small></div></template></Select></label>
                <template v-else>
                  <label class="field"><span>Customer Company / Parent * <small class="source-badge source-manual">Manual</small></span><InputText v-model="form.parentCompanyName" :disabled="form.parentMethod === 'MATCH_CUSTOMER_NAME'" :invalid="submitted && !!form.parentMethod && !form.parentCompanyName.trim()" fluid /></label>
                  <div v-if="companySuggestions.length" class="company-suggestions">
                    <p class="suggestion-label"><i class="pi pi-lightbulb" /> Existing companies that match:</p>
                    <button v-for="suggestion in companySuggestions" :key="suggestion.id" type="button" class="suggestion-item" @click="selectCompanySuggestion(suggestion)">
                      <strong>{{ suggestion.name }}</strong><span>{{ suggestion.parentCode }} — click to link as Existing Company</span>
                    </button>
                  </div>
                </template>
                <div v-if="selectedParent" class="master-preview"><span>Locked master preview</span><strong>{{ selectedParent.parentCode }} · {{ selectedParent.name }}</strong><p>{{ selectedParent.address.previewAddress }}</p></div>
              </article>
            </div>
          </section>

          <section id="sec-02" class="form-section">
            <div class="section-heading"><span>02</span><div><h2>Address Information</h2><p>Google snapshot prefill with explicit Site and Company scopes.</p></div></div>
            <div class="dual-cards">
              <article class="form-card"><div class="card-label"><strong>Site Address</strong><Tag value="Required" severity="danger" /></div>
                <label class="field"><span>Input Method * <small class="source-badge source-google">Google</small></span><Select v-model="form.siteAddress.mode" :options="addressModes" option-label="label" option-value="value" :invalid="submitted && !form.siteAddress.mode" fluid /></label>
                <label v-if="form.siteAddress.mode === 'GMAPS_SIMULATION'" class="field"><span>Local Gmaps Suggestions</span><Select :options="data.options.addressSuggestions" option-label="previewAddress" placeholder="Choose local suggestion" fluid @change="applySuggestion('site', $event.value)" /></label>
                <div class="address-grid"><label class="field"><span>Province * <small class="source-badge source-parsed">Parsed</small></span><InputText v-model="form.siteAddress.province" :invalid="submitted && !form.siteAddress.province" fluid /></label><label class="field"><span>District * <small class="source-badge source-parsed">Parsed</small></span><InputText v-model="form.siteAddress.district" :invalid="submitted && !form.siteAddress.district" fluid /></label><label class="field"><span>Sub-District * <small class="source-badge source-parsed">Parsed</small></span><InputText v-model="form.siteAddress.subDistrict" :invalid="submitted && !form.siteAddress.subDistrict" fluid /></label><label class="field"><span>Village * <small class="source-badge source-parsed">Parsed</small></span><InputText v-model="form.siteAddress.village" :invalid="submitted && !form.siteAddress.village" fluid /></label><label class="field"><span>Latitude <em>Optional</em> <small class="source-badge source-google">Google</small></span><InputNumber v-model="form.siteAddress.latitude" :min-fraction-digits="4" :max-fraction-digits="7" fluid /></label><label class="field"><span>Longitude <em>Optional</em> <small class="source-badge source-google">Google</small></span><InputNumber v-model="form.siteAddress.longitude" :min-fraction-digits="4" :max-fraction-digits="7" fluid /></label></div>
                <label class="field"><span>Preview Address * <small class="source-badge source-google">Google</small></span><Textarea v-model="form.siteAddress.previewAddress" :invalid="submitted && !form.siteAddress.previewAddress.trim()" rows="3" fluid /></label>
                <Button label="Auto-parse Province / District / SubDistrict / Village" icon="pi pi-wand" severity="secondary" outlined type="button" size="small" @click="autoParseAddress" />
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

          <section id="sec-03" class="form-section">
            <div class="section-heading"><span>03</span><div><h2>Contact Information</h2><p>Optional repeatable contacts. A public Google phone is not assumed to be a named PIC.</p></div></div>
            <div class="dual-cards"><article class="form-card"><div class="card-label"><strong>Customer Site Contacts</strong><Tag value="Optional" severity="secondary" /></div><div v-for="(contact, index) in form.siteContacts" :key="index" class="repeat-card"><div class="repeat-heading"><span>Site Contact {{ index + 1 }}</span><Button icon="pi pi-trash" severity="danger" text rounded type="button" @click="form.siteContacts.splice(index, 1)" /></div><div class="two-fields"><label class="field"><span>Contact Name <small class="source-badge source-manual">Manual</small></span><InputText v-model="contact.name" fluid /></label><label class="field"><span>Position <small class="source-badge source-manual">Manual</small></span><InputText v-model="contact.position" fluid /></label><label class="field"><span>Phone Number <small class="source-badge source-google">Google</small></span><InputText v-model="contact.phone" fluid /></label><label class="field"><span>Email Address <small class="source-badge source-manual">Manual</small></span><InputText v-model="contact.email" type="email" fluid /></label></div></div><Button label="Add Another Site Contact" icon="pi pi-plus" severity="secondary" outlined type="button" @click="addContact('site')" /></article>
              <article class="form-card"><div class="card-label"><strong>Company Contacts</strong><Tag :value="isExistingParent ? 'Master locked' : 'Optional'" severity="secondary" /></div><div v-for="(contact, index) in form.companyContacts" :key="index" class="repeat-card"><div class="repeat-heading"><span>Company Contact {{ index + 1 }}</span><Button v-if="!isExistingParent" icon="pi pi-trash" severity="danger" text rounded type="button" @click="form.companyContacts.splice(index, 1)" /></div><div class="two-fields"><label class="field"><span>Contact Name</span><InputText v-model="contact.name" :disabled="isExistingParent" fluid /></label><label class="field"><span>Position</span><InputText v-model="contact.position" :disabled="isExistingParent" fluid /></label><label class="field"><span>Phone Number</span><InputText v-model="contact.phone" :disabled="isExistingParent" fluid /></label><label class="field"><span>Email Address</span><InputText v-model="contact.email" :disabled="isExistingParent" type="email" fluid /></label></div></div><Button v-if="!isExistingParent" label="Add Another Company Contact" icon="pi pi-plus" severity="secondary" outlined type="button" @click="addContact('company')" /></article></div>
          </section>

          <section id="sec-04" class="form-section"><div class="section-heading"><span>04</span><div><h2>Tax Information</h2><p>Optional data; never sourced from Google Places.</p></div></div><div class="dual-cards"><article class="form-card"><div class="card-label"><strong>Site Tax and Identity</strong><Tag value="Optional" severity="secondary" /></div><label class="field"><span>PPN <small class="source-badge source-manual">Manual</small></span><Select v-model="form.ppn" :options="['PKP', 'Non-PKP']" show-clear fluid /></label><label class="field"><span>ID TKU Number <small class="source-badge source-manual">Manual</small></span><InputText v-model="form.idTkuNumber" fluid /></label><label class="field"><span>NIK <small class="source-badge source-manual">Manual</small></span><InputText v-model="form.nik" fluid /></label></article><article class="form-card"><div class="card-label"><strong>Company Tax</strong><Tag :value="isExistingParent ? 'Master synced' : 'Optional'" severity="secondary" /></div><label class="field"><span>Company NPWP Name</span><InputText v-model="form.companyNpwpName" :disabled="isExistingParent" fluid /></label><label class="field"><span>Company NPWP Address</span><Textarea v-model="form.companyNpwpAddress" :disabled="isExistingParent" rows="2" fluid /></label><label class="field"><span>Company NPWP Number</span><InputText v-model="form.companyNpwpNumber" :disabled="isExistingParent" fluid /></label></article></div></section>

          <section id="sec-05" class="form-section"><div class="section-heading"><span>05</span><div><h2>Other Master Data</h2><p>Local options simulate ERP master selections.</p></div></div><div class="dual-cards"><article class="form-card"><div class="card-label"><strong>Customer Site</strong><Tag value="Optional" severity="secondary" /></div><label class="field"><span>Shipment Cost <small class="source-badge source-manual">Manual</small></span><Select v-model="form.shipmentCost" :options="data.options.shipmentCosts" show-clear fluid /></label><label class="field"><span>Invoice Type <small class="source-badge source-manual">Manual</small></span><Select v-model="form.invoiceType" :options="data.options.invoiceTypes" show-clear fluid /></label><label class="field"><span>Bank Account <small class="source-badge source-manual">Manual</small></span><InputText v-model="form.bankAccount" placeholder="Simulation reference" fluid /></label></article><article class="form-card"><div class="card-label"><strong>Company</strong><Tag :value="isExistingParent ? 'Master synced' : 'Optional'" severity="secondary" /></div><label class="field"><span>Term of Payment</span><Select v-model="form.termOfPayment" :options="data.options.termsOfPayment" :disabled="isExistingParent" show-clear fluid /></label></article></div></section>

          <section id="sec-06" class="form-section"><div class="section-heading"><span>06</span><div><h2>Billing and Shipment Configuration</h2><p>Optional Document Header preview.</p></div></div><article class="form-card"><div class="document-header"><div><span>Seller Identity</span><strong>{{ data.sellerIdentity }}</strong></div><div><span>Customer ID</span><strong>{{ customerCodePreview }}</strong></div></div><div class="two-fields"><label class="field"><span>Bill To Source <small class="source-badge source-manual">Manual</small></span><Select v-model="form.billToSource" :options="documentSources" show-clear fluid /></label><label class="field"><span>Ship To Source <small class="source-badge source-manual">Manual</small></span><Select v-model="form.shipToSource" :options="documentSources" show-clear fluid /></label></div><div class="dual-preview"><div><span>Billing address preview</span><p>{{ billPreview }}</p></div><div><span>Shipment address preview</span><p>{{ shipPreview }}</p></div></div></article></section>

          <section id="sec-07" class="form-section"><div class="section-heading"><span>07</span><div><h2>Sales Assignment and Company KAM</h2><p>Sales users come from active application users; KAM remains ERP/master data only.</p></div></div><div class="dual-cards"><article class="form-card"><div class="card-label"><strong>Sales Assignment</strong><Tag value="Primary required" severity="warn" /></div><label class="field"><span>Sales Executive * <small class="source-badge source-google">Google</small></span><Select v-model="form.salesExecutiveId" :options="data.salesExecutives" option-label="fullName" option-value="id" :invalid="submitted && !form.salesExecutiveId" filter fluid /></label><div v-for="(assignment, index) in form.salesAssignments" :key="index" class="repeat-card"><div class="repeat-heading"><span>Additional Sales Assignment {{ index + 1 }}</span><Button icon="pi pi-trash" severity="danger" text rounded type="button" @click="form.salesAssignments.splice(index, 1)" /></div><label class="field"><span>Sales Executive</span><Select v-model="assignment.ownerId" :options="data.salesExecutives" option-label="fullName" option-value="id" @change="assignment.ownerName = salesName($event.value)" fluid /></label><div class="period-grid"><Select v-model="assignment.startMonth" :options="months" option-label="label" option-value="value" /><Select v-model="assignment.startYear" :options="years" /><InputText v-model="assignment.end" placeholder="UNTIL_NOW or YYYY-MM" /></div></div><Button label="Add Another Sales Assignment" icon="pi pi-plus" severity="secondary" outlined type="button" @click="addAssignment('sales')" /></article><article class="form-card"><div class="card-label"><strong>Company KAM</strong><Tag value="Optional master data" severity="secondary" /></div><div v-for="(assignment, index) in form.kamAssignments" :key="index" class="repeat-card"><div class="repeat-heading"><span>KAM Assignment {{ index + 1 }}</span><Button v-if="!isExistingParent" icon="pi pi-trash" severity="danger" text rounded type="button" @click="form.kamAssignments.splice(index, 1)" /></div><label class="field"><span>KAM Name</span><Select v-model="assignment.ownerName" :options="data.options.kams" :disabled="isExistingParent" fluid /></label><div class="period-grid"><Select v-model="assignment.startMonth" :options="months" option-label="label" option-value="value" /><Select v-model="assignment.startYear" :options="years" /><InputText v-model="assignment.end" placeholder="UNTIL_NOW or YYYY-MM" /></div></div><Button v-if="!isExistingParent" label="Add Another KAM Assignment" icon="pi pi-plus" severity="secondary" outlined type="button" @click="addAssignment('kam')" /></article></div></section>

          <div class="conversion-actions"><Button label="Cancel" severity="secondary" outlined type="button" @click="router.push(`/admin/prospects/${route.params.id}/review`)" /><Button label="Convert to Customer Existing" icon="pi pi-check" type="submit" :loading="saving" /></div>
        </div>

        <aside class="scope-panel">
          <p class="eyebrow">Live status</p>
          <h3>{{ requiredSectionsDone ? 'Ready to convert' : `${progressPercent}% complete` }}</h3>
          <div v-for="s in sectionStatus" :key="s.id" class="scope-status-row">
            <i :class="s.done ? 'pi pi-check-circle scope-ok' : s.required ? 'pi pi-exclamation-circle scope-warn' : 'pi pi-circle scope-muted'" />
            <span :class="{ 'scope-done-text': s.done }">{{ s.label }}<template v-if="s.required"> *</template></span>
          </div>
          <div v-if="googleDataIssues.length" class="scope-warning" style="margin-top:0.8rem"><i class="pi pi-exclamation-triangle" /><p>Google snapshot is missing {{ googleDataIssues.length }} field(s). Verify address manually.</p></div>
          <div class="scope-warning"><i class="pi pi-info-circle" /><p>No live Google/ERP call is made. KAM is not an application login role.</p></div>
        </aside>
      </form>
    </template>

    <Dialog v-model:visible="showConfirmDialog" header="Confirm Conversion" :modal="true" :closable="true" :style="{ width: '480px' }">
      <div class="confirm-body">
        <p>You are about to convert this Won prospect into a Customer Existing record. This action is <strong>atomic and irreversible</strong>.</p>
        <div class="confirm-summary">
          <div><span>Customer Name</span><strong>{{ form.customerName }}</strong></div>
          <div><span>Parent Company</span><strong>{{ isExistingParent ? selectedParent?.name || '—' : form.parentCompanyName || '—' }}</strong></div>
          <div><span>Site Address</span><strong>{{ form.siteAddress.previewAddress || '—' }}</strong></div>
          <div><span>Sales Executive</span><strong>{{ data?.salesExecutives.find(e => e.id === form.salesExecutiveId)?.fullName || '—' }}</strong></div>
          <div><span>Category</span><strong>{{ form.customerCategory || '—' }}</strong></div>
          <div><span>Segment</span><strong>{{ form.customerSegment || '—' }}</strong></div>
        </div>
        <Message severity="info" :closable="false">A Parent Company and Customer Site record will be created in a single database transaction.</Message>
        <Message v-if="googleDataIssues.length" severity="warn" :closable="false">Google snapshot gaps: {{ googleDataIssues.join(', ') }}. You may want to verify address data.</Message>
      </div>
      <template #footer>
        <Button label="Go Back" severity="secondary" outlined @click="showConfirmDialog = false" />
        <Button label="Convert Now" icon="pi pi-check" :loading="saving" @click="executeConvert" />
      </template>
    </Dialog>
  </section>
</template>
