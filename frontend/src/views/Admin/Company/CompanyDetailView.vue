<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import Button from 'primevue/button'
import Message from 'primevue/message'
import Tag from 'primevue/tag'
import { useCustomerListStore } from '../../../stores/customerList'

const route = useRoute()
const router = useRouter()
const store = useCustomerListStore()
const error = ref('')
const loading = ref(true)
const activeTab = ref('sites')

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
  } catch (e) { error.value = store.errorMessage(e) }
  finally { loading.value = false }
})
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
          <button class="back-link" @click="router.push('/admin/companies')">
            <i class="pi pi-arrow-left" /> Back to Company List
          </button>
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
          <Button label="Delete" icon="pi pi-trash" severity="danger" text size="small" />
        </div>
      </header>

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
.back-link {
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
  color: var(--brand-blue);
  font-size: 0.8rem;
  font-weight: 600;
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
.code-blue { background: #eff6ff; color: #2563eb; }

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
  color: #2563eb;
  transition: color 0.15s;
}
.link-btn:hover { color: #1d4ed8; text-decoration: underline; }
.cell-text { font-size: 0.84rem; color: var(--text-secondary); }
.cell-date { font-size: 0.8rem; color: var(--text-muted); white-space: nowrap; }

/* ── ROW ACTIONS ──────────────────────────────────────────────────── */
.td-action { text-align: center; }
.row-actions { display: flex; align-items: center; justify-content: center; gap: 0.15rem; }
.act-view { color: #2563eb !important; }
.act-view:hover { background: #eff6ff !important; }
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
@media (max-width: 1024px) { .detail-grid { grid-template-columns: 1fr; } }
@media (max-width: 768px) {
  .admin-page { padding: 1.25rem 1rem; }
  .page-heading { flex-direction: column; }
  .summary-strip { flex-direction: column; }
  .strip-item { border-right: none; border-bottom: 1px solid #f0f3f7; }
  .strip-item:last-child { border-bottom: none; }
  .info-grid { grid-template-columns: 1fr; }
}
</style>
