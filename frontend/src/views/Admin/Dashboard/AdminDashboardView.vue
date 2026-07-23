<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import Message from 'primevue/message'
import Tag from 'primevue/tag'
import { useAuthStore } from '../../../stores/auth'
import { useCrmStore } from '../../../stores/crm'
const auth=useAuthStore(); const crm=useCrmStore(); const error=ref('')
const active=computed(()=>crm.pipeline.filter(v=>!['LOST','CONVERTED'].includes(v.status)).length)
const won=computed(()=>crm.pipeline.filter(v=>v.status==='WON').length)
const assignedToday=computed(()=>crm.pipeline.slice(0,3))
const stageCounts=computed(()=>['NEW_LEAD','CONTACTED','INTERESTED','QUALIFIED','PROPOSAL_SENT','NEGOTIATION','WON'].map(status=>({status,count:crm.pipeline.filter(v=>v.status===status).length})))
onMounted(async()=>{try{await Promise.all([crm.loadPipeline(),crm.loadAdminCustomers()])}catch(e){error.value=crm.errorMessage(e)}})
</script>
<template><section class="admin-dashboard"><Message v-if="error" severity="error">{{ error }}</Message><div class="dashboard-header"><div><h1>Admin Dashboard</h1><p>A clear view of field sales momentum, {{ auth.user?.fullName }}.</p></div><RouterLink class="date-control" to="/admin/prospects/pipeline"><i class="pi pi-calendar" /> {{ new Date().toLocaleDateString() }}</RouterLink></div><div class="metric-grid"><RouterLink to="/admin/customers" class="metric-card"><span>Total Customers<i class="pi pi-users" /></span><strong>{{ crm.adminCustomers.length }}</strong><small>Converted existing customers</small></RouterLink><RouterLink to="/admin/visit-monitoring" class="metric-card"><span>Today's Visits<i class="pi pi-map-marker" /></span><strong>{{ assignedToday.length }}</strong><small>Simulation assignments</small></RouterLink><RouterLink to="/admin/prospects/pipeline" class="metric-card"><span>Total Prospects<i class="pi pi-briefcase" /></span><strong>{{ crm.pipeline.length }}</strong><small>{{ active }} active pipeline</small></RouterLink><RouterLink to="/admin/prospects/won" class="metric-card"><span>Won Prospect<i class="pi pi-star" /></span><strong>{{ won }}</strong><small>Ready for review</small></RouterLink></div><div class="dashboard-grid"><article class="dashboard-panel trend-panel"><header><div><strong>Pipeline Trend</strong><span>Current distribution by stage</span></div></header><div class="bar-chart"><div v-for="entry in stageCounts" :key="entry.status"><span>{{ entry.count }}</span><i :style="{height:`${Math.max(12,entry.count*28)}px`}" /><small>{{ entry.status.split('_').map(v=>v[0]).join('') }}</small></div></div></article><article class="dashboard-panel pipeline-summary"><header><div><strong>Prospect Pipeline</strong><span>Current stage distribution</span></div><RouterLink to="/admin/prospects/pipeline">•••</RouterLink></header><div class="open-total"><span>Open prospects</span><strong>{{ active }}</strong></div><div v-for="entry in stageCounts.slice(-4)" :key="entry.status" class="pipeline-line"><span>{{ entry.status.replaceAll('_',' ') }}</span><i><b :style="{width:`${Math.max(5,(entry.count/Math.max(1,crm.pipeline.length))*100)}%`}" /></i><strong>{{ entry.count }}</strong></div></article><article class="dashboard-panel recent-panel"><header><div><strong>Recent Pipeline Activity</strong><span>Latest records updated by the sales team</span></div><RouterLink to="/admin/prospects/pipeline">View all</RouterLink></header><table><thead><tr><th>Sales Executive</th><th>Prospect</th><th>Industry</th><th>Status</th></tr></thead><tbody><tr v-for="item in crm.pipeline.slice(0,5)" :key="item.id"><td>{{ item.assignedSalesExecutive }}</td><td>{{ item.placeName }}</td><td>{{ item.industryGroup }}</td><td><Tag :value="item.status.replaceAll('_',' ')" :severity="item.status==='WON'?'success':'info'" /></td></tr></tbody></table></article><article class="dashboard-panel assignments"><header><div><strong>Today's Assignment</strong><span>{{ assignedToday.length }} active records</span></div></header><RouterLink v-for="item in assignedToday" :key="item.id" :to="`/admin/prospects/pipeline?search=${encodeURIComponent(item.placeName)}`"><span>{{ new Date(item.updatedAt).toLocaleTimeString([],{hour:'2-digit',minute:'2-digit'}) }}</span><div><strong>{{ item.placeName }}</strong><small>{{ item.assignedSalesExecutive }}</small></div><i class="pi pi-arrow-right" /></RouterLink></article></div></section></template>

