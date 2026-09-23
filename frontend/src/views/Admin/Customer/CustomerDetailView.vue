<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import Button from 'primevue/button'
import Message from 'primevue/message'
import Tag from 'primevue/tag'
import Dialog from 'primevue/dialog'
import { useToast } from 'primevue/usetoast'
import { useCrmStore } from '../../../stores/crm'
import { deleteCustomer, getAdminCustomerPlaceDetails, getProspectInitialAnalysis } from '../../../api/crm'
import type { CustomerDetail, PlaceDetails, ProspectInitialAnalysis, ProspectMenuDocument, ProspectMenuFinding, ProspectMenuProfile } from '../../../types/crm'
import { useAuthStore } from '../../../stores/auth'
import PlacePhotoGallery from '../../../components/PlacePhotoGallery.vue'
import ProspectComments from '../../../components/ProspectComments.vue'
import AISummaryCard from '../../../components/prospect-ai/AISummaryCard.vue'
import CustomerProductOpportunityCard from '../../../components/customer-ai/CustomerProductOpportunityCard.vue'
import CustomerSnapshotCard from '../../../components/customer-ai/CustomerSnapshotCard.vue'
import TanyaAICard from '../../../components/prospect-ai/TanyaAICard.vue'
import { priceLevelLabel, businessStatusLabel, businessStatusSeverity, stars, utcOffsetLabel } from '../../../utils/placeLabels'

const fieldSources = {
  customerCode: { source: 'system' as const, tooltip: 'Generated automatically by the system during prospect conversion.' },
  siteName: { source: 'google' as const, tooltip: 'Retrieved from Google Places. Can still be edited by administrators.' },
  segment: { source: 'manual' as const, tooltip: 'Selected manually by the administrator.' },
  category: { source: 'google' as const, tooltip: 'Retrieved from Google Places and may be adjusted manually.' },
  region: { source: 'system' as const, tooltip: 'Derived from the selected location.' },
  salesExec: { source: 'manual' as const, tooltip: 'Assigned manually by administrator.' },
  convertedAt: { source: 'system' as const, tooltip: 'Generated automatically when the prospect is converted.' },
  lastUpdated: { source: 'system' as const, tooltip: 'Automatically updated whenever the customer data changes.' },
  companyName: { source: 'system' as const, tooltip: 'Linked to the parent company.' },
  companyCode: { source: 'system' as const, tooltip: 'Generated automatically by the system.' },
  sourceProspect: { source: 'system' as const, tooltip: 'Automatically references the original prospect.' },
}

function getFieldSource(key: string) {
  return fieldSources[key as keyof typeof fieldSources] ?? null
}

const fsLabels: Record<string, string> = { google: 'GOOGLE', manual: 'MANUAL', system: 'SYSTEM' }

const route = useRoute()
const router = useRouter()
const crm = useCrmStore()
const auth = useAuthStore()
const toast = useToast()
const error = ref('')
const detail = ref<CustomerDetail | null>(null)
const customerExtra = computed(() => (detail.value?.customer || {}) as Record<string, any>)
const placeDetails = ref<PlaceDetails | null>(null)
const initialAnalysis = ref<ProspectInitialAnalysis | null>(null)
const sourceProspectId = computed(() => detail.value?.customer.sourceProspectId || '')
const storedMenu = computed(() => initialAnalysis.value?.menu as ProspectMenuDocument | null | undefined)
const storedDiscovery = computed(() => (storedMenu.value?.discovery ?? storedMenu.value?.finding ?? null) as ProspectMenuFinding | null)
const storedProfiling = computed(() => (storedMenu.value?.profiling ?? storedMenu.value?.profile ?? null) as ProspectMenuProfile | null)
const activeTab = ref('overview')
const showAllHours = ref(false)
const deleteDialogVisible = ref(false)
const deleting = ref(false)

const siteGoogleMapsUrl = computed(() => {
  if (placeDetails.value?.googleMapsUrl) return placeDetails.value.googleMapsUrl
  const lat = detail.value?.customer.address?.latitude
  const lng = detail.value?.customer.address?.longitude
  if (lat != null && lng != null) return `https://www.google.com/maps/search/?api=1&query=${lat},${lng}`
  if (detail.value?.customer.address?.previewAddress) return `https://www.google.com/maps/search/?api=1&query=${encodeURIComponent(detail.value.customer.address.previewAddress)}`
  return ''
})

const tabs = [
  { key: 'overview', label: 'Overview', icon: 'pi pi-id-card' },
  { key: 'google', label: 'Google Maps', icon: 'pi pi-map' },
  { key: 'contacts', label: 'Contacts', icon: 'pi pi-users' },
  { key: 'company', label: 'Company', icon: 'pi pi-building' },
  { key: 'address', label: 'Address', icon: 'pi pi-map-marker' },
]

function formatDate(dateStr: string) {
  if (!dateStr) return '—'
  return new Date(dateStr).toLocaleDateString('en-GB', { day: '2-digit', month: 'short', year: 'numeric' })
}

function formatDateTime(dateStr: string) {
  if (!dateStr) return '—'
  return new Date(dateStr).toLocaleDateString('en-GB', { day: '2-digit', month: 'short', year: 'numeric', hour: '2-digit', minute: '2-digit' })
}

function segmentSeverity(seg: string) {
  switch (seg) {
    case 'Key Account': return 'warn'
    case 'Modern Trade': return 'info'
    case 'Food Service': return 'success'
    default: return 'secondary'
  }
}

onMounted(async () => {
  try {
    const customerId = String(route.params.id)
    const [cust, place] = await Promise.all([
      crm.loadAdminCustomer(customerId),
      getAdminCustomerPlaceDetails(customerId).catch(() => null),
    ])
    detail.value = cust
    placeDetails.value = place
    if (cust.customer.sourceProspectId && (auth.hasPermission('view_ai_summary') || auth.hasPermission('view_ai_menu_profiling'))) {
      initialAnalysis.value = await getProspectInitialAnalysis(cust.customer.sourceProspectId).catch(() => null)
    }
  } catch (e) {
    error.value = crm.errorMessage(e)
  }
})

function confirmDelete() {
  deleteDialogVisible.value = true
}

async function executeDelete() {
  if (!detail.value) return
  deleting.value = true
  try {
    await deleteCustomer(detail.value.customer.id)
    toast.add({ severity: 'success', summary: 'Customer deleted', detail: `${detail.value.customer.name} has been removed.`, life: 3000 })
    router.push('/admin/customers')
  } catch (e) {
    error.value = crm.errorMessage(e)
    deleteDialogVisible.value = false
  } finally {
    deleting.value = false
  }
}
</script>

