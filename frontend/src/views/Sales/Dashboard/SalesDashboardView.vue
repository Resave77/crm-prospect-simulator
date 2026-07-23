<script setup lang="ts">import {computed,onMounted,ref} from 'vue'; import Message from 'primevue/message'; import Tag from 'primevue/tag'; import {useAuthStore} from '../../../stores/auth'; import {useCrmStore} from '../../../stores/crm'; const auth=useAuthStore(),crm=useCrmStore(),error=ref(''); const active=computed(()=>crm.myProspects.filter(v=>!['LOST','CONVERTED'].includes(v.status))); const completed=computed(()=>crm.myProspects.filter(v=>['WON','LOST','CONVERTED'].includes(v.status)).length); onMounted(async()=>{try{await Promise.all([crm.loadMyProspects(),crm.loadMyCustomers()])}catch(e){error.value=crm.errorMessage(e)}})</script>
<template>
  <section class="sales-home">
    <Message v-if="error" severity="error">{{ error }}</Message>

    <RouterLink class="ready-card" to="/sales/my-prospects">
      <div>
        <strong>Your day is ready</strong>
        <span>{{ active.length }} visits planned · First at 09:30</span>
      </div>
      <i class="pi pi-arrow-right" />
    </RouterLink>

    <div class="section-title">
      <strong>Quick statistics</strong>
      <RouterLink to="/sales/history">View all</RouterLink>
    </div>

    <div class="quick-stats">
      <RouterLink to="/sales/my-customers">
        <div class="quick-stats-info">
          <span class="stat-icon blue-dot">C</span>
          <small>Today's customers</small>
        </div>
        <strong>{{ crm.myCustomers.length }}</strong>
      </RouterLink>

      <RouterLink to="/sales/my-prospects">
        <div class="quick-stats-info">
          <span class="stat-icon amber-dot">P</span>
          <small>Today's prospects</small>
        </div>
        <strong>{{ active.length }}</strong>
      </RouterLink>

      <RouterLink to="/sales/history">
        <div class="quick-stats-info">
          <span class="stat-icon green-dot"><i class="pi pi-check" /></span>
          <small>Completed visits</small>
        </div>
        <strong>{{ completed }}</strong>
      </RouterLink>

      <RouterLink to="/sales/my-prospects">
        <div class="quick-stats-info">
          <span class="stat-icon red-dot"><i class="pi pi-clock" /></span>
          <small>Pending visits</small>
        </div>
        <strong>{{ active.filter(v => v.status === 'NEGOTIATION' || v.status === 'CONTACTED').length }}</strong>
      </RouterLink>
    </div>

    <div class="section-title">
      <strong>Quick actions</strong>
    </div>

    <div class="quick-actions">
      <RouterLink to="/sales/my-prospects" class="action-primary">
        <span class="action-icon"><i class="pi pi-play" /></span>
        <span>Start visit</span>
      </RouterLink>

      <RouterLink to="/sales/my-prospects">
        <span class="action-icon"><i class="pi pi-map-marker" /></span>
        <span>Open maps</span>
      </RouterLink>

      <RouterLink to="/sales/my-customers">
        <span class="action-icon text-icon">C</span>
        <span>Customer</span>
      </RouterLink>

      <RouterLink to="/sales/my-prospects">
        <span class="action-icon text-icon">P</span>
        <span>Prospect</span>
      </RouterLink>
    </div>

    <div class="section-title">
      <strong>Today's visits</strong>
      <RouterLink to="/sales/my-prospects">See route</RouterLink>
    </div>

    <div class="today-list">
      <RouterLink
        v-for="(item, index) in active.slice(0, 3)"
        :key="item.id"
        :to="`/sales/my-prospects/${item.id}`"
      >
        <time>
          {{ `${9 + index * 2}:30` }}
          <small>Today</small>
        </time>
        <i class="visit-dot" :class="index % 2 === 0 ? 'dot-amber' : 'dot-blue'" />
        <div>
          <strong>{{ item.placeName }}</strong>
          <small>{{ item.placeCategory || item.industryGroup }} · 1.8 km away</small>
        </div>
        <span class="visit-badge">Pending</span>
      </RouterLink>

      <div v-if="!active.length" class="empty-state">
        <strong>No visits queued</strong>
      </div>
    </div>
  </section>
</template>

