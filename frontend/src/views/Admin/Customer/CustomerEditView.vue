<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import Button from 'primevue/button'
import InputText from 'primevue/inputtext'
import Select from 'primevue/select'
import Textarea from 'primevue/textarea'
import Message from 'primevue/message'
import type { Contact, CustomerSite } from '../../../types/crm'
import { useCustomerListStore } from '../../../stores/customerList'
import { getAdminCustomer, getParentCompany, getSalesExecutives, updateAdminCustomer, updateParentCompany } from '../../../api/crm'
import { listCategories, listSegments, type MasterDataCategory, type MasterDataSegment } from '../../../api/masterData'

const route = useRoute()
const router = useRouter()
const store = useCustomerListStore()
const error = ref('')
const saved = ref(false)
const saving = ref(false)
const parentCompanyMode = ref<'manual' | 'match' | 'existing'>('manual')

const segments = ref<MasterDataSegment[]>([])
const categories = ref<MasterDataCategory[]>([])
const salesExecutives = ref<{ id: string; fullName: string }[]>([])
const optionsLoading = ref(false)
const optionsError = ref('')
const similarCustomerDismissed = ref(false)
const existingCustomerCode = ref('')
const segmentOptions = computed(() => segments.value.map((s) => ({ label: s.name, value: s.name })))
const categoryOptions = computed(() => {
  const segment = segments.value.find((item) => item.name === form.customerSegment)
  const available = segment ? categories.value.filter((item) => item.segmentId === segment.id) : categories.value
  return available.map((c) => ({ label: c.name, value: c.name }))
})
const defaultCityProvinceOptions = [
  'Surabaya (Kota), Jawa Timur',
  'Jakarta Pusat (Kota), DKI Jakarta',
  'Jakarta Selatan (Kota), DKI Jakarta',
  'Kediri (Kabupaten), Jawa Timur',
  'Sidoarjo (Kabupaten), Jawa Timur',
  'Bandung (Kota), Jawa Barat',
  'Semarang (Kota), Jawa Tengah',
  'Denpasar (Kota), Bali',
  'Medan (Kota), Sumatera Utara',
  'Makassar (Kota), Sulawesi Selatan',
  'Tangerang (Kota), Banten',
  'Bekasi (Kota), Jawa Barat',
  'Yogyakarta (Kota), DI Yogyakarta',
]
const cityProvinceOptions = computed(() => {
  const cityMap = new Map<string, Set<string>>()
  for (const city of defaultCityProvinceOptions) {
    cityMap.set(city, new Set<string>())
  }
  for (const customer of store.allCustomers) {
    const city = customer.address?.district?.trim()
    const province = customer.address?.province?.trim()
    if (!city || !province) continue
    const label = `${city}, ${province}`
    const detail = [customer.address?.subDistrict, customer.address?.village].filter(Boolean).join(', ') || customer.address?.previewAddress || customer.id
    if (!cityMap.has(label)) cityMap.set(label, new Set<string>())
    cityMap.get(label)?.add(detail)
  }
  return [...cityMap.entries()]
    .map(([label, details]) => ({
      label,
      postalCount: details.size,
      order: defaultCityProvinceOptions.indexOf(label),
    }))
    .sort((a, b) => {
      if (a.order !== -1 && b.order !== -1) return a.order - b.order
      if (a.order !== -1) return -1
      if (b.order !== -1) return 1
      return a.label.localeCompare(b.label)
    })
})
const cityDropdownOpen = ref<'site' | 'company' | ''>('')
const postalDropdownOpen = ref<'site' | 'company' | ''>('')
const bankAccountOpen = ref(false)
const invoiceTypeOpen = ref(false)
const invoiceTypeOptions = ['Invoice 1', 'Invoice 2', 'Invoice 3', 'Invoice 4']
const bankAccountOptions = ref<string[]>([])
const salesAssignmentCount = ref(1)
const monthOptions = ['January', 'February', 'March', 'April', 'May', 'June', 'July', 'August', 'September', 'October', 'November', 'December'].map((label) => ({ label, value: label }))
const yearOptions = Array.from({ length: 6 }, (_, index) => ({ label: String(new Date().getFullYear() - index), value: String(new Date().getFullYear() - index) }))
const billToSource = ref<'company' | 'site'>('company')
const shipToSource = ref<'company' | 'site' | 'delivery'>('site')
const siteCitySearch = ref('')
const companyCitySearch = ref('')
const sitePostalSearch = ref('')
const companyPostalSearch = ref('')
const filteredSiteCityOptions = computed(() => filterCityOptions(siteCitySearch.value))
const filteredCompanyCityOptions = computed(() => filterCityOptions(companyCitySearch.value))
const filteredSitePostalDetailOptions = computed(() => filterPostalDetailOptions('site', sitePostalSearch.value))
const filteredCompanyPostalDetailOptions = computed(() => filterPostalDetailOptions('company', companyPostalSearch.value))
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
  latitude: '',
  longitude: '',
  province: '',
  district: '',
  subDistrict: '',
  village: '',
  companyAddress: '',
  companyLatitude: '',
  companyLongitude: '',
  companyProvince: '',
  companyDistrict: '',
  companySubDistrict: '',
  companyVillage: '',
  pph: 'PPN 12%',
  idTkuNumber: '',
  nik: '',
  companyNpwpName: '',
  companyNpwpAddress: '',
  companyNpwpNumber: '',
  shipmentCost: '0',
  invoiceType: 'Invoice 2',
  bankAccount: '',
  termPayment: '30',
  startMonth: '',
  startYear: '',
  endPeriod: 'Until Now',
  deliveryName: '',
  deliveryAddress: '',
  notes: '',
  contacts: [blankContact()] as Contact[],
  companyContacts: [blankContact()] as Contact[],
})

const siteSourceAddress = computed(() => [form.address, form.province, form.district].filter(Boolean).join(' - '))
const companySourceAddress = computed(() => [form.companyAddress, form.companyProvince, form.companyDistrict].filter(Boolean).join(' - '))

const previewCompanyCode = computed(() => {
  const typedCode = form.parentCode.trim()
  if (typedCode) return typedCode
  return store.allCustomers[0]?.parentCode || 'PC-000024'
})

const previewCustomerSiteCode = computed(() => {
  const companyCode = previewCompanyCode.value
  const siteNumbers = store.allCustomers
    .map((customer) => customer.customerCode)
    .filter((code) => code.startsWith(`${companyCode}-S`))
    .map((code) => {
      const parts = code.split('-S')
      return Number(parts[parts.length - 1])
    })
    .filter((num) => Number.isFinite(num))

  const nextSiteNumber = siteNumbers.length ? Math.max(...siteNumbers) + 1 : 24
  return `${companyCode}-S${String(nextSiteNumber).padStart(3, '0')}`
})

const existingParentCompanies = computed(() => {
  const companies = new Map<string, { code: string; name: string }>()
  for (const customer of store.allCustomers) {
    if (customer.parentCode && customer.parentCompanyName && !companies.has(customer.parentCode)) {
      companies.set(customer.parentCode, { code: customer.parentCode, name: customer.parentCompanyName })
    }
  }
  return [...companies.values()]
})

function selectParentCompanyMode(mode: 'manual' | 'match' | 'existing') {
  parentCompanyMode.value = mode
  if (mode === 'match') {
    form.parentCompanyName = form.name
  }
  if (mode === 'existing') {
    const company = existingParentCompanies.value[0]
    if (company) {
      form.parentCompanyName = company.name
      form.parentCode = company.code
    }
  }
}

function filterCityOptions(keyword: string) {
  const normalizedKeyword = keyword.trim().toLowerCase()
  if (!normalizedKeyword) return cityProvinceOptions.value
  return cityProvinceOptions.value.filter((option) => option.label.toLowerCase().includes(normalizedKeyword))
}

function getPostalDetailOptions(scope: 'site' | 'company') {
  const selectedCity = scope === 'site' ? form.province : form.companyProvince
  const detailMap = new Map<string, string>()
  for (const customer of store.allCustomers) {
    const city = customer.address?.district?.trim()
    const province = customer.address?.province?.trim()
    if (!city || !province || `${city}, ${province}` !== selectedCity) continue
    const label = [customer.address?.village, customer.address?.subDistrict, customer.address?.district]
      .map((value) => value?.trim())
      .filter(Boolean)
      .join(', ')
    if (!label) continue
    detailMap.set(label, customer.address?.previewAddress || `${city}, ${province}`)

  }
  return [...detailMap.entries()]
    .map(([label, description]) => ({ label, description }))
    .sort((a, b) => a.label.localeCompare(b.label))
}

function filterPostalDetailOptions(scope: 'site' | 'company', keyword: string) {
  const options = getPostalDetailOptions(scope)
  const normalizedKeyword = keyword.trim().toLowerCase()
  if (!normalizedKeyword) return options
  return options.filter((option) => `${option.label} ${option.description}`.toLowerCase().includes(normalizedKeyword))
}

function selectCityProvince(scope: 'site' | 'company', value: string) {
  if (scope === 'site') {
    form.province = value
    form.district = ''
    sitePostalSearch.value = ''
  } else {
    form.companyProvince = value
    form.companyDistrict = ''
    companyPostalSearch.value = ''
  }
  cityDropdownOpen.value = ''
  postalDropdownOpen.value = ''
}

function clearCityProvince(scope: 'site' | 'company') {
  if (scope === 'site') {
    form.province = ''
    siteCitySearch.value = ''
    form.district = ''
    sitePostalSearch.value = ''
  } else {
    form.companyProvince = ''
    companyCitySearch.value = ''
    form.companyDistrict = ''
    companyPostalSearch.value = ''
  }
  cityDropdownOpen.value = ''
  postalDropdownOpen.value = ''
}

function selectPostalDetail(scope: 'site' | 'company', value: string) {
  if (scope === 'site') {
    form.district = value
  } else {
    form.companyDistrict = value
  }
  postalDropdownOpen.value = ''
}

function clearPostalDetail(scope: 'site' | 'company') {
  if (scope === 'site') {
    form.district = ''
    sitePostalSearch.value = ''
  } else {
    form.companyDistrict = ''
    companyPostalSearch.value = ''
  }
  postalDropdownOpen.value = ''
}

function setupBillingSourceInteractions() {
  const section = document.querySelector('.billing-shipment-section')
  if (!section) return
  const sellerLabel = [...section.querySelectorAll('p')].find((item) => item.textContent?.trim() === 'Seller Identity')
  const seller = sellerLabel?.nextElementSibling as HTMLElement | null
  if (seller && seller.textContent?.trim() === 'Company Name') seller.textContent = 'PT Yummy Food Utama'
  const sellerAddress = seller?.nextElementSibling as HTMLElement | null
  if (sellerAddress && sellerAddress.textContent?.trim() === '[Company Address]') sellerAddress.textContent = 'Jl. Raya Bogor Km. 22 No. 40, RT.5/RW.5, Rambutan, Ciracas, East Jakarta, Special Capital Region of Jakarta 13750, Indonesia'
  const sellerMeta = section.querySelectorAll('p')
  sellerMeta.forEach((item) => {
    if (item.textContent?.includes('Phone:')) item.innerHTML = 'Phone: <span class="font-medium text-[#334155]">(021) 87790947</span>'
    if (item.textContent?.includes('NPWP:')) item.innerHTML = 'NPWP: <span class="font-medium text-[#334155]">0013470778007000</span>'
    if (item.textContent?.includes('Website:')) item.innerHTML = 'Website: <span class="font-medium text-[#334155]">https://www.yummydairy.com/</span>'
  })
  section.addEventListener('click', (event) => {
    const target = event.target as HTMLElement
    const button = target.closest('button') as HTMLButtonElement | null
    if (!button || !button.parentElement?.classList.contains('flex')) return
    const group = button.parentElement
    const buttons = [...group.querySelectorAll('button')]
    if (buttons.length < 2 || buttons.length > 3) return
    if (buttons.length === 3 && button.textContent?.trim() === 'Other Delivery') shipToSource.value = 'delivery'
    if (buttons.length === 3 && button.textContent?.trim() === 'Site') shipToSource.value = 'site'
    if (buttons.length === 3 && button.textContent?.trim() === 'Company') shipToSource.value = 'company'
    buttons.forEach((item) => {
      item.classList.remove('bg-[#dc2626]', 'text-white', 'font-semibold', 'shadow-[0px_4px_12px_0px_rgba(220,38,38,0.22)]')
      item.classList.add('bg-white', 'text-[#64748b]')
    })
    button.classList.remove('bg-white', 'text-[#64748b]')
    button.classList.add('bg-[#dc2626]', 'text-white', 'font-semibold', 'shadow-[0px_4px_12px_0px_rgba(220,38,38,0.22)]')
    const card = group.parentElement?.parentElement
    const content = card?.lastElementChild
    const contentLines = content?.querySelectorAll('p')
    if (contentLines && contentLines.length >= 2) {
      const selected = button.textContent?.trim()
      if (buttons.length === 2) {
        contentLines[0].textContent = selected === 'Site' ? (form.name || '[Name]') : (form.parentCompanyName || '[Company Name]')

        contentLines[1].textContent = selected === 'Site' ? (siteSourceAddress.value || '[Street Address]') : (companySourceAddress.value || '[Street Address]')
      } else if (selected === 'Other Delivery') {
        contentLines[0].textContent = form.deliveryName || 'Other Delivery'
        contentLines[1].textContent = form.deliveryAddress || '[Street Address]'
      } else {
        contentLines[0].textContent = selected === 'Company' ? (form.parentCompanyName || '[Company Name]') : (form.name || '[Name]')
        contentLines[1].textContent = selected === 'Company' ? (companySourceAddress.value || '[Street Address]') : (siteSourceAddress.value || '[Street Address]')
      }
    }
  })
}

function syncSellerIdentity() {
  const section = document.querySelector('.billing-shipment-section')
  if (!section) return
  const sellerLabel = [...section.querySelectorAll('p')].find((item) => item.textContent?.trim() === 'Seller Identity')
  const seller = sellerLabel?.nextElementSibling as HTMLElement | null
  const sellerAddress = seller?.nextElementSibling as HTMLElement | null
  if (seller) seller.textContent = 'PT Yummy Food Utama'
  if (sellerAddress) sellerAddress.textContent = 'Jl. Raya Bogor Km. 22 No. 40, RT.5/RW.5, Rambutan, Ciracas, East Jakarta, Special Capital Region of Jakarta 13750, Indonesia'
  section.querySelectorAll('p').forEach((item) => {
    if (item.textContent?.includes('Phone:')) item.innerHTML = 'Phone: <span class="font-medium text-[#334155]">(021) 87790947</span>'
    if (item.textContent?.includes('NPWP:')) item.innerHTML = 'NPWP: <span class="font-medium text-[#334155]">0013470778007000</span>'
    if (item.textContent?.includes('Website:')) item.innerHTML = 'Website: <span class="font-medium text-[#334155]">https://www.yummydairy.com/</span>'
  })
}

function syncBillingSourceVisuals() {
  const section = document.querySelector('.billing-shipment-section')
  if (!section) return
  const groups = [...section.querySelectorAll('.mt\\[8px\\].flex')].filter((group) => group.querySelectorAll('button').length >= 2)
  groups.forEach((group, index) => {
    const buttons = [...group.querySelectorAll('button')] as HTMLButtonElement[]
    const selected = index === 0 ? (billToSource.value === 'company' ? 'Company' : 'Site') : (shipToSource.value === 'delivery' ? 'Other Delivery' : shipToSource.value === 'company' ? 'Company' : 'Site')
    buttons.forEach((button) => {
      const active = button.textContent?.trim() === selected
      button.classList.toggle('bg-[#dc2626]', active)
      button.classList.toggle('text-white', active)
      button.classList.toggle('font-semibold', active)
      button.classList.toggle('shadow-[0px_4px_12px_0px_rgba(220,38,38,0.22)]', active)
      button.classList.toggle('bg-white', !active)
      button.classList.toggle('text-[#475569]', !active)
    })
    const card = group.parentElement?.parentElement
    const lines = card?.lastElementChild?.querySelectorAll('p')
    if (!lines || lines.length < 2) return
    if (index === 0) {
      lines[0].textContent = billToSource.value === 'company' ? (form.parentCompanyName || '[Company Name]') : (form.name || '[Name]')
      lines[1].textContent = billToSource.value === 'company' ? (companySourceAddress.value || '[Street Address]') : (siteSourceAddress.value || '[Street Address]')
    } else if (shipToSource.value === 'delivery') {
      lines[0].textContent = form.deliveryName || 'Other Delivery'
      lines[1].textContent = form.deliveryAddress || '[Street Address]'
    } else {
      lines[0].textContent = shipToSource.value === 'company' ? (form.parentCompanyName || '[Company Name]') : (form.name || '[Name]')
      lines[1].textContent = shipToSource.value === 'company' ? (companySourceAddress.value || '[Street Address]') : (siteSourceAddress.value || '[Street Address]')
    }
  })
}

watch(() => [form.parentCompanyName, form.companyAddress], () => {
  window.setTimeout(syncSellerIdentity, 0)
})

watch([billToSource, shipToSource], () => {
  window.setTimeout(syncBillingSourceVisuals, 0)
})

function addContact() {
  form.contacts.push(blankContact())
}
function addCompanyContact() {
  form.companyContacts.push(blankContact())
}
function togglePpn() {
  form.pph = form.pph === 'PPN 12%' ? 'Non PPN' : 'PPN 12%'
}
function removeContact(index: number) {
  if (form.contacts.length > 1) form.contacts.splice(index, 1)
}
function removeCompanyContact(index: number) {
  if (form.companyContacts.length > 1) form.companyContacts.splice(index, 1)
}

