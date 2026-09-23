<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import Button from 'primevue/button'
import Message from 'primevue/message'
import Tag from 'primevue/tag'
import Dialog from 'primevue/dialog'
import { useToast } from 'primevue/usetoast'
import { useCustomerListStore } from '../../../stores/customerList'
import { deleteCustomer, getParentCompany } from '../../../api/crm'
import type { ParentCompany } from '../../../types/crm'

const route = useRoute()
const router = useRouter()
const store = useCustomerListStore()
const toast = useToast()
const error = ref('')
const loading = ref(true)
const activeTab = ref('sites')
const deleteDialogVisible = ref(false)
const deleting = ref(false)
const company = ref<ParentCompany | null>(null)

const tabs = [
  { key: 'sites', label: 'Sites', icon: 'pi pi-map-marker' },
  { key: 'info', label: 'Company Info', icon: 'pi pi-building' },
]

const code = computed(() => route.params.id as string)

const sites = computed(() => store.allCustomers.filter((c) => c.parentCode === code.value))

const companyName = computed(() => sites.value[0]?.parentCompanyName || code.value)

const totalSites = computed(() => sites.value.length)
const regions = computed(() => [...new Set(sites.value.map((s) => s.region).filter(Boolean))])
const segments = computed(() => [...new Set(sites.value.map((s) => s.segment).filter(Boolean))])

function formatDate(dateStr: string) {
  if (!dateStr) return '—'
  return new Date(dateStr).toLocaleDateString('en-GB', { day: '2-digit', month: 'short', year: 'numeric' })
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
    if (store.allCustomers.length === 0) await store.fetchCustomers()
    company.value = await getParentCompany(code.value)
  } catch (e) { error.value = store.errorMessage(e) }
  finally { loading.value = false }
})

function confirmDelete() {
  deleteDialogVisible.value = true
}

async function executeDelete() {
  deleting.value = true
  try {
    for (const site of sites.value) {
      await deleteCustomer(site.id)
    }
    store.allCustomers = store.allCustomers.filter((c) => c.parentCode !== code.value)
    toast.add({ severity: 'success', summary: 'Company deleted', detail: `${companyName.value} and its ${sites.value.length} site(s) have been removed.`, life: 3000 })
    router.push('/admin/customers')
  } catch (e) {
    error.value = store.errorMessage(e)
    deleteDialogVisible.value = false
  } finally {
    deleting.value = false
  }
}
</script>