<style scoped>
.sales-home { display: flex; flex-direction: column; gap: 1.25rem; }
.ready-card {
  padding: 1.15rem 1.25rem; display: flex; align-items: center; justify-content: space-between;
  color: #ffffff; background: linear-gradient(135deg, #2563eb, #1d4ed8); border-radius: 20px;
  text-decoration: none; box-shadow: 0 8px 22px -4px rgba(37, 99, 235, 0.35);
  transition: transform 0.2s ease, box-shadow 0.2s ease;
}
.ready-card:hover { transform: translateY(-2px); box-shadow: 0 12px 28px -4px rgba(37, 99, 235, 0.45); }
.ready-card div { display: flex; flex-direction: column; gap: 0.2rem; }
.ready-card strong { font-size: 1.1rem; font-weight: 800; color: #ffffff; }
.ready-card span { font-size: 0.8rem; color: #dbeafe; }
.ready-card i { width: 42px; height: 42px; display: grid; place-items: center; color: #2563eb; background: #ffffff; border-radius: 50%; font-size: 1.05rem; flex-shrink: 0; box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08); }
.section-title { display: flex; justify-content: space-between; align-items: center; margin-top: 0.15rem; }
.section-title strong { font-size: 0.95rem; font-weight: 800; color: #0f172a; }
.section-title a { color: #2563eb; font-size: 0.8rem; text-decoration: none; font-weight: 700; }
.quick-stats { display: grid; grid-template-columns: 1fr 1fr; gap: 0.75rem; }
.quick-stats > a {
  padding: 1rem 1.15rem; display: flex; align-items: center; justify-content: space-between;
  color: #0f172a; background: #ffffff; border: 1px solid #e2e8f0; border-radius: 18px;
  box-shadow: 0 2px 6px rgba(0, 0, 0, 0.02); text-decoration: none; transition: all 0.2s ease;
}
.quick-stats > a:hover { border-color: #cbd5e1; box-shadow: 0 4px 12px rgba(0, 0, 0, 0.05); transform: translateY(-1px); }
.quick-stats-info { display: flex; flex-direction: column; gap: 0.4rem; }
.quick-stats-info small { color: #64748b; font-size: 0.75rem; font-weight: 600; }
.quick-stats strong { font-size: 1.5rem; font-weight: 800; color: #0f172a; line-height: 1; }
.stat-icon { width: 36px; height: 36px; display: grid; place-items: center; border-radius: 50%; font-size: 0.85rem; font-weight: 800; flex-shrink: 0; }
.blue-dot { color: #2563eb; background: #eff6ff; }
.amber-dot { color: #d97706; background: #fff7ed; }
.green-dot { color: #16a34a; background: #f0fdf4; }
.red-dot { color: #dc2626; background: #fef2f2; }
.quick-actions { display: grid; grid-template-columns: repeat(4, 1fr); gap: 0.6rem; }
.quick-actions > a {
  padding: 0.85rem 0.35rem; display: flex; flex-direction: column; align-items: center;
  justify-content: center; gap: 0.45rem; color: #0f172a; background: #ffffff;
  border: 1px solid #e2e8f0; border-radius: 18px; text-decoration: none; font-size: 0.72rem;
  font-weight: 700; text-align: center; transition: all 0.2s ease;
}
.quick-actions > a.action-primary { background: #2563eb; color: #ffffff; border-color: #2563eb; box-shadow: 0 4px 14px rgba(37, 99, 235, 0.3); }
.quick-actions > a.action-primary:hover { background: #1d4ed8; }
.quick-actions > a:not(.action-primary):hover { border-color: #2563eb; background: #eff6ff; color: #2563eb; }
.action-icon { width: 36px; height: 36px; border-radius: 50%; display: grid; place-items: center; font-size: 0.9rem; }
.action-primary .action-icon { background: #ffffff; color: #2563eb; }
.quick-actions > a:not(.action-primary) .action-icon { background: #f1f5f9; color: #0f172a; }
.quick-actions > a:not(.action-primary) .action-icon.text-icon { background: #eff6ff; color: #2563eb; font-weight: 800; }
.today-list { display: flex; flex-direction: column; gap: 0.75rem; }
.today-list > a {
  display: grid; grid-template-columns: auto auto 1fr auto; gap: 0.85rem; align-items: center;
  padding: 1rem 1.15rem; color: #0f172a; background: #ffffff; border: 1px solid #e2e8f0;
  border-radius: 18px; text-decoration: none; box-shadow: 0 2px 6px rgba(0,0,0,0.02); transition: all 0.2s ease;
}
.today-list > a:hover { border-color: #cbd5e1; box-shadow: 0 4px 12px rgba(0, 0, 0, 0.05); }
.today-list time { display: flex; flex-direction: column; font-weight: 800; font-size: 0.92rem; color: #0f172a; line-height: 1.2; }
.today-list time small { color: #94a3b8; font-size: 0.72rem; font-weight: 600; }
.visit-dot { width: 8px; height: 8px; border-radius: 50%; flex-shrink: 0; background: var(--brand-blue); }
.dot-amber { background: #f59e0b; }
.dot-blue { background: #2563eb; }
.today-list div { display: grid; gap: 0.15rem; }
.today-list strong { font-size: 0.92rem; font-weight: 700; }
.today-list span, .today-list small { color: var(--text-muted); font-size: 0.8rem; }
.visit-badge { background: #fef3c7; color: #b45309; font-size: 0.72rem; font-weight: 700; border-radius: 9999px; padding: 0.35rem 0.85rem; }
</style>