const isFormValid = computed(() =>
  form.name.trim() !== '' &&
  form.customerSegment !== '' &&
  form.customerCategory !== ''
)

async function handleSubmit() {
  if (!isFormValid.value) return
  saving.value = true
  try {
    await new Promise((resolve) => setTimeout(resolve, 500))
    const id = String(route.params.id || '')
    const existing = store.allCustomers.find((customer) => customer.id === id || customer.customerCode === id)
    if (!existing) throw new Error('Customer not found')
    const updated: CustomerSite = {
      ...existing,
      name: form.name,
      segment: form.customerSegment,
      category: form.customerCategory,
      region: form.region || existing.region,
      parentCode: form.parentCode || existing.parentCode,
      parentCompanyName: form.parentCompanyName || existing.parentCompanyName,
      address: {
        ...existing.address,
        previewAddress: form.address || existing.address?.previewAddress || '',
        latitude: Number(form.latitude) || existing.address?.latitude || 0,
        longitude: Number(form.longitude) || existing.address?.longitude || 0,
        province: form.province || existing.address?.province || '',
        district: form.district || existing.address?.district || '',
        subDistrict: form.subDistrict || existing.address?.subDistrict || '',
        village: form.village || existing.address?.village || '',
      },
      contacts: JSON.parse(JSON.stringify(form.contacts)),
      salesExecutiveId: form.salesExecutiveId || existing.salesExecutiveId,
      updatedAt: new Date().toISOString(),
    }
    const response = await updateAdminCustomer(existing.id, {
      name: updated.name,
      segment: updated.segment,
      category: updated.category,
      region: updated.region,
      address: updated.address,
      contacts: updated.contacts,
      ppn: form.pph,
      idTkuNumber: form.idTkuNumber,
      nik: form.nik,
      shipmentCost: form.shipmentCost,
      invoiceType: form.invoiceType,
      bankAccount: form.bankAccount,
      billToSource: billToSource.value,
      shipToSource: shipToSource.value,
      billingAddressPreview: billToSource.value === 'company' ? form.companyAddress : form.address,
      shippingAddressPreview: shipToSource.value === 'company' ? form.companyAddress : shipToSource.value === 'delivery' ? form.deliveryAddress : form.address,
      salesExecutiveId: updated.salesExecutiveId,
    })
    const parentCode = String(existing.parentCode || form.parentCode || '')
    if (parentCode) {
      await updateParentCompany(parentCode, {
        name: form.parentCompanyName || existing.parentCompanyName,
        termOfPayment: form.termPayment,
        npwpName: form.companyNpwpName,
        npwpAddress: form.companyNpwpAddress,
        npwpNumber: form.companyNpwpNumber,
        companyAddress: {
          mode: 'manual',
          province: form.companyProvince,
          district: form.companyDistrict,
          subDistrict: form.companySubDistrict,
          village: form.companyVillage,
          latitude: Number(form.companyLatitude) || 0,
          longitude: Number(form.companyLongitude) || 0,
          previewAddress: form.companyAddress,
        },
        companyContacts: JSON.parse(JSON.stringify(form.companyContacts)),
      })
    }
    const persisted = response?.customer || updated
    store.allCustomers = store.allCustomers.map((customer) => customer.id === existing.id ? persisted : customer)
    store.items = store.items.map((customer) => customer.id === existing.id ? persisted : customer)
    const edits = JSON.parse(localStorage.getItem('crm_customer_edits') || '{}') as Record<string, Partial<CustomerSite>>
    edits[existing.id] = persisted
    localStorage.setItem('crm_customer_edits', JSON.stringify(edits))
    saved.value = true
  } catch (cause) {
    console.error('Failed to save customer', cause)
    error.value = 'Failed to save customer. Please try again.'
  } finally {
    saving.value = false
  }
}
const addSalesAssignment = () => {
  salesAssignmentCount.value += 1
}
const removeSalesAssignment = () => {
  if (salesAssignmentCount.value > 1) salesAssignmentCount.value -= 1
}


async function loadExistingCustomer() {
  const id = String(route.params.id || '')
  let customer: CustomerSite | undefined
  let parentCompany: Awaited<ReturnType<typeof getParentCompany>> | undefined
  try {
    customer = (await getAdminCustomer(id))?.customer
  } catch {
    customer = store.allCustomers.find((item) => item.id === id || item.customerCode === id)
  }
  if (!customer) {
    error.value = 'Customer not found.'
    return
  }
  // Keep the edit form consistent with the latest saved customer state while
  // the list/detail views are still hydrating from the local edit cache.
  const edits = JSON.parse(localStorage.getItem('crm_customer_edits') || '{}') as Record<string, Partial<CustomerSite>>
  const cachedEdit = edits[customer.id] || edits[customer.customerCode]
  if (cachedEdit) customer = { ...customer, ...cachedEdit }
  if (customer.parentCompanyId || customer.parentCode) {
    try {
      parentCompany = await getParentCompany(customer.parentCompanyId || customer.parentCode)
    } catch {
      parentCompany = undefined
    }
  }
  form.name = customer.name || ''
  existingCustomerCode.value = customer.customerCode || ''
  const segment = String(customer.segment || '').trim().toUpperCase()
  form.customerSegment = segment === 'B2C' ? 'B2C' : segment === 'B2B' ? 'B2B' : String(customer.segment || '')
  form.customerCategory = customer.category || ''
  form.region = customer.region || ''
  form.salesExecutiveId = customer.salesExecutiveId || ''
  form.parentCompanyName = customer.parentCompanyName || ''
  form.parentCode = customer.parentCode || ''
  form.address = customer.address?.previewAddress || ''
  form.latitude = customer.address?.latitude ? String(customer.address.latitude) : ''
  form.longitude = customer.address?.longitude ? String(customer.address.longitude) : ''
  form.province = customer.address?.province || ''
  form.district = customer.address?.district || ''
  form.subDistrict = customer.address?.subDistrict || ''
  form.village = customer.address?.village || ''
  form.companyAddress = parentCompany?.address?.previewAddress || customer.address?.previewAddress || ''
  form.companyLatitude = parentCompany?.address?.latitude ? String(parentCompany.address.latitude) : ''
  form.companyLongitude = parentCompany?.address?.longitude ? String(parentCompany.address.longitude) : ''
  form.companyProvince = parentCompany?.address?.province || ''
  form.companyDistrict = parentCompany?.address?.district || ''
  form.companySubDistrict = parentCompany?.address?.subDistrict || ''
  form.companyVillage = parentCompany?.address?.village || ''
  form.pph = customer.ppn || 'PPN 12%'
  form.idTkuNumber = customer.idTkuNumber || ''
  form.nik = customer.nik || ''
  form.shipmentCost = customer.shipmentCost || '0'
  form.invoiceType = customer.invoiceType || 'Invoice 2'
  form.bankAccount = customer.bankAccount || ''
  billToSource.value = customer.billToSource === 'site' ? 'site' : 'company'
  shipToSource.value = customer.shipToSource === 'company' ? 'company' : customer.shipToSource === 'delivery' ? 'delivery' : 'site'
  form.termPayment = customer.termOfPayment || '30'
  form.companyNpwpName = parentCompany?.npwpName || customer.parentCompanyName || ''
  form.companyNpwpAddress = parentCompany?.npwpAddress || form.companyAddress
  form.companyNpwpNumber = parentCompany?.npwpNumber || ''
  form.termPayment = parentCompany?.termOfPayment || customer.termOfPayment || '30'
  form.companyContacts = parentCompany?.contacts?.length ? JSON.parse(JSON.stringify(parentCompany.contacts)) : [blankContact()]
  form.contacts = customer.contacts?.length ? JSON.parse(JSON.stringify(customer.contacts)) : [blankContact()]
}

