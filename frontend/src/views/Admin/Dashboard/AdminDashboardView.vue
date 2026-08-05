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

<template><section class="admin-dashboard"><Message v-if="error" severity="error">{{ error }}</Message><div class="dashboard-header"><div><h1>Admin Dashboard</h1><p>A clear view of field sales momentum, {{ auth.user?.fullName }}.</p></div><RouterLink class="date-control" to="/admin/prospects/pipeline"><i class="pi pi-calendar" /> {{ new Date().toLocaleDateString() }}</RouterLink></div><div class="metric-grid"><RouterLink to="/admin/customers" class="metric-card"><span>Total Customers<i class="pi pi-users" /></span><strong>{{ crm.adminCustomers.length }}</strong><small>Converted existing customers</small></RouterLink><RouterLink to="/admin/visit-monitoring" class="metric-card"><span>Today's Visits<i class="pi pi-map-marker" /></span><strong>{{ assignedToday.length }}</strong><small>Simulation assignments</small></RouterLink><RouterLink to="/admin/prospects/pipeline" class="metric-card"><span>Total Prospects<i class="pi pi-briefcase" /></span><strong>{{ crm.pipeline.length }}</strong><small>{{ active }} active pipeline</small></RouterLink><RouterLink to="/admin/prospects/list" class="metric-card"><span>Won Prospect<i class="pi pi-star" /></span><strong>{{ won }}</strong><small>Use WON filter in list</small></RouterLink></div><div class="dashboard-grid"><article class="dashboard-panel trend-panel"><header><div><strong>Pipeline Trend</strong><span>Current distribution by stage</span></div></header><div class="bar-chart"><div v-for="entry in stageCounts" :key="entry.status"><span>{{ entry.count }}</span><i :style="{height:`${Math.max(12,entry.count*28)}px`}" /><small>{{ entry.status.split('_').map(v=>v[0]).join('') }}</small></div></div></article><article class="dashboard-panel pipeline-summary"><header><div><strong>Prospect Pipeline</strong><span>Current stage distribution</span></div><RouterLink to="/admin/prospects/pipeline">•••</RouterLink></header><div class="open-total"><span>Open prospects</span><strong>{{ active }}</strong></div><div v-for="entry in stageCounts.slice(-4)" :key="entry.status" class="pipeline-line"><span>{{ entry.status.replaceAll('_',' ') }}</span><i><b :style="{width:`${Math.max(5,(entry.count/Math.max(1,crm.pipeline.length))*100)}%`}" /></i><strong>{{ entry.count }}</strong></div></article><article class="dashboard-panel recent-panel"><header><div><strong>Recent Pipeline Activity</strong><span>Latest records updated by the sales team</span></div><RouterLink to="/admin/prospects/pipeline">View all</RouterLink></header><table><thead><tr><th>Sales Executive</th><th>Prospect</th><th>Industry</th><th>Status</th></tr></thead><tbody><tr v-for="item in crm.pipeline.slice(0,5)" :key="item.id"><td>{{ item.assignedSalesExecutive }}</td><td>{{ item.placeName }}</td><td>{{ item.industryGroup }}</td><td><Tag :value="item.status.replaceAll('_',' ')" :severity="item.status==='WON'?'success':'info'" /></td></tr></tbody></table></article><article class="dashboard-panel assignments"><header><div><strong>Today's Assignment</strong><span>{{ assignedToday.length }} active records</span></div></header><RouterLink v-for="item in assignedToday" :key="item.id" :to="`/admin/prospects/pipeline?search=${encodeURIComponent(item.placeName)}`"><span>{{ new Date(item.updatedAt).toLocaleTimeString([],{hour:'2-digit',minute:'2-digit'}) }}</span><div><strong>{{ item.placeName }}</strong><small>{{ item.assignedSalesExecutive }}</small></div><i class="pi pi-arrow-right" /></RouterLink></article></div></section></template>

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
  border-color: var(--brand-blue, #2563eb); 
  color: var(--brand-blue, #2563eb);
  box-shadow: 0 4px 12px rgba(37, 99, 235, 0.12);
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
  background: linear-gradient(90deg, transparent, rgba(37, 99, 235, 0.3), transparent);
  opacity: 0;
  transition: opacity 0.25s ease;
}
.metric-card:hover { 
  box-shadow: 0 10px 20px -5px rgba(15, 23, 42, 0.08), 0 4px 6px -2px rgba(15, 23, 42, 0.03); 
  border-color: rgba(37, 99, 235, 0.3); 
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
  color: var(--brand-blue, #2563eb); 
  background: var(--brand-blue-bg, #eff6ff); 
  border-radius: var(--radius-sm, 8px); 
  font-size: 0.9rem;
  transition: transform 0.2s ease;
}
.metric-card:hover > span i {
  transform: scale(1.08);
}

/* Aksentuasi warna ikon berdasarkan kartu */
.metric-card:nth-child(1) > span i { color: #2563eb; background: #eff6ff; }
.metric-card:nth-child(2) > span i { color: #059669; background: #ecfdf5; }
.metric-card:nth-child(3) > span i { color: #7c3aed; background: #f5f3ff; }
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
  color: var(--brand-blue, #2563eb); 
  text-decoration: none; 
  font-size: 0.72rem; 
  font-weight: 600; 
  padding: 0.2rem 0.5rem;
  border-radius: 4px;
  background: var(--brand-blue-50, #eff6ff);
  transition: all 0.2s ease; 
}
.dashboard-panel header a:hover { 
  background: var(--brand-blue, #2563eb);
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
  background: linear-gradient(180deg, #3b82f6 0%, #1d4ed8 100%); 
  border-radius: 6px 6px 2px 2px; 
  box-shadow: 0 2px 4px rgba(37, 99, 235, 0.25);
  transition: height 0.3s cubic-bezier(0.4, 0, 0.2, 1), transform 0.2s ease; 
}
.bar-chart div:hover i {
  transform: scaleY(1.03);
  background: linear-gradient(180deg, #60a5fa 0%, #2563eb 100%);
}

/* Summary Pipeline */
.open-total { 
  margin: 1rem 0; 
  padding: 0.75rem 1rem; 
  display: flex;
  justify-content: space-between;
  align-items: center;
  color: #1e40af; 
  background: linear-gradient(135deg, #eff6ff 0%, #dbeafe 100%); 
  border: 1px solid #bfdbfe;
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
  background: linear-gradient(90deg, #3b82f6, #1d4ed8); 
  border-radius: 10px; 
  transition: width 0.4s ease;
}
.pipeline-line strong {
  text-align: right;
  color: var(--text-primary, #0f172a);
}

/* Recent Activity Table */
.recent-panel { overflow-x: auto; }
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
  background: #eff6ff; 
  border-color: #bfdbfe;
  transform: translateX(2px);
}
.assignments > a > span { 
  font-size: 0.65rem; 
  font-weight: 700;
  color: #2563eb;
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
  color: #2563eb;
  transform: translateX(3px);
}

@media (max-width: 1100px) { .metric-grid { grid-template-columns: repeat(2, 1fr); } }
@media (max-width: 900px) { .dashboard-grid { grid-template-columns: 1fr; } }
</style>