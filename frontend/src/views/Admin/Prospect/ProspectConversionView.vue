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
import axios from 'axios'
import { convertProspect, getConversionForm, searchParentCompanies } from '../../../api/crm'
import { useCrmStore } from '../../../stores/crm'
import type { Address, Contact, ConversionFormData, ConversionInput, ParentCompany, PeriodAssignment } from '../../../types/crm'

function clone<T>(obj: T): T {
  return JSON.parse(JSON.stringify(obj))
}

const route = useRoute()
const router = useRouter()
const crm = useCrmStore()
const data = ref<ConversionFormData | null>(null)
const error = ref('')
const submitted = ref(false)
const saving = ref(false)
const loading = ref(true)
const timedOut = ref(false)
const companySuggestions = ref<ParentCompany[]>([])
const showConfirmDialog = ref(false)
let searchTimeout: ReturnType<typeof setTimeout> | null = null

const blankAddress = (): Address => ({ mode: '', province: '', district: '', subDistrict: '', village: '', latitude: null, longitude: null, previewAddress: '' })
const blankContact = (phone = ''): Contact => ({ name: '', position: '', phone, email: '' })
const blankAssignment = (): PeriodAssignment => ({ ownerId: '', ownerName: '', startMonth: new Date().getMonth() + 1, startYear: new Date().getFullYear(), end: 'UNTIL_NOW' })

