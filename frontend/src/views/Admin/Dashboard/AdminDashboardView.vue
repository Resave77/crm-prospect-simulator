<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import Message from 'primevue/message'
import Tag from 'primevue/tag'
import { useAuthStore } from '../../../stores/auth'
import YummySocialLinks from '../../../components/layout/YummySocialLinks.vue'
import { useCrmStore } from '../../../stores/crm'
import { getAdminVisits } from '../../../api/crm'
import type { ProspectStatus, VisitMonitoringItem } from '../../../types/crm'

const auth = useAuthStore()
const crm = useCrmStore()
const error = ref('')
const loading = ref(true)
const todayVisits = ref<VisitMonitoringItem[]>([])

const todayKey = new Date().toLocaleDateString('en-CA')
const formattedDate = new Intl.DateTimeFormat(undefined, { weekday: 'short', day: 'numeric', month: 'short', year: 'numeric' }).format(new Date())
const stages: ProspectStatus[] = ['NEW_LEAD', 'CONTACTED', 'INTERESTED', 'QUALIFIED', 'PROPOSAL_SENT', 'NEGOTIATION', 'WON']

const active = computed(() => crm.pipeline.filter((item) => !['LOST', 'CONVERTED', 'WON'].includes(item.status)).length)
const won = computed(() => crm.pipeline.filter((item) => item.status === 'WON' || item.status === 'CONVERTED').length)
const stageCounts = computed(() => stages.map((status) => ({ status, count: crm.pipeline.filter((item) => item.status === status).length })))
const maxStageCount = computed(() => Math.max(1, ...stageCounts.value.map((entry) => entry.count)))
const recentPipeline = computed(() => [...crm.pipeline].sort((a, b) => new Date(b.updatedAt).getTime() - new Date(a.updatedAt).getTime()).slice(0, 5))
const todayAssignments = computed(() => todayVisits.value.slice(0, 4))

function statusSeverity(status: ProspectStatus): 'success' | 'danger' | 'warn' | 'info' | 'secondary' {
  if (status === 'WON' || status === 'CONVERTED') return 'success'
  if (status === 'LOST') return 'danger'
  if (status === 'QUALIFIED' || status === 'NEGOTIATION') return 'warn'
  if (status === 'NEW_LEAD') return 'secondary'
  return 'info'
}

function stageLabel(status: string): string {
  const labels: Record<string, string> = { NEW_LEAD: 'Lead Baru', CONTACTED: 'Dihubungi', INTERESTED: 'Tertarik', QUALIFIED: 'Terkualifikasi', PROPOSAL_SENT: 'Proposal Dikirim', NEGOTIATION: 'Negosiasi', WON: 'Berhasil' }
  return labels[status] || status.replaceAll('_', ' ')
}

onMounted(async () => {
  try {
    const [, , visits] = await Promise.all([
      crm.loadPipeline(),
      crm.loadAdminCustomers(),
      getAdminVisits({ dateFrom: todayKey, dateTo: todayKey, salesExecutiveId: '', customerName: '', radiusStatus: 'ALL' }),
    ])
    todayVisits.value = visits
  } catch (caught) {
    error.value = crm.errorMessage(caught)
  } finally {
    loading.value = false
  }
})
</script>

