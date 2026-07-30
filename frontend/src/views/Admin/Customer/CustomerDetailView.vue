<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import Button from 'primevue/button'
import Message from 'primevue/message'
import Tag from 'primevue/tag'
import { useCrmStore } from '../../../stores/crm'
import { getAdminCustomerPlaceDetails } from '../../../api/crm'
import type { CustomerDetail, PlaceDetails } from '../../../types/crm'
import { priceLevelLabel, businessStatusLabel, businessStatusSeverity, stars } from '../../../utils/placeLabels'

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
const error = ref('')
const detail = ref<CustomerDetail | null>(null)
const placeDetails = ref<PlaceDetails | null>(null)
const activeTab = ref('overview')
const activePhotoIdx = ref(0)
const showAllHours = ref(false)

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
  } catch (e) {
    error.value = crm.errorMessage(e)
  }
})
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
      <header class="page-heading">
        <div class="page-title-wrapper">
          <button class="back-link" @click="router.push('/admin/customers')">
            <i class="pi pi-arrow-left" />
          </button>
          <span class="eyebrow">Customer Detail</span>
          <div class="title-row">
            <h1>{{ detail.customer.name }}</h1>
            <Tag value="Active" severity="success" />
          </div>
          <div class="subtitle-row">
            <code class="code-tag code-blue">{{ detail.customer.customerCode }}</code>
            <span class="muted">&mdash;</span>
            <button class="link-btn" @click="router.push(`/admin/companies/${detail.parentCompany.parentCode}`)">{{ detail.parentCompany.name }}</button>
          </div>
        </div>
        <div class="page-heading-actions">
          <Button label="Edit" icon="pi pi-pencil" size="small" @click="router.push(`/admin/customers/${route.params.id}/edit`)" />
          <Button label="Delete" icon="pi pi-trash" severity="danger" text size="small" />
        </div>
      </header>

      <!-- SUMMARY STRIP -->
      <div class="summary-strip">
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
          <!-- Editorial Summary -->
          <div v-if="placeDetails.editorialSummary" class="detail-card" style="grid-column: 1 / -1;">
            <h3 class="card-heading"><i class="pi pi-info-circle" /> About this place</h3>
            <p class="editorial-text">{{ placeDetails.editorialSummary }}</p>
          </div>

          <!-- Photos -->
          <div v-if="placeDetails.photos?.length" class="detail-card" style="grid-column: 1 / -1;">
            <h3 class="card-heading"><i class="pi pi-images" /> Photos</h3>
            <div class="photo-scroll">
              <div
                v-for="(photo, idx) in placeDetails.photos"
                :key="photo.name"
                class="photo-item"
                :class="{ active: idx === activePhotoIdx }"
                @click="activePhotoIdx = idx"
              >
                <img :src="photo.photoUrl" :alt="`Photo ${idx + 1}`" loading="lazy" @error="($event.target as HTMLImageElement).style.display='none'" />
              </div>
            </div>
            <div v-if="placeDetails.photos[activePhotoIdx]?.attribution" class="photo-attribution">
              Photo: {{ placeDetails.photos[activePhotoIdx].attribution }}
            </div>
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
              <div v-if="placeDetails.googleMapsUrl" class="info-item">
                <span class="info-label">Google Maps</span>
                <a :href="placeDetails.googleMapsUrl" target="_blank" rel="noopener" class="info-link"><i class="pi pi-map" /> View on Google Maps</a>
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
            <h3 class="card-heading"><i class="pi pi-comments" /> Reviews</h3>
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
            <span>{{ (contact.name || 'U').charAt(0).toUpperCase() }}</span>
          </div>
          <div class="contact-info">
            <h4>{{ contact.name || 'Unnamed Contact' }}</h4>
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
.back-link {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 2rem;
  height: 2rem;
  color: var(--brand-blue);
  background: var(--brand-blue-bg);
  border: 1px solid transparent;
  border-radius: var(--radius-md);
  text-decoration: none;
  cursor: pointer;
  font-size: 0.9rem;
  transition: background 0.15s, border-color 0.15s;
}
.back-link:hover { background: #dbeafe; border-color: var(--brand-blue); }
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
  background: #eff6ff;
  color: #2563eb;
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
  color: #2563eb;
  transition: color 0.15s;
}
.link-btn:hover { color: #1d4ed8; text-decoration: underline; }

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
  background: #eff6ff;
  color: #2563eb;
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
  background: #eef2ff;
  color: #6366f1;
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
.info-link { color: #2563eb; text-decoration: none; font-size: 0.85rem; font-weight: 600; display: inline-flex; align-items: center; gap: 0.3rem; }
.info-link:hover { text-decoration: underline; }
.types-wrap { display: flex; flex-wrap: wrap; gap: 0.35rem; }
.type-tag { font-size: 0.68rem !important; }
.hours-status { display: flex; align-items: center; gap: 0.4rem; margin-bottom: 0.75rem; }
.hours-dot { width: 8px; height: 8px; border-radius: 50%; }
.hours-dot.open { background: #22c55e; box-shadow: 0 0 6px rgba(34, 197, 94, 0.4); }
.hours-dot.closed { background: #ef4444; }
.hours-list { display: grid; gap: 0.3rem; }
.hours-row { font-size: 0.82rem; color: var(--text-secondary); line-height: 1.4; }
.hours-toggle { background: none; border: none; color: #2563eb; font-size: 0.75rem; font-weight: 600; cursor: pointer; padding: 0.2rem 0; text-align: left; }
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
  background: #3b82f6;
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
</style>