<template>
  <section class="admin-page">
    <Message v-if="error" severity="error" :closable="false">{{ error }}</Message>

    <!-- LOADING -->
    <div v-if="!detail && !error" class="state-box">
      <i class="pi pi-spin pi-spinner state-icon" />
      <span>Loading customer details...</span>
    </div>

    <template v-if="detail">
      <!-- PAGE HEADER -->
      <nav class="customer-detail-navbar">
        <button type="button" class="customer-detail-back" @click="router.push('/admin/customers')"><i class="pi pi-arrow-left" /> Back to Customer Site List</button>
        <span class="customer-detail-navbar-title">Customer Site Detail</span>
        <span class="customer-detail-breadcrumb">Customer List &gt; Customer Site &gt; Detail</span>
        <Button label="Edit Customer Site" icon="pi pi-pencil" size="small" @click="router.push(`/admin/customers/${route.params.id}/edit`)" />
      </nav>

      <!-- SUMMARY STRIP -->
      <div class="summary-strip customer-detail-card">
        <div class="strip-item">
          <i class="pi pi-tag" />
          <div>
            <span>Segment <span v-if="getFieldSource('segment')" class="fs-badge" :class="'fs-' + getFieldSource('segment')!.source" :title="getFieldSource('segment')!.tooltip">{{ fsLabels[getFieldSource('segment')!.source] }}</span></span>
            <Tag :value="detail.customer.segment" :severity="segmentSeverity(detail.customer.segment)" />
          </div>
        </div>
        <div class="strip-item">
          <i class="pi pi-folder" />
          <div>
            <span>Category <span v-if="getFieldSource('category')" class="fs-badge" :class="'fs-' + getFieldSource('category')!.source" :title="getFieldSource('category')!.tooltip">{{ fsLabels[getFieldSource('category')!.source] }}</span></span>
            <strong>{{ detail.customer.category || '—' }}</strong>
          </div>
        </div>
        <div class="strip-item">
          <i class="pi pi-map-marker" />
          <div>
            <span>Region <span v-if="getFieldSource('region')" class="fs-badge" :class="'fs-' + getFieldSource('region')!.source" :title="getFieldSource('region')!.tooltip">{{ fsLabels[getFieldSource('region')!.source] }}</span></span>
            <strong>{{ detail.customer.region || '—' }}</strong>
          </div>
        </div>
        <div class="strip-item">
          <i class="pi pi-user" />
          <div>
            <span>Sales Executive <span v-if="getFieldSource('salesExec')" class="fs-badge" :class="'fs-' + getFieldSource('salesExec')!.source" :title="getFieldSource('salesExec')!.tooltip">{{ fsLabels[getFieldSource('salesExec')!.source] }}</span></span>
            <strong>{{ detail.customer.salesExecutiveName || 'Unassigned' }}</strong>
          </div>
        </div>
        <div class="strip-item">
          <i class="pi pi-calendar" />
          <div>
            <span>Converted <span v-if="getFieldSource('convertedAt')" class="fs-badge" :class="'fs-' + getFieldSource('convertedAt')!.source" :title="getFieldSource('convertedAt')!.tooltip">{{ fsLabels[getFieldSource('convertedAt')!.source] }}</span></span>
            <strong>{{ formatDate(detail.customer.convertedAt) }}</strong>
          </div>
        </div>
      </div>

      <CustomerSnapshotCard class="customer-admin-snapshot" :customer="detail.customer" :operating-status="placeDetails?.businessStatus" />

      <section class="customer-detail-information-grid">
        <div>
        <div class="detail-card customer-information-card">
          <div class="detail-card-header"><h3>Customer Information</h3><p>Site identity for this outlet, branch, or store.</p></div>
          <div class="detail-list">
            <div><span>Customer Name / Outlet / Branch / Store</span><strong>{{ detail.customer.name }}</strong></div>
            <div><span>Site Code</span><strong>{{ detail.customer.customerCode }}</strong></div>
            <div><span>Region</span><strong>{{ detail.customer.region || '—' }}</strong></div>
            <div><span>Head Sales</span><strong>{{ detail.customer.salesExecutiveName || 'Unassigned' }}</strong></div>
            <div><span>Customer Segment</span><strong>{{ detail.customer.segment || '—' }}</strong></div>
            <div><span>Customer Category</span><strong>{{ detail.customer.category || '—' }}</strong></div>
          </div>
        </div>
        <div class="detail-card address-information-card">
          <div class="detail-card-header"><h3>Address Information</h3><p>Address for customer transactions and site-level documents.</p></div>
          <div class="detail-list">
            <div><span>Site Address</span><strong>{{ detail.customer.address?.previewAddress || '—' }}</strong></div>
            <div><span>City, Province</span><strong>{{ [detail.customer.address?.district, detail.customer.address?.province].filter(Boolean).join(', ') || '—' }}</strong></div>
            <div><span>Postal Code, District</span><strong>{{ [detail.customer.address?.subDistrict, detail.customer.address?.village].filter(Boolean).join(' - ') || '—' }}</strong></div>
            <div><span>Latitude</span><strong>{{ detail.customer.address?.latitude ?? '—' }}</strong></div>
            <div><span>Longitude</span><strong>{{ detail.customer.address?.longitude ?? '—' }}</strong></div>
          </div>
        </div>
        <div class="detail-card contact-information-card">
          <div class="detail-card-header"><h3>Contact Information</h3><p>General contacts for this customer site.</p></div>
          <div class="contact-list">
            <div v-for="(contact, index) in (detail.customer.contacts || [])" :key="`${contact.email || contact.phone || contact.name}-${index}`" class="contact-item">
              <div class="contact-item-heading"><div><span>CONTACT {{ index + 1 }}</span><strong>{{ contact.name || '—' }}</strong></div><em>{{ contact.position || 'Contact' }}</em></div>
              <div class="contact-item-fields"><div><span>PHONE NUMBER</span><strong>{{ contact.phone || '—' }}</strong></div><div><span>EMAIL ADDRESS</span><strong>{{ contact.email || '—' }}</strong></div></div>
            </div>
            <div v-if="!detail.customer.contacts?.length" class="empty-contact">No contact information available.</div>
          </div>
        </div>
        <div class="company-data-card">
          <div class="company-data-card-header"><div><p>Company Data</p><span>Data parent company untuk customer site ini.</span></div><div class="company-data-card-actions"><button type="button" @click="router.push(`/admin/companies/${detail.parentCompany.parentCode}`)">Open Company Detail <i class="pi pi-arrow-up-right" /></button><button type="button" @click="router.push(`/admin/companies/${detail.parentCompany.parentCode}/edit`)">Edit Company Detail <i class="pi pi-pencil" /></button></div><div class="company-data-card-title"><div><h3>{{ detail.parentCompany.name }}</h3><span>{{ detail.parentCompany.parentCode }}</span></div><em>Key Account</em></div></div>
          <div class="company-data-card-body"><div class="company-nested-card"><div class="detail-card-header"><h3>Company Information</h3><p>Legal identity of the parent company.</p></div><div class="detail-list"><div><span>Company Name</span><strong>{{ detail.parentCompany.name }}</strong></div><div><span>Company Code</span><strong>{{ detail.parentCompany.parentCode }}</strong></div><div><span>Company Tier</span><strong>Key Account</strong></div></div></div><div class="company-nested-card"><div class="detail-card-header"><h3>Tax Information</h3><p>Company NPWP data synced from company name and address.</p></div><div class="detail-list"><div><span>Company NPWP Name</span><strong>{{ detail.parentCompany.npwpName || detail.parentCompany.name }}</strong></div><div><span>Company NPWP Address</span><strong>{{ detail.parentCompany.npwpAddress || detail.parentCompany.address?.previewAddress || '—' }}</strong></div><div><span>Company NPWP Number</span><strong>{{ detail.parentCompany.npwpNumber || '—' }}</strong></div></div></div></div>
        </div>
        <CustomerProductOpportunityCard v-if="sourceProspectId && auth.hasPermission('view_ai_menu_profiling')" class="customer-product-opportunity-below" :discovery="storedDiscovery" :profiling="storedProfiling" />
        </div>
        <div class="customer-detail-side-information">
          <AISummaryCard v-if="sourceProspectId && auth.hasPermission('view_ai_summary')" class="customer-ai-summary-card" :prospect-name="detail.customer.name" :analysis="initialAnalysis" context="customer" />
          <ProspectComments v-if="sourceProspectId" class="customer-discuss-card" :prospect-id="sourceProspectId" role="ADMINISTRATOR" embedded />
          <TanyaAICard v-if="sourceProspectId && auth.hasPermission('use_prospect_ai_chat')" class="customer-tanya-ai-card" :prospect-id="sourceProspectId" />
          <div class="detail-card tax-information-card">
            <div class="detail-card-header"><h3>Tax Information</h3><p>Site-level tax identity and individual buyer identification.</p></div>
            <div class="detail-list">
              <div><span>PPN</span><strong>{{ customerExtra.ppn || '—' }}</strong></div>
              <div><span>ID TKU Number</span><strong>{{ customerExtra.idTkuNumber || '—' }}</strong></div>
              <div><span>NIK</span><strong>{{ customerExtra.nik || '—' }}</strong></div>
            </div>
          </div>
          <div class="detail-card other-master-data-card">
            <div class="detail-card-header"><h3>Other Master Data</h3><p>Site-level document defaults.</p></div>
            <div class="detail-list">
              <div><span>Term of Payment</span><strong>{{ customerExtra.termOfPayment || '—' }}</strong></div>
              <div><span>Shipment Cost</span><strong>Rp {{ customerExtra.shipmentCost || '0' }}</strong></div>
              <div><span>Invoice Type</span><strong>{{ customerExtra.invoiceType || '—' }}</strong></div>
              <div><span>Bank Account</span><strong>{{ customerExtra.bankAccount || '—' }}</strong></div>
            </div>
          </div>
          <div class="detail-card billing-shipment-card">
            <div class="detail-card-header"><h3>Billing &amp; Shipment Configuration</h3><p>Default document header sources for billing and delivery.</p></div>
            <div class="billing-shipment-list">
              <div class="billing-shipment-block"><div class="billing-shipment-label"><span>BILL TO</span><em>{{ customerExtra.billToSource || 'company' }}</em></div><div><span>Name</span><strong>{{ customerExtra.billToSource === 'site' ? detail.customer.name : detail.parentCompany.name }}</strong></div><div><span>Address</span><strong>{{ customerExtra.billingAddressPreview || (customerExtra.billToSource === 'site' ? detail.customer.address?.previewAddress : detail.parentCompany.address?.previewAddress) || '—' }}</strong></div></div>
              <div class="billing-shipment-block"><div class="billing-shipment-label"><span>SHIP TO</span><em>{{ customerExtra.shipToSource || 'site' }}</em></div><div><span>Name</span><strong>{{ customerExtra.shipToSource === 'company' ? detail.parentCompany.name : detail.customer.name }}</strong></div><div><span>Address</span><strong>{{ customerExtra.shippingAddressPreview || (customerExtra.shipToSource === 'company' ? detail.parentCompany.address?.previewAddress : detail.customer.address?.previewAddress) || '—' }}</strong></div></div>
            </div>
          </div>
          <section class="sales-assignment-card"><div class="detail-card-header"><h3>Sales Assignment</h3><p>Site-level sales ownership for this branch, outlet, or store.</p></div><div class="sales-assignment-body"><div v-if="customerExtra.assignedSales || detail.customer.salesExecutiveName" class="sales-assignment-timeline"><div class="sales-assignment-item"><div class="sales-assignment-dot"></div><div class="sales-assignment-entry"><div><strong>{{ customerExtra.assignedSales?.fullName || customerExtra.assignedSales?.name || detail.customer.salesExecutiveName }}</strong><em>Aktif</em></div><p>{{ customerExtra.assignedSales?.startMonth || '—' }} – sekarang</p></div></div></div><div v-else class="empty-contact">No sales assignment available.</div></div></section>
        </div>
      </section>

      <section class="company-data-card company-data-card-standalone">
        <div class="company-contact-card company-contact-card-inside"><div class="detail-card-header"><h3>Company Contact Information</h3><p>General company contacts for coordination and follow up.</p></div><div class="company-contact-grid"><div v-for="(contact, index) in (detail.parentCompany.contacts || [])" :key="`inside-${contact.email || contact.phone || contact.name}-${index}`" class="company-contact-item"><div class="company-contact-heading"><div><span>CONTACT {{ index + 1 }}</span><strong>{{ contact.name || '—' }}</strong></div><em>{{ contact.position || 'Contact' }}</em></div><div class="company-contact-fields"><div><span>PHONE NUMBER</span><strong>{{ contact.phone || '—' }}</strong></div><div><span>EMAIL ADDRESS</span><strong>{{ contact.email || '—' }}</strong></div></div></div><div v-if="!detail.parentCompany.contacts?.length" class="empty-contact">No company contact information available.</div></div></div>
        <div class="company-data-card-header"><div><p>Company Data</p><span>Data parent company untuk customer site ini.</span></div><div class="company-data-card-actions"><button type="button" @click="router.push(`/admin/companies/${detail.parentCompany.parentCode}`)">Open Company Detail <i class="pi pi-arrow-up-right" /></button><button type="button" @click="router.push(`/admin/companies/${detail.parentCompany.parentCode}/edit`)">Edit Company Detail <i class="pi pi-pencil" /></button></div><div class="company-data-card-title"><div><h3>{{ detail.parentCompany.name }}</h3><span>{{ detail.parentCompany.parentCode }}</span></div><em>Key Account</em></div></div>
        <div class="company-data-card-body"><div class="company-nested-card"><div class="detail-card-header"><h3>Company Information</h3><p>Legal identity of the parent company.</p></div><div class="detail-list"><div><span>Company Name</span><strong>{{ detail.parentCompany.name }}</strong></div><div><span>Company Code</span><strong>{{ detail.parentCompany.parentCode }}</strong></div><div><span>Company Tier</span><strong>Key Account</strong></div></div></div><div class="company-nested-card"><div class="detail-card-header"><h3>Tax Information</h3><p>Company NPWP data synced from company name and address.</p></div><div class="detail-list"><div><span>Company NPWP Name</span><strong>{{ detail.parentCompany.npwpName || detail.parentCompany.name }}</strong></div><div><span>Company NPWP Address</span><strong>{{ detail.parentCompany.npwpAddress || detail.parentCompany.address?.previewAddress || '—' }}</strong></div><div><span>Company NPWP Number</span><strong>{{ detail.parentCompany.npwpNumber || '—' }}</strong></div></div></div></div>
        <div class="company-address-card company-address-card-inline"><div class="detail-card-header"><h3>Company Address</h3><p>Legal and billing address used for company and tax records.</p></div><div class="detail-list"><div><span>Company Address</span><strong>{{ detail.parentCompany.address?.previewAddress || '—' }}</strong></div><div><span>City, Province</span><strong>{{ [detail.parentCompany.address?.district, detail.parentCompany.address?.province].filter(Boolean).join(', ') || '—' }}</strong></div><div><span>Postal Code, District</span><strong>{{ [detail.parentCompany.address?.subDistrict, detail.parentCompany.address?.village].filter(Boolean).join(' - ') || '—' }}</strong></div><div><span>Latitude</span><strong>{{ detail.parentCompany.address?.latitude ?? '—' }}</strong></div><div><span>Longitude</span><strong>{{ detail.parentCompany.address?.longitude ?? '—' }}</strong></div></div></div>
      <section class="company-kam-card company-kam-card-inline"><div class="detail-card-header"><h3>Company KAM</h3><p>Company-level key account manager ownership.</p></div><div class="company-kam-body"><div v-if="detail.parentCompany.kamAssignments?.length" class="company-kam-timeline"><div v-for="(assignment, index) in detail.parentCompany.kamAssignments" :key="`${assignment.userId || assignment.id || index}`" class="company-kam-item"><div class="company-kam-dot"></div><div class="company-kam-entry"><div><strong>{{ assignment.userName || assignment.fullName || 'Unassigned' }}</strong><em>{{ assignment.isActive === false ? 'Tidak Aktif' : 'Aktif' }}</em></div><p>{{ assignment.startMonth || '—' }} – {{ assignment.end || 'sekarang' }}</p></div></div></div><div v-else class="empty-contact">No company KAM assignment available.</div></div></section>
      </section>

      <section class="company-kam-card company-kam-card-standalone"><div class="detail-card-header"><h3>Company KAM</h3><p>Company-level key account manager ownership.</p></div><div class="company-kam-body"><div v-if="detail.parentCompany.kamAssignments?.length" class="company-kam-timeline"><div v-for="(assignment, index) in detail.parentCompany.kamAssignments" :key="`${assignment.userId || assignment.id || index}`" class="company-kam-item"><div class="company-kam-dot"></div><div class="company-kam-entry"><div><strong>{{ assignment.userName || assignment.fullName || 'Unassigned' }}</strong><em>{{ assignment.isActive === false ? 'Tidak Aktif' : 'Aktif' }}</em></div><p>{{ assignment.startMonth || '—' }} – {{ assignment.end || 'sekarang' }}</p></div></div></div><div v-else class="empty-contact">No company KAM assignment available.</div></div></section>
      <section class="company-contact-card"><div class="detail-card-header"><h3>Company Contact Information</h3><p>General company contacts for coordination and follow up.</p></div><div class="company-contact-grid"><div v-for="(contact, index) in (detail.parentCompany.contacts || [])" :key="`${contact.email || contact.phone || contact.name}-${index}`" class="company-contact-item"><div class="company-contact-heading"><div><span>CONTACT {{ index + 1 }}</span><strong>{{ contact.name || '—' }}</strong></div><em>{{ contact.position || 'Contact' }}</em></div><div class="company-contact-fields"><div><span>PHONE NUMBER</span><strong>{{ contact.phone || '—' }}</strong></div><div><span>EMAIL ADDRESS</span><strong>{{ contact.email || '—' }}</strong></div></div></div><div v-if="!detail.parentCompany.contacts?.length" class="empty-contact">No company contact information available.</div></div></section>

      <section v-if="sourceProspectId" class="customer-intelligence">
      </section>
      <section v-else class="detail-card customer-ai-unavailable">
        <h3>Intelligence Customer</h3>
        <p>Riwayat AI tidak tersedia karena customer ini tidak memiliki relasi prospect sumber.</p>
      </section>

      <!-- TABS -->
      <nav class="tabs-bar">
        <button
          v-for="tab in tabs"
          :key="tab.key"
          :class="['tab-item', { active: activeTab === tab.key }]"
          @click="activeTab = tab.key"
        >
          <i :class="tab.icon" />
          {{ tab.label }}
        </button>
      </nav>

      <!-- =================== OVERVIEW TAB =================== -->
      <div v-if="activeTab === 'overview'" class="detail-grid">
        <div class="detail-card">
          <h3 class="card-heading">
            <i class="pi pi-map-marker" /> Site Information
          </h3>
          <div class="info-grid">
            <div class="info-item">
              <span class="info-label">Customer Code <span v-if="getFieldSource('customerCode')" class="fs-badge" :class="'fs-' + getFieldSource('customerCode')!.source" :title="getFieldSource('customerCode')!.tooltip">{{ fsLabels[getFieldSource('customerCode')!.source] }}</span></span>
              <code class="code-tag code-blue">{{ detail.customer.customerCode }}</code>
            </div>
            <div class="info-item">
              <span class="info-label">Site Name <span v-if="getFieldSource('siteName')" class="fs-badge" :class="'fs-' + getFieldSource('siteName')!.source" :title="getFieldSource('siteName')!.tooltip">{{ fsLabels[getFieldSource('siteName')!.source] }}</span></span>
              <strong>{{ detail.customer.name }}</strong>
            </div>
            <div class="info-item">
              <span class="info-label">Segment <span v-if="getFieldSource('segment')" class="fs-badge" :class="'fs-' + getFieldSource('segment')!.source" :title="getFieldSource('segment')!.tooltip">{{ fsLabels[getFieldSource('segment')!.source] }}</span></span>
              <Tag :value="detail.customer.segment" :severity="segmentSeverity(detail.customer.segment)" />
            </div>
            <div class="info-item">
              <span class="info-label">Category <span v-if="getFieldSource('category')" class="fs-badge" :class="'fs-' + getFieldSource('category')!.source" :title="getFieldSource('category')!.tooltip">{{ fsLabels[getFieldSource('category')!.source] }}</span></span>
              <strong>{{ detail.customer.category || '—' }}</strong>
            </div>
            <div class="info-item">
              <span class="info-label">Region <span v-if="getFieldSource('region')" class="fs-badge" :class="'fs-' + getFieldSource('region')!.source" :title="getFieldSource('region')!.tooltip">{{ fsLabels[getFieldSource('region')!.source] }}</span></span>
              <strong>{{ detail.customer.region || '—' }}</strong>
            </div>
            <div class="info-item">
              <span class="info-label">Sales Executive <span v-if="getFieldSource('salesExec')" class="fs-badge" :class="'fs-' + getFieldSource('salesExec')!.source" :title="getFieldSource('salesExec')!.tooltip">{{ fsLabels[getFieldSource('salesExec')!.source] }}</span></span>
              <strong>{{ detail.customer.salesExecutiveName || 'Unassigned' }}</strong>
            </div>
            <div class="info-item">
              <span class="info-label">Converted At <span v-if="getFieldSource('convertedAt')" class="fs-badge" :class="'fs-' + getFieldSource('convertedAt')!.source" :title="getFieldSource('convertedAt')!.tooltip">{{ fsLabels[getFieldSource('convertedAt')!.source] }}</span></span>
              <strong>{{ formatDateTime(detail.customer.convertedAt) }}</strong>
            </div>
            <div class="info-item">
              <span class="info-label">Last Updated <span v-if="getFieldSource('lastUpdated')" class="fs-badge" :class="'fs-' + getFieldSource('lastUpdated')!.source" :title="getFieldSource('lastUpdated')!.tooltip">{{ fsLabels[getFieldSource('lastUpdated')!.source] }}</span></span>
              <strong>{{ formatDateTime(detail.customer.updatedAt) }}</strong>
            </div>
          </div>
        </div>

        <div class="detail-card">
          <h3 class="card-heading">
            <i class="pi pi-building" /> Parent Company
          </h3>
          <div class="info-grid">
            <div class="info-item full">
              <span class="info-label">Company Name <span v-if="getFieldSource('companyName')" class="fs-badge" :class="'fs-' + getFieldSource('companyName')!.source" :title="getFieldSource('companyName')!.tooltip">{{ fsLabels[getFieldSource('companyName')!.source] }}</span></span>
              <strong>{{ detail.parentCompany.name }}</strong>
            </div>
            <div class="info-item">
              <span class="info-label">Company Code <span v-if="getFieldSource('companyCode')" class="fs-badge" :class="'fs-' + getFieldSource('companyCode')!.source" :title="getFieldSource('companyCode')!.tooltip">{{ fsLabels[getFieldSource('companyCode')!.source] }}</span></span>
              <code class="code-tag">{{ detail.parentCompany.parentCode }}</code>
            </div>
            <div class="info-item">
              <span class="info-label">Source Prospect <span v-if="getFieldSource('sourceProspect')" class="fs-badge" :class="'fs-' + getFieldSource('sourceProspect')!.source" :title="getFieldSource('sourceProspect')!.tooltip">{{ fsLabels[getFieldSource('sourceProspect')!.source] }}</span></span>
              <strong>{{ detail.sourceProspectName }}</strong>
            </div>
            <div class="info-item" v-if="detail.parentCompany.termOfPayment">
              <span class="info-label">Term of Payment</span>
              <strong>{{ detail.parentCompany.termOfPayment }}</strong>
            </div>
          </div>
          <Button label="View Full Company" icon="pi pi-arrow-right" severity="secondary" text size="small" class="card-footer-link" @click="router.push(`/admin/companies/${detail.parentCompany.parentCode}`)" />
        </div>
      </div>

      <!-- =================== GOOGLE MAPS TAB =================== -->
      <div v-if="activeTab === 'google'" class="detail-grid">
        <!-- No Google Data -->
        <div v-if="!placeDetails" class="empty-card" style="grid-column: 1 / -1;">
          <i class="pi pi-map" />
          <strong>No Google Maps data</strong>
           <span class="muted">This customer was not created from a Google Place.</span>
        </div>

        <template v-else>
          <!-- Google Place Identity -->
          <div class="detail-card">
            <h3 class="card-heading"><i class="pi pi-building" /> Google Place</h3>
            <div class="info-grid">
              <div v-if="placeDetails.placeName" class="info-item full">
                <span class="info-label">Google Place Name <span class="fs-badge fs-google" title="Retrieved from Google Places.">GOOGLE</span></span>
                <strong>{{ placeDetails.placeName }}</strong>
              </div>
              <div v-if="placeDetails.formattedAddress" class="info-item full">
                <span class="info-label">Google Address <span class="fs-badge fs-google" title="Retrieved from Google Places.">GOOGLE</span></span>
                <strong>{{ placeDetails.formattedAddress }}</strong>
              </div>
              <div v-if="placeDetails.latitude != null" class="info-item">
                <span class="info-label">Coordinates <span class="fs-badge fs-google" title="Retrieved from Google Places.">GOOGLE</span></span>
                <strong>{{ placeDetails.latitude.toFixed(6) }}, {{ placeDetails.longitude.toFixed(6) }}</strong>
              </div>
              <div v-if="placeDetails.googlePlaceId" class="info-item full">
                <span class="info-label">Google Place ID <span class="fs-badge fs-google" title="Retrieved from Google Places.">GOOGLE</span></span>
                <code class="code-tag" style="word-break: break-all;">{{ placeDetails.googlePlaceId }}</code>
              </div>
              <div v-if="placeDetails.googleMapsUrl" class="info-item full">
                <span class="info-label">Google Maps <span class="fs-badge fs-google" title="Retrieved from Google Places.">GOOGLE</span></span>
                <a :href="placeDetails.googleMapsUrl" target="_blank" rel="noopener" class="info-link"><i class="pi pi-map" /> View on Google Maps</a>
              </div>
            </div>
          </div>

          <!-- Editorial Summary -->
          <div v-if="placeDetails.editorialSummary" class="detail-card" style="grid-column: 1 / -1;">
            <h3 class="card-heading"><i class="pi pi-info-circle" /> About this place</h3>
            <p class="editorial-text">{{ placeDetails.editorialSummary }}</p>
          </div>

          <!-- Photos -->
          <div v-if="placeDetails.photos?.length" class="detail-card" style="grid-column: 1 / -1;">
            <h3 class="card-heading"><i class="pi pi-images" /> Photos</h3>
            <PlacePhotoGallery :photos="placeDetails.photos" :prospect-id="detail.customer.sourceProspectId" role="ADMINISTRATOR" />
          </div>

          <!-- Rating & Business Status -->
          <div class="detail-card">
            <h3 class="card-heading"><i class="pi pi-star" /> Rating & Status</h3>
            <div class="info-grid">
              <div v-if="placeDetails.rating > 0" class="info-item full">
                <span class="info-label">Google Rating</span>
                <div class="rating-row">
                  <span class="rating-num">{{ placeDetails.rating.toFixed(1) }}</span>
                  <div class="rating-stars">
                    <i v-for="(s, i) in stars(placeDetails.rating)" :key="i" :class="['pi', s]" />
                  </div>
                  <span class="rating-count">({{ placeDetails.userRatingCount.toLocaleString() }} reviews)</span>
                </div>
              </div>
              <div v-if="placeDetails.businessStatus" class="info-item">
                <span class="info-label">Business Status</span>
                <Tag :value="businessStatusLabel(placeDetails.businessStatus)" :severity="businessStatusSeverity(placeDetails.businessStatus)" />
              </div>
              <div v-if="placeDetails.priceLevel" class="info-item">
                <span class="info-label">Price Level</span>
                <strong>{{ priceLevelLabel(placeDetails.priceLevel) }}</strong>
              </div>
              <div v-if="placeDetails.utcOffsetMinutes != null" class="info-item">
                <span class="info-label">Time Zone</span>
                <strong>{{ utcOffsetLabel(placeDetails.utcOffsetMinutes) }} <span class="muted">({{ placeDetails.utcOffsetMinutes >= 0 ? '+' : '' }}{{ placeDetails.utcOffsetMinutes }} min)</span></strong>
              </div>
              <div v-if="placeDetails.phoneNumber" class="info-item">
                <span class="info-label">Phone</span>
                <a :href="`tel:${placeDetails.phoneNumber}`" class="info-link">{{ placeDetails.phoneNumber }}</a>
              </div>
              <div v-if="placeDetails.internationalPhone" class="info-item">
                <span class="info-label">International Phone</span>
                <strong>{{ placeDetails.internationalPhone }}</strong>
              </div>
              <div v-if="placeDetails.websiteUrl" class="info-item">
                <span class="info-label">Website</span>
                <a :href="placeDetails.websiteUrl" target="_blank" rel="noopener" class="info-link"><i class="pi pi-external-link" /> {{ placeDetails.websiteUrl }}</a>
              </div>
            </div>
          </div>

          <!-- Place Types -->
          <div v-if="placeDetails.placeTypes?.length" class="detail-card">
            <h3 class="card-heading"><i class="pi pi-tags" /> Place Categories</h3>
            <div class="types-wrap">
              <Tag v-for="t in placeDetails.placeTypes" :key="t" :value="t.replace(/_/g, ' ')" severity="secondary" class="type-tag" />
            </div>
          </div>

          <!-- Opening Hours -->
          <div v-if="placeDetails.openingHours" class="detail-card">
            <h3 class="card-heading"><i class="pi pi-clock" /> Opening Hours</h3>
            <div class="hours-status">
              <span :class="['hours-dot', placeDetails.openingHours.openNow ? 'open' : 'closed']" />
              <strong>{{ placeDetails.openingHours.openNow ? 'Open now' : 'Closed' }}</strong>
            </div>
            <div v-if="placeDetails.openingHours.weekdays?.length" class="hours-list">
              <div
                v-for="(day, i) in (showAllHours ? placeDetails.openingHours.weekdays : placeDetails.openingHours.weekdays.slice(0, 3))"
                :key="i"
                class="hours-row"
                v-html="day"
              />
              <button v-if="placeDetails.openingHours.weekdays.length > 3" class="hours-toggle" @click="showAllHours = !showAllHours">
                {{ showAllHours ? 'Show less' : `Show all ${placeDetails.openingHours.weekdays.length} days` }}
              </button>
            </div>
          </div>

          <!-- Reviews -->
          <div v-if="placeDetails.reviews?.length" class="detail-card" style="grid-column: 1 / -1;">
            <h3 class="card-heading"><i class="pi pi-comments" /> Reviews
              <span v-if="placeDetails.userRatingCount > 0" class="rating-count">({{ placeDetails.userRatingCount.toLocaleString() }} total on Google)</span>
            </h3>
            <div class="reviews-list">
              <div v-for="(review, i) in placeDetails.reviews.slice(0, 5)" :key="i" class="review-item">
                <div class="review-header">
                  <img v-if="review.authorPhoto" :src="review.authorPhoto" class="review-avatar" :alt="review.authorName" @error="($event.target as HTMLImageElement).style.display='none'" />
                  <div v-else class="review-avatar-placeholder">{{ review.authorName?.charAt(0) || '?' }}</div>
                  <div class="review-meta">
                    <strong>{{ review.authorName }}</strong>
                    <div class="review-stars">
                      <i v-for="(s, j) in stars(review.rating)" :key="j" :class="['pi', s]" />
                      <span class="review-time">{{ review.time }}</span>
                    </div>
                  </div>
                </div>
                <p v-if="review.text" class="review-text">{{ review.text }}</p>
              </div>
            </div>
            <a v-if="placeDetails.googleMapsUrl" :href="placeDetails.googleMapsUrl" target="_blank" rel="noopener" class="reviews-maps-link">
              <i class="pi pi-external-link" /> See all reviews on Google Maps
            </a>
          </div>

          <!-- Service Options -->
          <div v-if="placeDetails.delivery || placeDetails.dineIn || placeDetails.takeout || placeDetails.curbsidePickup" class="detail-card">
            <h3 class="card-heading"><i class="pi pi-cog" /> Service Options</h3>
            <div class="info-grid">
              <div v-if="placeDetails.dineIn" class="info-item"><span class="info-label">Dine In</span><strong>Available</strong></div>
              <div v-if="placeDetails.takeout" class="info-item"><span class="info-label">Takeout</span><strong>Available</strong></div>
              <div v-if="placeDetails.delivery" class="info-item"><span class="info-label">Delivery</span><strong>Available</strong></div>
              <div v-if="placeDetails.curbsidePickup" class="info-item"><span class="info-label">Curbside Pickup</span><strong>Available</strong></div>
            </div>
          </div>

          <!-- Parking Options -->
          <div v-if="placeDetails.parkingOptions" class="detail-card">
            <h3 class="card-heading"><i class="pi pi-directions" /> Parking</h3>
            <div class="info-grid">
              <div v-if="placeDetails.parkingOptions.freeParkingLot" class="info-item"><span class="info-label">Free Parking Lot</span><strong>Yes</strong></div>
              <div v-if="placeDetails.parkingOptions.freeStreetParking" class="info-item"><span class="info-label">Free Street Parking</span><strong>Yes</strong></div>
              <div v-if="placeDetails.parkingOptions.paidParkingLot" class="info-item"><span class="info-label">Paid Parking Lot</span><strong>Yes</strong></div>
              <div v-if="placeDetails.parkingOptions.paidStreetParking" class="info-item"><span class="info-label">Paid Street Parking</span><strong>Yes</strong></div>
              <div v-if="placeDetails.parkingOptions.garageParking" class="info-item"><span class="info-label">Garage Parking</span><strong>Yes</strong></div>
              <div v-if="placeDetails.parkingOptions.valetParking" class="info-item"><span class="info-label">Valet Parking</span><strong>Yes</strong></div>
            </div>
          </div>

          <!-- Payment Options -->
          <div v-if="placeDetails.paymentOptions && (placeDetails.paymentOptions.cashOnly || placeDetails.paymentOptions.creditCardOnly || placeDetails.paymentOptions.debitCardOnly || placeDetails.paymentOptions.nfcOnly)" class="detail-card">
            <h3 class="card-heading"><i class="pi pi-wallet" /> Payment Options</h3>
            <div class="info-grid">
              <div v-if="placeDetails.paymentOptions.cashOnly" class="info-item"><span class="info-label">Cash Only</span><strong>Yes</strong></div>
              <div v-if="placeDetails.paymentOptions.creditCardOnly" class="info-item"><span class="info-label">Credit Card Only</span><strong>Yes</strong></div>
              <div v-if="placeDetails.paymentOptions.debitCardOnly" class="info-item"><span class="info-label">Debit Card Only</span><strong>Yes</strong></div>
              <div v-if="placeDetails.paymentOptions.nfcOnly" class="info-item"><span class="info-label">NFC Only</span><strong>Yes</strong></div>
            </div>
          </div>

          <!-- Accessibility -->
          <div v-if="placeDetails.accessibilityOptions" class="detail-card">
            <h3 class="card-heading"><i class="pi pi-verified" /> Accessibility</h3>
            <div class="info-grid">
              <div v-if="placeDetails.accessibilityOptions.wheelchairAccessibleEntrance" class="info-item"><span class="info-label">Wheelchair Entrance</span><strong>Yes</strong></div>
              <div v-if="placeDetails.accessibilityOptions.wheelchairAccessibleParking" class="info-item"><span class="info-label">Wheelchair Parking</span><strong>Yes</strong></div>
              <div v-if="placeDetails.accessibilityOptions.wheelchairAccessibleRestroom" class="info-item"><span class="info-label">Wheelchair Restroom</span><strong>Yes</strong></div>
              <div v-if="placeDetails.accessibilityOptions.wheelchairAccessibleSeating" class="info-item"><span class="info-label">Wheelchair Seating</span><strong>Yes</strong></div>
            </div>
          </div>
        </template>
      </div>

      <!-- =================== CONTACTS TAB =================== -->
      <div v-if="activeTab === 'contacts'" class="detail-grid">
        <div v-if="!detail.customer.contacts?.length" class="empty-card">
          <i class="pi pi-users" />
          <strong>No contacts registered</strong>
          <span class="muted">Contacts will appear here once added to this customer site.</span>
        </div>
        <div v-for="(contact, idx) in detail.customer.contacts" :key="idx" class="contact-card">
          <div class="contact-avatar">
            <span>{{ (contact.name || detail.customer.name || 'U').charAt(0).toUpperCase() }}</span>
          </div>
          <div class="contact-info">
            <h4>{{ contact.name || detail.customer.name || 'Unnamed Contact' }}</h4>
            <span class="contact-position" v-if="contact.position">{{ contact.position }} <span class="fs-badge fs-manual" title="Entered manually by the administrator.">MANUAL</span></span>
            <div class="contact-details">
              <span v-if="contact.phone"><i class="pi pi-phone" /> {{ contact.phone }} <span class="fs-badge fs-manual" title="Entered manually by the administrator.">MANUAL</span></span>
              <span v-if="contact.email"><i class="pi pi-envelope" /> {{ contact.email }} <span class="fs-badge fs-manual" title="Entered manually by the administrator.">MANUAL</span></span>
            </div>
          </div>
        </div>
      </div>

      <!-- =================== COMPANY TAB =================== -->
      <div v-if="activeTab === 'company'" class="detail-grid">
        <div class="detail-card">
          <h3 class="card-heading">
            <i class="pi pi-building" /> Corporate Details
          </h3>
          <div class="info-grid">
            <div class="info-item full">
              <span class="info-label">Company Name <span v-if="getFieldSource('companyName')" class="fs-badge" :class="'fs-' + getFieldSource('companyName')!.source" :title="getFieldSource('companyName')!.tooltip">{{ fsLabels[getFieldSource('companyName')!.source] }}</span></span>
              <button class="link-btn" @click="router.push(`/admin/companies/${detail.parentCompany.parentCode}`)">{{ detail.parentCompany.name }}</button>
            </div>
            <div class="info-item">
              <span class="info-label">Parent Code <span v-if="getFieldSource('companyCode')" class="fs-badge" :class="'fs-' + getFieldSource('companyCode')!.source" :title="getFieldSource('companyCode')!.tooltip">{{ fsLabels[getFieldSource('companyCode')!.source] }}</span></span>
              <code class="code-tag">{{ detail.parentCompany.parentCode }}</code>
            </div>
            <div class="info-item full" v-if="detail.parentCompany.address.previewAddress">
              <span class="info-label">Company Address</span>
              <strong>{{ detail.parentCompany.address.previewAddress }}</strong>
            </div>
            <div class="info-item" v-if="detail.parentCompany.npwpNumber">
              <span class="info-label">NPWP Number <span class="fs-badge fs-manual" title="Entered manually by the administrator.">MANUAL</span></span>
              <code class="code-tag">{{ detail.parentCompany.npwpNumber }}</code>
            </div>
            <div class="info-item" v-if="detail.parentCompany.npwpName">
              <span class="info-label">NPWP Name <span class="fs-badge fs-manual" title="Entered manually by the administrator.">MANUAL</span></span>
              <strong>{{ detail.parentCompany.npwpName }}</strong>
            </div>
            <div class="info-item" v-if="detail.parentCompany.npwpAddress">
              <span class="info-label">NPWP Address <span class="fs-badge fs-manual" title="Entered manually by the administrator.">MANUAL</span></span>
              <strong>{{ detail.parentCompany.npwpAddress }}</strong>
            </div>
            <div class="info-item" v-if="detail.parentCompany.termOfPayment">
              <span class="info-label">Term of Payment <span class="fs-badge fs-manual" title="Configured manually by the administrator.">MANUAL</span></span>
              <strong>{{ detail.parentCompany.termOfPayment }}</strong>
            </div>
          </div>
        </div>

        <div v-if="detail.parentCompany.contacts?.length" class="detail-card">
          <h3 class="card-heading">
            <i class="pi pi-users" /> Company Contacts
          </h3>
          <div class="contacts-grid">
            <div v-for="(contact, idx) in detail.parentCompany.contacts" :key="idx" class="contact-card compact">
              <div class="contact-avatar small">
                <span>{{ (contact.name || 'U').charAt(0).toUpperCase() }}</span>
              </div>
              <div class="contact-info">
                <h4>{{ contact.name || 'Unnamed' }}</h4>
                <span class="contact-position" v-if="contact.position">{{ contact.position }}</span>
                <div class="contact-details">
                  <span v-if="contact.phone"><i class="pi pi-phone" /> {{ contact.phone }} <span class="fs-badge fs-manual" title="Entered manually by the administrator.">MANUAL</span></span>
                  <span v-if="contact.email"><i class="pi pi-envelope" /> {{ contact.email }} <span class="fs-badge fs-manual" title="Entered manually by the administrator.">MANUAL</span></span>
                </div>
              </div>
            </div>
          </div>
        </div>

        <div v-if="detail.parentCompany.kamAssignments?.length" class="detail-card">
          <h3 class="card-heading">
            <i class="pi pi-user" /> KAM Assignments
          </h3>
          <div class="assignment-list">
            <div v-for="(kam, idx) in detail.parentCompany.kamAssignments" :key="idx" class="assignment-row">
              <div class="assignment-avatar">
                <span>{{ (kam.ownerName || 'U').charAt(0).toUpperCase() }}</span>
              </div>
              <div class="assignment-info">
                <strong>{{ kam.ownerName }}</strong>
                <span class="assignment-period">
                  {{ kam.startMonth }}/{{ kam.startYear }} &mdash; {{ kam.end === 'UNTIL_NOW' ? 'Present' : kam.end }}
                </span>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- =================== ADDRESS TAB =================== -->
      <div v-if="activeTab === 'address'" class="detail-grid">
        <div class="detail-card">
          <h3 class="card-heading">
            <i class="pi pi-map" /> Site Address
          </h3>
          <div v-if="detail.customer.address?.previewAddress" class="address-block">
            <p class="address-full">{{ detail.customer.address.previewAddress }}</p>
            <div class="info-grid">
              <div class="info-item" v-if="detail.customer.address.province">
                <span class="info-label">Province <span class="fs-badge fs-google" title="Derived from Google Places data.">GOOGLE</span></span>
                <strong>{{ detail.customer.address.province }}</strong>
              </div>
              <div class="info-item" v-if="detail.customer.address.district">
                <span class="info-label">District <span class="fs-badge fs-google" title="Derived from Google Places data.">GOOGLE</span></span>
                <strong>{{ detail.customer.address.district }}</strong>
              </div>
              <div class="info-item" v-if="detail.customer.address.subDistrict">
                <span class="info-label">Sub-District <span class="fs-badge fs-google" title="Derived from Google Places data.">GOOGLE</span></span>
                <strong>{{ detail.customer.address.subDistrict }}</strong>
              </div>
              <div class="info-item" v-if="detail.customer.address.village">
                <span class="info-label">Village <span class="fs-badge fs-google" title="Derived from Google Places data.">GOOGLE</span></span>
                <strong>{{ detail.customer.address.village }}</strong>
              </div>
              <div class="info-item" v-if="detail.customer.address.latitude != null">
                <span class="info-label">Coordinates <span class="fs-badge fs-google" title="Retrieved from Google Places.">GOOGLE</span></span>
                <strong>{{ detail.customer.address.latitude?.toFixed(6) }}, {{ detail.customer.address.longitude?.toFixed(6) }}</strong>
              </div>
            </div>
            <a v-if="siteGoogleMapsUrl" :href="siteGoogleMapsUrl" target="_blank" rel="noopener" class="info-link address-maps-link">
              <i class="pi pi-map" /> Open in Google Maps
            </a>
          </div>
          <div v-else class="empty-inline">
            <i class="pi pi-map" />
            <span>No address information available for this customer site.</span>
          </div>
        </div>

        <div v-if="detail.parentCompany.address?.previewAddress" class="detail-card">
          <h3 class="card-heading">
            <i class="pi pi-building" /> Company Address
          </h3>
          <div class="address-block">
            <p class="address-full">{{ detail.parentCompany.address.previewAddress }}</p>
            <div class="info-grid">
              <div class="info-item" v-if="detail.parentCompany.address.province">
                <span class="info-label">Province <span class="fs-badge fs-system" title="Linked from the parent company record.">SYSTEM</span></span>
                <strong>{{ detail.parentCompany.address.province }}</strong>
              </div>
              <div class="info-item" v-if="detail.parentCompany.address.district">
                <span class="info-label">District <span class="fs-badge fs-system" title="Linked from the parent company record.">SYSTEM</span></span>
                <strong>{{ detail.parentCompany.address.district }}</strong>
              </div>
              <div class="info-item" v-if="detail.parentCompany.address.subDistrict">
                <span class="info-label">Sub-District <span class="fs-badge fs-system" title="Linked from the parent company record.">SYSTEM</span></span>
                <strong>{{ detail.parentCompany.address.subDistrict }}</strong>
              </div>
              <div class="info-item" v-if="detail.parentCompany.address.village">
                <span class="info-label">Village <span class="fs-badge fs-system" title="Linked from the parent company record.">SYSTEM</span></span>
                <strong>{{ detail.parentCompany.address.village }}</strong>
              </div>
            </div>
          </div>
        </div>
      </div>
    </template>

    <Dialog v-model:visible="deleteDialogVisible" header="Delete Customer" modal :style="{ width: '400px' }">
      <p>Are you sure you want to delete <strong>{{ detail?.customer.name }}</strong>? This action cannot be undone.</p>
      <template #footer>
        <Button label="Cancel" severity="secondary" text @click="deleteDialogVisible = false" :disabled="deleting" />
        <Button label="Delete" severity="danger" icon="pi pi-trash" :loading="deleting" @click="executeDelete" />
      </template>
    </Dialog>
  </section>