<template>
  <section class="erp-dashboard-shell">
    <Message v-if="error" severity="error" closable @close="error = ''">{{ error }}</Message>
    <div class="erp-dashboard-hero"><div class="erp-hero-icon"><i class="pi pi-chart-bar" /></div><div class="erp-hero-copy"><span>OVERVIEW</span><h1>Dashboard CRM</h1><p>Ringkasan cepat aktivitas CRM dan perkembangan pipeline untuk navigasi harian.</p></div><div class="erp-hero-actions"><YummySocialLinks /><RouterLink class="erp-date-pill" to="/admin/visit-monitoring"><i class="pi pi-calendar" /> {{ formattedDate }}</RouterLink></div></div>
    <div class="erp-stat-grid">
      <RouterLink to="/admin/customers" class="erp-stat-card"><span>Total Pelanggan</span><strong>{{ crm.adminCustomers.length }}</strong></RouterLink>
      <RouterLink to="/admin/visit-monitoring" class="erp-stat-card"><span>Kunjungan Hari Ini</span><strong>{{ todayVisits.length }}</strong></RouterLink>
      <RouterLink to="/admin/prospects/pipeline" class="erp-stat-card"><span>Total Prospek</span><strong>{{ crm.pipeline.length }}</strong></RouterLink>
      <RouterLink to="/admin/prospects/list" class="erp-stat-card"><span>Prospek Berhasil</span><strong>{{ won }}</strong></RouterLink>
    </div>
    <section class="erp-dashboard-panel quick-links"><header><div><h2>Quick Links</h2><p>Langsung lompat ke menu CRM yang paling sering dipakai.</p></div><RouterLink to="/admin/prospects/pipeline">View All <i class="pi pi-arrow-right" /></RouterLink></header><div class="erp-link-grid"><RouterLink to="/admin/customers"><i class="pi pi-users" /><span>Customer Existing</span><i class="pi pi-arrow-right" /></RouterLink><RouterLink to="/admin/prospects/list"><i class="pi pi-list" /><span>Customer Prospect</span><i class="pi pi-arrow-right" /></RouterLink><RouterLink to="/admin/prospects/pipeline"><i class="pi pi-chart-bar" /><span>Prospect Pipeline</span><i class="pi pi-arrow-right" /></RouterLink><RouterLink to="/admin/visit-monitoring"><i class="pi pi-map-marker" /><span>Visit Monitoring</span><i class="pi pi-arrow-right" /></RouterLink></div></section>
    <div class="erp-bottom-grid"><section class="erp-dashboard-panel"><h2><i class="pi pi-chart-bar" /> Pipeline Activity</h2><div class="erp-list"><div v-for="entry in stageCounts.slice(0, 4)" :key="entry.status"><span>{{ stageLabel(entry.status) }}</span><strong>{{ entry.count }}</strong></div></div></section><section class="erp-dashboard-panel"><h2><i class="pi pi-calendar" /> Recent Visits</h2><div class="erp-list"><div v-for="item in todayAssignments" :key="item.id"><span>{{ item.customerName }}</span><strong>{{ new Date(item.checkInAt).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' }) }}</strong></div><div v-if="!todayAssignments.length"><span>Belum ada kunjungan hari ini</span></div></div></section></div>
  </section>
  <section class="admin-dashboard">
    <Message v-if="error" severity="error" closable @close="error = ''">{{ error }}</Message>
    <div class="dashboard-header">
      <div><span class="dashboard-eyebrow">Ringkasan workspace</span><h1>Dashboard CRM</h1><p>Pantau aktivitas sales dan perkembangan pipeline, {{ auth.user?.fullName }}.</p></div>
      <div class="dashboard-header-actions"><YummySocialLinks /><RouterLink class="date-control" to="/admin/visit-monitoring"><i class="pi pi-calendar" /> {{ formattedDate }}</RouterLink></div>
    </div>
    <div v-if="loading" class="metric-grid dashboard-skeletons"><div v-for="n in 4" :key="n" class="metric-card skeleton-metric"><i /><strong /><small /></div></div>
    <div v-else class="metric-grid">
      <RouterLink to="/admin/customers" class="metric-card"><span>Total Pelanggan<i class="pi pi-users" /></span><strong>{{ crm.adminCustomers.length }}</strong><small>Data pelanggan yang sudah dikonversi</small></RouterLink>
      <RouterLink to="/admin/visit-monitoring" class="metric-card"><span>Kunjungan Hari Ini<i class="pi pi-map-marker" /></span><strong>{{ todayVisits.length }}</strong><small>Kunjungan lapangan yang tercatat</small></RouterLink>
      <RouterLink to="/admin/prospects/pipeline" class="metric-card"><span>Total Prospek<i class="pi pi-briefcase" /></span><strong>{{ crm.pipeline.length }}</strong><small>{{ active }} prospek masih aktif</small></RouterLink>
      <RouterLink to="/admin/prospects/list" class="metric-card"><span>Prospek Berhasil<i class="pi pi-star" /></span><strong>{{ won }}</strong><small>Prospek won dan terkonversi</small></RouterLink>
    </div>
    <div class="dashboard-grid">
      <article class="dashboard-panel trend-panel">
        <header><div><strong>Distribusi Pipeline</strong><span>Jumlah prospek berdasarkan tahap saat ini</span></div><RouterLink to="/admin/prospects/pipeline">Buka pipeline</RouterLink></header>
        <div v-if="!loading && crm.pipeline.length" class="bar-chart"><div v-for="entry in stageCounts" :key="entry.status"><span>{{ entry.count }}</span><i :style="{ height: `${Math.max(10, (entry.count / maxStageCount) * 120)}px` }" /><small>{{ entry.status.split('_').map((value) => value[0]).join('') }}</small></div></div>
        <div v-else-if="!loading" class="panel-empty"><i class="pi pi-chart-bar" /><strong>No pipeline data</strong><span>Pipeline distribution will appear here.</span></div><div v-else class="chart-skeleton" />
      </article>
      <article class="dashboard-panel pipeline-summary">
        <header><div><strong>Ringkasan Pipeline</strong><span>Perkembangan prospek di setiap tahap</span></div><RouterLink to="/admin/prospects/pipeline"><i class="pi pi-arrow-up-right" /></RouterLink></header>
        <div class="open-total"><span>Prospek aktif</span><strong>{{ active }}</strong></div>
        <div v-for="entry in stageCounts.slice(-4)" :key="entry.status" class="pipeline-line"><span>{{ entry.status.replaceAll('_', ' ') }}</span><i><b :style="{ width: `${Math.max(5, (entry.count / maxStageCount) * 100)}%` }" /></i><strong>{{ entry.count }}</strong></div>
      </article>
      <article class="dashboard-panel recent-panel">
        <header><div><strong>Recent Pipeline Activity</strong><span>Latest prospect records updated by the team</span></div><RouterLink to="/admin/prospects/pipeline">View all</RouterLink></header>
        <div v-if="recentPipeline.length" class="activity-table-wrap"><table><thead><tr><th>Sales Executive</th><th>Prospect</th><th>Industry</th><th>Status</th></tr></thead><tbody><tr v-for="item in recentPipeline" :key="item.id"><td>{{ item.assignedSalesExecutive || 'Unassigned' }}</td><td>{{ item.placeName }}</td><td>{{ item.industryGroup || '—' }}</td><td><Tag :value="item.status.replaceAll('_', ' ')" :severity="statusSeverity(item.status)" /></td></tr></tbody></table></div>
        <div v-else-if="!loading" class="panel-empty compact"><i class="pi pi-inbox" /><strong>No recent activity</strong></div>
      </article>
      <article class="dashboard-panel assignments">
        <header><div><strong>Kunjungan Hari Ini</strong><span>{{ todayVisits.length }} kunjungan tercatat</span></div><RouterLink to="/admin/visit-monitoring">Lihat semua</RouterLink></header>
        <RouterLink v-for="item in todayAssignments" :key="item.id" :to="`/admin/visit-monitoring?customerName=${encodeURIComponent(item.customerName)}`"><span>{{ new Date(item.checkInAt).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' }) }}</span><div><strong>{{ item.customerName }}</strong><small>{{ item.salesExecutiveName }}</small></div><i class="pi pi-arrow-right" /></RouterLink>
        <div v-if="!loading && !todayAssignments.length" class="panel-empty compact"><i class="pi pi-calendar-times" /><strong>No visits recorded today</strong></div>
      </article>
    </div>
  </section>
</template>

<style scoped>
.admin-dashboard { display: none; }
.erp-dashboard-shell { width:min(100%,1400px); margin:0 auto; padding:16px; background:#f8fafc; color:#0f172a; font-family:Inter,-apple-system,BlinkMacSystemFont,"Segoe UI",sans-serif; }
.erp-dashboard-hero,.erp-dashboard-panel,.erp-stat-card { border:1px solid #e2e8f0; background:#fff; box-shadow:0 10px 30px rgba(15,23,42,.04); }
.erp-dashboard-hero { display:flex; flex-wrap:nowrap; gap:12px; align-items:center; padding:20px; border-radius:20px; }.erp-hero-copy { min-width:0; }.erp-hero-actions { display:flex; flex-wrap:nowrap; align-items:center; gap:10px; margin-left:auto; align-self:center; white-space:nowrap; }.erp-hero-actions :deep(.yummy-social-links) { flex-wrap:nowrap; white-space:nowrap; }.erp-date-pill { display:inline-flex; align-items:center; gap:7px; padding:10px 14px; border:1px solid #e2e8f0; border-radius:10px; background:#fff; color:#475569; font-size:13px; text-decoration:none; white-space:nowrap; }
.erp-hero-icon { display:grid; place-items:center; width:44px; height:44px; border-radius:14px; background:#eff6ff; color:#2563eb; font-size:20px; }
.erp-dashboard-hero span,.erp-stat-card span { color:#94a3b8; font-size:11px; font-weight:700; letter-spacing:.08em; text-transform:uppercase; }
.erp-dashboard-hero h1 { margin:4px 0; font-size:24px; }.erp-dashboard-hero p { margin:0; color:#64748b; font-size:13px; }
.erp-stat-grid,.erp-link-grid,.erp-bottom-grid { display:grid; gap:12px; margin-top:12px; }.erp-stat-grid { grid-template-columns:repeat(4,minmax(0,1fr)); }
.erp-stat-card { display:grid; gap:8px; padding:16px; border-radius:18px; text-decoration:none; }.erp-stat-card strong { font-size:28px; color:#0f172a; }
.erp-dashboard-panel { padding:16px; border-radius:20px; }.quick-links { margin-top:12px; }.erp-dashboard-panel header { display:flex; justify-content:space-between; align-items:center; }.erp-dashboard-panel h2 { margin:0; font-size:16px; }.erp-dashboard-panel p { margin:2px 0 0; color:#64748b; font-size:12px; }.erp-dashboard-panel header a { padding:10px 14px; border:1px solid #e2e8f0; border-radius:10px; color:#475569; font-size:13px; text-decoration:none; }
.erp-link-grid { grid-template-columns:repeat(4,minmax(0,1fr)); }.erp-link-grid a { display:flex; align-items:center; gap:10px; padding:14px; border:1px solid #e2e8f0; border-radius:16px; background:#f8fafc; color:#0f172a; font-size:13px; font-weight:600; text-decoration:none; }.erp-link-grid a i:first-child { padding:9px; border-radius:12px; background:#fff1f2; color:#dc2626; }.erp-link-grid a i:last-child { margin-left:auto; color:#94a3b8; }
.erp-bottom-grid { grid-template-columns:repeat(2,minmax(0,1fr)); }.erp-dashboard-panel > h2 i { margin-right:8px; color:#475569; }.erp-list { display:grid; gap:10px; margin-top:14px; }.erp-list div { display:flex; justify-content:space-between; padding:12px 14px; border:1px solid #eef2f7; border-radius:14px; background:#f8fafc; color:#334155; font-size:13px; }
@media (max-width:900px) { .erp-dashboard-hero { overflow-x:auto; }.erp-hero-copy { flex:1 0 230px; }.erp-hero-actions { flex:0 0 auto; margin-left:auto; } }
@media (max-width:800px) { .erp-stat-grid,.erp-link-grid,.erp-bottom-grid { grid-template-columns:repeat(2,minmax(0,1fr)); } }
.admin-dashboard {
  width: min(100%, 1440px);
  margin: 0 auto;
  padding: 1rem 0;
  color: #0f172a;
}

/* Header */
.dashboard-header {
  margin-bottom: 1rem;
  padding: 1.2rem 1.35rem;
  border-radius: 20px;
  background: linear-gradient(115deg, #401326 0%, #721b36 48%, #c13b58 100%);
  box-shadow: 0 12px 30px rgba(83, 23, 45, .18);
  color: #fff;
  display: flex;
  align-items: center;
  justify-content: space-between;
}
.dashboard-header-actions { display:flex; flex-direction:column; align-items:flex-end; gap:.65rem; flex-shrink:0; }
.dashboard-header h1 {
  margin: 0;
  font-size: 1.65rem;
  font-weight: 800;
  letter-spacing: -0.03em;
  color: #fff;
  min-height: 132px;
}
.dashboard-header p {
  margin: 0.25rem 0 0;
  color: rgba(255,255,255,.72);
  font-size: 0.8rem;
}
.dashboard-eyebrow { display:block; margin-bottom:.3rem; color:#f9b4c1; font-size:.62rem; font-weight:800; letter-spacing:.14em; text-transform:uppercase; }
.date-control {
  padding: 0.55rem 0.9rem;
  display: flex;
  gap: 0.5rem;
  align-items: center;
  color: var(--text-secondary, #334155);
  background: var(--surface-card, #ffffff);
  border: 1px solid var(--border-default, #e2e8f0);
  border-radius: var(--radius-sm, 8px);
  text-decoration: none;
  font-size: 0.75rem;
  font-weight: 600;
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.04);
  transition: all 0.2s cubic-bezier(0.4, 0, 0.2, 1);
}
.date-control:hover {
  border-color: var(--brand-blue, #e63946);
  color: var(--brand-blue, #e63946);
  box-shadow: 0 4px 12px rgba(230, 57, 70, 0.12);
  transform: translateY(-1px);
}

/* Metric Cards */
.metric-grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 12px;
}
.metric-card {
  min-width: 0;
  padding: 16px;
  display: grid;
  color: var(--text-primary, #0f172a);
  background: var(--surface-card, #ffffff);
  border: 1px solid #e2e8f0;
  border-radius: 18px;
  box-shadow: 0 2px 8px -2px rgba(15, 23, 42, 0.05), 0 1px 3px -1px rgba(15, 23, 42, 0.03);
  text-decoration: none;
  position: relative;
  overflow: hidden;
  transition: all 0.25s cubic-bezier(0.4, 0, 0.2, 1);
}
.metric-card::before {
  content: '';
  position: absolute;
  top: 0; left: 0; right: 0;
  height: 3px;
  background: linear-gradient(90deg, transparent, rgba(230, 57, 70, 0.3), transparent);
  opacity: 0;
  transition: opacity 0.25s ease;
}
.metric-card:hover {
  box-shadow: 0 10px 20px -5px rgba(15, 23, 42, 0.08), 0 4px 6px -2px rgba(15, 23, 42, 0.03);
  border-color: rgba(230, 57, 70, 0.3);
  transform: translateY(-3px);
}
.metric-card:hover::before {
  opacity: 1;
}
.metric-card > span {
  display: flex;
  justify-content: space-between;
  align-items: center;
  color: var(--text-muted, #64748b);
  font-size: 0.7rem;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.04em;
}
.metric-card > span i {
  width: 2.1rem;
  height: 2.1rem;
  display: grid;
  place-items: center;
  color: var(--brand-blue, #e63946);
  background: var(--brand-blue-bg, #fff0f1);
  border-radius: var(--radius-sm, 8px);
  font-size: 0.9rem;
  transition: transform 0.2s ease;
}
.metric-card:hover > span i {
  transform: scale(1.08);
}

/* Aksentuasi warna ikon berdasarkan kartu */
.metric-card:nth-child(1) > span i { color: #e63946; background: #fff0f1; }
.metric-card:nth-child(2) > span i { color: #059669; background: #ecfdf5; }
.metric-card:nth-child(3) > span i { color: #c54b59; background: #f5f3ff; }
.metric-card:nth-child(4) > span i { color: #d97706; background: #fffbe2; }

.metric-card > strong {
  margin: 0.5rem 0 0.2rem;
  font-size: 1.75rem;
  font-weight: 800;
  letter-spacing: -0.03em;
  line-height: 1.1;
}
.metric-card small {
  color: #10b981;
  font-size: 0.68rem;
  font-weight: 600;
  display: flex;
  align-items: center;
  gap: 0.25rem;
}

/* Dashboard Panels */
.dashboard-grid {
  margin-top: 16px;
  display: grid;
  grid-template-columns: minmax(0, 2fr) minmax(300px, 0.9fr);
  gap: 12px;
}
.dashboard-panel {
  padding: 16px;
  background: var(--surface-card, #ffffff);
  border: 1px solid #e2e8f0;
  border-radius: 20px;
  box-shadow: 0 2px 8px -2px rgba(15, 23, 42, 0.05), 0 1px 3px -1px rgba(15, 23, 42, 0.03);
}
.dashboard-panel header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 0.5rem;
}
.dashboard-panel header div { display: grid; }
.dashboard-panel header strong {
  font-size: 0.9rem;
  font-weight: 700;
  color: var(--text-primary, #0f172a);
}
.dashboard-panel header span {
  margin-top: 0.1rem;
  color: var(--text-muted, #64748b);
  font-size: 0.68rem;
}
.dashboard-panel header a {
  color: var(--brand-blue, #e63946);
  text-decoration: none;
  font-size: 0.72rem;
  font-weight: 600;
  padding: 0.2rem 0.5rem;
  border-radius: 4px;
  background: var(--brand-blue-50, #fff0f1);
  transition: all 0.2s ease;
}
.dashboard-panel header a:hover {
  background: var(--brand-blue, #e63946);
  color: #ffffff;
}

/* Bar Chart */
.bar-chart {
  height: 200px;
  margin-top: 1rem;
  padding: 1.25rem 1.25rem 0.5rem;
  display: flex;
  align-items: flex-end;
  justify-content: space-around;
  background: linear-gradient(180deg, #f8fafc 0%, #edf2f7 100%);
  border-radius: var(--radius-md, 10px);
  border: 1px solid #e2e8f0;
}
.bar-chart div {
  height: 100%;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: flex-end;
  gap: 0.35rem;
}
.bar-chart span {
  color: #334155;
  font-size: 0.68rem;
  font-weight: 700;
}
.bar-chart small {
  color: #64748b;
  font-size: 0.62rem;
  font-weight: 600;
}
.bar-chart i {
  width: 24px;
  max-height: 130px;
  background: linear-gradient(180deg, #ef4e5d 0%, #d62839 100%);
  border-radius: 6px 6px 2px 2px;
  box-shadow: 0 2px 4px rgba(230, 57, 70, 0.25);
  transition: height 0.3s cubic-bezier(0.4, 0, 0.2, 1), transform 0.2s ease;
}
.bar-chart div:hover i {
  transform: scaleY(1.03);
  background: linear-gradient(180deg, #ef4e5d 0%, #e63946 100%);
}

/* Summary Pipeline */
.open-total {
  margin: 1rem 0;
  padding: 0.75rem 1rem;
  display: flex;
  justify-content: space-between;
  align-items: center;
  color: #d62839;
  background: linear-gradient(135deg, #fff0f1 0%, #ffd9dc 100%);
  border: 1px solid #f4b3ba;
  border-radius: var(--radius-sm, 8px);
}
.open-total span {
  font-size: 0.72rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.03em;
}
.open-total strong {
  font-size: 1.35rem;
  font-weight: 800;
}
.pipeline-line {
  margin: 0.65rem 0;
  display: grid;
  grid-template-columns: 100px 1fr 28px;
  gap: 0.6rem;
  align-items: center;
  font-size: 0.68rem;
  font-weight: 600;
  color: var(--text-secondary, #475569);
}
.pipeline-line i {
  height: 6px;
  background: #f1f5f9;
  border-radius: 10px;
  overflow: hidden;
  box-shadow: inset 0 1px 2px rgba(0,0,0,0.05);
}
.pipeline-line b {
  height: 100%;
  display: block;
  background: linear-gradient(90deg, #ef4e5d, #d62839);
  border-radius: 10px;
  transition: width 0.4s ease;
}
.pipeline-line strong {
  text-align: right;
  color: var(--text-primary, #0f172a);
}

/* Recent Activity Table */
.recent-panel { overflow-x: auto; }
.activity-table-wrap { overflow-x:auto; }
.recent-panel table {
  width: 100%;
  margin-top: 0.85rem;
  border-collapse: separate;
  border-spacing: 0;
}
.recent-panel th {
  padding: 0.6rem 0.75rem;
  font-size: 0.65rem;
  font-weight: 700;
  color: var(--text-muted, #64748b);
  text-transform: uppercase;
  letter-spacing: 0.04em;
  border-bottom: 2px solid var(--border-light, #f1f5f9);
  text-align: left;
}
.recent-panel td {
  padding: 0.65rem 0.75rem;
  font-size: 0.72rem;
  color: var(--text-secondary, #334155);
  border-bottom: 1px solid var(--border-light, #f8fafc);
  transition: background 0.15s ease;
}
.recent-panel tr:last-child td {
  border-bottom: none;
}
.recent-panel tr:hover td {
  background: #f8fafc;
}

/* Assignments */
.assignments > a {
  margin-top: 0.6rem;
  padding: 0.65rem 0.85rem;
  display: grid;
  grid-template-columns: 42px 1fr auto;
  gap: 0.65rem;
  align-items: center;
  color: #334155;
  background: #f8fafc;
  border: 1px solid #f1f5f9;
  border-radius: var(--radius-sm, 8px);
  text-decoration: none;
  transition: all 0.2s ease;
}
.assignments > a:hover {
  background: #fff0f1;
  border-color: #f4b3ba;
  transform: translateX(2px);
}
.assignments > a > span {
  font-size: 0.65rem;
  font-weight: 700;
  color: #e63946;
  background: #ffffff;
  padding: 0.2rem 0.35rem;
  border-radius: 4px;
  text-align: center;
  border: 1px solid #e2e8f0;
}
.assignments > a div { display: grid; }
.assignments > a strong {
  font-size: 0.72rem;
  color: var(--text-primary, #0f172a);
}
.assignments > a small {
  color: var(--text-muted, #64748b);
  font-size: 0.62rem;
}
.assignments > a i {
  color: #94a3b8;
  font-size: 0.75rem;
  transition: transform 0.2s ease, color 0.2s ease;
}
.assignments > a:hover i {
  color: #e63946;
  transform: translateX(3px);
}

.panel-empty { min-height:190px; display:flex; flex-direction:column; align-items:center; justify-content:center; gap:.35rem; color:var(--text-muted); text-align:center; }
.panel-empty i { width:44px; height:44px; display:grid; place-items:center; margin-bottom:.25rem; border-radius:13px; color:#e63946; background:#fff0f1; }
.panel-empty strong { color:var(--text-primary); font-size:.78rem; }
.panel-empty span { font-size:.68rem; }
.panel-empty.compact { min-height:110px; }
.chart-skeleton { height:200px; margin-top:1rem; border-radius:10px; background:linear-gradient(90deg,#f8f4f5 25%,#fff 50%,#f8f4f5 75%); background-size:200% 100%; animation:dashboard-pulse 1.4s infinite; }
.skeleton-metric { min-height:118px; pointer-events:none; }
.skeleton-metric > i,.skeleton-metric > strong,.skeleton-metric > small { display:block; border-radius:6px; background:#f1e8ea; animation:dashboard-pulse 1.4s infinite; }
.skeleton-metric > i { width:34px; height:34px; margin-left:auto; }
.skeleton-metric > strong { width:42%; height:24px; margin:.4rem 0; }
.skeleton-metric > small { width:70%; height:9px; }
@keyframes dashboard-pulse { 0%,100%{opacity:.55} 50%{opacity:1} }

/* CRM command-center refresh */
.admin-dashboard { max-width: 1480px; padding: 1rem 0 2rem; }
.dashboard-header {
  margin-bottom: 1.25rem; padding: 1.45rem 1.55rem; align-items: flex-end;
  border-radius: 20px; color: #fff; overflow: hidden; position: relative;
  background: radial-gradient(circle at 90% 15%, rgba(255, 126, 137, .28), transparent 30%), linear-gradient(120deg, #351522 0%, #761f32 62%, #b6324d 100%);
  box-shadow: 0 18px 35px -22px rgba(16, 28, 61, .85);
}
.dashboard-header::after { content:''; position:absolute; width:180px; height:180px; right:15%; bottom:-120px; border:1px solid rgba(255,255,255,.14); border-radius:50%; box-shadow:0 0 0 28px rgba(255,255,255,.04),0 0 0 58px rgba(255,255,255,.03); }
.dashboard-header > div:first-child, .dashboard-header-actions { position:relative; z-index:1; }
.dashboard-header h1 { color:#fff; font-size:1.8rem; }
.dashboard-header p { color:rgba(255,255,255,.68); }
.dashboard-eyebrow { color:#ffb0b8; letter-spacing:.14em; }
.dashboard-header-actions { flex-direction:row; align-items:center; }
.date-control { color:#fff; background:rgba(255,255,255,.1); border-color:rgba(255,255,255,.2); box-shadow:none; }
.date-control:hover { color:#fff; border-color:#ffb0b8; background:rgba(255,255,255,.18); }
.metric-grid { gap:.85rem; }
.metric-card { padding:1.2rem 1.25rem; border-radius:16px; border-color:#e7edf5; box-shadow:0 7px 18px -14px rgba(16,28,61,.5); }
.metric-card > span { color:#718096; }
.metric-card > strong { color:#17233e; font-size:2rem; }
.metric-card small { color:#7a879b; }
.metric-card:nth-child(1) { border-top:3px solid #e63946; }
.metric-card:nth-child(2) { border-top:3px solid #16b8a6; }
.metric-card:nth-child(3) { border-top:3px solid #f08a62; }
.metric-card:nth-child(4) { border-top:3px solid #e65b70; }
.metric-card:nth-child(1) > span i { color:#e63946; background:#fff0f1; }
.metric-card:nth-child(2) > span i { color:#16a894; background:#e8faf7; }
.metric-card:nth-child(3) > span i { color:#d66a45; background:#fff3ed; }
.metric-card:nth-child(4) > span i { color:#df5269; background:#fff0f3; }
.dashboard-grid { gap:.85rem; margin-top:1rem; }
.dashboard-panel { padding:1.2rem; border-radius:16px; border-color:#e7edf5; box-shadow:0 7px 18px -14px rgba(16,28,61,.5); }
.dashboard-panel header strong { color:#17233e; font-size:.95rem; }
.dashboard-panel header span { color:#8290a5; }
.dashboard-panel header a { color:#d52f43; background:#fff0f1; }
.dashboard-panel header a:hover { background:#e63946; }
.bar-chart { background:linear-gradient(180deg,#f7f9fd,#eef3f9); border:1px solid #edf1f7; }
.bar-chart i { background:linear-gradient(180deg,#ff6975,#b51f36); border-radius:8px 8px 3px 3px; box-shadow:0 5px 10px -7px #b51f36; }
.bar-chart small { color:#7c899e; font-weight:700; }
.open-total { padding:1rem; margin:.75rem 0 1rem; border-radius:12px; background:linear-gradient(120deg,#fff0f1,#fff9f9); }
.open-total strong { color:#b51f36; font-size:1.65rem; }
.pipeline-line i { background:#eef1f6; }
.pipeline-line i b { background:linear-gradient(90deg,#e63946,#f69b8a); }
.recent-panel, .assignments { border-top:3px solid #edf1f7; }
.recent-panel th { color:#91a0b5; }
.recent-panel td { color:#43516a; }
.assignments > a { background:#f8faff; border-color:#edf1f7; }
.assignments > a:hover { background:#fff0f1; border-color:#f4b3ba; }
.assignments > a > span { color:#d52f43; background:#fff0f1; border-color:#ffd9dc; }

@media (max-width: 1100px) { .metric-grid { grid-template-columns: repeat(2, 1fr); } }
@media (max-width: 900px) { .dashboard-grid { grid-template-columns: 1fr; } }
@media (max-width: 640px) {
  .admin-dashboard { padding:.15rem 0; }
  .dashboard-header { align-items:flex-start; flex-direction:column; gap:.85rem; margin-bottom:1rem; }
  .dashboard-header-actions { width:100%; align-items:stretch; gap:.65rem; }
  .dashboard-header-actions :deep(.yummy-social-links) { width:100%; display:grid; grid-template-columns:repeat(3,minmax(0,1fr)); gap:.5rem; }
  .dashboard-header-actions :deep(.yummy-social-link) { width:auto; min-height:46px; justify-content:center; padding:.45rem .35rem; border-radius:12px; }
  .dashboard-header-actions :deep(.yummy-social-label) { display:inline; font-size:.64rem; }
  .dashboard-header-actions :deep(.yummy-social-link > i) { display:none; }
  .dashboard-header-actions :deep(.yummy-social-icon) { width:20px; height:20px; flex-shrink:0; }
  .dashboard-header-actions .date-control { align-self:flex-end; }
  .dashboard-header h1 { font-size:1.35rem; }
  .dashboard-header p { font-size:.72rem; }
  .date-control { padding:.48rem; font-size:0; }
  .date-control i { font-size:.85rem; }
  .metric-grid { grid-template-columns:repeat(2,minmax(0,1fr)); gap:.65rem; }
  .metric-card { padding:.85rem; border-radius:12px; }
  .metric-card > span { font-size:.58rem; }
  .metric-card > span i { width:1.8rem; height:1.8rem; }
  .metric-card > strong { font-size:1.45rem; }
  .metric-card small { font-size:.58rem; }
  .dashboard-grid { gap:.7rem; margin-top:.8rem; }
  .dashboard-panel { padding:.9rem; border-radius:12px; }
  .bar-chart { height:180px; padding:.9rem .35rem .45rem; overflow-x:auto; justify-content:space-between; }
  .bar-chart div { min-width:36px; }
  .bar-chart i { width:18px; }
  .activity-table-wrap table { min-width:620px; }
  .pipeline-line { grid-template-columns:88px 1fr 24px; gap:.4rem; }
}

/* Mobile-first dashboard: simple, readable, and touch-friendly */
@media (max-width: 640px) {
  .admin-dashboard { padding: .25rem 0 1.25rem; }
  .dashboard-header { margin: 0 0 .8rem; padding: 1rem; display: block; border-radius: 16px; }
  .dashboard-header h1 { font-size: 1.3rem; margin-top: .15rem; }
  .dashboard-header p { max-width: 270px; margin-top: .3rem; font-size: .7rem; line-height: 1.45; }
  .dashboard-header-actions { margin-top: .8rem; display: flex; align-items: stretch; justify-content: space-between; gap: .5rem; }
  .dashboard-header-actions :deep(.yummy-social-links) { display: none; }
  .date-control { flex: 1; justify-content: center; min-height: 40px; padding: .55rem .7rem; font-size: .68rem; border-radius: 10px; }
  .metric-grid { grid-template-columns: repeat(2, minmax(0, 1fr)); gap: .55rem; }
  .metric-card { min-height: 112px; padding: .8rem; border-radius: 13px; }
  .metric-card > span { font-size: .58rem; line-height: 1.25; text-transform: none; letter-spacing: 0; gap: .35rem; }
  .metric-card > span i { width: 1.75rem; height: 1.75rem; font-size: .75rem; }
  .metric-card > strong { margin: .45rem 0 .1rem; font-size: 1.55rem; }
  .metric-card small { font-size: .58rem; line-height: 1.3; }
  .dashboard-grid { display: flex; flex-direction: column; gap: .65rem; margin-top: .7rem; }
  .dashboard-panel { padding: .85rem; border-radius: 13px; }
  .dashboard-panel header { align-items: center; margin-bottom: .35rem; }
  .dashboard-panel header strong { font-size: .82rem; }
  .dashboard-panel header span { font-size: .6rem; }
  .dashboard-panel header a { padding: .35rem .45rem; font-size: .62rem; white-space: nowrap; }
  .bar-chart { height: 165px; margin-top: .65rem; padding: .8rem .35rem .4rem; }
  .bar-chart div { min-width: 34px; }
  .bar-chart i { width: 16px; }
  .bar-chart span { font-size: .6rem; }
  .bar-chart small { font-size: .55rem; }
  .open-total { display: flex; align-items: center; justify-content: space-between; padding: .7rem .8rem; margin: .55rem 0 .7rem; }
  .open-total span { font-size: .65rem; }
  .open-total strong { font-size: 1.35rem; }
  .pipeline-line { grid-template-columns: 108px 1fr 22px; min-height: 26px; font-size: .62rem; }
  .recent-panel { overflow: hidden; }
  .activity-table-wrap { overflow-x: auto; margin: 0 -.85rem; padding: 0 .85rem .2rem; -webkit-overflow-scrolling: touch; }
  .activity-table-wrap table { min-width: 540px; font-size: .65rem; }
  .activity-table-wrap th, .activity-table-wrap td { padding: .55rem .45rem; }
  .assignments > a { min-height: 52px; margin-top: .45rem; padding: .55rem .65rem; grid-template-columns: 48px 1fr 18px; gap: .5rem; }
  .assignments > a > span { font-size: .6rem; }
  .assignments > a strong { font-size: .68rem; }
  .assignments > a small { font-size: .59rem; }
}
</style>