<style scoped>
.admin-dashboard {
  width: min(100%, 1440px);
  margin: 0 auto;
}
.dashboard-header {
  margin-bottom: 1.25rem;
  display: flex;
  align-items: center;
  justify-content: space-between;
}
.dashboard-header h1 { margin: 0; font-size: 1.5rem; font-weight: 800; letter-spacing: -0.03em; }
.dashboard-header p { margin: 0.2rem 0 0; color: var(--text-muted); font-size: 0.72rem; }
.date-control {
  padding: 0.5rem 0.75rem;
  display: flex;
  gap: 0.45rem;
  align-items: center;
  color: var(--text-secondary);
  background: var(--surface-card);
  border: 1px solid var(--border-default);
  border-radius: var(--radius-sm);
  text-decoration: none;
  font-size: 0.7rem;
  transition: border-color var(--transition-fast), box-shadow var(--transition-fast);
}
.date-control:hover { border-color: var(--brand-blue); box-shadow: 0 0 0 3px rgba(37, 99, 235, 0.06); }
.metric-grid { display: grid; grid-template-columns: repeat(4, minmax(0, 1fr)); gap: 0.85rem; }
.metric-card {
  min-width: 0; padding: 1rem; display: grid;
  color: var(--text-primary); background: var(--surface-card);
  border: 1px solid var(--border-light); border-radius: var(--radius-lg);
  box-shadow: var(--shadow-sm); text-decoration: none;
  transition: box-shadow var(--transition-base), border-color var(--transition-base), transform var(--transition-base);
}
.metric-card:hover { box-shadow: var(--shadow-md); border-color: var(--border-default); transform: translateY(-1px); }
.metric-card > span { display: flex; justify-content: space-between; color: var(--text-muted); font-size: 0.62rem; font-weight: 600; }
.metric-card > span i {
  width: 1.75rem; height: 1.75rem; display: grid; place-items: center;
  color: var(--brand-blue); background: var(--brand-blue-bg); border-radius: var(--radius-sm); font-size: 0.75rem;
}
.metric-card > strong { margin: 0.35rem 0; font-size: 1.6rem; font-weight: 800; letter-spacing: -0.03em; }
.metric-card small { color: #22a65a; font-size: 0.58rem; font-weight: 500; }
.dashboard-grid { margin-top: 1rem; display: grid; grid-template-columns: minmax(0, 2fr) minmax(280px, 0.9fr); gap: 0.85rem; }
.dashboard-panel {
  padding: 1rem; background: var(--surface-card);
  border: 1px solid var(--border-light); border-radius: var(--radius-lg); box-shadow: var(--shadow-sm);
}
.dashboard-panel header { display: flex; justify-content: space-between; align-items: flex-start; }
.dashboard-panel header div { display: grid; }
.dashboard-panel header strong { font-size: 0.78rem; font-weight: 700; }
.dashboard-panel header span { margin-top: 0.15rem; color: var(--text-muted); font-size: 0.54rem; }
.dashboard-panel header a { color: var(--brand-blue); text-decoration: none; font-size: 0.6rem; font-weight: 600; transition: opacity var(--transition-fast); }
.dashboard-panel header a:hover { opacity: 0.75; }
.bar-chart {
  height: 190px; margin-top: 0.85rem; padding: 1rem 1.25rem 0.4rem;
  display: flex; align-items: flex-end; justify-content: space-around;
  background: var(--brand-blue-50); border-radius: var(--radius-md);
}
.bar-chart div { height: 100%; display: flex; flex-direction: column; align-items: center; justify-content: flex-end; gap: 0.2rem; }
.bar-chart span, .bar-chart small { color: #70809a; font-size: 0.52rem; }
.bar-chart i { width: 22px; max-height: 130px; background: linear-gradient(to top, #2563eb, #3d7bf2); border-radius: 4px; transition: height var(--transition-smooth); }
.open-total { margin: 0.75rem 0; padding: 0.6rem; display: grid; color: var(--brand-blue); background: var(--brand-blue-50); border-radius: var(--radius-sm); }
.open-total span { font-size: 0.55rem; font-weight: 500; }
.open-total strong { font-size: 1.15rem; font-weight: 800; }
.pipeline-line { margin: 0.5rem 0; display: grid; grid-template-columns: 80px 1fr 22px; gap: 0.45rem; align-items: center; font-size: 0.55rem; }
.pipeline-line i { height: 4px; background: #e6ebf2; border-radius: 10px; }
.pipeline-line b { height: 100%; display: block; background: var(--brand-blue); border-radius: 10px; }
.recent-panel { overflow-x: auto; }
.recent-panel table { margin-top: 0.65rem; }
.recent-panel th, .recent-panel td { padding: 0.5rem; font-size: 0.58rem; }
.assignments > a {
  margin-top: 0.5rem; padding: 0.5rem; display: grid; grid-template-columns: 36px 1fr auto;
  gap: 0.5rem; align-items: center; color: #314059; background: var(--brand-blue-50);
  border-radius: var(--radius-sm); text-decoration: none; transition: background var(--transition-fast);
}
.assignments > a:hover { background: var(--brand-blue-bg); }
.assignments > a > span { font-size: 0.52rem; }
.assignments > a div { display: grid; }
.assignments > a strong { font-size: 0.62rem; }
.assignments > a small { color: var(--text-muted); font-size: 0.52rem; }
@media (max-width: 1100px) { .metric-grid { grid-template-columns: repeat(2, 1fr); } }
@media (max-width: 900px) { .dashboard-grid { grid-template-columns: 1fr; } }
</style>