</template>

<style scoped>
.customer-intelligence {
  display: grid;
  grid-template-columns: minmax(0, 1fr) minmax(300px, 350px);
  gap: 1rem;
  min-width: 0;
  margin: 1rem 0;
}
.customer-admin-opportunity { grid-column: 1; grid-row: 1 / span 3; }
.customer-admin-insight { grid-column: 2; grid-row: 1; }
.customer-admin-discussion { grid-column: 2; grid-row: 2; }
.customer-admin-chat { grid-column: 2; grid-row: 3; }
.customer-admin-hero {
  min-width: 0;
  padding: .3rem 0 .8rem;
  border-bottom: 1px solid #eee3e5;
}
.customer-admin-hero .page-title-wrapper { min-width: 0; }
.customer-admin-hero .title-row { min-width: 0; display: flex; align-items: center; flex-wrap: wrap; gap: .55rem; }
.customer-admin-hero h1 { min-width: 0; overflow-wrap: anywhere; font-size: clamp(1.45rem, 2vw, 1.85rem); }
.customer-admin-hero .page-heading-actions { flex: 0 0 auto; }
.customer-detail-card { min-width: 0; border: 1px solid #eee3e5; border-radius: 14px; background: #fff; box-shadow: 0 5px 16px rgba(73, 34, 41, .05); }
.customer-detail-card .strip-item { min-width: 0; }
.customer-detail-card strong { overflow-wrap: anywhere; }
.customer-intelligence > * { min-width: 0; max-width: 100%; }
.customer-ai-unavailable { margin: 1rem 0; }
@media (max-width: 1199px) {
  .customer-intelligence { grid-template-columns: 1fr; }
  .customer-admin-insight { grid-column: 1; grid-row: 1; }
  .customer-admin-opportunity { grid-column: 1; grid-row: 2; }
  .customer-admin-discussion { grid-column: 1; grid-row: 3; }
  .customer-admin-chat { grid-column: 1; grid-row: 4; }
}
@media (max-width: 767px) {
  .customer-admin-hero { align-items: flex-start; }
  .customer-admin-hero .page-heading-actions { width: 100%; flex-wrap: wrap; }
}
.admin-page {
  display: flex;
  flex-direction: column;
  gap: 1.25rem;
  padding: 1.75rem 2rem;
  min-height: 100vh;
}

.admin-comments-section {
  width: 100%;
}

.admin-comments-section :deep(.pc-list) {
  max-height: 420px;
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
.title-row {
  display: flex;
  align-items: center;
  gap: 0.65rem;
  flex-wrap: wrap;
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
.code-blue {
  background: #fff0f1;
  color: #e63946;
}

/* ── SUMMARY STRIP ─────────────────────────────────────────────────── */
.summary-strip {
  display: flex;
  gap: 0;
  background: var(--surface-card);
  border: 1px solid var(--border-light);
  border-radius: var(--radius-lg);
  box-shadow: var(--shadow-xs);
  overflow: hidden;
}
.strip-item {
  flex: 1;
  display: flex;
  align-items: center;
  gap: 0.7rem;
  padding: 0.85rem 1.1rem;
  border-right: 1px solid #f0f3f7;
}
.strip-item:last-child { border-right: none; }
.strip-item i {
  color: var(--text-faint);
  font-size: 0.95rem;
  flex-shrink: 0;
}
.strip-item div {
  display: flex;
  flex-direction: column;
  min-width: 0;
}
.strip-item span {
  font-size: 0.65rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: var(--text-muted);
  line-height: 1.3;
}
.strip-item strong {
  font-size: 0.82rem;
  font-weight: 600;
  color: var(--text-primary);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

/* ── TABS ─────────────────────────────────────────────────────────── */
.tabs-bar {
  display: flex;
  gap: 0;
  border-bottom: 1px solid var(--border-light);
  padding: 0 0.15rem;
}
.tab-item {
  display: flex;
  align-items: center;
  gap: 0.45rem;
  padding: 0.7rem 1rem;
  font-size: 0.85rem;
  font-weight: 600;
  color: var(--text-muted);
  background: none;
  border: none;
  border-bottom: 2px solid transparent;
  margin-bottom: -1px;
  cursor: pointer;
  transition: color 0.15s, border-color 0.15s;
  white-space: nowrap;
}
.tab-item i { font-size: 0.85rem; }
.tab-item:hover { color: var(--text-primary); }
.tab-item.active {
  color: var(--brand-blue);
  border-bottom-color: var(--brand-blue);
}

/* ── DETAIL GRID ───────────────────────────────────────────────────── */
.detail-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 1.25rem;
}

/* ── DETAIL CARD ───────────────────────────────────────────────────── */
.customer-detail-information-grid{display:grid;grid-template-columns:minmax(0,1fr) minmax(320px,.96fr);gap:18px;align-items:start}.customer-detail-side-information{display:grid;gap:18px}.detail-card-header{padding:20px;border-bottom:1px solid #e2e8f0}.detail-card-header h3{margin:0;color:#172033;font-size:17px;font-weight:700}.detail-card-header p{margin:5px 0 0;color:#64748b;font-size:12px}.detail-list{display:grid;gap:10px;padding:20px}.detail-list>div{display:flex;min-height:48px;align-items:center;justify-content:space-between;gap:20px;border-radius:14px;background:#f8fafc;padding:0 16px}.detail-list span{color:#64748b;font-size:12px}.detail-list strong{max-width:65%;color:#172033;font-size:12px;text-align:right;overflow-wrap:anywhere}
.customer-detail-navbar{display:flex;min-height:52px;align-items:center;gap:12px;border-top:1px solid #e2e8f0;border-bottom:1px solid #e2e8f0;background:#fff;padding:0 10px}.customer-detail-back{display:inline-flex;align-items:center;gap:7px;border:0;border-right:1px solid #e2e8f0;background:transparent;padding:0 12px 0 0;color:#475569;font-size:12px;cursor:pointer}.customer-detail-back:hover{color:#dc2626}.customer-detail-navbar-title{color:#172033;font-size:14px;font-weight:700}.customer-detail-breadcrumb{color:#94a3b8;font-size:11px}.customer-detail-navbar :deep(.p-button){margin-left:auto;border-radius:10px;background:#dc2626;border-color:#dc2626;font-size:12px;font-weight:700}.customer-detail-navbar :deep(.p-button:hover){background:#b91c1c;border-color:#b91c1c}
.admin-page:has(.customer-detail-navbar){padding-top:0;padding-left:1rem;padding-right:1rem}.customer-detail-navbar{margin:0 -1rem .85rem}
.summary-strip,.customer-admin-snapshot{display:none !important}
.customer-information-card{border-color:#dfe6ef !important}
.company-contact-card-inside{display:block !important;visibility:visible !important;margin:20px;border-color:#dfe6ef}.company-data-card-standalone~.company-contact-card{display:none !important}
.company-data-card-standalone{display:flex !important;flex-direction:column}.company-data-card-standalone>.company-data-card-header{order:1}.company-data-card-standalone>.company-data-card-body{order:2}.company-data-card-standalone>.company-address-card-inline,.company-data-card-standalone>.company-kam-card-inline{order:3}.company-data-card-standalone>.company-contact-card-inside{order:4}
.company-data-card-standalone>.company-data-card-header{order:1!important}.company-data-card-standalone>.company-data-card-body{order:2!important}.company-data-card-standalone>.company-address-card-inline,.company-data-card-standalone>.company-kam-card-inline{order:3!important}.company-data-card-standalone>.company-contact-card-inside{order:99!important}
.company-contact-card{display:none !important}.company-data-card-standalone>.company-contact-card-inside{display:block !important;visibility:visible !important}
.company-data-card-standalone{font-family:Inter,ui-sans-serif,system-ui,sans-serif;border-color:#fecaca!important;border-radius:20px!important;background:#fff!important;box-shadow:0 14px 34px rgba(15,23,42,.06)!important}.company-data-card-standalone>.company-data-card-header{border-bottom:1px solid #fee2e2;padding:20px}.company-data-card-standalone .company-data-card-header p{font-size:11px;font-weight:700;letter-spacing:.08em;color:#94a3b8}.company-data-card-standalone .company-data-card-header>div:first-child span{margin-top:4px;color:#64748b;font-size:12px}.company-data-card-standalone .company-data-card-title{margin-top:24px}.company-data-card-standalone .company-data-card-title h3{font-size:17px;font-weight:700;color:#0f172a}.company-data-card-standalone .company-data-card-title span{font-size:12px;color:#64748b}.company-data-card-standalone .company-data-card-title em{font-size:11px;font-weight:600}.company-data-card-standalone .company-data-card-body{display:grid;grid-template-columns:repeat(2,minmax(0,1fr));gap:18px;padding:20px}.company-data-card-standalone .company-nested-card{border-color:#dfe6ef;border-radius:20px}.company-data-card-standalone .company-nested-card .detail-card-header{padding:20px;border-bottom-color:#eef2f7}.company-data-card-standalone .company-nested-card h3{font-size:17px;font-weight:700;color:#0f172a}.company-data-card-standalone .company-nested-card p{font-size:12px;color:#64748b}.company-data-card-standalone .company-nested-card .detail-list>div{background:#f8fafc;padding:14px 16px}.company-data-card-standalone .company-nested-card .detail-list span{font-size:12px;color:#64748b}.company-data-card-standalone .company-nested-card .detail-list strong{font-size:13px;font-weight:600;color:#0f172a}
.company-kam-card{margin:20px 0;border:1px solid #dfe6ef;border-radius:20px;background:#fff;overflow:hidden}.company-kam-card .detail-card-header{border-bottom:1px solid #eef2f7;padding:20px}.company-kam-body{padding:20px}.company-kam-timeline{position:relative;padding-left:20px}.company-kam-timeline:before{position:absolute;left:7px;top:10px;bottom:10px;width:1px;background:#e2e8f0;content:''}.company-kam-item{position:relative;margin-bottom:10px}.company-kam-dot{position:absolute;left:-20px;top:10px;width:8px;height:8px;border:2px solid #dc2626;border-radius:50%;background:#dc2626}.company-kam-entry{border:1px solid #fecdd3;border-radius:12px;background:#fff1f2;padding:11px 14px}.company-kam-entry>div{display:flex;align-items:center;justify-content:space-between;gap:6px}.company-kam-entry strong{color:#dc2626;font-size:13px}.company-kam-entry em{border-radius:999px;background:#dc2626;padding:2px 8px;color:#fff;font-size:10px;font-style:normal;font-weight:700;text-transform:uppercase}.company-kam-entry p{margin:3px 0 0;color:#dc2626b3;font-size:11px}
.company-contact-card{width:100%;margin:20px 0;overflow:hidden;border:1px solid #dfe6ef;border-radius:20px;background:#fff;box-shadow:0 8px 24px rgba(15,23,42,.04)}.company-contact-card .detail-card-header{border-bottom:1px solid #eef2f7;padding:20px}.company-contact-grid{display:grid;grid-template-columns:repeat(2,minmax(0,1fr));gap:12px;padding:20px}.company-contact-item{border:1px solid #e2e8f0;border-radius:16px;background:#f8fafc;padding:14px 16px}.company-contact-heading{display:flex;align-items:flex-start;justify-content:space-between;gap:12px}.company-contact-heading>div{min-width:0}.company-contact-heading span,.company-contact-fields span{display:block;color:#94a3b8;font-size:10px;font-weight:700;letter-spacing:.08em}.company-contact-heading strong{display:block;margin-top:4px;color:#0f172a;font-size:14px}.company-contact-heading em{flex-shrink:0;border:1px solid #dbeafe;border-radius:999px;background:#fff;padding:3px 10px;color:#1d4ed8;font-size:10px;font-style:normal;font-weight:700}.company-contact-fields{display:grid;grid-template-columns:repeat(2,minmax(0,1fr));gap:10px;margin-top:12px}.company-contact-fields>div{border-radius:14px;background:#fff;padding:10px 12px}.company-contact-fields strong{display:block;margin-top:5px;color:#0f172a;font-size:13px;line-height:20px;overflow-wrap:anywhere}
.sales-assignment-card{overflow:hidden;border:1px solid #dfe6ef;border-radius:20px;background:#fff;box-shadow:0 8px 24px rgba(15,23,42,.04)}.sales-assignment-card .detail-card-header{border-bottom:1px solid #eef2f7;padding:20px}.sales-assignment-body{padding:20px}.sales-assignment-timeline{position:relative;padding-left:20px}.sales-assignment-timeline:before{position:absolute;left:7px;top:10px;bottom:10px;width:1px;background:#e2e8f0;content:''}.sales-assignment-item{position:relative}.sales-assignment-dot{position:absolute;left:-20px;top:10px;width:8px;height:8px;border:2px solid #dc2626;border-radius:50%;background:#dc2626}.sales-assignment-entry{border:1px solid #fecdd3;border-radius:12px;background:#fff1f2;padding:11px 14px}.sales-assignment-entry>div{display:flex;align-items:center;justify-content:space-between;gap:6px}.sales-assignment-entry strong{color:#dc2626;font-size:13px}.sales-assignment-entry em{border-radius:999px;background:#dc2626;padding:2px 8px;color:#fff;font-size:10px;font-style:normal;font-weight:700;text-transform:uppercase}.sales-assignment-entry p{margin:3px 0 0;color:#dc2626b3;font-size:11px}
.company-address-card{margin:0 20px 20px;overflow:hidden;border:1px solid #dfe6ef;border-radius:20px;background:#fff}.company-address-card .detail-card-header{border-bottom:1px solid #eef2f7;padding:20px}.company-address-card .detail-list{gap:10px;padding:20px}.company-address-card .detail-list>div{min-height:48px;border-radius:14px;padding:14px 16px;background:#f8fafc}.company-address-card .detail-list span{color:#64748b;font-size:12px}.company-address-card .detail-list strong{max-width:62%;color:#0f172a;font-size:13px;text-align:right;overflow-wrap:anywhere}
.company-address-card-inline{display:inline-block;width:calc(50% - 9px);margin:0 0 20px 0;vertical-align:top;box-sizing:border-box}
.company-kam-card-inline{display:inline-block;width:calc(50% - 9px);margin:0 0 20px 18px;vertical-align:top;box-sizing:border-box}.company-kam-card-standalone{display:none}
.customer-detail-information-grid .company-data-card{display:none}.company-data-card-standalone{display:block !important;width:100%;margin-top:20px}
.company-data-card-standalone{padding-bottom:20px}.company-data-card-standalone>.company-address-card-inline{width:calc(50% - 29px);margin:20px 0 0 20px}.company-data-card-standalone>.company-kam-card-inline{width:calc(50% - 29px);margin:20px 20px 0 18px}
.customer-information-card,.address-information-card,.contact-information-card{width:100%;max-width:592px;box-sizing:border-box}
.customer-detail-information-grid{grid-template-columns:minmax(0,592px) minmax(320px,592px);justify-content:center}
.customer-detail-information-grid>.customer-detail-side-information{grid-column:2;grid-row:1}.customer-detail-information-grid>.company-data-card{grid-column:1/-1;grid-row:2}
.customer-detail-information-grid>div:first-child{display:contents}.customer-detail-information-grid>.customer-information-card{grid-column:1;grid-row:1}.customer-detail-information-grid>.address-information-card{grid-column:1;grid-row:2}.customer-detail-information-grid>.contact-information-card{grid-column:1;grid-row:3}.customer-detail-information-grid>.customer-product-opportunity-below{grid-column:1;grid-row:4}.customer-detail-information-grid>.company-data-card{grid-column:1/-1;grid-row:5;margin-top:0}
.customer-detail-information-grid>div:first-child{display:block;grid-column:1;grid-row:1}.customer-detail-information-grid>div:first-child>.company-data-card{margin-top:20px}
.customer-detail-information-grid>.company-data-card{margin-top:20px}
@media (max-width:900px){.customer-detail-information-grid>.customer-detail-side-information{grid-column:1;grid-row:auto}.customer-detail-information-grid>.company-data-card{grid-column:1;grid-row:auto}}
.tax-information-card{overflow:hidden;border:1px solid #dfe6ef !important;border-radius:20px;background:#fff;box-shadow:0 8px 24px rgba(15,23,42,.04)}.tax-information-card .detail-card-header{border-bottom:1px solid #eef2f7;padding:20px}.tax-information-card .detail-list{gap:10px;padding:20px}.tax-information-card .detail-list>div{min-height:48px;border-radius:14px;padding:14px 16px;background:#f8fafc}.tax-information-card .detail-list span{font-size:12px;font-weight:500;line-height:18px;color:#64748b}.tax-information-card .detail-list strong{max-width:58%;font-size:13px;font-weight:600;line-height:20px;color:#0f172a}
.other-master-data-card{overflow:hidden;border:1px solid #dfe6ef !important;border-radius:20px;background:#fff;box-shadow:0 8px 24px rgba(15,23,42,.04)}.other-master-data-card .detail-card-header{border-bottom:1px solid #eef2f7;padding:20px}.other-master-data-card .detail-list{gap:10px;padding:20px}.other-master-data-card .detail-list>div{min-height:48px;border-radius:14px;padding:14px 16px;background:#f8fafc}.other-master-data-card .detail-list span{font-size:12px;font-weight:500;line-height:18px;color:#64748b}.other-master-data-card .detail-list strong{max-width:58%;font-size:13px;font-weight:600;line-height:20px;color:#0f172a}
.customer-product-opportunity-below{margin-top:18px;border:1px solid #dfe6ef;border-radius:20px;background:#fff;box-shadow:0 8px 24px rgba(15,23,42,.04);overflow:hidden}
.address-information-card{margin-top:18px;overflow:hidden;border:1px solid #dfe6ef !important;border-radius:20px;background:#fff;box-shadow:0 8px 24px rgba(15,23,42,.04)}.address-information-card .detail-card-header{border-bottom:1px solid #eef2f7;padding:20px}.address-information-card .detail-list{gap:10px;padding:20px}.address-information-card .detail-list>div{min-height:48px;border-radius:14px;padding:14px 16px;background:#f8fafc}.address-information-card .detail-list span{font-size:12px;font-weight:500;line-height:18px;color:#64748b}.address-information-card .detail-list strong{max-width:58%;font-size:13px;font-weight:600;line-height:20px;color:#0f172a}
.billing-shipment-card{margin-top:18px;overflow:hidden;border:1px solid #dfe6ef !important;border-radius:20px;background:#fff;box-shadow:0 8px 24px rgba(15,23,42,.04)}.billing-shipment-card .detail-card-header{border-bottom:1px solid #eef2f7;padding:20px}.billing-shipment-list{display:grid;gap:12px;padding:20px}.billing-shipment-block{border:1px solid #dfe6ef;border-radius:16px;background:#f8fafc;padding:16px}.billing-shipment-label{display:flex;align-items:center;justify-content:space-between;margin-bottom:12px;color:#7890ad;font-size:10px;letter-spacing:.08em}.billing-shipment-label em{border-radius:999px;background:#e2e8f0;padding:4px 10px;color:#475569;font-size:10px;font-style:normal;letter-spacing:0;text-transform:capitalize}.billing-shipment-block>div:not(.billing-shipment-label){display:flex;justify-content:space-between;gap:16px;padding:8px 0}.billing-shipment-block>div:not(.billing-shipment-label) span{color:#64748b;font-size:12px}.billing-shipment-block>div:not(.billing-shipment-label) strong{max-width:62%;color:#0f172a;font-size:12px;text-align:right;overflow-wrap:anywhere}
.contact-information-card{margin-top:18px;overflow:hidden;border:1px solid #dfe6ef !important;border-radius:20px;background:#fff;box-shadow:0 8px 24px rgba(15,23,42,.04)}.contact-information-card .detail-card-header{border-bottom:1px solid #eef2f7;padding:20px}.contact-list{display:grid;gap:10px;padding:20px}.contact-item{border:1px solid #dfe6ef;border-radius:16px;background:#f8fafc;padding:16px}.contact-item-heading{display:flex;align-items:flex-start;justify-content:space-between;gap:12px}.contact-item-heading>div{display:grid;gap:6px}.contact-item-heading span,.contact-item-fields span{color:#7890ad;font-size:10px;font-weight:600;letter-spacing:.06em}.contact-item-heading strong{color:#0f172a;font-size:14px}.contact-item-heading em{border:1px solid #bfdbfe;border-radius:999px;background:#fff;padding:4px 12px;color:#2563eb;font-size:10px;font-style:normal}.contact-item-fields{display:grid;grid-template-columns:1fr 1fr;gap:10px;margin-top:12px}.contact-item-fields>div{display:grid;gap:5px;border-radius:14px;background:#fff;padding:12px}.contact-item-fields strong{color:#0f172a;font-size:13px;font-weight:600;overflow-wrap:anywhere}.empty-contact{padding:20px;text-align:center;color:#64748b;font-size:12px}
.company-data-card{grid-column:1/-1;overflow:hidden;border:1px solid #fecaca;border-radius:20px;background:#fff}.company-data-card-header{border-bottom:1px solid #fee2e2;padding:20px}.company-data-card-header>div:first-child p{margin:0;color:#94a3b8;font-size:11px;font-weight:700;letter-spacing:.08em;text-transform:uppercase}.company-data-card-header>div:first-child span{display:block;margin-top:4px;color:#64748b;font-size:12px}.company-data-card-actions{display:flex;flex-direction:column;align-items:flex-end;gap:6px;float:right;margin-top:-32px}.company-data-card-actions button{border:0;background:transparent;color:#dc2626;font-size:13px;font-weight:600;cursor:pointer}.company-data-card-actions button+button{color:#64748b}.company-data-card-title{display:flex;align-items:center;justify-content:space-between;gap:12px;margin-top:24px}.company-data-card-title h3{margin:0;color:#0f172a;font-size:17px}.company-data-card-title span{display:block;margin-top:4px;color:#64748b;font-size:12px}.company-data-card-title em{border:1px solid #ddd6fe;border-radius:999px;background:#ede9fe;padding:4px 10px;color:#7c3aed;font-size:11px;font-style:normal;font-weight:600}.company-data-card-body{display:grid;grid-template-columns:1fr 1fr;gap:18px;padding:20px}.company-nested-card{overflow:hidden;border:1px solid #dfe6ef;border-radius:20px;background:#fff}.company-nested-card .detail-card-header{border-bottom:1px solid #eef2f7;padding:20px}.company-nested-card .detail-list{gap:10px;padding:20px}.company-nested-card .detail-list>div{min-height:48px;border-radius:14px;padding:14px 16px;background:#f8fafc}.company-nested-card .detail-list span{color:#64748b;font-size:12px}.company-nested-card .detail-list strong{max-width:62%;color:#0f172a;font-size:13px;text-align:right;overflow-wrap:anywhere}
.customer-detail-navbar ~ .tabs-bar,.customer-detail-navbar ~ .tabs-bar ~ .detail-grid{display:none !important}
.customer-ai-summary-card,.customer-discuss-card{width:100%;min-height:180px;flex:none;overflow:hidden;border:1px solid #dfe6ef !important;border-radius:20px;background:#fff;box-shadow:0 8px 24px rgba(15,23,42,.04);margin:0 !important}.customer-ai-summary-card :deep(.card),.customer-ai-summary-card :deep(.ai-summary-card),.customer-discuss-card :deep(.card){width:100%;min-height:178px;border:0 !important;border-radius:0 !important;box-shadow:none !important;margin:0 !important}
.customer-tanya-ai-card{width:100%;min-height:180px;flex:none;overflow:hidden;border:1px solid #dfe6ef !important;border-radius:20px;background:#fff;box-shadow:0 8px 24px rgba(15,23,42,.04);margin:0 !important}.customer-tanya-ai-card :deep(.card){width:100%;min-height:178px;border:0 !important;border-radius:0 !important;box-shadow:none !important;margin:0 !important}
.customer-information-card{overflow:hidden;border:1px solid #dfe6ef;border-radius:20px;background:#fff;box-shadow:0 8px 24px rgba(15,23,42,.04)}.customer-information-card .detail-card-header{border-bottom:1px solid #eef2f7;padding:20px}.customer-information-card .detail-card-header h3{color:#0f172a;font-size:17px;font-weight:700}.customer-information-card .detail-card-header p{margin-top:4px;font-size:12px;color:#64748b}.customer-information-card .detail-list{gap:10px;padding:20px}.customer-information-card .detail-list>div{min-height:48px;border-radius:14px;padding:14px 16px;background:#f8fafc}.customer-information-card .detail-list span{font-size:12px;font-weight:500;line-height:18px;color:#64748b}.customer-information-card .detail-list strong{max-width:58%;font-size:13px;font-weight:600;line-height:20px;color:#0f172a}
.detail-card {
  background: var(--surface-card);
  border: 1px solid var(--border-light);
  border-radius: var(--radius-lg);
  padding: 1.35rem 1.5rem;
  box-shadow: var(--shadow-xs);
}
.card-heading {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  margin: 0 0 1.15rem;
  padding-bottom: 0.85rem;
  border-bottom: 1px solid #f0f3f7;
  font-size: 0.9rem;
  font-weight: 700;
  color: var(--text-primary);
}
.card-heading i {
  color: var(--brand-blue);
  font-size: 0.9rem;
}

/* ── INFO GRID ─────────────────────────────────────────────────────── */
.info-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 0.9rem;
}
.info-item {
  display: flex;
  flex-direction: column;
  gap: 0.2rem;
}
.info-item.full {
  grid-column: 1 / -1;
}
.info-label {
  font-size: 0.68rem;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: var(--text-muted);
}
.info-item strong {
  font-size: 0.85rem;
  font-weight: 600;
  color: var(--text-primary);
}

.card-footer-link {
  margin-top: 1rem;
  padding-top: 0.85rem;
  border-top: 1px solid #f0f3f7;
}

.link-btn {
  background: none;
  border: none;
  padding: 0;
  cursor: pointer;
  text-align: left;
  font: inherit;
  font-weight: 600;
  color: #e63946;
  transition: color 0.15s;
}
.link-btn:hover { color: #d62839; text-decoration: underline; }

/* ── EMPTY / INLINE ────────────────────────────────────────────────── */
.empty-card {
  grid-column: 1 / -1;
  min-height: 200px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 0.4rem;
  background: var(--surface-card);
  border: 1px solid var(--border-light);
  border-radius: var(--radius-lg);
  padding: 2rem;
  text-align: center;
}
.empty-card i {
  font-size: 2rem;
  color: var(--text-faint);
  margin-bottom: 0.25rem;
}
.empty-card strong {
  color: var(--text-primary);
  font-size: 0.95rem;
}
.empty-inline {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 1rem;
  background: #f8fafc;
  border-radius: var(--radius-md);
  color: var(--text-muted);
  font-size: 0.85rem;
}

/* ── CONTACT CARDS ─────────────────────────────────────────────────── */
.contact-card {
  background: var(--surface-card);
  border: 1px solid var(--border-light);
  border-radius: var(--radius-lg);
  padding: 1.25rem 1.5rem;
  box-shadow: var(--shadow-xs);
  display: flex;
  align-items: flex-start;
  gap: 1rem;
}
.contact-card.compact {
  padding: 1rem 1.25rem;
}
.contact-avatar {
  width: 48px;
  height: 48px;
  border-radius: 50%;
  background: #fff0f1;
  color: #e63946;
  display: grid;
  place-content: center;
  font-weight: 800;
  font-size: 1.1rem;
  flex-shrink: 0;
}
.contact-avatar.small {
  width: 40px;
  height: 40px;
  font-size: 0.9rem;
}
.contact-info {
  display: flex;
  flex-direction: column;
  gap: 0.2rem;
  min-width: 0;
}
.contact-info h4 {
  margin: 0;
  font-size: 0.9rem;
  font-weight: 700;
  color: var(--text-primary);
}
.contact-position {
  font-size: 0.75rem;
  color: var(--text-muted);
}
.contact-details {
  display: flex;
  flex-wrap: wrap;
  gap: 0.75rem;
  margin-top: 0.35rem;
}
.contact-details span {
  display: inline-flex;
  align-items: center;
  gap: 0.3rem;
  font-size: 0.8rem;
  color: var(--text-secondary);
}
.contact-details i {
  font-size: 0.75rem;
  color: var(--text-faint);
}
.contacts-grid {
  display: grid;
  gap: 0.75rem;
}

/* ── ASSIGNMENTS ───────────────────────────────────────────────────── */
.assignment-list {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}
.assignment-row {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.75rem;
  background: #f8fafc;
  border-radius: var(--radius-md);
}
.assignment-avatar {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  background: #fff5f5;
  color: #ef4e5d;
  display: grid;
  place-content: center;
  font-weight: 700;
  font-size: 0.85rem;
  flex-shrink: 0;
}
.assignment-info {
  display: flex;
  flex-direction: column;
  gap: 0.1rem;
}
.assignment-info strong {
  font-size: 0.85rem;
  font-weight: 600;
  color: var(--text-primary);
}
.assignment-period {
  font-size: 0.75rem;
  color: var(--text-muted);
}

/* ── ADDRESS BLOCK ─────────────────────────────────────────────────── */
.address-block {
  display: flex;
  flex-direction: column;
  gap: 0.85rem;
}
.address-full {
  margin: 0;
  padding: 0.75rem 1rem;
  background: #f8fafc;
  border: 1px solid #eef1f5;
  border-radius: var(--radius-md);
  font-size: 0.88rem;
  color: var(--text-primary);
  line-height: 1.55;
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

/* ── GOOGLE MAPS TAB ──────────────────────────────────────────────── */
.editorial-text { margin: 0; color: var(--text-secondary); font-size: 0.88rem; line-height: 1.6; font-style: italic; }
.photo-scroll { display: flex; gap: 0.5rem; overflow-x: auto; scroll-snap-type: x mandatory; -webkit-overflow-scrolling: touch; padding-bottom: 0.3rem; }
.photo-scroll::-webkit-scrollbar { height: 4px; }
.photo-scroll::-webkit-scrollbar-thumb { background: #cbd5e1; border-radius: 4px; }
.photo-item { flex: 0 0 220px; height: 160px; border-radius: 10px; overflow: hidden; cursor: pointer; scroll-snap-align: start; border: 2px solid transparent; transition: border-color 0.15s, transform 0.15s; }
.photo-item:hover { transform: scale(1.02); }
.photo-item.active { border-color: var(--brand-blue); }
.photo-item img { width: 100%; height: 100%; object-fit: cover; }
.photo-attribution { margin-top: 0.35rem; color: var(--text-muted); font-size: 0.65rem; font-style: italic; }
.rating-row { display: flex; align-items: center; gap: 0.4rem; }
.rating-num { font-size: 1.15rem; font-weight: 800; color: #f59e0b; }
.rating-stars { display: flex; gap: 1px; }
.rating-stars .pi { font-size: 0.7rem; color: #f59e0b; }
.rating-count { color: var(--text-muted); font-size: 0.75rem; }
.info-link { color: #e63946; text-decoration: none; font-size: 0.85rem; font-weight: 600; display: inline-flex; align-items: center; gap: 0.3rem; }
.info-link:hover { text-decoration: underline; }
.types-wrap { display: flex; flex-wrap: wrap; gap: 0.35rem; }
.type-tag { font-size: 0.68rem !important; }
.hours-status { display: flex; align-items: center; gap: 0.4rem; margin-bottom: 0.75rem; }
.hours-dot { width: 8px; height: 8px; border-radius: 50%; }
.hours-dot.open { background: #22c55e; box-shadow: 0 0 6px rgba(34, 197, 94, 0.4); }
.hours-dot.closed { background: #ef4444; }
.hours-list { display: grid; gap: 0.3rem; }
.hours-row { font-size: 0.82rem; color: var(--text-secondary); line-height: 1.4; }
.hours-toggle { background: none; border: none; color: #e63946; font-size: 0.75rem; font-weight: 600; cursor: pointer; padding: 0.2rem 0; text-align: left; }
.hours-toggle:hover { text-decoration: underline; }
.reviews-list { display: grid; gap: 0.85rem; }
.review-item { padding-bottom: 0.75rem; border-bottom: 1px solid #f0f3f7; }
.review-item:last-child { border-bottom: none; padding-bottom: 0; }
.review-header { display: flex; align-items: center; gap: 0.6rem; }
.review-avatar { width: 34px; height: 34px; border-radius: 50%; object-fit: cover; }
.review-avatar-placeholder { width: 34px; height: 34px; border-radius: 50%; display: grid; place-items: center; background: #e2e8f0; color: var(--text-muted); font-size: 0.75rem; font-weight: 700; flex-shrink: 0; }
.review-meta { flex: 1; min-width: 0; }
.review-meta strong { font-size: 0.82rem; color: var(--text-primary); }
.review-stars { display: flex; align-items: center; gap: 1px; }
.review-stars .pi { font-size: 0.55rem; color: #f59e0b; }
.review-time { color: var(--text-muted); font-size: 0.68rem; margin-left: 0.4rem; }
.review-text { margin: 0.35rem 0 0; color: var(--text-secondary); font-size: 0.82rem; line-height: 1.5; }
.reviews-maps-link {
  display: inline-flex; align-items: center; gap: 0.35rem;
  margin-top: 0.75rem; padding-top: 0.75rem; border-top: 1px solid #f0f3f7;
  color: #e63946; text-decoration: none; font-size: 0.78rem; font-weight: 600;
}
.reviews-maps-link:hover { text-decoration: underline; }
.address-maps-link { margin-top: 0.85rem; }

/* ── FIELD SOURCE BADGE ────────────────────────────────────────────── */
.fs-badge {
  display: inline-flex;
  align-items: center;
  padding: 0.1rem 0.35rem;
  border-radius: 4px;
  font-size: 0.52rem;
  font-weight: 700;
  letter-spacing: 0.04em;
  text-transform: uppercase;
  line-height: 1.3;
  vertical-align: middle;
  white-space: nowrap;
  margin-left: 0.25rem;
}
.fs-system {
  background: #10b981;
  color: #fff;
}
.fs-google {
  background: #ef4e5d;
  color: #fff;
}
.fs-manual {
  background: #f59e0b;
  color: #78350f;
}

/* ── RESPONSIVE ────────────────────────────────────────────────────── */
@media (max-width: 1024px) {
  .detail-grid { grid-template-columns: 1fr; }
}
@media (max-width: 768px) {
  .admin-page { padding: 1.25rem 1rem; }
  .page-heading { flex-direction: column; }
  .summary-strip { flex-direction: column; }
  .strip-item { border-right: none; border-bottom: 1px solid #f0f3f7; }
  .strip-item:last-child { border-bottom: none; }
  .info-grid { grid-template-columns: 1fr; }
  .contact-card { flex-direction: column; align-items: center; text-align: center; }
  .contact-details { justify-content: center; }
}
.company-data-card-standalone{display:flex !important;flex-direction:column !important}.company-data-card-standalone>.company-data-card-header{order:1!important}.company-data-card-standalone>.company-data-card-body{order:2!important}.company-data-card-standalone>.company-address-card-inline,.company-data-card-standalone>.company-kam-card-inline{order:3!important}.company-data-card-standalone>.company-contact-card-inside{order:99!important}
.company-data-card-standalone{gap:20px!important;padding-bottom:20px!important}
.company-data-card-standalone>.company-data-card-header{width:100%;margin:0!important}
.company-data-card-standalone>.company-data-card-body{width:100%;margin:0!important;padding:20px!important;gap:20px!important}
.company-data-card-standalone>.company-address-card-inline,.company-data-card-standalone>.company-kam-card-inline{margin:0 20px!important;width:calc(50% - 30px)!important;align-self:stretch}
.company-data-card-standalone>.company-contact-card-inside{width:calc(100% - 40px)!important;margin:0 20px!important}
.company-data-card-standalone{display:grid!important;grid-template-columns:minmax(0,1fr) minmax(0,1fr);align-items:start}
.company-data-card-standalone>.company-data-card-header,.company-data-card-standalone>.company-data-card-body,.company-data-card-standalone>.company-contact-card-inside{grid-column:1/-1}
.company-data-card-standalone>.company-address-card-inline{grid-column:1;margin:0 0 0 20px!important;width:calc(100% - 10px)!important}
.company-data-card-standalone>.company-kam-card-inline{grid-column:2;margin:0 20px 0 10px!important;width:calc(100% - 10px)!important}
@media (max-width:768px){.company-data-card-standalone{display:flex!important}.company-data-card-standalone>.company-address-card-inline,.company-data-card-standalone>.company-kam-card-inline{width:calc(100% - 40px)!important;margin:0 20px!important}}
@media (max-width:768px){.company-data-card-standalone>.company-data-card-body{grid-template-columns:1fr!important}.company-data-card-standalone>.company-address-card-inline,.company-data-card-standalone>.company-kam-card-inline{width:calc(100% - 40px)!important}}
.customer-detail-side-information :deep(.ai-card),.customer-detail-side-information :deep(.tanya-card),.customer-detail-side-information :deep(.pc-wrap),.customer-detail-information-grid :deep(.customer-opportunity-card){font-family:Inter,ui-sans-serif,system-ui,sans-serif;color:#0f172a}
.customer-detail-side-information :deep(.ai-card-head h2),.customer-detail-side-information :deep(.tanya-head h2),.customer-detail-side-information :deep(.pc-header),.customer-detail-information-grid :deep(.customer-opportunity-card h2){font-size:17px!important;font-weight:700!important;line-height:1.35!important;letter-spacing:-.01em}
.customer-detail-side-information :deep(.ai-eyebrow),.customer-detail-side-information :deep(.tanya-head .ai-eyebrow),.customer-detail-information-grid :deep(.customer-opportunity-card header p){font-size:11px!important;font-weight:700!important;letter-spacing:.08em!important;line-height:1.3!important}
.customer-detail-side-information :deep(p),.customer-detail-side-information :deep(span),.customer-detail-information-grid :deep(.customer-opportunity-card p),.customer-detail-information-grid :deep(.customer-opportunity-card span){line-height:1.55}
.customer-detail-side-information :deep(button),.customer-detail-side-information :deep(textarea),.customer-detail-information-grid :deep(.customer-opportunity-card button){font-family:inherit}
</style>
