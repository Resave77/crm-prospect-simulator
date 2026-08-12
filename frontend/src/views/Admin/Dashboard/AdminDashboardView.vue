<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import Message from 'primevue/message'
import Tag from 'primevue/tag'
import { useAuthStore } from '../../../stores/auth'
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
  <section class="admin-dashboard">
    <Message v-if="error" severity="error" closable @close="error = ''">{{ error }}</Message>
    <div class="dashboard-header">
      <div><span class="dashboard-eyebrow">Workspace overview</span><h1>Admin Dashboard</h1><p>Monitor sales activity and pipeline momentum, {{ auth.user?.fullName }}.</p></div>
      <RouterLink class="date-control" to="/admin/visit-monitoring"><i class="pi pi-calendar" /> {{ formattedDate }}</RouterLink>
    </div>
    <div v-if="loading" class="metric-grid dashboard-skeletons"><div v-for="n in 4" :key="n" class="metric-card skeleton-metric"><i /><strong /><small /></div></div>
    <div v-else class="metric-grid">
      <RouterLink to="/admin/customers" class="metric-card"><span>Total Customers<i class="pi pi-users" /></span><strong>{{ crm.adminCustomers.length }}</strong><small>Converted customer records</small></RouterLink>
      <RouterLink to="/admin/visit-monitoring" class="metric-card"><span>Today's Visits<i class="pi pi-map-marker" /></span><strong>{{ todayVisits.length }}</strong><small>Recorded field visits today</small></RouterLink>
      <RouterLink to="/admin/prospects/pipeline" class="metric-card"><span>Total Prospects<i class="pi pi-briefcase" /></span><strong>{{ crm.pipeline.length }}</strong><small>{{ active }} currently active</small></RouterLink>
      <RouterLink to="/admin/prospects/list" class="metric-card"><span>Won Prospects<i class="pi pi-star" /></span><strong>{{ won }}</strong><small>Won and converted records</small></RouterLink>
    </div>
    <div class="dashboard-grid">
      <article class="dashboard-panel trend-panel">
        <header><div><strong>Pipeline Distribution</strong><span>Prospects grouped by current stage</span></div><RouterLink to="/admin/prospects/pipeline">Open pipeline</RouterLink></header>
        <div v-if="!loading && crm.pipeline.length" class="bar-chart"><div v-for="entry in stageCounts" :key="entry.status"><span>{{ entry.count }}</span><i :style="{ height: `${Math.max(10, (entry.count / maxStageCount) * 120)}px` }" /><small>{{ entry.status.split('_').map((value) => value[0]).join('') }}</small></div></div>
        <div v-else-if="!loading" class="panel-empty"><i class="pi pi-chart-bar" /><strong>No pipeline data</strong><span>Pipeline distribution will appear here.</span></div><div v-else class="chart-skeleton" />
      </article>
      <article class="dashboard-panel pipeline-summary">
        <header><div><strong>Prospect Pipeline</strong><span>Progress across later stages</span></div><RouterLink to="/admin/prospects/pipeline"><i class="pi pi-arrow-up-right" /></RouterLink></header>
        <div class="open-total"><span>Open prospects</span><strong>{{ active }}</strong></div>
        <div v-for="entry in stageCounts.slice(-4)" :key="entry.status" class="pipeline-line"><span>{{ entry.status.replaceAll('_', ' ') }}</span><i><b :style="{ width: `${Math.max(5, (entry.count / maxStageCount) * 100)}%` }" /></i><strong>{{ entry.count }}</strong></div>
      </article>
      <article class="dashboard-panel recent-panel">
        <header><div><strong>Recent Pipeline Activity</strong><span>Latest prospect records updated by the team</span></div><RouterLink to="/admin/prospects/pipeline">View all</RouterLink></header>
        <div v-if="recentPipeline.length" class="activity-table-wrap"><table><thead><tr><th>Sales Executive</th><th>Prospect</th><th>Industry</th><th>Status</th></tr></thead><tbody><tr v-for="item in recentPipeline" :key="item.id"><td>{{ item.assignedSalesExecutive || 'Unassigned' }}</td><td>{{ item.placeName }}</td><td>{{ item.industryGroup || '—' }}</td><td><Tag :value="item.status.replaceAll('_', ' ')" :severity="statusSeverity(item.status)" /></td></tr></tbody></table></div>
        <div v-else-if="!loading" class="panel-empty compact"><i class="pi pi-inbox" /><strong>No recent activity</strong></div>
      </article>
      <article class="dashboard-panel assignments">
        <header><div><strong>Today's Visits</strong><span>{{ todayVisits.length }} recorded today</span></div><RouterLink to="/admin/visit-monitoring">View all</RouterLink></header>
        <RouterLink v-for="item in todayAssignments" :key="item.id" :to="`/admin/visit-monitoring?customerName=${encodeURIComponent(item.customerName)}`"><span>{{ new Date(item.checkInAt).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' }) }}</span><div><strong>{{ item.customerName }}</strong><small>{{ item.salesExecutiveName }}</small></div><i class="pi pi-arrow-right" /></RouterLink>
        <div v-if="!loading && !todayAssignments.length" class="panel-empty compact"><i class="pi pi-calendar-times" /><strong>No visits recorded today</strong></div>
      </article>
    </div>
  </section>
</template>

<style scoped>
.admin-dashboard {
  width: min(100%, 1440px);
  margin: 0 auto;
  padding: 0.5rem 0;
}

/* Header */
.dashboard-header {
  margin-bottom: 1.5rem;
  display: flex;
  align-items: center;
  justify-content: space-between;
}
.dashboard-header h1 {
  margin: 0;
  font-size: 1.65rem;
  font-weight: 800;
  letter-spacing: -0.03em;
  color: var(--text-primary, #0f172a);
}
.dashboard-header p {
  margin: 0.25rem 0 0;
  color: var(--text-muted, #64748b);
  font-size: 0.8rem;
}
.dashboard-eyebrow { display:block; margin-bottom:.3rem; color:#e63946; font-size:.62rem; font-weight:800; letter-spacing:.1em; text-transform:uppercase; }
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
  gap: 1rem;
}
.metric-card {
  min-width: 0;
  padding: 1.15rem;
  display: grid;
  color: var(--text-primary, #0f172a);
  background: var(--surface-card, #ffffff);
  border: 1px solid var(--border-light, #f1f5f9);
  border-radius: var(--radius-lg, 12px);
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
  margin-top: 1.25rem;
  display: grid;
  grid-template-columns: minmax(0, 2fr) minmax(300px, 0.9fr);
  gap: 1rem;
}
.dashboard-panel {
  padding: 1.25rem;
  background: var(--surface-card, #ffffff);
  border: 1px solid var(--border-light, #f1f5f9);
  border-radius: var(--radius-lg, 12px);
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

@media (max-width: 1100px) { .metric-grid { grid-template-columns: repeat(2, 1fr); } }
@media (max-width: 900px) { .dashboard-grid { grid-template-columns: 1fr; } }
@media (max-width: 640px) {
  .admin-dashboard { padding:.15rem 0; }
  .dashboard-header { align-items:flex-start; gap:.8rem; margin-bottom:1rem; }
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
</style>