<template>
  <section class="admin-page">
    <Message v-if="error" severity="error" :closable="false">{{ error }}</Message>

    <div v-if="loading" class="state-box">
      <i class="pi pi-spin pi-spinner state-icon" />
      <span>Loading company details...</span>
    </div>

    <template v-else-if="sites.length">
      <!-- PAGE HEADER -->
      <header class="page-heading">
        <div class="page-title-wrapper">
          <span class="eyebrow">Company Detail</span>
          <div class="title-row">
            <h1>{{ companyName }}</h1>
            <Tag value="Active" severity="success" />
          </div>
          <div class="subtitle-row">
            <code class="code-tag">{{ code }}</code>
          </div>
        </div>
        <div class="page-heading-actions">
          <Button label="Edit" icon="pi pi-pencil" size="small" @click="router.push(`/admin/companies/${code}/edit`)" />
          <Button label="Delete" icon="pi pi-trash" severity="danger" text size="small" @click="confirmDelete" />
        </div>
      </header>

      <section v-if="company" class="company-reference-card">
        <div class="company-reference-header">
          <div><p class="company-eyebrow">Company Data</p><p class="company-description">Data parent company untuk customer site ini.</p></div>
          <div class="company-reference-actions"><button type="button" @click="router.push(`/admin/companies/${code}/edit`)">Edit Company Detail <i class="pi pi-pencil" /></button></div>
          <div class="company-reference-title"><div><h2>{{ company.name }}</h2><p>{{ company.parentCode }}</p></div><span>Key Account</span></div>
        </div>
        <div class="company-reference-grid">
          <div class="company-reference-section"><div class="company-section-header"><h3>Company Address</h3><p>Legal and billing address used for company and tax records.</p></div><div class="company-data-list"><div><span>Company Address</span><strong>{{ company.address?.previewAddress || '—' }}</strong></div><div><span>Province</span><strong>{{ company.address?.province || '—' }}</strong></div><div><span>District</span><strong>{{ company.address?.district || '—' }}</strong></div><div><span>Latitude</span><strong>{{ company.address?.latitude ?? '—' }}</strong></div><div><span>Longitude</span><strong>{{ company.address?.longitude ?? '—' }}</strong></div></div></div>
          <div class="company-reference-section"><div class="company-section-header"><h3>Company KAM</h3><p>Company-level key account manager ownership.</p></div><div class="company-data-list"><div><span>Assigned KAM</span><strong>{{ company.kamAssignments?.[0]?.ownerName || '—' }}</strong></div><div><span>Status</span><strong>{{ company.kamAssignments?.[0]?.end ? 'Selesai' : 'Aktif' }}</strong></div></div></div>
          <div class="company-reference-section company-contact-reference"><div class="company-section-header"><h3>Company Contact Information</h3><p>General company contacts for coordination and follow up.</p></div><div class="company-contact-list"><div v-for="(contact, index) in company.contacts" :key="`${contact.email}-${index}`" class="company-contact-item"><div class="company-contact-heading"><div><span>Contact {{ index + 1 }}</span><strong>{{ contact.name || '—' }}</strong></div><em>{{ contact.position || 'General' }}</em></div><div class="company-contact-fields"><div><span>Phone Number</span><strong>{{ contact.phone || '—' }}</strong></div><div><span>Email Address</span><strong>{{ contact.email || '—' }}</strong></div></div></div><p v-if="!company.contacts?.length" class="company-contact-empty">No company contact information available.</p></div></div>
          <div class="company-reference-section"><div class="company-section-header"><h3>Company Information</h3><p>Legal identity of the parent company.</p></div><div class="company-data-list"><div><span>Company Name</span><strong>{{ company.name }}</strong></div><div><span>Company Code</span><strong>{{ company.parentCode }}</strong></div><div><span>Company Tier</span><strong>Key Account</strong></div><div><span>Total Sites</span><strong>{{ totalSites }}</strong></div></div></div>
          <div class="company-reference-section"><div class="company-section-header"><h3>Tax Information</h3><p>Company NPWP data synced from company name and address.</p></div><div class="company-data-list"><div><span>Company NPWP Name</span><strong>{{ company.npwpName || company.name }}</strong></div><div><span>Company NPWP Address</span><strong>{{ company.npwpAddress || company.address?.previewAddress || '—' }}</strong></div><div><span>Company NPWP Number</span><strong>{{ company.npwpNumber || '—' }}</strong></div></div></div>
        </div>
      </section>

      <section v-if="company" class="company-sites-reference">
        <div class="company-sites-header">
          <h2>Customer Sites ({{ sites.length }})</h2>
          <p>All outlet, branch, store, or office records under this company.</p>
        </div>
        <div class="company-sites-list">
          <button v-for="site in sites" :key="site.id" type="button" class="company-site-item" @click="router.push(`/admin/customers/${site.id}`)">
            <div class="company-site-main"><strong>{{ site.name }}</strong><span>{{ site.customerCode }}</span></div>
            <div class="company-site-meta"><span>{{ site.category || '—' }}</span><span>{{ site.region || '—' }}</span><i class="pi pi-chevron-right" /></div>
          </button>
          <p v-if="!sites.length" class="company-sites-empty">No customer sites available.</p>
        </div>
      </section>

      <!-- SUMMARY STRIP -->
      <div class="summary-strip">
        <div class="strip-item">
          <i class="pi pi-map-marker" />
          <div>
            <span>Total Sites</span>
            <strong>{{ totalSites }}</strong>
          </div>
        </div>
        <div class="strip-item">
          <i class="pi pi-tag" />
          <div>
            <span>Segments</span>
            <strong>{{ segments.length ? segments.join(', ') : '—' }}</strong>
          </div>
        </div>
        <div class="strip-item">
          <i class="pi pi-map" />
          <div>
            <span>Regions</span>
            <strong>{{ regions.length ? regions.join(', ') : '—' }}</strong>
          </div>
        </div>
        <div class="strip-item">
          <i class="pi pi-shield" />
          <div>
            <span>Tier</span>
            <strong>Tier 1</strong>
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

      <!-- =================== SITES TAB =================== -->
      <div v-if="activeTab === 'sites'" class="panel-stack">
        <div class="table-panel">
          <div v-if="!sites.length" class="state-box">
            <div class="state-icon-wrap"><i class="pi pi-map-marker" /></div>
            <strong>No sites found</strong>
            <span class="muted">This company has no associated customer sites.</span>
          </div>
          <div v-else class="table-scroll">
            <table class="data-table">
              <thead>
                <tr>
                  <th>Code</th>
                  <th>Site Name</th>
                  <th>Region</th>
                  <th>Segment</th>
                  <th>Category</th>
                  <th>Sales Executive</th>
                  <th>Converted</th>
                  <th class="th-action">Actions</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="site in sites" :key="site.id">
                  <td><code class="code-tag code-blue">{{ site.customerCode }}</code></td>
                  <td>
                    <button class="link-btn" @click="router.push(`/admin/customers/${site.id}`)">
                      {{ site.name }}
                    </button>
                  </td>
                  <td><span class="cell-text">{{ site.region || '—' }}</span></td>
                  <td><Tag :value="site.segment" :severity="segmentSeverity(site.segment)" /></td>
                  <td><span class="cell-text">{{ site.category || '—' }}</span></td>
                  <td><span class="cell-text">{{ site.salesExecutiveName || 'Unassigned' }}</span></td>
                  <td><span class="cell-date">{{ formatDate(site.convertedAt) }}</span></td>
                  <td class="td-action">
                    <div class="row-actions">
                      <Button icon="pi pi-eye" text rounded size="small" class="act-view" title="View" @click="router.push(`/admin/customers/${site.id}`)" />
                      <Button icon="pi pi-pencil" text rounded size="small" class="act-edit" title="Edit" @click="router.push(`/admin/customers/${site.id}/edit`)" />
                    </div>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </div>

      <!-- =================== COMPANY INFO TAB =================== -->
      <div v-if="activeTab === 'info'" class="detail-grid">
        <div class="detail-card">
          <h3 class="card-heading">
            <i class="pi pi-building" /> Corporate Details
          </h3>
          <div class="info-grid">
            <div class="info-item full">
              <span class="info-label">Company Name</span>
              <strong>{{ companyName }}</strong>
            </div>
            <div class="info-item">
              <span class="info-label">Company Code</span>
              <code class="code-tag">{{ code }}</code>
            </div>
            <div class="info-item">
              <span class="info-label">Total Sites</span>
              <strong>{{ totalSites }}</strong>
            </div>
            <div class="info-item">
              <span class="info-label">Tier</span>
              <strong>Tier 1</strong>
            </div>
            <div class="info-item">
              <span class="info-label">Status</span>
              <Tag value="Active" severity="success" />
            </div>
            <div class="info-item">
              <span class="info-label">NPWP Number</span>
              <code class="code-tag">00.000.000.0-000.000</code>
            </div>
          </div>
        </div>

        <div class="detail-card">
          <h3 class="card-heading">
            <i class="pi pi-map" /> Regional Presence
          </h3>
          <div v-if="regions.length" class="region-list">
            <div v-for="region in regions" :key="region" class="region-row">
              <i class="pi pi-map-marker" />
              <strong>{{ region }}</strong>
              <span class="region-count">{{ sites.filter((s) => s.region === region).length }} sites</span>
            </div>
          </div>
          <div v-else class="empty-inline">
            <span>No regional data available.</span>
          </div>
        </div>
      </div>
    </template>

    <!-- EMPTY STATE -->
    <div v-else class="empty-card">
      <i class="pi pi-building" />
      <strong>Company not found</strong>
      <span class="muted">The requested company could not be located.</span>
    </div>

    <Dialog v-model:visible="deleteDialogVisible" header="Delete Company" modal :style="{ width: '400px' }">
      <p>Are you sure you want to delete <strong>{{ companyName }}</strong> and all <strong>{{ totalSites }}</strong> of its site(s)? This action cannot be undone.</p>
      <template #footer>
        <Button label="Cancel" severity="secondary" text @click="deleteDialogVisible = false" :disabled="deleting" />
        <Button label="Delete" severity="danger" icon="pi pi-trash" :loading="deleting" @click="executeDelete" />
      </template>
    </Dialog>
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
.page-title-wrapper { display: flex; flex-direction: column; gap: 0.15rem; }
.page-title-wrapper .eyebrow {
  font-size: 0.68rem;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.1em;
  color: var(--brand-green-light, #0b7766);
  margin-top: 0.5rem;
}
.title-row { display: flex; align-items: center; gap: 0.65rem; flex-wrap: wrap; }
.page-title-wrapper h1 {
  font-size: 1.65rem;
  font-weight: 800;
  color: var(--text-primary);
  margin: 0.2rem 0 0.15rem;
  letter-spacing: -0.03em;
}
.subtitle-row { display: flex; align-items: center; gap: 0.5rem; margin-top: 0.1rem; }
.page-heading-actions { display: flex; gap: 0.5rem; align-items: center; padding-top: 0.15rem; }
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
.strip-item i { color: var(--text-faint); font-size: 0.95rem; flex-shrink: 0; }
.strip-item div { display: flex; flex-direction: column; min-width: 0; }
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

/* ── PANEL STACK ──────────────────────────────────────────────────── */
.panel-stack { display: flex; flex-direction: column; gap: 1rem; }

/* ── TABLE ─────────────────────────────────────────────────────────── */
.table-panel {
  background: var(--surface-card);
  border: 1px solid var(--border-light);
  border-radius: var(--radius-lg);
  box-shadow: var(--shadow-xs);
  overflow: hidden;
}
.table-scroll { overflow-x: auto; }
.data-table { width: 100%; border-collapse: collapse; font-size: 0.85rem; }
.data-table thead th {
  background: var(--surface-subtle);
  color: var(--text-muted);
  font-size: 0.68rem;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.06em;
  padding: 0.75rem 0.95rem;
  border-bottom: 1px solid var(--border-light);
  white-space: nowrap;
  text-align: left;
}
.data-table tbody td {
  padding: 0.75rem 0.95rem;
  border-bottom: 1px solid #f0f3f7;
  color: var(--text-primary);
  vertical-align: middle;
}
.data-table tbody tr:last-child td { border-bottom: none; }
.data-table tbody tr { transition: background var(--transition-fast); }
.data-table tbody tr:hover { background: #f8fafc; }
.th-action { width: 100px; text-align: center; }

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
.cell-text { font-size: 0.84rem; color: var(--text-secondary); }
.cell-date { font-size: 0.8rem; color: var(--text-muted); white-space: nowrap; }

/* ── ROW ACTIONS ──────────────────────────────────────────────────── */
.td-action { text-align: center; }
.row-actions { display: flex; align-items: center; justify-content: center; gap: 0.15rem; }
.act-view { color: #e63946 !important; }
.act-view:hover { background: #fff0f1 !important; }
.act-edit { color: #059669 !important; }
.act-edit:hover { background: #ecfdf5 !important; }

/* ── DETAIL GRID ───────────────────────────────────────────────────── */
.detail-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 1.25rem; }
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
.card-heading i { color: var(--brand-blue); font-size: 0.9rem; }
.info-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 0.9rem; }
.info-item { display: flex; flex-direction: column; gap: 0.2rem; }
.info-item.full { grid-column: 1 / -1; }
.info-label { font-size: 0.68rem; font-weight: 700; text-transform: uppercase; letter-spacing: 0.05em; color: var(--text-muted); }
.info-item strong { font-size: 0.85rem; font-weight: 600; color: var(--text-primary); }

/* ── REGION LIST ───────────────────────────────────────────────────── */
.region-list { display: flex; flex-direction: column; gap: 0.5rem; }
.region-row {
  display: flex;
  align-items: center;
  gap: 0.65rem;
  padding: 0.65rem 0.85rem;
  background: #f8fafc;
  border: 1px solid #eef1f5;
  border-radius: var(--radius-md);
}
.region-row i { color: var(--brand-blue); font-size: 0.85rem; }
.region-row strong { font-size: 0.85rem; color: var(--text-primary); flex: 1; }
.region-count { font-size: 0.75rem; font-weight: 600; color: var(--text-muted); }

/* ── STATE / EMPTY ─────────────────────────────────────────────────── */
.state-box {
  min-height: 260px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 0.4rem;
  padding: 2rem;
  text-align: center;
  color: var(--text-muted);
}
.state-icon { font-size: 1.75rem; color: var(--brand-blue); margin-bottom: 0.25rem; }
.state-icon-wrap {
  width: 56px;
  height: 56px;
  border-radius: var(--radius-lg);
  background: var(--surface-subtle);
  display: grid;
  place-content: center;
  margin-bottom: 0.35rem;
}
.state-icon-wrap i { font-size: 1.4rem; color: var(--text-faint); }
.state-box strong { color: var(--text-primary); font-size: 0.95rem; }

.empty-card {
  min-height: 300px;
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
.empty-card i { font-size: 2rem; color: var(--text-faint); margin-bottom: 0.25rem; }
.empty-card strong { color: var(--text-primary); font-size: 0.95rem; }

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

/* ── RESPONSIVE ────────────────────────────────────────────────────── */
 .admin-page:has(.company-reference-card)>.page-heading,.admin-page:has(.company-reference-card)>.summary-strip,.admin-page:has(.company-reference-card)>.tabs-bar,.admin-page:has(.company-reference-card)>.panel-stack,.admin-page:has(.company-reference-card)>.detail-grid{display:none}.company-reference-card{overflow:hidden;border:1px solid #fecaca;border-radius:20px;background:#fff}.company-reference-header{border-bottom:1px solid #fee2e2;padding:20px}.company-eyebrow{margin:0;color:#94a3b8;font-size:11px;font-weight:700;letter-spacing:.08em;text-transform:uppercase}.company-description{margin:4px 0 0;color:#64748b;font-size:12px}.company-reference-actions{display:flex;flex-direction:column;align-items:flex-end;gap:6px;float:right;margin-top:-34px}.company-reference-actions button{border:0;background:transparent;color:#dc2626;font-size:13px;font-weight:600;cursor:pointer}.company-reference-title{display:flex;align-items:center;justify-content:space-between;gap:12px;margin-top:24px}.company-reference-title h2{margin:0;color:#0f172a;font-size:17px}.company-reference-title p{margin:4px 0 0;color:#64748b;font-size:12px}.company-reference-title span{border:1px solid #ddd6fe;border-radius:999px;background:#ede9fe;padding:4px 10px;color:#7c3aed;font-size:11px;font-weight:600}.company-reference-grid{display:grid;grid-template-columns:1fr 1fr;gap:18px;padding:20px}.company-reference-section{overflow:hidden;border:1px solid #dfe6ef;border-radius:20px;background:#fff}.company-section-header{border-bottom:1px solid #eef2f7;padding:20px}.company-section-header h3{margin:0;color:#0f172a;font-size:17px}.company-section-header p{margin:4px 0 0;color:#64748b;font-size:12px}.company-data-list{display:grid;gap:10px;padding:20px}.company-data-list>div{display:flex;min-height:48px;align-items:center;justify-content:space-between;gap:16px;border-radius:14px;background:#f8fafc;padding:14px 16px}.company-data-list span{color:#64748b;font-size:12px}.company-data-list strong{max-width:62%;color:#0f172a;font-size:13px;text-align:right;overflow-wrap:anywhere}
@media (max-width: 1024px) { .detail-grid { grid-template-columns: 1fr; } }
@media (max-width: 768px) {
  .admin-page { padding: 1.25rem 1rem; }
  .page-heading { flex-direction: column; }
  .summary-strip { flex-direction: column; }
  .strip-item { border-right: none; border-bottom: 1px solid #f0f3f7; }
  .strip-item:last-child { border-bottom: none; }
  .info-grid { grid-template-columns: 1fr; }
}
.admin-page:has(.company-reference-card){padding:0!important;background:#f8fafc}.company-reference-card{border:0!important;border-radius:0!important;background:transparent!important;box-shadow:none!important}.company-reference-card>.company-reference-header{display:none!important}.company-reference-grid{grid-template-columns:minmax(0,1fr) minmax(0,1fr)!important;gap:18px!important;padding:20px 30px!important}.company-reference-section{border:1px solid #dfe6ef!important;border-radius:20px!important;box-shadow:0 14px 34px rgba(15,23,42,.06)}.company-reference-section .company-section-header{padding:20px!important}.company-reference-section .company-data-list{gap:10px!important;padding:20px!important}.company-reference-section .company-data-list>div{min-height:48px!important;padding:14px 16px!important}.company-reference-section .company-data-list span{font-size:12px!important}.company-reference-section .company-data-list strong{font-size:13px!important}@media(max-width:800px){.company-reference-grid{grid-template-columns:1fr!important;padding:16px!important}}
.company-reference-grid>.company-reference-section:nth-child(1){grid-column:1;grid-row:2!important}.company-reference-grid>.company-reference-section:nth-child(2){grid-column:2;grid-row:2!important}.company-reference-grid>.company-reference-section:nth-child(3){grid-column:1/-1;grid-row:3!important}.company-reference-grid>.company-reference-section:nth-child(4){grid-column:1;grid-row:1!important}.company-reference-grid>.company-reference-section:nth-child(5){grid-column:2;grid-row:1!important}.company-reference-grid>.company-contact-reference{grid-column:1/-1!important;grid-row:3!important}
.company-reference-card,.company-reference-card *{font-family:Inter,ui-sans-serif,system-ui,-apple-system,"Segoe UI",sans-serif}.company-contact-reference{grid-column:1/-1!important}.company-contact-list{display:grid;grid-template-columns:repeat(2,minmax(0,1fr));gap:12px;padding:20px}.company-contact-item{border:1px solid #dfe6ef;border-radius:16px;background:#f8fafc;padding:14px 16px}.company-contact-heading{display:flex;align-items:flex-start;justify-content:space-between;gap:12px}.company-contact-heading span,.company-contact-fields span{display:block;color:#94a3b8;font-size:10px;font-weight:700;letter-spacing:.08em;text-transform:uppercase}.company-contact-heading strong{display:block;margin-top:4px;color:#0f172a;font-size:14px;font-weight:700;line-height:20px}.company-contact-heading em{border:1px solid #dbeafe;border-radius:999px;background:#fff;padding:3px 10px;color:#1d4ed8;font-size:10px;font-style:normal;font-weight:600}.company-contact-fields{display:grid;grid-template-columns:repeat(2,minmax(0,1fr));gap:10px;margin-top:12px}.company-contact-fields>div{border-radius:14px;background:#fff;padding:10px 12px}.company-contact-fields strong{display:block;margin-top:5px;color:#0f172a;font-size:13px;font-weight:600;line-height:20px;overflow-wrap:anywhere}.company-contact-empty{margin:0;padding:20px;color:#64748b;font-size:12px}@media(max-width:700px){.company-contact-list{grid-template-columns:1fr}.company-contact-fields{grid-template-columns:1fr}}
@media(max-width:800px){.company-reference-grid>.company-reference-section{grid-column:1!important;grid-row:auto!important}}
.company-sites-reference{margin:0 30px 20px;overflow:hidden;border:1px solid #dfe6ef;border-radius:20px;background:#fff;box-shadow:0 14px 34px rgba(15,23,42,.06);font-family:Inter,ui-sans-serif,system-ui,sans-serif}.company-sites-header{border-bottom:1px solid #eef2f7;padding:18px 20px}.company-sites-header h2{margin:0;color:#0f172a;font-size:17px;font-weight:700}.company-sites-header p{margin:4px 0 0;color:#64748b;font-size:12px}.company-sites-list{display:grid;grid-template-columns:repeat(2,minmax(0,1fr));gap:12px;padding:20px}.company-site-item{display:flex;width:100%;align-items:center;justify-content:space-between;gap:16px;border:1px solid #dfe6ef;border-radius:16px;background:#f8fafc;padding:14px 16px;text-align:left;cursor:pointer;transition:border-color .2s,background .2s}.company-site-item:hover{border-color:#cbd5e1;background:#f1f5f9}.company-site-main{display:grid;min-width:0;gap:4px}.company-site-main strong{overflow:hidden;color:#0f172a;font-size:13px;font-weight:700;text-overflow:ellipsis;white-space:nowrap}.company-site-main span,.company-site-meta span{color:#64748b;font-size:11px}.company-site-meta{display:flex;align-items:center;justify-content:flex-end;gap:8px;white-space:nowrap}.company-site-meta i{color:#94a3b8;font-size:12px}.company-sites-empty{margin:0;padding:20px;color:#64748b;font-size:12px}@media(max-width:800px){.company-sites-reference{margin:0 16px 16px}.company-sites-list{grid-template-columns:1fr;padding:16px}}
</style>