const form = reactive<ConversionInput>({
  customerName: '', customerSegment: 'General Trade', customerCategory: '', parentMethod: 'MATCH_CUSTOMER_NAME', existingParentCompanyId: null,
  parentCompanyName: '', sameAsSiteAddress: true, siteAddress: blankAddress(), companyAddress: blankAddress(),
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

const collapsedSections = reactive<Record<string, boolean>>({ 'sec-03': true, 'sec-04': true, 'sec-05': true, 'sec-06': true })

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
  console.log('[ConversionForm] mounted, loading id:', route.params.id)

  const failTimer = setTimeout(() => {
    if (!data.value && !error.value) {
      timedOut.value = true
      error.value = 'Conversion form took too long to load. Cek console browser (F12 > Console) untuk detail.'
    }
  }, 15000)

  try {
    console.log('[ConversionForm] calling getConversionForm...')
    data.value = await getConversionForm(String(route.params.id))
    console.log('[ConversionForm] getConversionForm succeeded')
    const prospect = data.value.prospect.prospect
    const detail = data.value.placeDetails

    form.customerName = detail?.placeName || prospect.placeName

    const category = detail?.placeCategory || prospect.placeCategory
    form.customerCategory = data.value.options.categories.includes(category) ? category : ''

    const address = detail?.formattedAddress || prospect.formattedAddress
    form.siteAddress = { mode: 'GMAPS_SIMULATION', province: '', district: '', subDistrict: '', village: '', latitude: detail?.latitude ?? prospect.latitude, longitude: detail?.longitude ?? prospect.longitude, previewAddress: address }

    const phone = detail?.internationalPhone || detail?.phoneNumber || prospect.phoneNumber
    form.siteContacts = phone ? [blankContact(phone)] : []



    form.salesExecutiveId = prospect.assignedSalesExecutiveId
    autoParseAddress()

    if (data.value.options.segments.includes('General Trade')) {
      form.customerSegment = 'General Trade'
    }

    form.siteContacts = form.siteContacts.filter((c) => c.phone || c.email)

    if (form.customerName) {
      try {
        const results = await searchParentCompanies(form.customerName)
        const exactMatch = results.find((c) => c.name.toLowerCase() === form.customerName.toLowerCase())
        if (exactMatch) {
          form.existingParentCompanyId = exactMatch.id
          form.parentMethod = 'EXISTING_COMPANY'
        } else {
          form.parentMethod = 'MATCH_CUSTOMER_NAME'
          form.parentCompanyName = form.customerName
        }
      } catch {
        form.parentMethod = 'MATCH_CUSTOMER_NAME'
        form.parentCompanyName = form.customerName
      }
    }
  } catch (caught) {
    const msg = crm.errorMessage(caught)
    console.error('[ConversionForm] error:', caught, 'message:', msg)
    error.value = msg
  } finally {
    clearTimeout(failTimer)
    loading.value = false
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
  if (same && !isExistingParent.value) form.companyAddress = clone(form.siteAddress)
})

watch(() => form.siteAddress, (address) => {
  if (form.sameAsSiteAddress && !isExistingParent.value) form.companyAddress = clone(address)
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
    const msg = crm.errorMessage(caught)
    console.error('[ConversionForm] convert failed:', caught, 'msg:', msg)
    if (axios.isAxiosError(caught)) {
      console.error('[ConversionForm] response data:', caught.response?.data)
    }
    error.value = msg
    window.scrollTo({ top: 0, behavior: 'smooth' })
  } finally {
    saving.value = false
  }
}
</script>

<template>
  <section class="admin-page conversion-page">
    <div class="cv-header">
      <Button icon="pi pi-arrow-left" severity="secondary" text rounded @click="router.back()" title="Back" />
      <div class="cv-header-text">
        <p class="eyebrow">Prospect Conversion</p>
        <h1>Convert to Customer</h1>
        <p class="muted">Review the Google snapshot, complete required fields, and create the customer atomically.</p>
      </div>
      <Tag value="WON" severity="success" />
    </div>

    <Message v-if="error" severity="error" :closable="false">{{ error }}</Message>
    <div v-if="!data && !error" class="empty-state">Preparing conversion form...</div>

    <template v-if="data">
      <!-- Google Snapshot -->
      <article class="snapshot-card">
        <div class="snapshot-header">
          <div><p class="eyebrow">Google Snapshot</p><h2>{{ data.prospect.prospect.placeName }}</h2></div>
          <Tag value="Read only" severity="secondary" />
        </div>
        <div class="snapshot-grid">
          <div><span>Place Name</span><strong>{{ data.prospect.prospect.placeName }}</strong></div>
          <div><span>Category</span><strong>{{ data.prospect.prospect.placeCategory }}</strong></div>
          <div><span>Industry</span><strong>{{ data.prospect.prospect.industryGroup }}</strong></div>
          <div class="wide"><span>Address</span><strong>{{ data.prospect.prospect.formattedAddress }}</strong></div>
          <div><span>Phone</span><strong>{{ data.prospect.prospect.phoneNumber || 'Not available' }}</strong></div>
          <div><span>Website</span><strong><a v-if="data.prospect.prospect.websiteUrl" :href="data.prospect.prospect.websiteUrl" target="_blank" rel="noopener">{{ data.prospect.prospect.websiteUrl }}</a><span v-else>Not available</span></strong></div>
          <div><span>Sales</span><strong>{{ data.prospect.prospect.assignedSalesExecutive }}</strong></div>
        </div>
      </article>

      <Message v-if="googleDataIssues.length" severity="warn" :closable="false">
        <strong>Google snapshot incomplete:</strong> {{ googleDataIssues.join(' · ') }}.
      </Message>

      <!-- Progress Rail -->
      <div class="progress-rail">
        <div class="progress-track"><div class="progress-fill" :style="{ width: progressPercent + '%' }" /></div>
        <div class="progress-labels">
          <button v-for="s in sectionStatus" :key="s.id" type="button" class="progress-label" :class="{ done: s.done }" @click="scrollToSection(s.id)">
            <i :class="s.done ? 'pi pi-check-circle' : s.required ? 'pi pi-exclamation-circle' : 'pi pi-circle'" />
            <span>{{ s.label }}</span>
          </button>
        </div>
      </div>

      <!-- Form -->
      <form novalidate class="conversion-layout" @submit.prevent="confirmConvert">
        <div class="conversion-main">

          <!-- 01 Customer Information -->
          <section id="sec-01" class="form-section">
            <div class="section-heading">
              <span>01</span>
              <div><h2>Customer Information</h2><p>Customer Site identity and Parent Company relationship.</p></div>
            </div>
            <div class="dual-cards">
              <article class="form-card">
                <div class="card-label"><strong>Customer Site</strong><Tag value="Required" severity="danger" /></div>
                <label class="field">
                  <span>Customer Name / Outlet / Branch / Store *</span>
                  <InputText v-model="form.customerName" :invalid="submitted && !form.customerName.trim()" fluid />
                </label>
                <div class="two-fields">
                  <label class="field">
                    <span>Customer Segment *</span>
                    <Select v-model="form.customerSegment" :options="data.options.segments" placeholder="Select segment" :invalid="submitted && !form.customerSegment" fluid />
                  </label>
                  <label class="field">
                    <span>Customer Category *</span>
                    <Select v-model="form.customerCategory" :options="data.options.categories" placeholder="Select category" :invalid="submitted && !form.customerCategory" fluid />
                  </label>
                </div>
                <div class="code-preview">
                  <div class="two-fields">
                    <label class="field"><span>Parent Code Preview</span><InputText :model-value="parentCodePreview" disabled fluid /></label>
                    <label class="field"><span>Customer Code Preview</span><InputText :model-value="customerCodePreview" disabled fluid /></label>
                  </div>
                  <small>Codes are generated by the backend on save.</small>
                </div>
              </article>
              <article class="form-card">
                <div class="card-label"><strong>Parent Company</strong><Tag value="Required" severity="danger" /></div>
                <div class="radio-stack" :class="{ 'field-invalid': submitted && !form.parentMethod }">
                  <label v-for="method in parentMethods" :key="method.value">
                    <RadioButton v-model="form.parentMethod" name="parentMethod" :value="method.value" /><span>{{ method.label }}</span>
                  </label>
                </div>
                <label v-if="isExistingParent" class="field">
                  <span>Search Existing Company *</span>
                  <Select v-model="form.existingParentCompanyId" :options="data.parentCompanies" option-label="name" option-value="id" :invalid="submitted && isExistingParent && !form.existingParentCompanyId" filter placeholder="Search name or code" fluid>
                    <template #option="slot"><div><strong>{{ slot.option.name }}</strong><small class="block-muted">{{ slot.option.parentCode }}</small></div></template>
                  </Select>
                </label>
                <template v-else>
                  <label class="field">
                    <span>Customer Company / Parent *</span>
                    <InputText v-model="form.parentCompanyName" :disabled="form.parentMethod === 'MATCH_CUSTOMER_NAME'" :invalid="submitted && !!form.parentMethod && !form.parentCompanyName.trim()" fluid />
                  </label>
                  <div v-if="companySuggestions.length" class="company-suggestions">
                    <p class="suggestion-label"><i class="pi pi-lightbulb" /> Matching companies:</p>
                    <button v-for="suggestion in companySuggestions" :key="suggestion.id" type="button" class="suggestion-item" @click="selectCompanySuggestion(suggestion)">
                      <strong>{{ suggestion.name }}</strong><span>{{ suggestion.parentCode }}</span>
                    </button>
                  </div>
                </template>
                <div v-if="selectedParent" class="master-preview">
                  <span>Locked master preview</span>
                  <strong>{{ selectedParent.parentCode }} · {{ selectedParent.name }}</strong>
                  <p>{{ selectedParent.address.previewAddress }}</p>
                </div>
              </article>
            </div>
          </section>

          <!-- 02 Address -->
          <section id="sec-02" class="form-section">
            <div class="section-heading">
              <span>02</span>
              <div><h2>Address Information</h2><p>Site and Company address details.</p></div>
            </div>
            <div class="dual-cards">
              <article class="form-card">
                <div class="card-label"><strong>Site Address</strong><Tag value="Required" severity="danger" /></div>
                <label class="field"><span>Input Method *</span><Select v-model="form.siteAddress.mode" :options="addressModes" option-label="label" option-value="value" :invalid="submitted && !form.siteAddress.mode" fluid /></label>
                <label v-if="form.siteAddress.mode === 'GMAPS_SIMULATION'" class="field"><span>Local Gmaps Suggestions</span><Select :options="data.options.addressSuggestions" option-label="previewAddress" placeholder="Choose local suggestion" fluid @change="applySuggestion('site', $event.value)" /></label>
                <div class="address-grid">
                  <label class="field"><span>Province *</span><InputText v-model="form.siteAddress.province" :invalid="submitted && !form.siteAddress.province" fluid /></label>
                  <label class="field"><span>District *</span><InputText v-model="form.siteAddress.district" :invalid="submitted && !form.siteAddress.district" fluid /></label>
                  <label class="field"><span>Sub-District *</span><InputText v-model="form.siteAddress.subDistrict" :invalid="submitted && !form.siteAddress.subDistrict" fluid /></label>
                  <label class="field"><span>Village *</span><InputText v-model="form.siteAddress.village" :invalid="submitted && !form.siteAddress.village" fluid /></label>
                  <label class="field"><span>Latitude <em>Optional</em></span><InputNumber v-model="form.siteAddress.latitude" :min-fraction-digits="4" :max-fraction-digits="7" fluid /></label>
                  <label class="field"><span>Longitude <em>Optional</em></span><InputNumber v-model="form.siteAddress.longitude" :min-fraction-digits="4" :max-fraction-digits="7" fluid /></label>
                </div>
                <label class="field"><span>Preview Address *</span><Textarea v-model="form.siteAddress.previewAddress" :invalid="submitted && !form.siteAddress.previewAddress.trim()" rows="3" fluid /></label>
                <Button label="Auto-parse address fields" icon="pi pi-wand" severity="secondary" outlined type="button" size="small" @click="autoParseAddress" />
              </article>
              <article class="form-card" :class="{ 'locked-card': isExistingParent }">
                <div class="card-label"><strong>Company Address</strong><Tag :value="isExistingParent ? 'Locked' : 'Optional'" severity="secondary" /></div>
                <label v-if="!isExistingParent" class="check-row"><Checkbox v-model="form.sameAsSiteAddress" binary /><span>Same as Site Address</span></label>
                <label class="field"><span>Input Method</span><Select v-model="form.companyAddress.mode" :options="addressModes" option-label="label" option-value="value" :disabled="isExistingParent || form.sameAsSiteAddress" fluid /></label>
                <label v-if="!isExistingParent && !form.sameAsSiteAddress && form.companyAddress.mode === 'GMAPS_SIMULATION'" class="field"><span>Local Gmaps Suggestions</span><Select :options="data.options.addressSuggestions" option-label="previewAddress" placeholder="Choose local suggestion" fluid @change="applySuggestion('company', $event.value)" /></label>
                <div class="address-grid">
                  <label class="field"><span>Province</span><InputText v-model="form.companyAddress.province" :disabled="isExistingParent || form.sameAsSiteAddress" fluid /></label>
                  <label class="field"><span>District</span><InputText v-model="form.companyAddress.district" :disabled="isExistingParent || form.sameAsSiteAddress" fluid /></label>
                  <label class="field"><span>Sub-District</span><InputText v-model="form.companyAddress.subDistrict" :disabled="isExistingParent || form.sameAsSiteAddress" fluid /></label>
                  <label class="field"><span>Village</span><InputText v-model="form.companyAddress.village" :disabled="isExistingParent || form.sameAsSiteAddress" fluid /></label>
                  <label class="field"><span>Latitude <em>Optional</em></span><InputNumber v-model="form.companyAddress.latitude" :disabled="isExistingParent || form.sameAsSiteAddress" :max-fraction-digits="7" fluid /></label>
                  <label class="field"><span>Longitude <em>Optional</em></span><InputNumber v-model="form.companyAddress.longitude" :disabled="isExistingParent || form.sameAsSiteAddress" :max-fraction-digits="7" fluid /></label>
                </div>
                <label class="field"><span>Preview Address</span><Textarea v-model="form.companyAddress.previewAddress" :disabled="isExistingParent || form.sameAsSiteAddress" rows="3" fluid /></label>
              </article>
            </div>
          </section>

          <!-- 03 Contacts -->
          <section id="sec-03" class="form-section collapsible-section" :class="{ collapsed: collapsedSections['sec-03'] }">
            <button type="button" class="section-toggle" @click="collapsedSections['sec-03'] = !collapsedSections['sec-03']">
              <span>03</span>
              <div><h2>Contact Information</h2><p>Optional repeatable contacts for Site and Company.</p></div>
              <i :class="collapsedSections['sec-03'] ? 'pi pi-chevron-down' : 'pi pi-chevron-up'" />
            </button>
            <div v-show="!collapsedSections['sec-03']" class="section-body">
              <div class="dual-cards">
                <article class="form-card">
                  <div class="card-label"><strong>Site Contacts</strong><Tag value="Optional" severity="secondary" /></div>
                  <div v-for="(contact, index) in form.siteContacts" :key="index" class="repeat-card">
                    <div class="repeat-heading">
                      <span>Contact {{ index + 1 }}</span>
                      <Button icon="pi pi-trash" severity="danger" text rounded type="button" @click="form.siteContacts.splice(index, 1)" />
                    </div>
                    <div class="two-fields">
                      <label class="field"><span>Name</span><InputText v-model="contact.name" fluid /></label>
                      <label class="field"><span>Position</span><InputText v-model="contact.position" fluid /></label>
                    </div>
                    <div class="two-fields">
                      <label class="field"><span>Phone</span><InputText v-model="contact.phone" fluid /></label>
                      <label class="field"><span>Email</span><InputText v-model="contact.email" type="email" fluid /></label>
                    </div>
                  </div>
                  <Button label="Add Site Contact" icon="pi pi-plus" severity="secondary" outlined type="button" @click="addContact('site')" />
                </article>
                <article class="form-card">
                  <div class="card-label"><strong>Company Contacts</strong><Tag :value="isExistingParent ? 'Locked' : 'Optional'" severity="secondary" /></div>
                  <div v-for="(contact, index) in form.companyContacts" :key="index" class="repeat-card">
                    <div class="repeat-heading">
                      <span>Contact {{ index + 1 }}</span>
                      <Button v-if="!isExistingParent" icon="pi pi-trash" severity="danger" text rounded type="button" @click="form.companyContacts.splice(index, 1)" />
                    </div>
                    <div class="two-fields">
                      <label class="field"><span>Name</span><InputText v-model="contact.name" :disabled="isExistingParent" fluid /></label>
                      <label class="field"><span>Position</span><InputText v-model="contact.position" :disabled="isExistingParent" fluid /></label>
                    </div>
                    <div class="two-fields">
                      <label class="field"><span>Phone</span><InputText v-model="contact.phone" :disabled="isExistingParent" fluid /></label>
                      <label class="field"><span>Email</span><InputText v-model="contact.email" :disabled="isExistingParent" type="email" fluid /></label>
                    </div>
                  </div>
                  <Button v-if="!isExistingParent" label="Add Company Contact" icon="pi pi-plus" severity="secondary" outlined type="button" @click="addContact('company')" />
                </article>
              </div>
            </div>
          </section>

          <!-- 04 Tax -->
          <section id="sec-04" class="form-section collapsible-section" :class="{ collapsed: collapsedSections['sec-04'] }">
            <button type="button" class="section-toggle" @click="collapsedSections['sec-04'] = !collapsedSections['sec-04']">
              <span>04</span>
              <div><h2>Tax Information</h2><p>Optional; never sourced from Google Places.</p></div>
              <i :class="collapsedSections['sec-04'] ? 'pi pi-chevron-down' : 'pi pi-chevron-up'" />
            </button>
            <div v-show="!collapsedSections['sec-04']" class="section-body">
              <div class="dual-cards">
                <article class="form-card">
                  <div class="card-label"><strong>Site Tax</strong><Tag value="Optional" severity="secondary" /></div>
                  <label class="field"><span>PPN</span><Select v-model="form.ppn" :options="['PKP', 'Non-PKP']" show-clear fluid /></label>
                  <label class="field"><span>ID TKU Number</span><InputText v-model="form.idTkuNumber" fluid /></label>
                  <label class="field"><span>NIK</span><InputText v-model="form.nik" fluid /></label>
                </article>
                <article class="form-card">
                  <div class="card-label"><strong>Company Tax</strong><Tag :value="isExistingParent ? 'Synced' : 'Optional'" severity="secondary" /></div>
                  <label class="field"><span>NPWP Name</span><InputText v-model="form.companyNpwpName" :disabled="isExistingParent" fluid /></label>
                  <label class="field"><span>NPWP Address</span><Textarea v-model="form.companyNpwpAddress" :disabled="isExistingParent" rows="2" fluid /></label>
                  <label class="field"><span>NPWP Number</span><InputText v-model="form.companyNpwpNumber" :disabled="isExistingParent" fluid /></label>
                </article>
              </div>
            </div>
          </section>

          <!-- 05 Master Data -->
          <section id="sec-05" class="form-section collapsible-section" :class="{ collapsed: collapsedSections['sec-05'] }">
            <button type="button" class="section-toggle" @click="collapsedSections['sec-05'] = !collapsedSections['sec-05']">
              <span>05</span>
              <div><h2>Other Master Data</h2><p>Local options simulate ERP master selections.</p></div>
              <i :class="collapsedSections['sec-05'] ? 'pi pi-chevron-down' : 'pi pi-chevron-up'" />
            </button>
            <div v-show="!collapsedSections['sec-05']" class="section-body">
              <div class="dual-cards">
                <article class="form-card">
                  <div class="card-label"><strong>Customer Site</strong><Tag value="Optional" severity="secondary" /></div>
                  <label class="field"><span>Shipment Cost</span><Select v-model="form.shipmentCost" :options="data.options.shipmentCosts" show-clear fluid /></label>
                  <label class="field"><span>Invoice Type</span><Select v-model="form.invoiceType" :options="data.options.invoiceTypes" show-clear fluid /></label>
                  <label class="field"><span>Bank Account</span><InputText v-model="form.bankAccount" placeholder="Simulation reference" fluid /></label>
                </article>
                <article class="form-card">
                  <div class="card-label"><strong>Company</strong><Tag :value="isExistingParent ? 'Synced' : 'Optional'" severity="secondary" /></div>
                  <label class="field"><span>Term of Payment</span><Select v-model="form.termOfPayment" :options="data.options.termsOfPayment" :disabled="isExistingParent" show-clear fluid /></label>
                </article>
              </div>
            </div>
          </section>

          <!-- 06 Billing -->
          <section id="sec-06" class="form-section collapsible-section" :class="{ collapsed: collapsedSections['sec-06'] }">
            <button type="button" class="section-toggle" @click="collapsedSections['sec-06'] = !collapsedSections['sec-06']">
              <span>06</span>
              <div><h2>Billing & Shipment</h2><p>Optional Document Header preview.</p></div>
              <i :class="collapsedSections['sec-06'] ? 'pi pi-chevron-down' : 'pi pi-chevron-up'" />
            </button>
            <div v-show="!collapsedSections['sec-06']" class="section-body">
              <article class="form-card">
                <div class="document-header">
                  <div><span>Seller Identity</span><strong>{{ data.sellerIdentity }}</strong></div>
                  <div><span>Customer ID</span><strong>{{ customerCodePreview }}</strong></div>
                </div>
                <div class="two-fields">
                  <label class="field"><span>Bill To Source</span><Select v-model="form.billToSource" :options="documentSources" show-clear fluid /></label>
                  <label class="field"><span>Ship To Source</span><Select v-model="form.shipToSource" :options="documentSources" show-clear fluid /></label>
                </div>
                <div class="dual-preview">
                  <div><span>Billing address</span><p>{{ billPreview }}</p></div>
                  <div><span>Shipment address</span><p>{{ shipPreview }}</p></div>
                </div>
              </article>
            </div>
          </section>

          <!-- 07 Sales Assignment -->
          <section id="sec-07" class="form-section">
            <div class="section-heading">
              <span>07</span>
              <div><h2>Sales Assignment</h2><p>Select the primary Sales Executive for this customer.</p></div>
            </div>
            <div class="dual-cards">
              <article class="form-card">
                <div class="card-label"><strong>Sales Assignment</strong><Tag value="Required" severity="warn" /></div>
                <label class="field"><span>Sales Executive *</span><Select v-model="form.salesExecutiveId" :options="data.salesExecutives" option-label="fullName" option-value="id" :invalid="submitted && !form.salesExecutiveId" filter fluid /></label>
                <div v-for="(assignment, index) in form.salesAssignments" :key="index" class="repeat-card">
                  <div class="repeat-heading">
                    <span>Additional Sales {{ index + 1 }}</span>
                    <Button icon="pi pi-trash" severity="danger" text rounded type="button" @click="form.salesAssignments.splice(index, 1)" />
                  </div>
                  <label class="field"><span>Sales Executive</span><Select v-model="assignment.ownerId" :options="data.salesExecutives" option-label="fullName" option-value="id" @change="assignment.ownerName = salesName($event.value)" fluid /></label>
                  <div class="period-grid">
                    <Select v-model="assignment.startMonth" :options="months" option-label="label" option-value="value" />
                    <Select v-model="assignment.startYear" :options="years" />
                    <InputText v-model="assignment.end" placeholder="UNTIL_NOW or YYYY-MM" />
                  </div>
                </div>
                <Button label="Add Sales Assignment" icon="pi pi-plus" severity="secondary" outlined type="button" @click="addAssignment('sales')" />
              </article>
              <article class="form-card">
                <div class="card-label"><strong>Company KAM</strong><Tag value="Optional" severity="secondary" /></div>
                <div v-for="(assignment, index) in form.kamAssignments" :key="index" class="repeat-card">
                  <div class="repeat-heading">
                    <span>KAM {{ index + 1 }}</span>
                    <Button v-if="!isExistingParent" icon="pi pi-trash" severity="danger" text rounded type="button" @click="form.kamAssignments.splice(index, 1)" />
                  </div>
                  <label class="field"><span>KAM Executive</span><Select v-model="assignment.ownerId" :options="data.salesExecutives" option-label="fullName" option-value="id" :disabled="isExistingParent" @change="assignment.ownerName = salesName($event.value)" fluid /></label>
                  <div class="period-grid">
                    <Select v-model="assignment.startMonth" :options="months" option-label="label" option-value="value" :disabled="isExistingParent" />
                    <Select v-model="assignment.startYear" :options="years" :disabled="isExistingParent" />
                    <InputText v-model="assignment.end" :disabled="isExistingParent" placeholder="UNTIL_NOW or YYYY-MM" />
                  </div>
                </div>
                <Button v-if="!isExistingParent" label="Add KAM Assignment" icon="pi pi-plus" severity="secondary" outlined type="button" @click="addAssignment('kam')" />
              </article>
            </div>
          </section>

          <div class="conversion-actions">
            <Button label="Cancel" severity="secondary" outlined type="button" @click="router.push(`/admin/prospects/${route.params.id}/review`)" />
            <Button label="Convert to Customer" icon="pi pi-check" type="submit" :loading="saving" />
          </div>
        </div>

        <!-- Sidebar Status -->
        <aside class="scope-panel">
          <p class="eyebrow">Live status</p>
          <h3>{{ requiredSectionsDone ? 'Ready to convert' : `${progressPercent}% complete` }}</h3>
          <div v-for="s in sectionStatus" :key="s.id" class="scope-status-row">
            <i :class="s.done ? 'pi pi-check-circle scope-ok' : s.required ? 'pi pi-exclamation-circle scope-warn' : 'pi pi-circle scope-muted'" />
            <span :class="{ 'scope-done-text': s.done }">{{ s.label }}<template v-if="s.required"> *</template></span>
          </div>
          <div v-if="googleDataIssues.length" class="scope-warning" style="margin-top:0.8rem">
            <i class="pi pi-exclamation-triangle" /><p>Google snapshot is missing {{ googleDataIssues.length }} field(s).</p>
          </div>
        </aside>
      </form>
    </template>

    <!-- Confirm Dialog -->
    <Dialog v-model:visible="showConfirmDialog" header="Confirm Conversion" :modal="true" :closable="true" :style="{ width: 'min(92vw, 480px)' }">
      <div class="confirm-body">
        <p>You are about to convert this prospect into a Customer. This action is <strong>atomic and irreversible</strong>.</p>
        <div class="confirm-summary">
          <div><span>Customer Name</span><strong>{{ form.customerName }}</strong></div>
          <div><span>Parent Company</span><strong>{{ isExistingParent ? selectedParent?.name || '—' : form.parentCompanyName || '—' }}</strong></div>
          <div><span>Site Address</span><strong>{{ form.siteAddress.previewAddress || '—' }}</strong></div>
          <div><span>Sales Executive</span><strong>{{ data?.salesExecutives.find(e => e.id === form.salesExecutiveId)?.fullName || '—' }}</strong></div>
          <div><span>Category</span><strong>{{ form.customerCategory || '—' }}</strong></div>
          <div><span>Segment</span><strong>{{ form.customerSegment || '—' }}</strong></div>
        </div>
        <Message severity="info" :closable="false">A Parent Company and Customer Site record will be created in a single database transaction.</Message>
      </div>
      <template #footer>
        <Button label="Go Back" severity="secondary" outlined @click="showConfirmDialog = false" />
        <Button label="Convert Now" icon="pi pi-check" :loading="saving" @click="executeConvert" />
      </template>
    </Dialog>
  </section>
</template>

<style scoped>
/* Header */
.cv-header { display: flex; align-items: flex-start; gap: 0.75rem; margin-bottom: 0.25rem; }
.cv-header-text { flex: 1; min-width: 0; }

/* Layout */
.conversion-layout { display: grid; grid-template-columns: minmax(0, 1fr) 260px; gap: 1.5rem; align-items: start; }
.conversion-main { min-width: 0; display: grid; gap: 1.75rem; }

/* Sections */
.form-section {
  padding: clamp(1.25rem, 2.5vw, 1.75rem);
  border: 1px solid var(--border-light);
  border-radius: var(--radius-xl);
  background: var(--surface-card);
  box-shadow: var(--shadow-xs);
}

.section-heading {
  margin-bottom: 1.25rem; display: flex; gap: 0.75rem; align-items: center;
  padding-bottom: 0.85rem; border-bottom: 1px solid var(--border-light);
}
.section-heading > span {
  width: 2rem; height: 2rem; flex-shrink: 0;
  display: grid; place-items: center; border-radius: var(--radius-sm);
  color: #fff; background: var(--brand-blue);
  font-size: 0.65rem; font-weight: 800;
}
.section-heading h2 { margin: 0; font-size: 1rem; letter-spacing: -0.02em; font-weight: 700; }
.section-heading p { margin: 0.15rem 0 0; color: var(--text-muted); font-size: 0.75rem; }

/* Cards */
.dual-cards { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 1rem; align-items: start; }
.form-card {
  min-width: 0; padding: 1.25rem; display: grid; gap: 1.15rem;
  border: 1px solid var(--border-light); border-radius: var(--radius-lg);
  background: var(--surface-subtle);
}
.locked-card { background: #f1f5f9; }

/* Card labels */
.card-label, .repeat-heading {
  display: flex; gap: 0.6rem; align-items: center; justify-content: space-between;
  padding-bottom: 0.5rem; border-bottom: 1px solid var(--border-light);
}
.card-label strong { font-size: 0.8rem; }

/* Field grids */
.two-fields, .address-grid { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 1rem; }
.field span em { color: #899590; font-size: 0.68rem; font-style: normal; font-weight: 600; }
.form-card .field { gap: 0.55rem; }
.radio-stack { display: grid; gap: 0.5rem; }
.radio-stack label, .check-row { display: flex; gap: 0.5rem; align-items: center; color: #41504c; font-size: 0.82rem; font-weight: 600; }

/* Code preview / master preview */
.code-preview, .master-preview {
  padding: 0.75rem; display: grid; gap: 0.5rem;
  border: 1px dashed #c4d1cc; border-radius: var(--radius-md); background: #f2f7f5;
}
.code-preview small, .block-muted { color: #72807c; font-size: 0.7rem; line-height: 1.5; }
.master-preview span { color: var(--brand-green-light); font-size: 0.62rem; font-weight: 800; letter-spacing: 0.07em; text-transform: uppercase; }
.master-preview p { margin: 0; color: #6b7874; font-size: 0.78rem; }

/* Repeat cards */
.repeat-card {
  padding: 0.9rem; display: grid; gap: 0.7rem;
  border: 1px solid var(--border-default); border-radius: var(--radius-md);
  background: var(--surface-card);
}
.repeat-card + .repeat-card { margin-top: 0.5rem; }
.period-grid { display: grid; grid-template-columns: 1.2fr 0.8fr 1.3fr; gap: 0.6rem; }

/* Document header */
.document-header {
  padding: 0.85rem; display: grid; grid-template-columns: 1fr 1fr; gap: 0.85rem;
  border-radius: var(--radius-md); color: #e8f3f0; background: #0b443b;
}
.document-header strong { display: block; color: #fff; }

/* Dual preview */
.dual-preview { display: grid; grid-template-columns: 1fr 1fr; gap: 0.75rem; }
.dual-preview > div { padding: 0.75rem; border-radius: var(--radius-md); background: #f2f6f5; }
.dual-preview span { display: block; margin-bottom: 0.25rem; color: var(--text-muted); font-size: 0.65rem; font-weight: 700; letter-spacing: 0.06em; text-transform: uppercase; }
.dual-preview p { margin: 0; color: #41504c; font-size: 0.8rem; line-height: 1.45; }

/* Actions */
.conversion-actions { padding: 0.5rem 0 2rem; display: flex; justify-content: flex-end; gap: 0.7rem; }

/* Sidebar */
.scope-panel {
  position: sticky; top: 1rem;
  padding: 1.25rem; display: grid; gap: 0.75rem;
  border: 1px solid var(--border-light); border-radius: var(--radius-xl);
  background: #f7faf9; box-shadow: var(--shadow-sm);
}
.scope-panel h3 { margin: -0.2rem 0 0; font-size: 1.05rem; font-weight: 800; }
.scope-panel > div { padding-top: 0.7rem; border-top: 1px solid var(--border-light); }
.scope-panel p { margin: 0.3rem 0 0; color: #5b6965; font-size: 0.78rem; line-height: 1.5; }
.scope-warning { display: grid; grid-template-columns: auto 1fr; gap: 0.4rem; }
.scope-warning i { color: #f59e0b; }
.scope-warning p { margin: 0; }
.scope-status-row { display: flex; align-items: center; gap: 0.45rem; padding: 0.3rem 0; font-size: 0.75rem; }
.scope-status-row i { font-size: 0.6rem; flex-shrink: 0; }
.scope-ok { color: #22c55e; }
.scope-warn { color: #f59e0b; }
.scope-muted { color: #cbd5e1; }
.scope-done-text { color: var(--text-muted); text-decoration: line-through; }

/* Progress */
.progress-rail { margin-bottom: 0.5rem; padding: 0.85rem 1.15rem; border: 1px solid var(--border-light); border-radius: var(--radius-xl); background: var(--surface-card); box-shadow: var(--shadow-xs); }
.progress-track { height: 3px; margin-bottom: 0.5rem; border-radius: 10px; background: #edf1f5; overflow: hidden; }
.progress-fill { height: 100%; border-radius: 10px; background: linear-gradient(90deg, var(--brand-blue), #22c55e); transition: width 0.4s cubic-bezier(0.4, 0, 0.2, 1); }
.progress-labels { display: flex; gap: 0.1rem; flex-wrap: wrap; }
.progress-label {
  display: flex; align-items: center; gap: 0.25rem;
  padding: 0.25rem 0.5rem; border: 0; border-radius: var(--radius-sm);
  background: transparent; color: var(--text-muted);
  font-size: 0.6rem; font-weight: 600; cursor: pointer;
  transition: background var(--transition-fast), color var(--transition-fast);
}
.progress-label:hover { background: var(--surface-hover); }
.progress-label.done { color: #059669; }
.progress-label.done i { color: #22c55e; }
.progress-label:not(.done) i { color: #f59e0b; }

.field-invalid { outline: 2px solid #ef4444; outline-offset: 2px; border-radius: var(--radius-sm); }
:deep(.p-invalid) { border-color: #ef4444 !important; box-shadow: 0 0 0 2px rgba(239, 68, 68, 0.12) !important; }

/* Confirm Dialog */
.confirm-body p { margin: 0 0 0.75rem; font-size: 0.85rem; line-height: 1.6; }
.confirm-body strong { font-weight: 700; }
.confirm-summary { margin: 0.75rem 0; padding: 0.8rem; border: 1px solid var(--border-light); border-radius: var(--radius-md); background: var(--surface-subtle); display: grid; gap: 0.4rem; }
.confirm-summary div { display: flex; justify-content: space-between; align-items: baseline; gap: 1rem; padding: 0.25rem 0; border-bottom: 1px solid var(--border-light); }
.confirm-summary div:last-child { border-bottom: 0; }
.confirm-summary span { color: var(--text-muted); font-size: 0.7rem; font-weight: 600; flex-shrink: 0; }
.confirm-summary strong { font-size: 0.8rem; text-align: right; word-break: break-word; }

/* Snapshot */
.snapshot-card {
  padding: 1.25rem; margin-bottom: 0.5rem;
  border: 1px solid #e0e7ff; border-radius: var(--radius-xl);
  background: linear-gradient(135deg, #f8faff 0%, #f0f5ff 100%);
  box-shadow: var(--shadow-xs);
}
.snapshot-header { margin-bottom: 0.85rem; display: flex; align-items: flex-start; justify-content: space-between; }
.snapshot-header h2 { margin: 0.1rem 0 0; font-size: 1.15rem; letter-spacing: -0.03em; font-weight: 800; }
.snapshot-grid { display: grid; grid-template-columns: repeat(3, minmax(0, 1fr)); gap: 0.65rem; }
.snapshot-grid > div { padding: 0.6rem; border-radius: var(--radius-md); background: rgba(255, 255, 255, 0.7); }
.snapshot-grid .wide { grid-column: 1 / -1; }
.snapshot-grid span { display: block; margin-bottom: 0.15rem; color: var(--text-muted); font-size: 0.6rem; font-weight: 700; letter-spacing: 0.06em; text-transform: uppercase; }
.snapshot-grid strong { display: block; font-size: 0.78rem; word-break: break-word; }
.snapshot-grid a { color: var(--brand-blue); text-decoration: none; font-size: 0.76rem; }
.snapshot-grid a:hover { text-decoration: underline; }

/* Suggestions */
.company-suggestions { padding: 0.5rem; border: 1px solid var(--border-default); border-radius: var(--radius-md); background: var(--surface-card); }
.suggestion-label { margin: 0 0 0.35rem; display: flex; align-items: center; gap: 0.3rem; color: #b45309; font-size: 0.65rem; font-weight: 600; }
.suggestion-label i { font-size: 0.7rem; }
.suggestion-item {
  display: block; width: 100%; padding: 0.5rem 0.6rem; margin-top: 0.25rem;
  text-align: left; color: var(--text-primary);
  background: var(--surface-subtle); border: 1px solid var(--border-light);
  border-radius: var(--radius-sm); cursor: pointer;
  transition: background var(--transition-fast), border-color var(--transition-fast);
}
.suggestion-item:hover { background: var(--brand-blue-50); border-color: #bfdbfe; }
.suggestion-item strong { display: block; font-size: 0.75rem; }
.suggestion-item span { display: block; color: var(--text-muted); font-size: 0.6rem; margin-top: 0.1rem; }

/* Responsive */
/* Collapsible sections */
.collapsible-section { padding: 0; overflow: hidden; }
.section-toggle {
  display: flex; width: 100%; gap: 0.75rem; align-items: center; padding: clamp(1.25rem, 2.5vw, 1.75rem);
  border: 0; background: transparent; cursor: pointer; text-align: left;
  font: inherit; color: inherit;
}
.section-toggle > span {
  width: 2rem; height: 2rem; flex-shrink: 0;
  display: grid; place-items: center; border-radius: var(--radius-sm);
  color: #fff; background: var(--brand-blue);
  font-size: 0.65rem; font-weight: 800;
}
.section-toggle div { flex: 1; }
.section-toggle h2 { margin: 0; font-size: 1rem; letter-spacing: -0.02em; font-weight: 700; }
.section-toggle p { margin: 0.15rem 0 0; color: var(--text-muted); font-size: 0.75rem; }
.section-toggle i { font-size: 0.8rem; color: var(--text-muted); transition: transform var(--transition-fast); }
.section-body { padding: 0 clamp(1.25rem, 2.5vw, 1.75rem) clamp(1.25rem, 2.5vw, 1.75rem); border-top: 1px solid var(--border-light); }

@media (max-width: 900px) {
  .conversion-layout { grid-template-columns: 1fr; }
  .scope-panel { position: static; order: -1; }
  .dual-cards { grid-template-columns: 1fr; }
  .snapshot-grid { grid-template-columns: repeat(2, minmax(0, 1fr)); }
}
@media (max-width: 560px) {
  .two-fields, .address-grid, .dual-preview, .document-header { grid-template-columns: 1fr; }
  .period-grid { grid-template-columns: 1fr; }
}
</style>