onMounted(async () => {
  setupBillingSourceInteractions()
  optionsLoading.value = true
  try {
    const [segmentData, categoryData, salesData] = await Promise.all([
      listSegments({ status: 'ACTIVE' }),
      listCategories({ status: 'ACTIVE' }),
      getSalesExecutives(),
      store.fetchCustomers(),
      store.fetchFilterOptions(),
    ])
    segments.value = segmentData
    categories.value = categoryData
    salesExecutives.value = salesData
    await loadExistingCustomer()
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

watch(() => form.name, (name) => {

  if (parentCompanyMode.value === 'match') {
    form.parentCompanyName = name
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
        <h2>Customer Updated Successfully</h2>
        <p class="muted">The customer site <strong>{{ form.name }}</strong> has been added to the system.</p>
        <div class="success-actions">
          <Button label="View Customer List" icon="pi pi-list" @click="router.push('/admin/customers')" />
          <Button label="Add Another" icon="pi pi-plus" severity="secondary" outlined @click="saved = false; form.name = ''; form.customerSegment = ''; form.customerCategory = ''; form.region = ''; form.salesExecutiveId = ''; form.parentCompanyName = ''; form.parentCode = ''; form.address = ''; form.latitude = ''; form.longitude = ''; form.province = ''; form.district = ''; form.subDistrict = ''; form.village = ''; form.companyAddress = ''; form.companyLatitude = ''; form.companyLongitude = ''; form.companyProvince = ''; form.companyDistrict = ''; form.companySubDistrict = ''; form.companyVillage = ''; form.notes = ''; form.contacts = [blankContact()]; form.companyContacts = [blankContact()]" />
        </div>
      </div>
    </template>

    <!-- FORM -->
    <template v-else>
      <!-- PAGE HEADER -->
      <header class="create-customer-header sticky top-0 z-40 border-b border-[#e2e8f0] bg-white/95 px-[16px] py-[9px] shadow-[0px_2px_8px_0px_rgba(15,23,42,0.04)] backdrop-blur sm:px-[24px]">
        <div class="create-customer-header-inner mx-auto flex max-w-[1180px] flex-wrap items-center justify-between gap-[12px]">
          <div class="min-w-0 flex items-center gap-[12px]">
            <button
              type="button"
              class="create-back-button flex shrink-0 items-center gap-[6px] font-['Inter'] text-[12px] font-semibold text-[#64748b] hover:text-[#dc2626]"
              @click="router.push('/admin/customers')"
            >
              <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="lucide lucide-arrow-left size-[14px]"><path d="m12 19-7-7 7-7"></path><path d="M19 12H5"></path></svg>Back to Customer Site
            </button>
            <div class="hidden h-[30px] w-px bg-[#e2e8f0] sm:block"></div>
            <div class="min-w-0">
              <h1 class="truncate font-['Inter'] text-[17px] font-bold leading-[22px] text-[#0f172a]">Edit Customer</h1>
              <p class="truncate font-['Inter'] text-[11px] leading-[15px] text-[#94a3b8]">Customer Management &gt; Customer Site &gt; Edit</p>
            </div>
          </div>
          <div class="flex items-center gap-[8px]">
            <button
              type="button"
              class="create-cancel-button inline-flex h-[40px] items-center justify-center gap-[7px] rounded-[10px] border border-[#e2e8f0] bg-white px-[14px] font-['Inter'] text-[12px] font-semibold text-[#475569] transition-all hover:bg-[#f8fafc]"
              @click="router.push('/admin/customers')"
            >
              Cancel
            </button>
            <button
              type="button"
              class="create-submit-button inline-flex h-[40px] items-center justify-center gap-[7px] rounded-[10px] bg-gradient-to-br from-[#dc2626] to-[#b91c1c] px-[16px] font-['Inter'] text-[12px] font-semibold text-white shadow-[0px_4px_16px_0px_rgba(220,38,38,0.22)] transition-all hover:brightness-110 disabled:cursor-not-allowed disabled:from-[#94a3b8] disabled:to-[#94a3b8] disabled:shadow-none disabled:hover:brightness-100"
              :disabled="!isFormValid || saving"
              @click="handleSubmit"
            >
              <i v-if="saving" class="pi pi-spin pi-spinner text-[12px] shrink-0"></i>
              <svg v-else xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="lucide lucide-send size-[14px]"><path d="M14.536 21.686a.5.5 0 0 0 .937-.024l6.5-19a.496.496 0 0 0-.635-.635l-19 6.5a.5.5 0 0 0-.024.937l7.93 3.18a2 2 0 0 1 1.112 1.11z"></path><path d="m21.854 2.147-10.94 10.939"></path></svg>Save Changes
            </button>
          </div>
        </div>
      </header>

      <Message v-if="error" severity="error">{{ error }}</Message>

      <div class="form-layout">
        <!-- LEFT COLUMN: FORM -->
        <div class="form-stack">
          <div class="create-form-heading mb-[18px]">
            <h2 class="font-['Inter'] text-[24px] font-bold tracking-tight text-[#0f172a]">Edit Customer</h2>
            <p class="mt-[5px] max-w-[760px] font-['Inter'] text-[13px] leading-[20px] text-[#64748b]">Create a customer location quickly, then link it to an existing parent company or create a new parent.</p>
          </div>
          <section class="overflow-visible rounded-[20px] border border-[#e2e8f0] bg-white/10 shadow-[0px_14px_34px_0px_rgba(15,23,42,0.06)]">
            <div class="border-b border-[#eef2f7] px-[20px] py-[18px]">
              <div class="flex flex-wrap items-center gap-[8px]">
                <h2 class="font-['Inter'] text-[16px] font-bold text-[#0f172a]">Customer Information</h2>
                <span class="rounded-full border border-[#ddd6fe] bg-[#f5f3ff] px-[8px] py-[3px] font-['Inter'] text-[11px] font-semibold text-[#7c3aed]">Site + Company</span>
              </div>
              <p class="mt-[4px] font-['Inter'] text-[12px] leading-[18px] text-[#64748b]">Site identity is on the left, while parent company ownership is on the right.</p>
            </div>
            <div class="grid grid-cols-1 gap-[14px] px-[20px] py-[20px] md:grid-cols-2">
              <div class="overflow-visible rounded-[18px] border border-[#dbe3ef] bg-white shadow-[0px_10px_24px_0px_rgba(15,23,42,0.04)]">
                <div class="border-b border-[#eef2f7] bg-[#f8fafc] px-[16px] py-[14px]">
                  <div class="flex flex-wrap items-center gap-[8px]">
                    <p class="font-['Inter'] text-[13px] font-bold text-[#0f172a]">Site Identity</p>
                    <span class="shrink-0 rounded-full border border-[#bbf7d0] bg-[#f0fdf4] px-[8px] py-[3px] font-['Inter'] text-[10px] font-bold tracking-[0.02em] text-[#166534]">Site</span>
                  </div>
                  <p class="mt-[3px] font-['Inter'] text-[11px] leading-[16px] text-[#64748b]">Operational identity for the outlet, branch, or store.</p>
                </div>
                <div class="space-y-[16px] p-[16px]">
                  <div class="relative">
                    <div class="flex items-start gap-[6px] font-['Inter'] text-[12px] font-semibold leading-[16px] text-[#334155]">
                      <span class="min-w-0 flex-1">Customer Name / Outlet / Branch / Store<span class="ml-[4px] text-[#dc2626]">*</span></span>
                      <span class="ml-[8px] shrink-0 rounded-full border border-[#e2e8f0] bg-[#f8fafc] px-[8px] py-[3px] text-[10px] font-bold text-[#64748b]">Manual</span>
                    </div>
                    <div class="mt-[8px]">
                      <div class="relative">
                        <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="lucide lucide-search pointer-events-none absolute left-[14px] top-[14px] size-[16px] text-[#94a3b8]"><circle cx="11" cy="11" r="8"></circle><path d="m21 21-4.3-4.3"></path></svg>
                        <InputText v-model="form.name" placeholder="Search place name or type customer name" class="customer-info-input customer-info-input-search h-[56px] w-full truncate rounded-[14px] border border-[#d9e2ec] bg-white px-[14px] py-[12px] pl-[42px] pr-[48px] font-['Inter'] text-[13px] font-medium text-[#0f172a] shadow-[0px_1px_2px_0px_rgba(15,23,42,0.04)] outline-none transition-all placeholder:text-[#94a3b8] focus:border-[#94a3b8] focus:ring-2 focus:ring-[#dbeafe]" />
                      </div>

                    </div>
                    <p class="mt-[8px] font-['Inter'] text-[11px] text-[#64748b]">Manual entry is available. Add a Google Maps API key to enable autocomplete.</p>
                  </div>
                  <label class="flex flex-col gap-[8px]">
                    <span class="flex items-start gap-[6px] font-['Inter'] text-[12px] font-semibold leading-[16px] text-[#334155]"><span class="min-w-0 flex-1">Customer Segment</span><span class="mt-[1px] shrink-0 whitespace-nowrap text-[#dc2626]">*</span></span>
                    <Select v-model="form.customerSegment" :options="segmentOptions" optionLabel="label" optionValue="value" placeholder="Select segment" filter filterPlaceholder="Search..." class="customer-info-select">
                      <template #option="{ option }">
                        <span class="customer-info-option-check" :class="{ active: form.customerSegment === option.value }">
                          <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M20 6 9 17l-5-5"></path></svg>
                        </span>
                        <span class="min-w-0 flex-1"><span class="block text-[13px] font-semibold text-[#334155]">{{ option.label }}</span></span>
                      </template>
                    </Select>
                  </label>
                  <label class="flex flex-col gap-[8px]">
                    <span class="flex items-start gap-[6px] font-['Inter'] text-[12px] font-semibold leading-[16px] text-[#334155]"><span class="min-w-0 flex-1">Customer Category</span><span class="mt-[1px] shrink-0 whitespace-nowrap text-[#dc2626]">*</span></span>
                    <Select v-model="form.customerCategory" :options="categoryOptions" optionLabel="label" optionValue="value" placeholder="Select category" filter filterPlaceholder="Search..." class="customer-info-select">
                      <template #option="{ option }">
                        <span class="customer-info-option-check" :class="{ active: form.customerCategory === option.value }">
                          <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M20 6 9 17l-5-5"></path></svg>
                        </span>
                        <span class="min-w-0 flex-1"><span class="block text-[13px] font-semibold text-[#334155]">{{ option.label }}</span></span>
                      </template>
                    </Select>
                  </label>
                  <div class="rounded-[18px] border border-[#e2e8f0] bg-[#f8fafc] p-[14px]">
                    <p class="font-['Inter'] text-[12px] font-bold text-[#0f172a]">Code Preview</p>
                    <div class="mt-[10px] grid grid-cols-1 gap-[8px] sm:grid-cols-2">
                      <div class="rounded-[14px] bg-white px-[12px] py-[10px]">
                        <p class="font-['Inter'] text-[10px] font-bold uppercase tracking-[0.08em] text-[#94a3b8]">Parent Code</p>
                        <p class="mt-[4px] font-['Inter'] text-[13px] font-bold text-[#0f172a]">{{ previewCompanyCode }}</p>
                      </div>
                      <div class="rounded-[14px] bg-white px-[12px] py-[10px]">
                        <p class="font-['Inter'] text-[10px] font-bold uppercase tracking-[0.08em] text-[#94a3b8]">Customer Code</p>
                        <p class="mt-[4px] font-['Inter'] text-[13px] font-bold text-[#0f172a]">{{ previewCustomerSiteCode }}</p>
                      </div>
                    </div>
                    <p class="mt-[9px] font-['Inter'] text-[11px] text-[#64748b]">Customer codes will be generated automatically when you save.</p>
                  </div>
                </div>
              </div>

              <div class="overflow-hidden rounded-[18px] border border-[#dbe3ef] bg-white shadow-[0px_10px_24px_0px_rgba(15,23,42,0.04)]">
                <div class="border-b border-[#eef2f7] bg-[#f8fafc] px-[16px] py-[14px]">
                  <div class="flex flex-wrap items-center gap-[8px]">
                    <p class="font-['Inter'] text-[13px] font-bold text-[#0f172a]">Parent Company</p>
                    <span class="shrink-0 rounded-full border border-[#bfdbfe] bg-[#eff6ff] px-[8px] py-[3px] font-['Inter'] text-[10px] font-bold tracking-[0.02em] text-[#1d4ed8]">Company</span>
                  </div>
                  <p class="mt-[3px] font-['Inter'] text-[11px] leading-[16px] text-[#64748b]">Linked parent company data that can be typed manually, mirrored from the site name, or selected from an existing company.</p>
                </div>
                <div class="space-y-[16px] p-[16px]">
                  <div class="relative">
                    <div class="flex flex-col gap-[8px]">
                      <div class="flex items-start gap-[6px] font-['Inter'] text-[12px] font-semibold leading-[16px] text-[#334155]">
                        <span class="min-w-0 flex-1">Customer Company / Parent</span>
                        <span class="mt-[1px] shrink-0 whitespace-nowrap text-[#dc2626]">*</span>
                      </div>
                      <div class="relative">
                        <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="lucide lucide-search pointer-events-none absolute left-[14px] top-1/2 size-[16px] -translate-y-1/2 text-[#94a3b8]"><circle cx="11" cy="11" r="8"></circle><path d="m21 21-4.3-4.3"></path></svg>
                        <InputText v-model="form.parentCompanyName" :readonly="parentCompanyMode === 'match'" :placeholder="parentCompanyMode === 'existing' ? 'Existing parent company selected automatically' : 'Type new parent company name'" class="customer-info-input customer-info-input-search h-[48px] w-full rounded-[14px] border border-[#d9e2ec] bg-white px-[14px] pl-[42px] font-['Inter'] text-[13px] font-medium text-[#0f172a] shadow-[0px_1px_2px_0px_rgba(15,23,42,0.04)] outline-none transition-all placeholder:text-[#94a3b8] focus:border-[#94a3b8] focus:ring-2 focus:ring-[#dbeafe]" />
                      </div>
                      <span class="font-['Inter'] text-[11px] text-[#64748b]">Type a new parent company name, or pick an existing company to switch modes.</span>
                    </div>
                  </div>
                  <label class="flex flex-col gap-[8px]">
                    <span class="flex items-start gap-[6px] font-['Inter'] text-[12px] font-semibold leading-[16px] text-[#334155]"><span class="min-w-0 flex-1">Company Code</span><span class="mt-[1px] shrink-0 whitespace-nowrap text-[#94a3b8]">(Optional)</span></span>
                    <InputText v-model="form.parentCode" placeholder="Auto-generated if empty" class="customer-info-input h-[48px] w-full rounded-[14px] border border-[#d9e2ec] bg-white px-[14px] font-['Inter'] text-[13px] font-medium text-[#0f172a] shadow-[0px_1px_2px_0px_rgba(15,23,42,0.04)] outline-none transition-all placeholder:text-[#94a3b8] focus:border-[#94a3b8] focus:ring-2 focus:ring-[#dbeafe]" />
                  </label>
                  <div class="rounded-[18px] border border-[#e2e8f0] bg-[#f8fafc] p-[10px]">
                    <div class="space-y-[8px]">
                      <button type="button" :class="['flex min-h-[54px] w-full items-center gap-[12px] rounded-[14px] border px-[14px] py-[10px] text-left transition-all', parentCompanyMode === 'manual' ? 'border-[#fecaca] bg-white shadow-[0px_8px_18px_0px_rgba(220,38,38,0.08)]' : 'border-[#e2e8f0] bg-white hover:border-[#cbd5e1]']" :aria-pressed="parentCompanyMode === 'manual'" @click="selectParentCompanyMode('manual')">
                        <span class="flex size-[22px] shrink-0 items-center justify-center rounded-full font-['Inter'] text-[11px] font-bold" :class="parentCompanyMode === 'manual' ? 'bg-[#dc2626] text-white' : 'border border-[#cbd5e1] bg-white text-[#64748b]'">1</span>
                        <span class="min-w-0 flex-1"><span class="block font-['Inter'] text-[12px] font-bold" :class="parentCompanyMode === 'manual' ? 'text-[#0f172a]' : 'text-[#334155]'">Manual entry</span><span class="mt-[3px] block font-['Inter'] text-[11px] leading-[16px]" :class="parentCompanyMode === 'manual' ? 'text-[#475569]' : 'text-[#64748b]'">Type a new parent company name.</span></span>
                      </button>
                      <button type="button" :class="['flex min-h-[54px] w-full items-center gap-[12px] rounded-[14px] border px-[14px] py-[10px] text-left transition-all', parentCompanyMode === 'match' ? 'border-[#fecaca] bg-white shadow-[0px_8px_18px_0px_rgba(220,38,38,0.08)]' : 'border-[#e2e8f0] bg-white hover:border-[#cbd5e1]']" :aria-pressed="parentCompanyMode === 'match'" @click="selectParentCompanyMode('match')">
                        <span class="flex size-[22px] shrink-0 items-center justify-center rounded-full font-['Inter'] text-[11px] font-bold" :class="parentCompanyMode === 'match' ? 'bg-[#dc2626] text-white' : 'border border-[#cbd5e1] bg-white text-[#64748b]'">2</span>
                        <span class="min-w-0 flex-1"><span class="block font-['Inter'] text-[12px] font-bold" :class="parentCompanyMode === 'match' ? 'text-[#0f172a]' : 'text-[#334155]'">Company name matches customer name</span><span class="mt-[3px] block font-['Inter'] text-[11px] leading-[16px]" :class="parentCompanyMode === 'match' ? 'text-[#475569]' : 'text-[#64748b]'">Auto-fill and lock the company name to the customer name.</span></span>
                      </button>
                      <button type="button" :class="['flex min-h-[54px] w-full items-center gap-[12px] rounded-[14px] border px-[14px] py-[10px] text-left transition-all', parentCompanyMode === 'existing' ? 'border-[#fecaca] bg-white shadow-[0px_8px_18px_0px_rgba(220,38,38,0.08)]' : 'border-[#e2e8f0] bg-white hover:border-[#cbd5e1]']" :aria-pressed="parentCompanyMode === 'existing'" @click="selectParentCompanyMode('existing')">
                        <span class="flex size-[22px] shrink-0 items-center justify-center rounded-full font-['Inter'] text-[11px] font-bold" :class="parentCompanyMode === 'existing' ? 'bg-[#dc2626] text-white' : 'border border-[#cbd5e1] bg-white text-[#64748b]'">3</span>
                        <span class="min-w-0 flex-1"><span class="block font-['Inter'] text-[12px] font-bold" :class="parentCompanyMode === 'existing' ? 'text-[#0f172a]' : 'text-[#334155]'">Existing company</span><span class="mt-[3px] block font-['Inter'] text-[11px] leading-[16px]" :class="parentCompanyMode === 'existing' ? 'text-[#475569]' : 'text-[#64748b]'">Pick from the saved company list.</span></span>
                      </button>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </section>

          <!-- ADDRESS -->
          <section class="overflow-visible rounded-[20px] border border-[#e2e8f0] bg-white shadow-[0px_14px_34px_0px_rgba(15,23,42,0.06)]">
            <div class="border-b border-[#eef2f7] px-[20px] py-[18px]">
              <div class="flex flex-wrap items-center gap-[8px]">
                <h2 class="font-['Inter'] text-[16px] font-bold text-[#0f172a]">Address Information</h2>
                <span class="rounded-full border px-[8px] py-[3px] font-['Inter'] text-[11px] font-semibold border-[#ddd6fe] bg-[#f5f3ff] text-[#7c3aed]">Site + Company</span>
              </div>
              <p class="mt-[4px] text-[12px] text-[#64748b]">Site address supports transactions; company address is used for legal and tax records.</p>
            </div>
            <div class="grid grid-cols-1 gap-[14px] px-[20px] py-[20px] md:grid-cols-2">
              <div class="overflow-hidden rounded-[18px] border border-[#dbe3ef] bg-white shadow-[0px_10px_24px_0px_rgba(15,23,42,0.04)]">

                <div class="border-b border-[#eef2f7] bg-[#f8fafc] px-[16px] py-[14px]">
                  <div class="flex flex-wrap items-center gap-[8px]">
                    <p class="font-['Inter'] text-[13px] font-bold text-[#0f172a]">Site Address</p>
                    <span class="shrink-0 rounded-full border border-[#bbf7d0] bg-[#f0fdf4] px-[8px] py-[3px] font-['Inter'] text-[10px] font-bold tracking-[0.02em] text-[#166534]">Site</span>
                  </div>
                  <p class="mt-[3px] font-['Inter'] text-[11px] leading-[16px] text-[#64748b]">Address for customer transactions and site-level documents.</p>
                </div>
                <div class="space-y-[16px] p-[16px]">
                  <div class="relative">
                    <div class="flex items-start gap-[6px] font-['Inter'] text-[12px] font-semibold leading-[16px] text-[#334155]">
                      <span class="min-w-0 flex-1">Search by Gmaps / Manual<span class="ml-[4px] text-[#dc2626]">*</span></span>
                      <span class="ml-[8px] shrink-0 rounded-full border border-[#e2e8f0] bg-[#f8fafc] px-[8px] py-[3px] text-[10px] font-bold text-[#64748b]">Manual</span>
                    </div>
                    <div class="mt-[8px]">
                      <div class="relative">
                        <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="lucide lucide-search pointer-events-none absolute left-[14px] top-[14px] size-[16px] text-[#94a3b8]"><circle cx="11" cy="11" r="8"></circle><path d="m21 21-4.3-4.3"></path></svg>
                        <InputText v-model="form.address" placeholder="Search or type customer address" class="customer-info-input customer-info-input-search h-[56px] w-full truncate rounded-[14px] border border-[#d9e2ec] bg-white px-[14px] py-[12px] pl-[42px] pr-[48px] font-['Inter'] text-[13px] font-medium text-[#0f172a] shadow-[0px_1px_2px_0px_rgba(15,23,42,0.04)] outline-none transition-all placeholder:text-[#94a3b8] focus:border-[#94a3b8] focus:ring-2 focus:ring-[#dbeafe]" />
                      </div>
                    </div>
                    <p class="mt-[8px] font-['Inter'] text-[11px] leading-[16px] text-[#64748b]">Manual entry is available. Add a Google Maps API key to enable autocomplete.</p>
                  </div>
                  <div class="rounded-[16px] border border-[#eef2f7] bg-[#fbfdff] p-[12px]">
                    <p class="mb-[12px] font-['Inter'] text-[11px] font-bold uppercase tracking-[0.08em] text-[#64748b]">Location Detail</p>
                    <div class="grid grid-cols-1 gap-[12px]">
                      <label class="flex flex-col gap-[8px]">
                        <span class="flex items-start gap-[6px] font-['Inter'] text-[12px] font-semibold leading-[16px] text-[#334155]"><span class="min-w-0 flex-1">City, Province</span><span class="mt-[1px] shrink-0 whitespace-nowrap text-[#dc2626]">*</span></span>
                        <div class="relative">
                          <button type="button" class="city-province-trigger" :class="{ open: cityDropdownOpen === 'site' }" @click="cityDropdownOpen = cityDropdownOpen === 'site' ? '' : 'site'">
                            <span :class="['truncate font-medium', form.province ? 'text-[#0f172a]' : 'text-[#94a3b8]']">{{ form.province || 'Select city, province' }}</span>
                            <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" :class="['lucide lucide-chevron-down size-[16px] shrink-0 text-[#64748b] transition-transform', cityDropdownOpen === 'site' ? 'rotate-180' : '']"><path d="m6 9 6 6 6-6"></path></svg>
                          </button>
                          <div v-if="cityDropdownOpen === 'site'" class="city-province-dropdown">
                            <div class="border-b border-[#eef2f7] p-[10px]">
                              <div class="relative">
                                <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="lucide lucide-search pointer-events-none absolute left-[12px] top-1/2 size-[15px] -translate-y-1/2 text-[#94a3b8]"><circle cx="11" cy="11" r="8"></circle><path d="m21 21-4.3-4.3"></path></svg>
                                <input v-model="siteCitySearch" placeholder="Search city, province..." class="h-[44px] w-full rounded-[12px] border border-[#d9e2ec] bg-white pl-[36px] pr-[12px] font-['Inter'] text-[13px] text-[#0f172a] outline-none transition-all placeholder:text-[#94a3b8] focus:border-[#94a3b8] focus:ring-2 focus:ring-[#e2e8f0]" />
                              </div>
                            </div>
                            <div class="max-h-[240px] overflow-y-auto p-[6px]">
                              <button type="button" class="city-province-option" :class="{ active: !form.province }" @click="clearCityProvince('site')">
                                <span class="city-province-check" :class="{ active: !form.province }"><svg v-if="!form.province" xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="size-[12px] text-white"><path d="M20 6 9 17l-5-5"></path></svg></span>
                <span class="min-w-0 flex-1"><span class="block font-['Inter'] text-[13px] font-semibold">Select city, province</span><span class="mt-[2px] block font-['Inter'] text-[11px] text-[#64748b]">Clear selection</span></span>
                              </button>
                              <button v-for="option in filteredSiteCityOptions" :key="option.label" type="button" class="city-province-option" :class="{ active: form.province === option.label }" @click="selectCityProvince('site', option.label)">
                                <span class="city-province-check" :class="{ active: form.province === option.label }"><svg v-if="form.province === option.label" xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="size-[12px] text-white"><path d="M20 6 9 17l-5-5"></path></svg></span>
                <span class="min-w-0 flex-1"><span class="block font-['Inter'] text-[13px] font-semibold text-[#0f172a]">{{ option.label }}</span><span class="mt-[2px] block font-['Inter'] text-[11px] text-[#64748b]">{{ option.postalCount ? `${option.postalCount} postal code${option.postalCount === 1 ? '' : 's'} available` : 'Postal code from CRM data' }}</span></span>
                              </button>
                            </div>
                          </div>
                        </div>
                      </label>
                      <label class="flex flex-col gap-[8px]">
                        <span class="flex items-start gap-[6px] font-['Inter'] text-[12px] font-semibold leading-[16px] text-[#334155]"><span class="min-w-0 flex-1">Postal Code, District, Sub-District</span><span class="mt-[1px] shrink-0 whitespace-nowrap text-[#94a3b8]">(Optional)</span></span>
                        <div class="relative">
                          <button type="button" class="city-province-trigger" :class="{ open: postalDropdownOpen === 'site' }" @click="postalDropdownOpen = postalDropdownOpen === 'site' ? '' : 'site'">
                            <span :class="['truncate font-medium', form.district ? 'text-[#0f172a]' : 'text-[#94a3b8]']">{{ form.district || 'Select postal code, district, sub-district' }}</span>
                            <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" :class="['lucide lucide-chevron-down size-[16px] shrink-0 text-[#64748b] transition-transform', postalDropdownOpen === 'site' ? 'rotate-180' : '']"><path d="m6 9 6 6 6-6"></path></svg>
                          </button>
                          <div v-if="postalDropdownOpen === 'site'" class="city-province-dropdown">
                            <div class="border-b border-[#eef2f7] p-[10px]">
                              <div class="relative">
                                <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="lucide lucide-search pointer-events-none absolute left-[12px] top-1/2 size-[15px] -translate-y-1/2 text-[#94a3b8]"><circle cx="11" cy="11" r="8"></circle><path d="m21 21-4.3-4.3"></path></svg>
                                <input v-model="sitePostalSearch" placeholder="Search postal code, district..." class="h-[44px] w-full rounded-[12px] border border-[#d9e2ec] bg-white pl-[36px] pr-[12px] font-['Inter'] text-[13px] text-[#0f172a] outline-none transition-all placeholder:text-[#94a3b8] focus:border-[#94a3b8] focus:ring-2 focus:ring-[#e2e8f0]" />
                              </div>
                            </div>
                            <div class="max-h-[240px] overflow-y-auto p-[6px]">
                              <button type="button" class="city-province-option" :class="{ active: !form.district }" @click="clearPostalDetail('site')">
                                <span class="city-province-check" :class="{ active: !form.district }"><svg v-if="!form.district" xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="size-[12px] text-white"><path d="M20 6 9 17l-5-5"></path></svg></span>
                                <span class="min-w-0 flex-1"><span class="block font-['Inter'] text-[13px] font-semibold">Select postal code, district, sub-district</span><span class="mt-[2px] block font-['Inter'] text-[11px] text-[#64748b]">Clear selection</span></span>
                              </button>
                              <button v-for="option in filteredSitePostalDetailOptions" :key="option.label" type="button" class="city-province-option" :class="{ active: form.district === option.label }" @click="selectPostalDetail('site', option.label)">
                                <span class="city-province-check" :class="{ active: form.district === option.label }"><svg v-if="form.district === option.label" xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="size-[12px] text-white"><path d="M20 6 9 17l-5-5"></path></svg></span>
                                <span class="min-w-0 flex-1"><span class="block font-['Inter'] text-[13px] font-semibold text-[#0f172a]">{{ option.label }}</span><span class="mt-[2px] block truncate font-['Inter'] text-[11px] text-[#64748b]">{{ option.description }}</span></span>
                              </button>
                            </div>
                          </div>
                        </div>
                      </label>
                    </div>
                  </div>
                  <div class="rounded-[16px] border border-[#eef2f7] bg-[#fbfdff] p-[12px]"><p class="mb-[12px] font-['Inter'] text-[11px] font-bold uppercase tracking-[0.08em] text-[#64748b]">Coordinate</p><div class="grid grid-cols-1 gap-[12px] sm:grid-cols-2"><label class="flex flex-col gap-[8px]"><span class="flex items-start gap-[6px] font-['Inter'] text-[12px] font-semibold leading-[16px] text-[#334155]"><span class="min-w-0 flex-1">Latitude</span><span class="mt-[1px] shrink-0 whitespace-nowrap text-[#94a3b8]">(Optional)</span></span><InputText v-model="form.latitude" placeholder="Example: -6.195026" class="h-[48px] w-full rounded-[14px] border border-[#d9e2ec] bg-white px-[14px] font-['Inter'] text-[13px] font-medium text-[#0f172a] shadow-[0px_1px_2px_0px_rgba(15,23,42,0.04)] outline-none transition-all placeholder:text-[#94a3b8] focus:border-[#94a3b8] focus:ring-2 focus:ring-[#dbeafe]" /></label><label class="flex flex-col gap-[8px]"><span class="flex items-start gap-[6px] font-['Inter'] text-[12px] font-semibold leading-[16px] text-[#334155]"><span class="min-w-0 flex-1">Longitude</span><span class="mt-[1px] shrink-0 whitespace-nowrap text-[#94a3b8]">(Optional)</span></span><InputText v-model="form.longitude" placeholder="Example: 106.821810" class="h-[48px] w-full rounded-[14px] border border-[#d9e2ec] bg-white px-[14px] font-['Inter'] text-[13px] font-medium text-[#0f172a] shadow-[0px_1px_2px_0px_rgba(15,23,42,0.04)] outline-none transition-all placeholder:text-[#94a3b8] focus:border-[#94a3b8] focus:ring-2 focus:ring-[#dbeafe]" /></label></div></div><div class="rounded-[16px] border border-[#e2e8f0] bg-white p-[12px]"><div class="flex items-center justify-between gap-[10px]"><p class="font-['Inter'] text-[11px] font-bold uppercase tracking-[0.08em] text-[#64748b]">Preview</p><span class="rounded-full border px-[8px] py-[3px] text-[10px] font-bold border-[#e2e8f0] bg-white text-[#64748b]">Manual</span></div><div class="mt-[8px] rounded-[12px] bg-[#f8fafc] px-[12px] py-[10px]"><p class="font-['Inter'] text-[13px] font-semibold leading-[20px]" :class="form.address ? 'text-[#334155]' : 'text-[#94a3b8]'">{{ form.address || 'Customer address preview will appear here.' }}</p></div></div>
                </div>
              </div>

              <div class="overflow-visible rounded-[18px] border border-[#dbe3ef] bg-white shadow-[0px_10px_24px_0px_rgba(15,23,42,0.04)]">
                <div class="border-b border-[#eef2f7] bg-[#f8fafc] px-[16px] py-[14px]">
                  <div class="flex flex-wrap items-center gap-[8px]">
                    <p class="font-['Inter'] text-[13px] font-bold text-[#0f172a]">Company Address</p>
                    <span class="shrink-0 rounded-full border border-[#bfdbfe] bg-[#eff6ff] px-[8px] py-[3px] font-['Inter'] text-[10px] font-bold tracking-[0.02em] text-[#1d4ed8]">Company</span>
                  </div>
                  <p class="mt-[3px] font-['Inter'] text-[11px] leading-[16px] text-[#64748b]">Legal and billing address used for company and tax records.</p>
                </div>
                <div class="space-y-[16px] p-[16px]">
                  <div class="relative">
                    <div class="flex items-start gap-[6px] font-['Inter'] text-[12px] font-semibold leading-[16px] text-[#334155]">
                      <span class="min-w-0 flex-1">Search by Gmaps / Manual<span class="ml-[4px] text-[#dc2626]">*</span></span>
                      <span class="ml-[8px] shrink-0 rounded-full border border-[#e2e8f0] bg-[#f8fafc] px-[8px] py-[3px] text-[10px] font-bold text-[#64748b]">Manual</span>
                    </div>
                    <div class="mt-[8px]">
                      <div class="relative">

                        <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="lucide lucide-search pointer-events-none absolute left-[14px] top-[14px] size-[16px] text-[#94a3b8]"><circle cx="11" cy="11" r="8"></circle><path d="m21 21-4.3-4.3"></path></svg>
                        <InputText v-model="form.companyAddress" placeholder="Search or type company address" class="customer-info-input customer-info-input-search h-[56px] w-full truncate rounded-[14px] border border-[#d9e2ec] bg-white px-[14px] py-[12px] pl-[42px] pr-[48px] font-['Inter'] text-[13px] font-medium text-[#0f172a] shadow-[0px_1px_2px_0px_rgba(15,23,42,0.04)] outline-none transition-all placeholder:text-[#94a3b8] focus:border-[#94a3b8] focus:ring-2 focus:ring-[#dbeafe]" />
                      </div>
                    </div>
                    <p class="mt-[8px] font-['Inter'] text-[11px] leading-[16px] text-[#64748b]">Manual entry is available. Add a Google Maps API key to enable autocomplete.</p>
                  </div>
                  <div class="rounded-[16px] border border-[#eef2f7] bg-[#fbfdff] p-[12px]">
                    <p class="mb-[12px] font-['Inter'] text-[11px] font-bold uppercase tracking-[0.08em] text-[#64748b]">Location Detail</p>
                    <div class="grid grid-cols-1 gap-[12px]">
                      <label class="flex flex-col gap-[8px]">
                        <span class="flex items-start gap-[6px] font-['Inter'] text-[12px] font-semibold leading-[16px] text-[#334155]"><span class="min-w-0 flex-1">City, Province</span><span class="mt-[1px] shrink-0 whitespace-nowrap text-[#dc2626]">*</span></span>
                        <div class="relative">
                          <button type="button" class="city-province-trigger" :class="{ open: cityDropdownOpen === 'company' }" @click="cityDropdownOpen = cityDropdownOpen === 'company' ? '' : 'company'">
                            <span :class="['truncate font-medium', form.companyProvince ? 'text-[#0f172a]' : 'text-[#94a3b8]']">{{ form.companyProvince || 'Select city, province' }}</span>
                            <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" :class="['lucide lucide-chevron-down size-[16px] shrink-0 text-[#64748b] transition-transform', cityDropdownOpen === 'company' ? 'rotate-180' : '']"><path d="m6 9 6 6 6-6"></path></svg>
                          </button>
                          <div v-if="cityDropdownOpen === 'company'" class="city-province-dropdown">
                            <div class="border-b border-[#eef2f7] p-[10px]">
                              <div class="relative">
                                <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="lucide lucide-search pointer-events-none absolute left-[12px] top-1/2 size-[15px] -translate-y-1/2 text-[#94a3b8]"><circle cx="11" cy="11" r="8"></circle><path d="m21 21-4.3-4.3"></path></svg>
                                <input v-model="companyCitySearch" placeholder="Search city, province..." class="h-[44px] w-full rounded-[12px] border border-[#d9e2ec] bg-white pl-[36px] pr-[12px] font-['Inter'] text-[13px] text-[#0f172a] outline-none transition-all placeholder:text-[#94a3b8] focus:border-[#94a3b8] focus:ring-2 focus:ring-[#e2e8f0]" />
                              </div>
                            </div>
                            <div class="max-h-[240px] overflow-y-auto p-[6px]">
                              <button type="button" class="city-province-option" :class="{ active: !form.companyProvince }" @click="clearCityProvince('company')">
                                <span class="city-province-check" :class="{ active: !form.companyProvince }"><svg v-if="!form.companyProvince" xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="size-[12px] text-white"><path d="M20 6 9 17l-5-5"></path></svg></span>
                                <span class="min-w-0 flex-1"><span class="block font-['Inter'] text-[13px] font-semibold">Select city, province</span><span class="mt-[2px] block font-['Inter'] text-[11px] text-[#64748b]">Clear selection</span></span>
                              </button>
                              <button v-for="option in filteredCompanyCityOptions" :key="option.label" type="button" class="city-province-option" :class="{ active: form.companyProvince === option.label }" @click="selectCityProvince('company', option.label)">
                                <span class="city-province-check" :class="{ active: form.companyProvince === option.label }"><svg v-if="form.companyProvince === option.label" xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="size-[12px] text-white"><path d="M20 6 9 17l-5-5"></path></svg></span>
                                <span class="min-w-0 flex-1"><span class="block font-['Inter'] text-[13px] font-semibold text-[#0f172a]">{{ option.label }}</span><span class="mt-[2px] block font-['Inter'] text-[11px] text-[#64748b]">{{ option.postalCount ? `${option.postalCount} postal code${option.postalCount === 1 ? '' : 's'} available` : 'Postal code from CRM data' }}</span></span>
                              </button>
                            </div>
                          </div>
                        </div>
                      </label>
                      <label class="flex flex-col gap-[8px]">
                        <span class="flex items-start gap-[6px] font-['Inter'] text-[12px] font-semibold leading-[16px] text-[#334155]"><span class="min-w-0 flex-1">Postal Code, District, Sub-District</span><span class="mt-[1px] shrink-0 whitespace-nowrap text-[#94a3b8]">(Optional)</span></span>
                        <div class="relative">
                          <button type="button" class="city-province-trigger" :class="{ open: postalDropdownOpen === 'company' }" @click="postalDropdownOpen = postalDropdownOpen === 'company' ? '' : 'company'">
                            <span :class="['truncate font-medium', form.companyDistrict ? 'text-[#0f172a]' : 'text-[#94a3b8]']">{{ form.companyDistrict || 'Select postal code, district, sub-district' }}</span>
                            <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" :class="['lucide lucide-chevron-down size-[16px] shrink-0 text-[#64748b] transition-transform', postalDropdownOpen === 'company' ? 'rotate-180' : '']"><path d="m6 9 6 6 6-6"></path></svg>
                          </button>
                          <div v-if="postalDropdownOpen === 'company'" class="city-province-dropdown">
                            <div class="border-b border-[#eef2f7] p-[10px]">
                              <div class="relative">
                                <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="lucide lucide-search pointer-events-none absolute left-[12px] top-1/2 size-[15px] -translate-y-1/2 text-[#94a3b8]"><circle cx="11" cy="11" r="8"></circle><path d="m21 21-4.3-4.3"></path></svg>
                                <input v-model="companyPostalSearch" placeholder="Search postal code, district..." class="h-[44px] w-full rounded-[12px] border border-[#d9e2ec] bg-white pl-[36px] pr-[12px] font-['Inter'] text-[13px] text-[#0f172a] outline-none transition-all placeholder:text-[#94a3b8] focus:border-[#94a3b8] focus:ring-2 focus:ring-[#e2e8f0]" />
                              </div>
                            </div>
                            <div class="max-h-[240px] overflow-y-auto p-[6px]">
                              <button type="button" class="city-province-option" :class="{ active: !form.companyDistrict }" @click="clearPostalDetail('company')">
                                <span class="city-province-check" :class="{ active: !form.companyDistrict }"><svg v-if="!form.companyDistrict" xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="size-[12px] text-white"><path d="M20 6 9 17l-5-5"></path></svg></span>
                                <span class="min-w-0 flex-1"><span class="block font-['Inter'] text-[13px] font-semibold">Select postal code, district, sub-district</span><span class="mt-[2px] block font-['Inter'] text-[11px] text-[#64748b]">Clear selection</span></span>
                              </button>
                              <button v-for="option in filteredCompanyPostalDetailOptions" :key="option.label" type="button" class="city-province-option" :class="{ active: form.companyDistrict === option.label }" @click="selectPostalDetail('company', option.label)">
                                <span class="city-province-check" :class="{ active: form.companyDistrict === option.label }"><svg v-if="form.companyDistrict === option.label" xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="size-[12px] text-white"><path d="M20 6 9 17l-5-5"></path></svg></span>
                                <span class="min-w-0 flex-1"><span class="block font-['Inter'] text-[13px] font-semibold text-[#0f172a]">{{ option.label }}</span><span class="mt-[2px] block truncate font-['Inter'] text-[11px] text-[#64748b]">{{ option.description }}</span></span>
                              </button>
                            </div>
                          </div>
                        </div>
                      </label>
                    </div>
                  </div>
                  <div class="rounded-[16px] border border-[#eef2f7] bg-[#fbfdff] p-[12px]"><p class="mb-[12px] font-['Inter'] text-[11px] font-bold uppercase tracking-[0.08em] text-[#64748b]">Coordinate</p><div class="grid grid-cols-1 gap-[12px] sm:grid-cols-2"><label class="flex flex-col gap-[8px]"><span class="flex items-start gap-[6px] font-['Inter'] text-[12px] font-semibold leading-[16px] text-[#334155]"><span class="min-w-0 flex-1">Latitude</span><span class="mt-[1px] shrink-0 whitespace-nowrap text-[#94a3b8]">(Optional)</span></span><InputText v-model="form.companyLatitude" placeholder="Example: -6.195026" class="h-[48px] w-full rounded-[14px] border border-[#d9e2ec] bg-white px-[14px] font-['Inter'] text-[13px] font-medium text-[#0f172a] shadow-[0px_1px_2px_0px_rgba(15,23,42,0.04)] outline-none transition-all placeholder:text-[#94a3b8] focus:border-[#94a3b8] focus:ring-2 focus:ring-[#dbeafe]" /></label><label class="flex flex-col gap-[8px]"><span class="flex items-start gap-[6px] font-['Inter'] text-[12px] font-semibold leading-[16px] text-[#334155]"><span class="min-w-0 flex-1">Longitude</span><span class="mt-[1px] shrink-0 whitespace-nowrap text-[#94a3b8]">(Optional)</span></span><InputText v-model="form.companyLongitude" placeholder="Example: 106.821810" class="h-[48px] w-full rounded-[14px] border border-[#d9e2ec] bg-white px-[14px] font-['Inter'] text-[13px] font-medium text-[#0f172a] shadow-[0px_1px_2px_0px_rgba(15,23,42,0.04)] outline-none transition-all placeholder:text-[#94a3b8] focus:border-[#94a3b8] focus:ring-2 focus:ring-[#dbeafe]" /></label></div></div><div class="rounded-[16px] border border-[#e2e8f0] bg-white p-[12px]"><div class="flex items-center justify-between gap-[10px]"><p class="font-['Inter'] text-[11px] font-bold uppercase tracking-[0.08em] text-[#64748b]">Preview</p><span class="rounded-full border px-[8px] py-[3px] text-[10px] font-bold border-[#e2e8f0] bg-white text-[#64748b]">Manual</span></div><div class="mt-[8px] rounded-[12px] bg-[#f8fafc] px-[12px] py-[10px]"><p class="font-['Inter'] text-[13px] font-semibold leading-[20px]" :class="form.companyAddress ? 'text-[#334155]' : 'text-[#94a3b8]'">{{ form.companyAddress || 'Company address preview will appear here.' }}</p></div></div>
                </div>
              </div>
            </div>
          </section>

          <!-- TAX INFORMATION -->
          <section class="tax-information-section overflow-hidden rounded-[20px] border border-[#e2e8f0] bg-white shadow-[0px_14px_34px_0px_rgba(15,23,42,0.06)]">
            <div class="border-b border-[#eef2f7] px-[20px] py-[18px]"><div class="flex flex-wrap items-center gap-[8px]"><h2 class="font-['Inter'] text-[16px] font-bold text-[#0f172a]">Tax Information</h2><span class="rounded-full border px-[8px] py-[3px] font-['Inter'] text-[11px] font-semibold border-[#ddd6fe] bg-[#f5f3ff] text-[#7c3aed]">Company + Site</span><span class="rounded-full border border-[#e2e8f0] bg-[#f8fafc] px-[8px] py-[3px] font-['Inter'] text-[11px] font-semibold text-[#64748b]">Optional</span></div><p class="mt-[4px] text-[12px] text-[#64748b]">Company NPWP data is synced from company details, while site-level tax and identity fields stay separate.</p></div>
            <div class="grid grid-cols-1 gap-[14px] px-[20px] py-[20px] md:grid-cols-2">
              <div class="overflow-hidden rounded-[18px] border border-[#dbe3ef] bg-white shadow-[0px_10px_24px_0px_rgba(15,23,42,0.04)]"><div class="border-b border-[#eef2f7] bg-[#f8fafc] px-[16px] py-[14px]"><div class="flex flex-wrap items-center gap-[8px]"><p class="font-['Inter'] text-[13px] font-bold text-[#0f172a]">Site Tax &amp; Identity</p><span class="rounded-full border border-[#bbf7d0] bg-[#f0fdf4] px-[8px] py-[3px] font-['Inter'] text-[10px] font-bold text-[#166534]">Site</span></div><p class="mt-[3px] font-['Inter'] text-[11px] leading-[16px] text-[#64748b]">Site-level document defaults and individual buyer identification.</p></div><div class="space-y-[16px] p-[16px]"><label class="contact-input-label">PPN<select v-model="form.pph" class="h-[48px] w-full rounded-[14px] border border-[#d9e2ec] bg-white px-[14px] font-['Inter'] text-[13px] font-medium text-[#0f172a] outline-none"><option>PPN 12%</option><option>PPN 11%</option><option>Non-PKP</option></select></label><label class="contact-input-label">ID TKU Number <span class="font-normal text-[#94a3b8]">(Optional)</span><div class="relative"><InputText v-model="form.idTkuNumber" placeholder="Input ID TKU number" class="contact-reference-input pr-[100px]" /><span class="pointer-events-none absolute right-[12px] top-1/2 -translate-y-1/2 rounded-full border border-[#bbf7d0] bg-[#f0fdf4] px-[8px] py-[3px] text-[10px] font-semibold text-[#166534]"># Numbers only</span></div></label><p class="font-['Inter'] text-[11px] leading-[16px] text-[#64748b]">Use this for outlets, branches, or stores if you have more than one location.</p><label class="contact-input-label">NIK <span class="font-normal text-[#94a3b8]">(Optional)</span><div class="relative"><InputText v-model="form.nik" placeholder="Input NIK" class="contact-reference-input pr-[100px]" /><span class="pointer-events-none absolute right-[12px] top-1/2 -translate-y-1/2 rounded-full border border-[#bbf7d0] bg-[#f0fdf4] px-[8px] py-[3px] text-[10px] font-semibold text-[#166534]"># Numbers only</span></div></label></div></div>
              <div class="overflow-hidden rounded-[18px] border border-[#dbe3ef] bg-white shadow-[0px_10px_24px_0px_rgba(15,23,42,0.04)]"><div class="border-b border-[#eef2f7] bg-[#f8fafc] px-[16px] py-[14px]"><div class="flex flex-wrap items-center gap-[8px]"><p class="font-['Inter'] text-[13px] font-bold text-[#0f172a]">Company Tax</p><span class="rounded-full border border-[#bfdbfe] bg-[#eff6ff] px-[8px] py-[3px] font-['Inter'] text-[10px] font-bold text-[#1d4ed8]">Company</span></div><p class="mt-[3px] font-['Inter'] text-[11px] leading-[16px] text-[#64748b]">Auto-filled from Company Name and Company Address when available.</p></div><div class="space-y-[16px] p-[16px]"><label class="contact-input-label">Company NPWP Name <span class="ml-auto font-normal text-[#94a3b8]">(Preview)</span><InputText v-model="form.companyNpwpName" :placeholder="form.parentCompanyName || 'Will be filled from Company Name.'" class="contact-reference-input bg-[#f8fafc]" /></label><p class="-mt-[8px] font-['Inter'] text-[11px] text-[#64748b]">Synced from Company Name.</p><label class="contact-input-label">Company NPWP Address <span class="ml-auto font-normal text-[#94a3b8]">(Preview)</span><InputText v-model="form.companyNpwpAddress" :placeholder="form.companyAddress || 'Will be filled from Company Address.'" class="contact-reference-input bg-[#f8fafc]" /></label><p class="-mt-[8px] font-['Inter'] text-[11px] text-[#64748b]">Synced from Company Address.</p><label class="contact-input-label">Company NPWP Number <span class="font-normal text-[#94a3b8]">(Optional)</span><InputText v-model="form.companyNpwpNumber" placeholder="Input company NPWP number" class="contact-reference-input" /></label></div></div>
            </div>
          </section>

          <!-- OTHER MASTER DATA -->
          <section class="other-master-data-section overflow-hidden rounded-[20px] border border-[#e2e8f0] bg-white shadow-[0px_14px_34px_0px_rgba(15,23,42,0.06)]">
            <div class="border-b border-[#eef2f7] px-[20px] py-[18px]"><div class="flex flex-wrap items-center gap-[8px]"><h2 class="font-['Inter'] text-[16px] font-bold text-[#0f172a]">Other Master Data</h2><span class="rounded-full border border-[#bbf7d0] bg-[#f0fdf4] px-[8px] py-[3px] font-['Inter'] text-[11px] font-semibold text-[#166534]">Site</span><span class="rounded-full border border-[#e2e8f0] bg-[#f8fafc] px-[8px] py-[3px] font-['Inter'] text-[11px] font-semibold text-[#64748b]">Optional</span></div><p class="mt-[4px] text-[12px] text-[#64748b]">Site-level document defaults prefilled from Master Data &gt; Customer Defaults.</p></div>
          <div class="space-y-[14px] p-[20px]"><div class="rounded-[16px] border border-[#dbe3ef] bg-white p-[14px]"><p class="font-['Inter'] text-[11px] font-bold uppercase tracking-[0.08em] text-[#64748b]">Delivery Master Data</p><p class="mt-[4px] font-['Inter'] text-[11px] text-[#64748b]">Used for delivery charge defaults and dispatch documents.</p><label class="mt-[14px] block font-['Inter'] text-[12px] font-semibold text-[#334155]">Shipment Cost<div class="relative mt-[8px] max-w-[380px]"><InputText v-model="form.shipmentCost" inputmode="numeric" class="contact-reference-input pl-[60px] pr-[115px]" /><span class="absolute left-0 top-0 flex h-[48px] w-[46px] items-center justify-center border-r border-[#e2e8f0] font-['Inter'] text-[13px] text-[#64748b]">Rp</span><span class="pointer-events-none absolute right-[12px] top-1/2 -translate-y-1/2 rounded-full border border-[#bbf7d0] bg-[#f0fdf4] px-[10px] py-[4px] font-['Inter'] text-[10px] font-bold text-[#166534]"># Numbers only</span></div></label></div><div class="rounded-[16px] border border-[#dbe3ef] bg-white p-[14px]"><p class="font-['Inter'] text-[11px] font-bold uppercase tracking-[0.08em] text-[#64748b]">Invoice Master Data</p><p class="mt-[4px] font-['Inter'] text-[11px] text-[#64748b]">Used for invoice type and settlement account defaults.</p><div class="mt-[14px] grid grid-cols-1 gap-[12px] md:grid-cols-2"><label class="contact-input-label">Invoice Type<Select v-model="form.invoiceType" :options="invoiceTypeOptions" placeholder="Select invoice type" filter filterPlaceholder="Search..." appendTo="body" class="master-data-select invoice-type-select"><template #option="slotProps"><div class="invoice-option-row"><span :class="['invoice-option-check', { checked: form.invoiceType === slotProps.option }]"><i v-if="form.invoiceType === slotProps.option" class="pi pi-check" /></span><span>{{ slotProps.option }}</span></div></template></Select></label><label class="contact-input-label">Bank Account<Select v-model="form.bankAccount" :options="bankAccountOptions" placeholder="Select bank account" filter filterPlaceholder="Search..." appendTo="body" class="master-data-select bank-account-select"><template #option="slotProps"><div class="invoice-option-row"><span :class="['invoice-option-check', { checked: form.bankAccount === slotProps.option }]"><i v-if="form.bankAccount === slotProps.option" class="pi pi-check" /></span><span>{{ slotProps.option }}</span></div></template></Select></label><label class="contact-input-label">Term of Payment<div class="flex"><InputText v-model="form.termPayment" inputmode="numeric" class="contact-reference-input rounded-r-none" /><span class="flex h-[48px] items-center rounded-r-[14px] border border-l-0 border-[#d9e2ec] bg-[#f8fafc] px-[12px] font-['Inter'] text-[12px] text-[#64748b]">Days</span></div></label></div></div></div>
          </section>

          <!-- BILLING & SHIPMENT CONFIGURATION -->
          <section class="billing-shipment-section overflow-hidden rounded-[20px] border border-[#e2e8f0] bg-white shadow-[0px_14px_34px_0px_rgba(15,23,42,0.06)]"><div class="border-b border-[#eef2f7] px-[20px] py-[18px]"><div class="flex flex-wrap items-center gap-[8px]"><h2 class="font-['Inter'] text-[16px] font-bold text-[#0f172a]">Billing &amp; Shipment Configuration</h2><span class="rounded-full border px-[8px] py-[3px] font-['Inter'] text-[11px] font-semibold border-[#bbf7d0] bg-[#f0fdf4] text-[#166534]">Site</span><span class="rounded-full border border-[#e2e8f0] bg-[#f8fafc] px-[8px] py-[3px] font-['Inter'] text-[11px] font-semibold text-[#64748b]">Optional</span></div><p class="mt-[4px] text-[12px] text-[#64748b]">Site-level document header settings. Default sources are prefilled from Master Data &gt; Customer Defaults.</p></div><div class="rounded-[18px] border border-[#e2e8f0] bg-[#f8fafc] p-[16px]"><p class="font-['Inter'] text-[13px] font-bold text-[#0f172a]">Document Header</p><p class="mt-[4px] font-['Inter'] text-[11px] leading-[16px] text-[#64748b]">Standard seller identity and recipient details for documents and delivery flows.</p><div class="mt-[14px] rounded-[16px] border border-[#dbe3ef] bg-white p-[16px]"><div class="flex items-start justify-between gap-[16px]"><div class="flex min-w-0 gap-[14px]"><span class="flex size-[48px] shrink-0 items-center justify-center rounded-[14px] border border-[#dbe3ef] bg-white font-['Inter'] text-[14px] font-black text-[#111827] shadow-[0px_6px_18px_0px_rgba(15,23,42,0.05)]">PY</span><div class="min-w-0"><p class="font-['Inter'] text-[11px] font-bold uppercase tracking-[0.08em] text-[#64748b]">Seller Identity</p><p class="mt-[6px] break-words font-['Inter'] text-[20px] font-bold leading-[26px] text-[#111827]">{{ form.parentCompanyName || 'Company Name' }}</p><p class="mt-[8px] max-w-[780px] break-words font-['Inter'] text-[13px] font-medium leading-[21px] text-[#334155]">{{ form.companyAddress || '[Company Address]' }}</p><div class="mt-[10px] flex flex-wrap gap-x-[18px] gap-y-[4px]"><p class="font-['Inter'] text-[12px] font-semibold text-[#475569]">Phone: <span class="font-medium text-[#334155]">â€”</span></p><p class="font-['Inter'] text-[12px] font-semibold text-[#475569]">NPWP: <span class="font-medium text-[#334155]">â€”</span></p><p class="font-['Inter'] text-[12px] font-semibold text-[#475569]">Website: <span class="font-medium text-[#334155]">â€”</span></p></div></div></div><div class="rounded-[14px] border border-[#dbe3ef] px-[14px] py-[12px]"><p class="font-['Inter'] text-[10px] uppercase tracking-[0.08em] text-[#64748b]">Customer ID</p><p class="mt-[8px] font-['Inter'] text-[13px] font-bold text-[#0f172a]">{{ previewCustomerSiteCode }}</p></div></div></div><div class="mt-[18px] grid grid-cols-1 gap-[14px] md:grid-cols-2"><div class="rounded-[16px] border border-[#dbe3ef] bg-white"><div class="border-b border-[#eef2f7] p-[14px]"><p class="font-['Inter'] text-[10px] uppercase tracking-[0.08em] text-[#64748b]">Source</p><div class="mt-[8px] flex gap-[6px] rounded-[12px] border border-[#fecaca] bg-[#fff7f7] p-[4px]"><button type="button" class="rounded-[9px] bg-[#dc2626] px-[10px] py-[7px] font-['Inter'] text-[11px] font-semibold text-white">Company</button><button type="button" class="rounded-[9px] bg-white px-[10px] py-[7px] font-['Inter'] text-[11px] text-[#475569]">Site</button></div><p class="mt-[12px] font-['Inter'] text-[12px] font-bold uppercase tracking-[0.08em] text-[#0f172a]">Bill To:</p></div><div class="p-[14px] font-['Inter'] text-[13px] text-[#0f172a]"><p class="font-semibold">{{ form.parentCompanyName || '[Company Name]' }}</p><p class="mt-[6px]">{{ form.companyAddress || '[Street Address]' }}</p></div></div><div class="rounded-[16px] border border-[#dbe3ef] bg-white"><div class="border-b border-[#eef2f7] p-[14px]"><p class="font-['Inter'] text-[10px] uppercase tracking-[0.08em] text-[#64748b]">Source</p><div class="mt-[8px] flex gap-[6px] rounded-[12px] border border-[#fecaca] bg-[#fff7f7] p-[4px]"><button type="button" class="rounded-[9px] bg-white px-[10px] py-[7px] font-['Inter'] text-[11px] text-[#475569]">Company</button><button type="button" class="rounded-[9px] bg-[#dc2626] px-[10px] py-[7px] font-['Inter'] text-[11px] font-semibold text-white">Site</button><button type="button" class="rounded-[9px] bg-white px-[10px] py-[7px] font-['Inter'] text-[11px] text-[#475569]">Other Delivery</button></div><p class="mt-[12px] font-['Inter'] text-[12px] font-bold uppercase tracking-[0.08em] text-[#0f172a]">Ship To:</p></div><div class="p-[14px] font-['Inter'] text-[13px] text-[#0f172a]"><p class="font-semibold">{{ form.name || '[Name]' }}</p><p class="mt-[6px]">{{ form.address || '[Street Address]' }}</p></div></div></div></div></section>

          <section v-if="shipToSource === 'delivery'" class="billing-delivery-fields rounded-[16px] border border-[#dbe3ef] bg-white p-[16px]"><p class="font-['Inter'] text-[12px] font-bold uppercase tracking-[0.08em] text-[#64748b]">Other Delivery</p><label class="mt-[14px] flex flex-col gap-[8px] font-['Inter'] text-[11px] font-bold uppercase tracking-[0.08em] text-[#64748b]">Delivery Name<InputText v-model="form.deliveryName" placeholder="Example: Warehouse A" class="delivery-field" /></label><label class="mt-[14px] flex flex-col gap-[8px] font-['Inter'] text-[11px] font-bold uppercase tracking-[0.08em] text-[#64748b]">Delivery Address<Textarea v-model="form.deliveryAddress" placeholder="Type the delivery destination street address" rows="3" class="delivery-field delivery-textarea" /></label></section>

          <section class="sales-assignment-section overflow-hidden rounded-[20px] border border-[#e2e8f0] bg-white shadow-[0px_14px_34px_0px_rgba(15,23,42,0.06)]"><div class="border-b border-[#eef2f7] px-[20px] py-[18px]"><div class="flex flex-wrap items-center gap-[8px]"><h2 class="font-['Inter'] text-[16px] font-bold text-[#0f172a]">Sales Assignment for Branch, Outlet, and Store</h2><span class="rounded-full border px-[8px] py-[3px] font-['Inter'] text-[11px] font-semibold border-[#bbf7d0] bg-[#f0fdf4] text-[#166534]">Site</span><span class="rounded-full border border-[#e2e8f0] bg-[#f8fafc] px-[8px] py-[3px] font-['Inter'] text-[11px] font-semibold text-[#64748b]">Optional</span></div><p class="mt-[4px] text-[12px] text-[#64748b]">Site-level ownership for branch, outlet, or store accounts with non-overlapping periods.</p></div><div class="md:col-span-2 space-y-[14px]"><div class="max-w-[560px] rounded-[16px] border border-[#e2e8f0] bg-[#f8fafc] px-[14px] py-[12px]"><p class="font-['Inter'] text-[12px] font-semibold text-[#0f172a]">How it works</p><p class="mt-[4px] text-[12px] leading-[18px] text-[#64748b]">Use this for branch, outlet, or store-level ownership. End can stay open as Until Now, or be closed with a specific month and year.</p></div><div v-for="period in salesAssignmentCount" :key="period" class="rounded-[18px] border border-[#e2e8f0] bg-white p-[14px]"><div class="flex items-start justify-between gap-[12px]"><div><p class="font-['Inter'] text-[12px] font-bold uppercase tracking-[0.08em] text-[#94a3b8]">Period {{ period }}</p><p class="mt-[4px] font-['Inter'] text-[13px] font-semibold text-[#0f172a]">Select sales owner and assignment period</p></div><button type="button" @click="removeSalesAssignment" class="rounded-full p-[7px] text-[#94a3b8]" aria-label="Remove sales assignment period"><svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="lucide lucide-x size-[14px]"><path d="M18 6 6 18"></path><path d="m6 6 12 12"></path></svg></button></div><div class="mt-[12px] grid grid-cols-1 gap-[12px] lg:grid-cols-2"><label class="contact-input-label">Sales Executive<Select v-model="form.salesExecutiveId" :options="salesExecutives" optionLabel="fullName" optionValue="id" placeholder="Select sales executive" showClear filter filterPlaceholder="Search..." appendTo="body" class="sales-assignment-select"><template #option="slotProps"><div class="invoice-option-row"><span :class="['invoice-option-check', { checked: form.salesExecutiveId === (slotProps.option.value || slotProps.option.id || slotProps.option) }]"><i v-if="form.salesExecutiveId === (slotProps.option.value || slotProps.option.id || slotProps.option)" class="pi pi-check" /></span><span>{{ slotProps.option.label || slotProps.option.fullName || slotProps.option }}</span></div></template></Select></label><label class="contact-input-label">Start Month<Select v-model="form.startMonth" :options="monthOptions" optionLabel="label" optionValue="value" placeholder="Month" showClear filter filterPlaceholder="Search..." appendTo="body" class="sales-assignment-select"><template #option="slotProps"><div class="invoice-option-row"><span :class="['invoice-option-check', { checked: form.startMonth === (slotProps.option.value || slotProps.option.id || slotProps.option) }]"><i v-if="form.startMonth === (slotProps.option.value || slotProps.option.id || slotProps.option)" class="pi pi-check" /></span><span>{{ slotProps.option.label || slotProps.option.fullName || slotProps.option }}</span></div></template></Select></label><label class="contact-input-label">Start Year<Select v-model="form.startYear" :options="yearOptions" optionLabel="label" optionValue="value" placeholder="Year" showClear filter filterPlaceholder="Search..." appendTo="body" class="sales-assignment-select"><template #option="slotProps"><div class="invoice-option-row"><span :class="['invoice-option-check', { checked: form.startYear === (slotProps.option.value || slotProps.option.id || slotProps.option) }]"><i v-if="form.startYear === (slotProps.option.value || slotProps.option.id || slotProps.option)" class="pi pi-check" /></span><span>{{ slotProps.option.label || slotProps.option.fullName || slotProps.option }}</span></div></template></Select></label><label class="contact-input-label lg:col-span-2">End<Select v-model="form.endPeriod" :options="[{ label: 'Until Now', value: 'Until Now' }]" optionLabel="label" optionValue="value" placeholder="Until Now" showClear appendTo="body" class="sales-assignment-select"><template #option="slotProps"><div class="invoice-option-row"><span :class="['invoice-option-check', { checked: form.endPeriod === (slotProps.option.value || slotProps.option.id || slotProps.option) }]"><i v-if="form.endPeriod === (slotProps.option.value || slotProps.option.id || slotProps.option)" class="pi pi-check" /></span><span>{{ slotProps.option.label || slotProps.option.fullName || slotProps.option }}</span></div></template></Select></label></div></div><button type="button" @click="addSalesAssignment" class="inline-flex h-[40px] items-center gap-[7px] rounded-[10px] border border-[#e2e8f0] bg-white px-[14px] font-['Inter'] text-[12px] font-semibold text-[#475569]"><i class="pi pi-plus text-[14px]" />Add Another Sales Assignment</button></div></section>

          <!-- CONTACTS -->
          <div class="contacts-information-section form-card">
            <div class="border-b border-[#eef2f7] px-[20px] py-[18px]">
              <div class="flex flex-wrap items-center gap-[8px]"><h2 class="font-['Inter'] text-[16px] font-bold text-[#0f172a]">Contact Information</h2><span class="rounded-full border px-[8px] py-[3px] font-['Inter'] text-[11px] font-semibold border-[#ddd6fe] bg-[#f5f3ff] text-[#7c3aed]">Site + Company</span><span class="rounded-full border border-[#e2e8f0] bg-[#f8fafc] px-[8px] py-[3px] font-['Inter'] text-[11px] font-semibold text-[#64748b]">Optional</span></div>
              <p class="mt-[4px] text-[12px] text-[#64748b]">Keep site contacts and company contacts separate so operational and billing communication does not get mixed.</p>
            </div>
            <div class="grid grid-cols-1 gap-[18px] p-[20px] md:grid-cols-2">

              <div class="overflow-hidden rounded-[18px] border border-[#dbe3ef] bg-white shadow-[0px_10px_24px_0px_rgba(15,23,42,0.04)]">
                <div class="border-b border-[#eef2f7] bg-[#f8fafc] px-[16px] py-[14px]"><div class="flex items-start gap-[12px]"><span class="flex size-[40px] shrink-0 items-center justify-center rounded-[12px] border border-[#dbe3ef] bg-white text-[#dc2626] shadow-[0px_6px_18px_0px_rgba(15,23,42,0.05)]"><i class="pi pi-user text-[17px]" /></span><div class="min-w-0"><div class="flex flex-wrap items-center gap-[8px]"><p class="font-['Inter'] text-[13px] font-bold text-[#0f172a]">Customer Site Contacts</p><span class="shrink-0 rounded-full border px-[8px] py-[3px] font-['Inter'] text-[10px] font-bold tracking-[0.02em] border-[#bbf7d0] bg-[#f0fdf4] text-[#166534]">Site</span></div><p class="mt-[3px] text-[11px] leading-[16px] text-[#64748b]">Use this list for outlet, branch, or site-level contacts.</p></div></div></div>
                <div class="space-y-[12px] p-[16px]"><div v-for="(contact, idx) in form.contacts" :key="idx" class="rounded-[16px] border border-[#e2e8f0] bg-[#fbfdff] p-[14px]"><div class="flex items-start justify-between gap-[12px]"><div class="min-w-0"><p class="font-['Inter'] text-[11px] font-bold uppercase tracking-[0.08em] text-[#94a3b8]">Contact {{ idx + 1 }}</p><p class="mt-[4px] truncate font-['Inter'] text-[13px] font-semibold text-[#0f172a]">{{ contact.name || 'New contact' }}</p></div><button v-if="form.contacts.length > 1" type="button" class="rounded-full p-[7px] text-[#94a3b8] transition-colors hover:bg-[#f8fafc] hover:text-[#dc2626]" aria-label="Remove customer site contact" @click="removeContact(idx)"><svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="lucide lucide-x size-[14px]"><path d="M18 6 6 18"></path><path d="m6 6 12 12"></path></svg></button></div><div class="mt-[12px] space-y-[12px]"><label class="contact-input-label">Contact Name<InputText v-model="contact.name" placeholder="Example: Dina Pratama" class="contact-reference-input" /></label><label class="contact-input-label">Position<InputText v-model="contact.position" placeholder="Example: Outlet Supervisor" class="contact-reference-input" /></label><label class="contact-input-label">Phone Number<InputText v-model="contact.phone" placeholder="Example: 0812-3456-7890" inputmode="tel" class="contact-reference-input" /></label><label class="contact-input-label">Email Address<InputText v-model="contact.email" placeholder="Example: outlet@company.com" type="email" class="contact-reference-input" /></label></div></div><button type="button" class="inline-flex h-[40px] items-center gap-[7px] rounded-[10px] border border-[#e2e8f0] bg-white px-[14px] font-['Inter'] text-[12px] font-semibold text-[#475569] transition-colors hover:bg-[#f8fafc]" @click="addContact"><i class="pi pi-plus text-[14px]" />Add Another Site Contact</button></div>
              </div>
              <div class="overflow-hidden rounded-[18px] border border-[#dbe3ef] bg-white shadow-[0px_10px_24px_0px_rgba(15,23,42,0.04)]">
                <div class="border-b border-[#eef2f7] bg-[#f8fafc] px-[16px] py-[14px]"><div class="flex items-start gap-[12px]"><span class="flex size-[40px] shrink-0 items-center justify-center rounded-[12px] border border-[#dbe3ef] bg-white text-[#dc2626] shadow-[0px_6px_18px_0px_rgba(15,23,42,0.05)]"><i class="pi pi-building text-[17px]" /></span><div class="min-w-0"><div class="flex flex-wrap items-center gap-[8px]"><p class="font-['Inter'] text-[13px] font-bold text-[#0f172a]">Company Contacts</p><span class="shrink-0 rounded-full border border-[#bfdbfe] bg-[#eff6ff] px-[8px] py-[3px] font-['Inter'] text-[10px] font-bold tracking-[0.02em] text-[#1d4ed8]">Company</span></div><p class="mt-[3px] text-[11px] leading-[16px] text-[#64748b]">Use this list for billing, tax, or general company contacts.</p></div></div></div>
                <div class="space-y-[12px] p-[16px]"><div v-for="(contact, idx) in form.companyContacts" :key="idx" class="rounded-[16px] border border-[#e2e8f0] bg-[#fbfdff] p-[14px]"><div class="flex items-start justify-between gap-[12px]"><div class="min-w-0"><p class="font-['Inter'] text-[11px] font-bold uppercase tracking-[0.08em] text-[#94a3b8]">Contact {{ idx + 1 }}</p><p class="mt-[4px] truncate font-['Inter'] text-[13px] font-semibold text-[#0f172a]">{{ contact.name || 'New contact' }}</p></div><button v-if="form.companyContacts.length > 1" type="button" class="rounded-full p-[7px] text-[#94a3b8] transition-colors hover:bg-[#f8fafc] hover:text-[#dc2626]" aria-label="Remove company contact" @click="removeCompanyContact(idx)"><svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="lucide lucide-x size-[14px]"><path d="M18 6 6 18"></path><path d="m6 6 12 12"></path></svg></button></div><div class="mt-[12px] space-y-[12px]"><label class="contact-input-label">Contact Name<InputText v-model="contact.name" placeholder="Example: Andi Wijaya" class="contact-reference-input" /></label><label class="contact-input-label">Position<InputText v-model="contact.position" placeholder="Example: Finance Manager" class="contact-reference-input" /></label><label class="contact-input-label">Phone Number<InputText v-model="contact.phone" placeholder="Example: 0813-9876-5432" inputmode="tel" class="contact-reference-input" /></label><label class="contact-input-label">Email Address<InputText v-model="contact.email" placeholder="Example: finance@company.com" type="email" class="contact-reference-input" /></label></div></div><button type="button" class="inline-flex h-[40px] items-center gap-[7px] rounded-[10px] border border-[#e2e8f0] bg-white px-[14px] font-['Inter'] text-[12px] font-semibold text-[#475569] transition-colors hover:bg-[#f8fafc]" @click="addCompanyContact"><i class="pi pi-plus text-[14px]" />Add Another Company Contact</button></div>
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
                <strong>{{ form.name || 'â€”' }}</strong>
              </div>
              <div class="summary-row">
                <span>Segment</span>
                <strong>{{ form.customerSegment || 'â€”' }}</strong>
              </div>
              <div class="summary-row">
                <span>Category</span>
                <strong>{{ form.customerCategory || 'â€”' }}</strong>
              </div>
              <div class="summary-row">
                <span>Region</span>
                <strong>{{ form.region || 'â€”' }}</strong>
              </div>
              <div class="summary-row">
                <span>Company</span>
                <strong>{{ form.parentCompanyName || 'â€”' }}</strong>
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
          <div v-if="form.parentCompanyName.trim()" class="sidebar-card selected-parent-card">
            <p class="selected-parent-label">Selected Parent</p>
            <div class="selected-parent-content"><div><p class="selected-parent-code">{{ form.parentCode || '—' }}</p><p class="selected-parent-name">{{ form.parentCompanyName }}</p><p class="selected-parent-total">Total Customer: {{ store.allCustomers.filter((customer) => customer.parentCode === form.parentCode).length || 1 }}</p></div><span class="selected-parent-status">Active</span></div>
          </div>
          <div v-if="!similarCustomerDismissed && form.name.trim() && form.parentCompanyName.trim()" class="sidebar-card similar-customer-card">
            <div class="similar-customer-heading"><i class="pi pi-exclamation-triangle" /><div><h4>Similar customer found</h4><p>We found similar parent/customer data. Please check before creating duplicate.</p></div></div>
            <div class="similar-customer-item"><strong>{{ existingCustomerCode || 'Existing customer' }}</strong><span>{{ form.name }}</span><small>Type: Customer</small><button type="button" @click="selectParentCompanyMode('existing'); similarCustomerDismissed = true">Use Existing Parent</button></div>
            <button type="button" class="similar-customer-ignore" @click="similarCustomerDismissed = true">Ignore and Continue</button>
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
                  <div class="scope-check-row"><span class="scope-check" :class="{ complete: form.province }"><i v-if="form.province" class="pi pi-check" /></span><span>City, Province</span></div>
                </div>
              </div>
              <div class="scope-group">
                <div class="scope-group-heading"><span>COMPANY</span><span class="scope-badge company">Company</span></div>
                <div class="scope-checklist">
                  <div class="scope-check-row"><span class="scope-check" :class="{ complete: form.companyAddress }"><i v-if="form.companyAddress" class="pi pi-check" /></span><span>Company Address</span></div>
                  <div class="scope-check-row"><span class="scope-check" :class="{ complete: form.companyProvince }"><i v-if="form.companyProvince" class="pi pi-check" /></span><span>Company City, Province</span></div>
                </div>
              </div>
              <div class="scope-row"><span class="scope-badge company">Reference</span><span>Parent company name and code are reference fields; create companies separately.</span></div>
            </div>
          </div>
          <div class="sidebar-actions">
            <Button label="Save Changes" icon="pi pi-check" class="full-width sidebar-create-button" :loading="saving" :disabled="!isFormValid || saving" @click="handleSubmit" />
            <Button label="Cancel" severity="secondary" text class="full-width" @click="router.push('/admin/customers')" />
          </div>
        </aside>
      </div>
    </template>
  </section>
</template>

<style scoped>
.form-sidebar{max-height:calc(100vh - 88px);overflow-y:auto;overscroll-behavior:contain;padding-right:4px}.form-sidebar .similar-customer-card{order:-2}.form-sidebar .field-scope-card{order:-1}
.similar-customer-card{border-color:#fde68a;background:#fffbeb}.similar-customer-heading{display:flex;align-items:flex-start;gap:10px;color:#92400e}.similar-customer-heading>i{margin-top:2px;color:#d97706}.similar-customer-heading h4{margin:0;color:#92400e;font-size:13px}.similar-customer-heading p{margin:3px 0 0;color:#92400e;font-size:12px;line-height:18px}.similar-customer-item{display:grid;gap:2px;margin-top:10px;padding:12px;border-radius:14px;background:#fff;color:#0f172a}.similar-customer-item strong{font-size:12px}.similar-customer-item span{font-size:12px;color:#475569}.similar-customer-item small{font-size:11px;color:#94a3b8}.similar-customer-item button{width:max-content;margin-top:6px;padding:7px 10px;border-radius:9px;background:#dc2626;color:#fff;font-size:11px;font-weight:700}.similar-customer-ignore{margin-top:10px;color:#92400e;font-size:11px;font-weight:700;text-decoration:underline}
.form-sidebar .similar-customer-card{order:-4!important}.form-sidebar .field-scope-card{order:-3!important}
.form-sidebar .selected-parent-card{order:-5!important}.selected-parent-card{background:#f8fafc}.selected-parent-label{margin:0;color:#94a3b8;font-size:11px;font-weight:700;letter-spacing:.08em;text-transform:uppercase}.selected-parent-content{display:flex;align-items:flex-start;justify-content:space-between;gap:12px;margin-top:8px}.selected-parent-code{margin:0;color:#dc2626;font-size:12px;font-weight:700}.selected-parent-name{margin:3px 0 0;color:#0f172a;font-size:14px;font-weight:700}.selected-parent-total{margin:3px 0 0;color:#64748b;font-size:12px}.selected-parent-status{display:inline-flex;border:1px solid #bbf7d0;border-radius:999px;background:#f0fdf4;padding:3px 9px;color:#166534;font-size:11px;font-weight:700}
.create-customer-header {
  box-sizing: border-box;
  width: calc(100% + 4rem);
  min-height: 59px;
  margin: -1.75rem -2rem 0;
  padding: 9px 16px !important;
}
.create-customer-header-inner {
  min-height: 40px;
}
.create-back-button {
  min-height: 0 !important;
  border: 0 !important;
  border-radius: 0 !important;
  background: transparent !important;
  padding: 0 !important;
  color: #64748b !important;
  box-shadow: none !important;
  font-family: Inter, ui-sans-serif, system-ui, sans-serif !important;

  font-size: 12px !important;
  font-weight: 600 !important;
  line-height: 16px !important;
}
.create-back-button:hover {
  color: #dc2626 !important;
}
.create-back-button svg {
  width: 14px !important;
  height: 14px !important;
}
.create-customer-header h1 {
  margin: 0 !important;
  font-family: Inter, ui-sans-serif, system-ui, sans-serif !important;
  font-size: 17px !important;
  font-weight: 700 !important;
  line-height: 22px !important;
  color: #0f172a !important;
}
.create-customer-header p {
  margin: 0 !important;
  font-family: Inter, ui-sans-serif, system-ui, sans-serif !important;
  font-size: 11px !important;
  line-height: 15px !important;
  color: #94a3b8 !important;
}
.create-cancel-button,
.create-submit-button {
  min-height: 40px !important;
  height: 40px !important;
  border-radius: 10px !important;
  font-family: Inter, ui-sans-serif, system-ui, sans-serif !important;
  font-size: 12px !important;
  font-weight: 600 !important;
  line-height: 16px !important;
}
.create-cancel-button {
  width: 69px;
  min-width: 70px;
  align-items: center !important;
  justify-content: center !important;
  gap: 7px !important;
  border: 1px solid #e2e8f0 !important;
  background: #fff !important;
  padding: 0 14px !important;
  color: #475569 !important;
  box-shadow: none !important;
  transition: all 150ms ease !important;
}
.create-cancel-button:hover {
  background: #f8fafc !important;
}
.create-submit-button {
  width: 151px;
  min-width: 152px;
  align-items: center !important;
  justify-content: center !important;
  gap: 7px !important;
  border: 0 !important;
  background: linear-gradient(to bottom right, #dc2626, #b91c1c) !important;
  background-color: #dc2626 !important;
  background-image: linear-gradient(to bottom right, #dc2626, #b91c1c) !important;
  padding: 0 16px !important;
  color: #fff !important;
  -webkit-text-fill-color: #fff !important;
  box-shadow: 0 4px 16px rgba(220, 38, 38, .22) !important;
  transition: all 150ms ease !important;
}
.create-submit-button:hover {
  filter: brightness(1.1);
}
.create-submit-button:disabled {
  cursor: not-allowed !important;
  background: linear-gradient(to bottom right, #dc2626, #b91c1c) !important;
  background-color: #dc2626 !important;
  background-image: linear-gradient(to bottom right, #dc2626, #b91c1c) !important;
  color: #fff !important;
  -webkit-text-fill-color: #fff !important;
  box-shadow: 0 4px 16px rgba(220, 38, 38, .22) !important;
  filter: none !important;
}
.create-submit-button svg {
  width: 14px !important;
  height: 14px !important;
}
.customer-info-input {
  min-height: inherit !important;
  border-color: #d9e2ec !important;
  border-radius: 14px !important;
  background: #fff !important;
  font-family: Inter, ui-sans-serif, system-ui, sans-serif !important;
  font-size: 13px !important;
  font-weight: 500 !important;
  color: #0f172a !important;
  box-shadow: 0 1px 2px rgba(15, 23, 42, .04) !important;
}
.customer-info-input::placeholder {
  color: #94a3b8 !important;
}
.customer-info-input:focus {

  border-color: #94a3b8 !important;
  box-shadow: 0 0 0 2px #dbeafe !important;
}
.customer-info-select {
  width: 100%;
  height: 48px !important;
  min-height: 48px !important;
  border: 1px solid #d9e2ec !important;
  border-radius: 14px !important;
  background: #fff !important;
  box-shadow: 0 1px 2px rgba(15, 23, 42, .04) !important;
}
.customer-info-select:hover {
  border-color: #cbd5e1 !important;
}
.customer-info-select :deep(.p-select-label) {
  display: flex;
  align-items: center;
  height: 46px;
  padding: 0 14px !important;
  font-family: Inter, ui-sans-serif, system-ui, sans-serif !important;
  font-size: 13px !important;
  font-weight: 500 !important;
  color: #0f172a !important;
}
.customer-info-select :deep(.p-placeholder) {
  color: #94a3b8 !important;
}
.customer-info-select :deep(.p-select-dropdown) {
  width: 42px;
  color: #64748b;
}
.customer-info-select :deep(.p-select-dropdown-icon) {
  width: 16px;
  height: 16px;
}
.customer-info-input-search {
  padding-left: 42px !important;
}
:deep(input[placeholder^="Example:"]) {
  box-sizing: border-box !important;
  width: 100% !important;
  height: 48px !important;
  padding: 0 14px !important;
  border: 1px solid #d9e2ec !important;
  border-radius: 14px !important;
  background: #fff !important;
  font-family: Inter, ui-sans-serif, system-ui, sans-serif !important;
  font-size: 13px !important;
  font-weight: 500 !important;
  line-height: 20px !important;
  color: #0f172a !important;
  box-shadow: 0 1px 2px rgba(15, 23, 42, .04) !important;
}
:deep(input[placeholder^="Example:"]::placeholder) {
  color: #94a3b8 !important;
  opacity: 1 !important;
}
:deep(input[placeholder^="Example:"]:focus) {
  border-color: #94a3b8 !important;
  outline: none !important;
  box-shadow: 0 0 0 2px #dbeafe !important;
}
.contact-input-label {
  display: flex;
  flex-direction: column;
  gap: 8px;
  font-family: Inter, ui-sans-serif, system-ui, sans-serif;
  font-size: 12px;
  font-weight: 600;
  line-height: 16px;
  color: #334155;
}
.contact-reference-input {
  width: 100% !important;
  height: 48px !important;
  border: 1px solid #d9e2ec !important;
  border-radius: 14px !important;
  background: #fff !important;
  padding: 0 14px !important;
  font-family: Inter, ui-sans-serif, system-ui, sans-serif !important;
  font-size: 13px !important;
  font-weight: 500 !important;
  color: #0f172a !important;
  box-shadow: 0 1px 2px rgba(15, 23, 42, .04) !important;
}
.contact-reference-input::placeholder { color: #94a3b8 !important; }
.tax-information-section select {
  box-sizing: border-box !important;
  width: 100% !important;
  height: 48px !important;
  padding: 0 14px !important;
  border: 1px solid #d9e2ec !important;
  border-radius: 14px !important;
  background: #fff !important;
  font-family: Inter, ui-sans-serif, system-ui, sans-serif !important;
  font-size: 13px !important;
  font-weight: 600 !important;
  color: #0f172a !important;
}

.tax-information-section select option:nth-child(2),
.tax-information-section select option:nth-child(3) {
  display: none;
}
.tax-information-section .contact-input-label > span {
  font-family: Inter, ui-sans-serif, system-ui, sans-serif;
}
.tax-information-section label.contact-input-label > span.font-normal {
  position: absolute !important;
  top: 0 !important;
  right: 0 !important;
  display: inline-flex !important;
  flex-shrink: 0 !important;
  margin-left: 0 !important;
  white-space: nowrap !important;
  font-size: 12px !important;
  font-weight: 400 !important;
  line-height: 16px !important;
  color: #94a3b8 !important;
  align-self: auto !important;
}
.tax-information-section label.contact-input-label {
  position: relative !important;
}
.form-card > .grid > div > div > button:last-child {
  height: 40px !important;
  gap: 7px !important;
  padding: 0 14px !important;
  border: 1px solid #e2e8f0 !important;
  border-radius: 10px !important;
  background: #fff !important;
  font-family: Inter, ui-sans-serif, system-ui, sans-serif !important;
  font-size: 12px !important;
  font-weight: 600 !important;
  line-height: 16px !important;
  color: #475569 !important;
}
.form-card > .grid > div > div > button:last-child i {
  font-size: 14px !important;
}
.city-province-trigger {
  display: flex;
  width: 100%;
  height: 48px;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
  border: 1px solid #d9e2ec;
  border-radius: 14px;
  background: #fff;
  padding: 0 14px;
  text-align: left;
  font-family: Inter, ui-sans-serif, system-ui, sans-serif;
  font-size: 13px;
  font-weight: 500;
  box-shadow: 0 1px 2px rgba(15, 23, 42, .04);
  outline: none;
  transition: all 150ms ease;
}
.city-province-trigger:hover {
  border-color: #cbd5e1;
}
.city-province-trigger.open {
  border-color: #94a3b8;
  box-shadow: 0 0 0 1px #cbd5e1;
}
.city-province-dropdown {
  position: absolute;
  top: calc(100% + 8px);
  left: 0;
  z-index: 60;
  width: 100%;
  max-height: 330px;
  overflow: hidden;
  border: 1px solid #dbe3ef;
  border-radius: 14px;
  background: #fff;
  box-shadow: 0 16px 36px rgba(15, 23, 42, .12);
}
.city-province-dropdown > div:first-child,
.invoice-type-dropdown > div:first-child,
.bank-account-dropdown > div:first-child {
  border-bottom: 1px solid #eef2f7;
  padding: 12px;
}
.city-province-dropdown input,
.invoice-type-search,
.bank-account-search {
  box-sizing: border-box !important;
  height: 40px !important;
  width: 100% !important;
  border: 1px solid #d9e2ec !important;
  border-radius: 12px !important;
  background: #fff !important;
  font-family: Inter, ui-sans-serif, system-ui, sans-serif !important;
  font-size: 12px !important;
  color: #0f172a !important;
}
.city-province-dropdown input {
  height: 44px;

  border-radius: 12px;
  font-family: Inter, ui-sans-serif, system-ui, sans-serif;
  font-size: 13px;
  font-weight: 500;
}
.city-province-dropdown > div:last-child {
  max-height: 232px !important;
  scrollbar-color: #8b8b8b transparent;
  scrollbar-width: thin;
}
.city-province-option {
  display: flex;
  width: 100%;
  align-items: flex-start;
  gap: 12px;
  min-height: 0;
  border: 0;
  border-radius: 14px;
  background: #fff;
  padding: 10px 12px;
  font-family: Inter, ui-sans-serif, system-ui, sans-serif;
  font-size: 12px;
  text-align: left;
  color: #334155;
  transition: background-color 150ms ease;
}
.city-province-option:hover,
.city-province-option.active {
  background: #f8fafc;
  color: #0f172a;
}
.city-province-check {
  display: flex;
  width: 18px;
  height: 18px;
  flex-shrink: 0;
  align-items: center;
  justify-content: center;
  margin-top: 2px;
  border: 1px solid #cbd5e1;
  border-radius: 5px;
  background: #fff;
}
.city-province-check.active {
  border-color: #ef4444;
  background: #ef4444;
}
.city-province-option span {
  line-height: 16px;
}
.city-province-option .block {
  overflow: visible;
  white-space: normal;
}
.sidebar-actions :deep(.sidebar-create-button) {
  border-color: #dc2626 !important;
  background: #dc2626 !important;
  background-color: #dc2626 !important;
  color: #fff !important;
  box-shadow: 0 4px 16px rgba(220, 38, 38, .22) !important;
}
.sidebar-actions :deep(.sidebar-create-button:not(:disabled):hover) {
  border-color: #b91c1c !important;
  background: #b91c1c !important;
  background-color: #b91c1c !important;
}
.sidebar-actions :deep(.sidebar-create-button:disabled) {
  border-color: #dc2626 !important;
  background: #dc2626 !important;
  background-color: #dc2626 !important;
  color: #fff !important;
  box-shadow: 0 4px 16px rgba(220, 38, 38, .22) !important;
}
@media (min-width: 640px) {
  .create-customer-header {
    padding-right: 24px !important;
    padding-left: 24px !important;
  }
}

.admin-page header.sticky > div > div:first-child > button {
  min-height: 0;
  border: 0;
  border-radius: 0;
  background: transparent;
  padding: 0;
  box-shadow: none;
}

.admin-page {
  display: flex;
  flex-direction: column;
  gap: 1.25rem;
  padding: 1.75rem 2rem;
  min-height: 100vh;
}

/* â”€â”€ PAGE HEADER â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€ */
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

/* â”€â”€ FORM LAYOUT â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€ */
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

/* â”€â”€ FORM CARDS â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€ */
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

/* â”€â”€ FORM GRID â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€ */
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

/* â”€â”€ CONTACTS â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€ */
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

/* â”€â”€ SIDEBAR â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€ */
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

/* â”€â”€ SUCCESS PANEL â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€ */
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

/* â”€â”€ RESPONSIVE â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€ */
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

.admin-page {
  gap: 0 !important;
  padding: 1.75rem 2rem 0 !important;
  background: #f8fafc !important;
}
.form-layout {
  width: min(100%, 1180px) !important;
  margin: 0 auto !important;
  display: grid !important;
  grid-template-columns: minmax(0, 1fr) 270px !important;

  gap: 18px !important;
  align-items: start !important;
  padding: 20px 16px !important;
}
.form-stack {
  gap: 18px !important;
  display: flex !important;
  flex-direction: column !important;
}
.contacts-information-section {
  order: 9;
}
.contacts-information-section > .grid > div > div:first-child {
  min-height: 88px !important;
  box-sizing: border-box !important;
  display: flex !important;
  align-items: center !important;
}
.tax-information-section {
  order: 10;
}
.other-master-data-section {
  order: 11;
}
.billing-shipment-section {
  order: 12;
}
.sales-assignment-section {
  order: 13;
}
.sales-assignment-section .sales-assignment-select {
  width: 100% !important;
  min-height: 48px !important;
  height: 48px !important;
  border: 1px solid #d9e2ec !important;
  border-radius: 14px !important;
  background: #fff !important;
  box-shadow: 0 1px 2px rgba(15, 23, 42, 0.04) !important;
}
.sales-assignment-section .sales-assignment-select :deep(.p-select-label) {
  display: flex !important;
  align-items: center !important;
  height: 46px !important;
  line-height: 1 !important;
  padding: 0 14px !important;
  font-family: Inter, ui-sans-serif, system-ui, sans-serif !important;
  font-size: 13px !important;
  font-weight: 500 !important;
  color: #0f172a !important;
}
.sales-assignment-section .sales-assignment-select :deep(.p-placeholder) { color: #94a3b8 !important; }
.sales-assignment-section .sales-assignment-select :deep(.p-select-dropdown) { width: 40px !important; color: #64748b !important; }
.master-data-select {
  width: 100% !important;
  min-height: 48px !important;
  height: 48px !important;
  border: 1px solid #d9e2ec !important;
  border-radius: 14px !important;
  background: #fff !important;
}
.master-data-select :deep(.p-select-label) {
  display: flex !important;
  align-items: center !important;
  height: 46px !important;
  line-height: 1 !important;
  padding: 0 14px !important;
  font-family: Inter, ui-sans-serif, system-ui, sans-serif !important;
  font-size: 13px !important;
  font-weight: 500 !important;
  color: #0f172a !important;
  text-transform: none !important;
  letter-spacing: normal !important;
}
.master-data-select :deep(.p-placeholder) { color: #94a3b8 !important; }
.master-data-select :deep(.p-select-dropdown) { width: 40px !important; color: #64748b !important; }
.invoice-type-select :deep(.p-select-label) { font-weight: 600 !important; }
.invoice-option-row { display: flex; align-items: center; gap: 12px; width: 100%; font: 500 13px Inter, sans-serif; color: #334155; }
.invoice-option-check { display: inline-flex; width: 18px; height: 18px; flex: 0 0 18px; align-items: center; justify-content: center; border: 1px solid #cbd5e1; border-radius: 5px; background: #fff; color: #fff; font-size: 11px; }
.invoice-option-check.checked { border-color: #ef4444; background: #ef4444; }
.master-data-select:not(.p-disabled):hover { border-color: #cbd5e1 !important; }
.master-data-select:not(.p-disabled).p-focus,
.master-data-select:not(.p-disabled):focus-within {
  border-color: #94a3b8 !important;
  box-shadow: 0 0 0 1px #cbd5e1 !important;
  outline: none !important;
}
.master-data-select .p-select-overlay {
  border: 1px solid #e2e8f0 !important;
  border-radius: 14px !important;
  box-shadow: 0 14px 32px rgba(15, 23, 42, 0.12) !important;
}
.master-data-multiselect {
  width: 100% !important;
  min-height: 48px !important;
  border: 1px solid #d9e2ec !important;
  border-radius: 14px !important;
  background: #fff !important;
}
.master-data-multiselect .p-multiselect-label {
  display: flex !important;

  align-items: center !important;
  min-height: 46px !important;
  padding: 0 14px !important;
  font: 500 13px Inter, sans-serif !important;
  color: #0f172a !important;
  text-transform: none !important;
  letter-spacing: normal !important;
}
.master-data-multiselect .p-placeholder { color: #94a3b8 !important; }
.master-data-multiselect .p-multiselect-dropdown { width: 40px !important; color: #64748b !important; }
.bank-selected-count { border-radius: 999px; background: #fff1f2; padding: 4px 10px; color: #dc2626; font: 600 12px Inter, sans-serif; }
.bank-option-row { display: flex; align-items: flex-start; gap: 10px; width: 100%; font: 500 12px Inter, sans-serif; color: #334155; }
.bank-option-check { width: 18px; height: 18px; flex: 0 0 18px; border: 1px solid #cbd5e1; border-radius: 5px; background: #fff; }
.bank-option-check.checked { border-color: #2563eb; background: #2563eb; color: #fff; display: inline-flex; align-items: center; justify-content: center; font-size: 11px; }
.bank-option-title { display: block; font-weight: 600; }
.bank-option-subtitle { display: block; margin-top: 2px; color: #94a3b8; font-size: 10px; }
.sales-assignment-section .rounded-\[18px\] > .mt-\[12px\] {
  display: grid !important;
  grid-template-columns: minmax(280px, 1.9fr) minmax(150px, .9fr) minmax(150px, .9fr) !important;
  gap: 12px !important;
}
.sales-assignment-section .rounded-\[18px\] > .mt-\[12px\] > label:first-child {
  grid-column: span 1;
}
.sales-assignment-section .rounded-\[18px\] > .mt-\[12px\] > label:last-child {
  grid-column: 1 / span 1 !important;
}
.sales-assignment-section .rounded-\[18px\] > .mt-\[12px\] > label {
  min-width: 0;
}
.sales-assignment-section .rounded-\[18px\] {
  border-color: #e2e8f0 !important;
  border-radius: 18px !important;
  background: #fff !important;
  padding: 14px !important;
  box-shadow: 0 1px 2px rgba(15, 23, 42, .04) !important;
}
.sales-assignment-section .rounded-\[18px\] > .flex:first-child p:first-child {
  font-family: Inter, ui-sans-serif, system-ui, sans-serif !important;
  font-size: 12px !important;
  font-weight: 700 !important;
  letter-spacing: .08em !important;
  text-transform: uppercase !important;
  color: #94a3b8 !important;
}
.sales-assignment-section .rounded-\[18px\] > .flex:first-child p:last-child {
  margin-top: 4px !important;
  font-size: 13px !important;
  font-weight: 600 !important;
  color: #0f172a !important;
}
.sales-assignment-section .contact-input-label {
  font-family: Inter, ui-sans-serif, system-ui, sans-serif !important;
  font-size: 12px !important;
  font-weight: 600 !important;
  letter-spacing: normal !important;
  text-transform: none !important;
  color: #334155 !important;
}
.sales-assignment-section .contact-reference-input,
.sales-assignment-section select {
  box-sizing: border-box !important;
  height: 48px !important;
  min-height: 48px !important;
  width: 100% !important;
  border: 1px solid #d9e2ec !important;
  border-radius: 14px !important;
  background: #fff !important;
  padding: 0 14px !important;
  font-family: Inter, ui-sans-serif, system-ui, sans-serif !important;
  font-size: 13px !important;
  font-weight: 500 !important;
  color: #0f172a !important;
}
.sales-assignment-section .md\:col-span-2 {
  max-width: none !important;
}
.sales-assignment-section > div:nth-child(2) > div:first-child {
  max-width: 560px !important;
}
.sales-assignment-section > div:nth-child(2) {
  padding: 20px !important;
}
.sales-assignment-section > div:nth-child(2) > button:last-child {
  height: 40px !important;
  gap: 7px !important;
  border: 1px solid #dbeafe !important;
  border-radius: 10px !important;
  background: #fff !important;
  padding: 0 14px !important;
  font-family: Inter, ui-sans-serif, system-ui, sans-serif !important;
  font-size: 12px !important;
  font-weight: 600 !important;
  color: #1d4ed8 !important;
  transition: background-color .2s;
}
.sales-assignment-section > div:nth-child(2) > button:last-child:hover {
  background: #eff6ff !important;
}
.sales-assignment-section > div:nth-child(2) > button:last-child i {

  font-size: 14px !important;
}
@media (max-width: 900px) {
  .sales-assignment-section .rounded-\[18px\] > .mt-\[12px\] {
    grid-template-columns: 1fr 1fr !important;
  }
  .sales-assignment-section .rounded-\[18px\] > .mt-\[12px\] > label:first-child,
  .sales-assignment-section .rounded-\[18px\] > .mt-\[12px\] > label:last-child {
    grid-column: span 2 !important;
  }
}
.billing-shipment-section .mt-\[18px\] {
  gap: 14px !important;
}
.billing-shipment-section .mt-\[18px\] > div {
  overflow: hidden;
  border: 1px solid #dbe3ef;
  border-radius: 16px;
  background: #fff;
  box-shadow: 0 8px 24px rgba(15, 23, 42, .04);
}
.billing-shipment-section .mt-\[18px\] > div > div:first-child {
  border-bottom: 1px solid #e2e8f0;
  background: #f8fafc;
}
.billing-shipment-section .mt-\[18px\] > div > div:first-child > p:first-child {
  position: relative;
  margin: 0;
  font-size: 10px;
  font-weight: 700;
  letter-spacing: .08em;
  text-transform: uppercase;
  color: #94a3b8;
}
.billing-shipment-section .mt-\[18px\] > div > div:first-child > p:first-child::after {
  content: 'âŒ„';
  position: absolute;
  right: 0;
  color: #dc2626;
  font-size: 16px;
  line-height: 10px;
}
.billing-shipment-section .mt-\[18px\] > div > div:first-child > div {
  display: flex;
  width: 100%;
  flex-wrap: wrap;
  gap: 6px;
  border: 1px solid #fecaca;
  border-radius: 12px;
  background: #fff7f7;
  padding: 4px;
}
.billing-shipment-section .mt-\[18px\] > div > div:first-child > div button {
  height: 30px;
  border-radius: 9px;
  padding: 0 10px;
  font-family: Inter, ui-sans-serif, system-ui, sans-serif;
  font-size: 11px;
  font-weight: 700;
}
.billing-shipment-section .mt-\[18px\] > div > div:first-child > p:last-child {
  padding-top: 2px;
  font-size: 12px;
  font-weight: 900;
  letter-spacing: .08em;
}
.billing-shipment-section .mt-\[18px\] > div:nth-child(2) > div:first-child > div button:last-child {
  display: none !important;
}
.delivery-field { width:100%; border:1px solid #fecaca !important; border-radius:12px !important; background:#fff7f7 !important; padding:0 12px !important; font:500 13px Inter, sans-serif !important; color:#0f172a !important; }
.delivery-field.p-inputtext { height:40px !important; }
.delivery-textarea { min-height:90px !important; padding:12px !important; resize:vertical; }
.billing-delivery-fields { border-color: #dbe3ef !important; background: #fff !important; }
.billing-delivery-fields .delivery-field { height: 42px !important; min-height: 42px !important; border-radius: 12px !important; background: #fffafa !important; padding: 0 12px !important; font: 600 12px Inter, sans-serif !important; }
.billing-delivery-fields .delivery-textarea { height: auto !important; min-height: 90px !important; padding: 10px 12px !important; line-height: 19px !important; }
.billing-delivery-fields label { gap: 6px !important; font-size: 11px !important; letter-spacing: .08em !important; text-transform: uppercase !important; color: #64748b !important; }
.other-master-data-section label > .relative:has(.contact-reference-input) {
  min-height: 48px;
  height: 48px;
  width: 100%;
  overflow: hidden;
  border: 1px solid #d9e2ec;
  border-radius: 14px;
  background: #fff;
  box-shadow: 0 1px 2px rgba(15, 23, 42, .04);
  transition: border-color .2s, box-shadow .2s;
}
.other-master-data-section label > .relative:has(.contact-reference-input):focus-within {
  border-color: #94a3b8;
  box-shadow: 0 0 0 2px #dbeafe;
}
.other-master-data-section label > .relative > .contact-reference-input {
  box-sizing: border-box !important;
  height: 46px !important;
  border: 0 !important;
  border-radius: 0 !important;
  box-shadow: none !important;
  padding-left: 60px !important;
  padding-right: 124px !important;
}

.other-master-data-section label > .relative > span:first-of-type {
  z-index: 1;
  height: 48px;
  display: flex;
  align-items: center;
  border-right: 1px solid #d9e2ec;
  background: #f8fafc;
  padding: 0 14px;
  font-family: Inter, ui-sans-serif, system-ui, sans-serif;
  font-size: 13px;
  font-weight: 700;
  color: #64748b;
}
.other-master-data-section .contact-input-label {
  gap: 8px !important;
  font-size: 11px !important;
  font-weight: 700 !important;
  letter-spacing: .08em !important;
  text-transform: uppercase !important;
  color: #64748b !important;
}
.other-master-data-section .contact-input-label .contact-reference-input,
.other-master-data-section .contact-input-label select {
  text-transform: none !important;
  letter-spacing: normal !important;
}
.bank-account-trigger {
  display: flex;
  width: 100%;
  height: 48px;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
  border: 1px solid #d9e2ec;
  border-radius: 14px;
  background: #fff;
  padding: 0 14px;
  font-family: Inter, ui-sans-serif, system-ui, sans-serif;
  font-size: 13px;
  font-weight: 500;
  color: #94a3b8;
  text-align: left;
  box-shadow: 0 1px 2px rgba(15, 23, 42, .04);
}
.bank-account-trigger:hover,
.bank-account-trigger:focus {
  border-color: #94a3b8;
  outline: none;
  box-shadow: 0 0 0 2px #dbeafe;
}
.other-master-data-section label:has(.bank-account-dropdown) .bank-account-trigger {
  border-color: #ef4444;
  box-shadow: 0 0 0 1px #ef4444;
}
.bank-account-trigger i { color: #64748b; transition: transform .2s; }
.other-master-data-section label:has(.bank-account-trigger) { position: relative; }
.bank-account-dropdown { position: absolute; z-index: 30; top: calc(100% + 6px); left: 0; width: min(360px, 100%); max-height: 300px; overflow: hidden; border: 1px solid #e2e8f0; border-radius: 14px; background: #fff; padding: 10px; box-shadow: 0 14px 32px rgba(15,23,42,.12); }
.bank-account-search { height: 40px; width: 100%; border: 1px solid #d9e2ec; border-radius: 12px; padding: 0 12px 0 36px; font-family: Inter, ui-sans-serif, system-ui, sans-serif; font-size: 12px; outline: none; }
.bank-account-search:focus { border-color: #dc2626; box-shadow: 0 0 0 1px #dc2626; }
.bank-account-search-icon { position: absolute; left: 12px; top: 12px; color: #94a3b8; }
.bank-account-empty { max-height: 176px; overflow-y: auto; padding: 18px 8px 10px; font-family: Inter, ui-sans-serif, system-ui, sans-serif; font-size: 12px; color: #94a3b8; text-align: center; }
.invoice-type-trigger { display:flex; width:100%; height:48px; align-items:center; justify-content:space-between; padding:0 14px; border:1px solid #d9e2ec; border-radius:14px; background:#fff; font:600 13px Inter, sans-serif; color:#0f172a; text-align:left; }
.invoice-type-trigger i { color:#64748b; transition:transform .2s; }
.other-master-data-section label:has(.invoice-type-trigger) { position:relative; }
.invoice-type-dropdown { position:absolute; z-index:30; top:calc(100% + 6px); left:0; width:100%; overflow:hidden; border:1px solid #e2e8f0; border-radius:14px; background:#fff; padding:12px; box-shadow:0 14px 32px rgba(15,23,42,.12); }
.invoice-type-search { height:40px; width:100%; border:1px solid #d9e2ec; border-radius:12px; padding:0 38px 0 12px; font:12px Inter, sans-serif; outline:none; }
.invoice-type-search-icon { position:absolute; right:12px; top:12px; color:#94a3b8; }
.invoice-type-options { max-height:176px; overflow-y:auto; padding:6px 0; }
.invoice-type-option { display:flex; min-height:44px; width:100%; align-items:center; gap:12px; padding:10px 12px; border:0; background:#fff; font:600 13px Inter, sans-serif; color:#334155; text-align:left; }
.invoice-type-option:hover, .invoice-type-option.active { background:#f8fafc; color:#0f172a; }
.invoice-type-check { display:flex; width:18px; height:18px; flex-shrink:0; align-items:center; justify-content:center; border:1px solid #cbd5e1; border-radius:5px; color:#fff; }
.invoice-type-option.active .invoice-type-check { border-color:#ef4444; background:#ef4444; }
.invoice-type-check i { font-size:12px; }
.create-form-heading {
  margin-bottom: 0 !important;
}
.create-form-heading h2 {
  margin: 0 !important;
  font-family: Inter, ui-sans-serif, system-ui, sans-serif !important;
  font-size: 24px !important;
  font-weight: 700 !important;
  line-height: 32px !important;
  letter-spacing: -0.025em !important;
  color: #0f172a !important;
}
.create-form-heading p {
  margin-top: 5px !important;
  max-width: 760px !important;
  font-family: Inter, ui-sans-serif, system-ui, sans-serif !important;
  font-size: 13px !important;
  line-height: 20px !important;
  color: #64748b !important;
}
.form-sidebar {
  top: 78px !important;
  gap: 18px !important;
}
.sidebar-card {
  border: 1px solid #e2e8f0 !important;
  border-radius: 20px !important;

  background: #fff !important;
  padding: 16px !important;
  box-shadow: 0 14px 34px rgba(15, 23, 42, .06) !important;
}
.field-scope-card h4 {
  margin: 0 !important;
  font-family: Inter, ui-sans-serif, system-ui, sans-serif !important;
  font-size: 16px !important;
  font-weight: 700 !important;
  line-height: 22px !important;
  letter-spacing: 0 !important;
  text-transform: none !important;
  color: #0f172a !important;
}
.scope-description {
  margin-top: 4px !important;
  font-family: Inter, ui-sans-serif, system-ui, sans-serif !important;
  font-size: 12px !important;
  line-height: 18px !important;
  color: #64748b !important;
}
.scope-legend {
  margin-top: 10px !important;
  margin-bottom: 14px !important;
}
@media (min-width: 640px) {
  .form-layout {
    padding-right: 24px !important;
    padding-left: 24px !important;
  }
}
@media (max-width: 1024px) {
  .form-layout {
    grid-template-columns: 1fr !important;
  }
  .form-sidebar {
    position: static !important;
  }
}
.admin-page .customer-info-select {
  display: flex !important;
  width: 100% !important;
  height: 48px !important;
  min-height: 48px !important;
  align-items: center !important;
  justify-content: space-between !important;
  gap: 10px !important;
  border: 1px solid #d9e2ec !important;
  border-radius: 14px !important;
  background: #fff !important;
  padding: 0 !important;
  text-align: left !important;
  box-shadow: 0 1px 2px rgba(15, 23, 42, .04) !important;
  outline: none !important;
  transition: all 150ms ease !important;
}
.admin-page .customer-info-select:hover {
  border-color: #cbd5e1 !important;
}
.admin-page .customer-info-select.p-select-open,
.admin-page .customer-info-select.p-focus {
  border-color: #94a3b8 !important;
  box-shadow: 0 0 0 2px #dbeafe !important;
}
.admin-page .customer-info-select :deep(.p-select-label) {
  display: flex !important;
  min-width: 0 !important;
  flex: 1 1 auto !important;
  height: 48px !important;
  align-items: center !important;
  padding: 0 0 0 14px !important;
  overflow: hidden !important;
  color: #0f172a !important;
  font-family: Inter, ui-sans-serif, system-ui, sans-serif !important;
  font-size: 13px !important;
  font-weight: 500 !important;
  line-height: 16px !important;
  letter-spacing: 0 !important;
  text-overflow: ellipsis !important;
  white-space: nowrap !important;
}
.admin-page .customer-info-select :deep(.p-select-label.p-placeholder),
.admin-page .customer-info-select :deep(.p-placeholder) {
  color: #94a3b8 !important;
  font-family: Inter, ui-sans-serif, system-ui, sans-serif !important;
  font-size: 13px !important;
  font-weight: 500 !important;
  line-height: 16px !important;
  letter-spacing: 0 !important;
}
.admin-page .customer-info-select :deep(.p-select-dropdown) {
  width: 42px !important;
  min-width: 42px !important;
  color: #64748b !important;
}
.admin-page .customer-info-select :deep(.p-select-dropdown-icon) {
  width: 16px !important;
  height: 16px !important;
  color: #64748b !important;
  transition: transform 150ms ease !important;

}
.admin-page .customer-info-select.p-select-open :deep(.p-select-dropdown-icon) {
  transform: rotate(180deg) !important;
}
.customer-info-option-check {
  display: flex;
  width: 18px;
  height: 18px;
  flex-shrink: 0;
  align-items: center;
  justify-content: center;
  border: 1px solid #cbd5e1;
  border-radius: 5px;
  background: #fff;
  color: transparent;
}
.customer-info-option-check svg {
  width: 12px;
  height: 12px;
}
.customer-info-option-check.active {
  border-color: #ef4444 !important;
  background: #ef4444 !important;
  color: #fff !important;
}
:global(.p-select-overlay) {
  z-index: 9999 !important;
  overflow: hidden !important;
  border: 1px solid #e2e8f0 !important;
  border-radius: 14px !important;
  background: #fff !important;
  box-shadow: 0 14px 32px rgba(15, 23, 42, .12) !important;
}
:global(.p-select-overlay .p-select-header) {
  border-bottom: 1px solid #eef2f7 !important;
  padding: 12px !important;
  background: #fff !important;
}
:global(.p-select-overlay .p-select-filter) {
  width: 100% !important;
  height: 40px !important;
  border: 1px solid #d9e2ec !important;
  border-radius: 12px !important;
  background: #fff !important;
  padding: 0 38px 0 12px !important;
  color: #0f172a !important;
  font-family: Inter, ui-sans-serif, system-ui, sans-serif !important;
  font-size: 12px !important;
  outline: none !important;
}
:global(.p-select-overlay .p-select-filter::placeholder) {
  color: #94a3b8 !important;
}
:global(.p-select-overlay .p-select-filter:focus) {
  border-color: #dc2626 !important;
  box-shadow: 0 0 0 1px #dc2626 !important;
}
:global(.p-select-overlay .p-select-filter-icon) {
  right: 12px !important;
  width: 16px !important;
  height: 16px !important;
  color: #94a3b8 !important;
}
:global(.p-select-overlay .p-select-list-container) {
  max-height: 132px !important;
  overscroll-behavior: contain !important;
  scrollbar-color: #8b8b8b transparent;
  scrollbar-width: thin;
}
:global(.p-select-overlay .p-select-list) {
  padding: 6px 0 !important;
}
:global(.p-select-overlay .p-select-option) {
  display: flex !important;
  min-height: 44px !important;
  width: 100% !important;
  align-items: center !important;
  gap: 12px !important;
  padding: 10px 12px !important;
  color: #334155 !important;
  font-family: Inter, ui-sans-serif, system-ui, sans-serif !important;
  font-size: 12px !important;
  text-align: left !important;
  transition: background-color 150ms ease !important;
}
:global(.p-select-overlay .p-select-option:hover),
:global(.p-select-overlay .p-select-option.p-focus) {
  background: #f8fafc !important;
}
:global(.p-select-overlay .p-select-option.p-select-option-selected) {
  background: #f8fafc !important;
  color: #334155 !important;
}
:global(.p-select-overlay .p-select-option span span) {
  color: #334155 !important;
  font-family: Inter, ui-sans-serif, system-ui, sans-serif !important;
  font-size: 13px !important;
  font-weight: 600 !important;
  line-height: 18px !important;
}

</style>
