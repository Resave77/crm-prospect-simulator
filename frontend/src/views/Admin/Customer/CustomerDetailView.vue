<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import Button from 'primevue/button'
import Message from 'primevue/message'
import Tag from 'primevue/tag'
import { useCrmStore } from '../../../stores/crm'
import type { CustomerDetail } from '../../../types/crm'

const route = useRoute()
const router = useRouter()
const crm = useCrmStore()
const error = ref('')
const detail = ref<CustomerDetail | null>(null)
const activeTab = ref('overview')

const tabs = [
  { key: 'overview', label: 'Overview', icon: 'pi pi-id-card' },
  { key: 'contacts', label: 'Contacts', icon: 'pi pi-users' },
  { key: 'company', label: 'Company', icon: 'pi pi-building' },
  { key: 'address', label: 'Address', icon: 'pi pi-map' },
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
    detail.value = await crm.loadAdminCustomer(String(route.params.id))
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
            <i class="pi pi-arrow-left" /> Back to Customer List
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
            <span>Segment</span>
            <Tag :value="detail.customer.segment" :severity="segmentSeverity(detail.customer.segment)" />
          </div>
        </div>
        <div class="strip-item">
          <i class="pi pi-folder" />
          <div>
            <span>Category</span>
            <strong>{{ detail.customer.category || '—' }}</strong>
          </div>
        </div>
        <div class="strip-item">
          <i class="pi pi-map-marker" />
          <div>
            <span>Region</span>
            <strong>{{ detail.customer.region || '—' }}</strong>
          </div>
        </div>
        <div class="strip-item">
          <i class="pi pi-user" />
          <div>
            <span>Sales Executive</span>
            <strong>{{ detail.customer.salesExecutiveName || 'Unassigned' }}</strong>
          </div>
        </div>
        <div class="strip-item">
          <i class="pi pi-calendar" />
          <div>
            <span>Converted</span>
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
              <span class="info-label">Customer Code</span>
              <code class="code-tag code-blue">{{ detail.customer.customerCode }}</code>
            </div>
            <div class="info-item">
              <span class="info-label">Site Name</span>
              <strong>{{ detail.customer.name }}</strong>
            </div>
            <div class="info-item">
              <span class="info-label">Segment</span>
              <Tag :value="detail.customer.segment" :severity="segmentSeverity(detail.customer.segment)" />
            </div>
            <div class="info-item">
              <span class="info-label">Category</span>
              <strong>{{ detail.customer.category || '—' }}</strong>
            </div>
            <div class="info-item">
              <span class="info-label">Region</span>
              <strong>{{ detail.customer.region || '—' }}</strong>
            </div>
            <div class="info-item">
              <span class="info-label">Sales Executive</span>
              <strong>{{ detail.customer.salesExecutiveName || 'Unassigned' }}</strong>
            </div>
            <div class="info-item">
              <span class="info-label">Converted At</span>
              <strong>{{ formatDateTime(detail.customer.convertedAt) }}</strong>
            </div>
            <div class="info-item">
              <span class="info-label">Last Updated</span>
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
              <span class="info-label">Company Name</span>
              <strong>{{ detail.parentCompany.name }}</strong>
            </div>
            <div class="info-item">
              <span class="info-label">Company Code</span>
              <code class="code-tag">{{ detail.parentCompany.parentCode }}</code>
            </div>
            <div class="info-item">
              <span class="info-label">Source Prospect</span>
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
            <span class="contact-position" v-if="contact.position">{{ contact.position }}</span>
            <div class="contact-details">
              <span v-if="contact.phone"><i class="pi pi-phone" /> {{ contact.phone }}</span>
              <span v-if="contact.email"><i class="pi pi-envelope" /> {{ contact.email }}</span>
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
              <span class="info-label">Company Name</span>
              <button class="link-btn" @click="router.push(`/admin/companies/${detail.parentCompany.parentCode}`)">{{ detail.parentCompany.name }}</button>
            </div>
            <div class="info-item">
              <span class="info-label">Parent Code</span>
              <code class="code-tag">{{ detail.parentCompany.parentCode }}</code>
            </div>
            <div class="info-item" v-if="detail.parentCompany.npwpNumber">
              <span class="info-label">NPWP Number</span>
              <code class="code-tag">{{ detail.parentCompany.npwpNumber }}</code>
            </div>
            <div class="info-item" v-if="detail.parentCompany.npwpName">
              <span class="info-label">NPWP Name</span>
              <strong>{{ detail.parentCompany.npwpName }}</strong>
            </div>
            <div class="info-item" v-if="detail.parentCompany.npwpAddress">
              <span class="info-label">NPWP Address</span>
              <strong>{{ detail.parentCompany.npwpAddress }}</strong>
            </div>
            <div class="info-item" v-if="detail.parentCompany.termOfPayment">
              <span class="info-label">Term of Payment</span>
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
                  <span v-if="contact.phone"><i class="pi pi-phone" /> {{ contact.phone }}</span>
                  <span v-if="contact.email"><i class="pi pi-envelope" /> {{ contact.email }}</span>
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
                <span class="info-label">Province</span>
                <strong>{{ detail.customer.address.province }}</strong>
              </div>
              <div class="info-item" v-if="detail.customer.address.district">
                <span class="info-label">District</span>
                <strong>{{ detail.customer.address.district }}</strong>
              </div>
              <div class="info-item" v-if="detail.customer.address.subDistrict">
                <span class="info-label">Sub-District</span>
                <strong>{{ detail.customer.address.subDistrict }}</strong>
              </div>
              <div class="info-item" v-if="detail.customer.address.village">
                <span class="info-label">Village</span>
                <strong>{{ detail.customer.address.village }}</strong>
              </div>
              <div class="info-item" v-if="detail.customer.address.latitude != null">
                <span class="info-label">Coordinates</span>
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
                <span class="info-label">Province</span>
                <strong>{{ detail.parentCompany.address.province }}</strong>
              </div>
              <div class="info-item" v-if="detail.parentCompany.address.district">
                <span class="info-label">District</span>
                <strong>{{ detail.parentCompany.address.district }}</strong>
              </div>
              <div class="info-item" v-if="detail.parentCompany.address.subDistrict">
                <span class="info-label">Sub-District</span>
                <strong>{{ detail.parentCompany.address.subDistrict }}</strong>
              </div>
              <div class="info-item" v-if="detail.parentCompany.address.village">
                <span class="info-label">Village</span>
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
  gap: 0.35rem;
  color: var(--brand-blue);
  font-size: 0.8rem;
  font-weight: 600;
  text-decoration: none;
  cursor: pointer;
  background: none;
  border: none;
  padding: 0;
  font: inherit;
  transition: opacity 0.15s;
}
.back-link:hover { opacity: 0.8; }
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
