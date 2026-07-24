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
      <RouterLink :to="{ name: 'SalesPipeline' }" class="pipeline-link">
        <i class="pi pi-chart-bar" />
        Sales Pipeline
      </RouterLink>
    </div>

    <div class="quick-actions">
      <RouterLink to="/sales/my-prospects" class="action-primary">
        <span class="action-icon action-icon-primary"><i class="pi pi-play" /></span>
        <span>Start visit</span>
      </RouterLink>

      <RouterLink to="/sales/my-prospects">
        <span class="action-icon action-icon-mint"><i class="pi pi-map-marker" /></span>
        <span>Open maps</span>
      </RouterLink>

      <RouterLink to="/sales/my-customers">
        <span class="action-icon action-icon-indigo"><i class="pi pi-users" /></span>
        <span>Customer</span>
      </RouterLink>

      <RouterLink to="/sales/my-prospects">
        <span class="action-icon action-icon-amber"><i class="pi pi-briefcase" /></span>
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
.sales-home { display: flex; flex-direction: column; gap: 1.5rem; }

.ready-card {
  padding: 1.25rem 1.35rem; display: flex; align-items: center; justify-content: space-between;
  color: #ffffff; background: linear-gradient(135deg, #2563eb 0%, #1e40af 50%, #1d4ed8 100%);
  border-radius: 20px; text-decoration: none;
  box-shadow: 0 8px 24px -4px rgba(37, 99, 235, 0.4), 0 2px 8px rgba(37, 99, 235, 0.15);
  transition: transform 0.25s ease, box-shadow 0.25s ease;
  position: relative; overflow: hidden;
}

.ready-card::before {
  content: ''; position: absolute; top: -40%; right: -20%; width: 200px; height: 200px;
  background: radial-gradient(circle, rgba(255,255,255,0.12) 0%, transparent 70%);
  border-radius: 50%; pointer-events: none;
}

.ready-card:hover { transform: translateY(-2px); box-shadow: 0 14px 32px -4px rgba(37, 99, 235, 0.5), 0 4px 12px rgba(37, 99, 235, 0.2); }
.ready-card div { display: flex; flex-direction: column; gap: 0.3rem; }
.ready-card strong { font-size: 1.1rem; font-weight: 800; color: #ffffff; }
.ready-card span { font-size: 0.8rem; color: #bfdbfe; font-weight: 500; }
.ready-card > i {
  width: 44px; height: 44px; display: grid; place-items: center; color: #2563eb;
  background: #ffffff; border-radius: 50%; font-size: 1rem; flex-shrink: 0;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1); transition: transform 0.2s ease;
}
.ready-card:hover > i { transform: translateX(2px); }

.section-title { display: flex; justify-content: space-between; align-items: center; }
.section-title strong { font-size: 0.92rem; font-weight: 800; color: #0f172a; }
.section-title a { color: #2563eb; font-size: 0.78rem; text-decoration: none; font-weight: 700; transition: opacity 0.15s ease; }
.section-title a:hover { opacity: 0.75; }

.pipeline-link { display: inline-flex; align-items: center; gap: 0.3rem; padding: 0.3rem 0.65rem; border-radius: 9999px; background: #eff6ff; color: #2563eb; font-size: 0.72rem; font-weight: 700; text-decoration: none; transition: all 0.15s ease; }
.pipeline-link:hover { background: #dbeafe; opacity: 1; }
.pipeline-link i { font-size: 0.75rem; }

.quick-stats { display: grid; grid-template-columns: 1fr 1fr; gap: 0.7rem; }
.quick-stats > a {
  padding: 1rem 1.1rem; display: flex; align-items: center; justify-content: space-between;
  color: #0f172a; background: #ffffff; border: 1px solid #eef1f6; border-radius: 18px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.03); text-decoration: none;
  transition: all 0.2s ease;
}
.quick-stats > a:hover { border-color: #d6dce6; box-shadow: 0 4px 14px rgba(0, 0, 0, 0.06); transform: translateY(-1px); }
.quick-stats-info { display: flex; flex-direction: column; gap: 0.5rem; }
.quick-stats-info small { color: #64748b; font-size: 0.73rem; font-weight: 600; }
.quick-stats strong { font-size: 1.6rem; font-weight: 800; color: #0f172a; line-height: 1; }

.stat-icon {
  width: 38px; height: 38px; display: grid; place-items: center; border-radius: 12px;
  font-size: 0.85rem; font-weight: 800; flex-shrink: 0;
}
.blue-dot { color: #2563eb; background: #eff6ff; }
.amber-dot { color: #d97706; background: #fffbeb; }
.green-dot { color: #16a34a; background: #f0fdf4; }
.red-dot { color: #dc2626; background: #fef2f2; }

.quick-actions { display: grid; grid-template-columns: repeat(4, 1fr); gap: 0.55rem; }
.quick-actions > a {
  padding: 0.9rem 0.3rem; display: flex; flex-direction: column; align-items: center;
  justify-content: center; gap: 0.5rem; color: #0f172a; background: #ffffff;
  border: 1px solid #eef1f6; border-radius: 18px; text-decoration: none; font-size: 0.7rem;
  font-weight: 700; text-align: center; transition: all 0.2s ease;
}
.quick-actions > a:not(.action-primary):hover { border-color: #d0d8e4; background: #f8fafc; transform: translateY(-1px); box-shadow: 0 4px 12px rgba(0, 0, 0, 0.05); }

.quick-actions > a.action-primary {
  background: linear-gradient(135deg, #2563eb 0%, #1d4ed8 100%); color: #ffffff;
  border-color: transparent; box-shadow: 0 6px 20px -3px rgba(37, 99, 235, 0.4);
}
.quick-actions > a.action-primary:hover { box-shadow: 0 8px 24px -3px rgba(37, 99, 235, 0.5); transform: translateY(-1px); }

.action-icon { width: 38px; height: 38px; border-radius: 14px; display: grid; place-items: center; font-size: 0.9rem; transition: transform 0.2s ease; }
.action-icon-primary { background: #ffffff; color: #2563eb; box-shadow: 0 2px 8px rgba(255, 255, 255, 0.25); }
.action-icon-mint { background: #ecfdf5; color: #059669; }
.action-icon-indigo { background: #eef2ff; color: #4f46e5; }
.action-icon-amber { background: #fffbeb; color: #d97706; }

.today-list { display: flex; flex-direction: column; gap: 0.65rem; }
.today-list > a {
  display: grid; grid-template-columns: auto auto 1fr auto; gap: 0.85rem; align-items: center;
  padding: 1rem 1.1rem; color: #0f172a; background: #ffffff; border: 1px solid #eef1f6;
  border-radius: 18px; text-decoration: none; box-shadow: 0 1px 3px rgba(0,0,0,0.02);
  transition: all 0.2s ease;
}
.today-list > a:hover { border-color: #d6dce6; box-shadow: 0 4px 14px rgba(0, 0, 0, 0.06); }
.today-list time { display: flex; flex-direction: column; font-weight: 800; font-size: 0.9rem; color: #0f172a; line-height: 1.2; }
.today-list time small { color: #94a3b8; font-size: 0.7rem; font-weight: 600; }

.visit-dot { width: 8px; height: 8px; border-radius: 50%; flex-shrink: 0; }
.dot-amber { background: #f59e0b; box-shadow: 0 0 0 3px rgba(245, 158, 11, 0.15); }
.dot-blue { background: #2563eb; box-shadow: 0 0 0 3px rgba(37, 99, 235, 0.15); }

.today-list div { display: grid; gap: 0.15rem; }
.today-list strong { font-size: 0.88rem; font-weight: 700; }
.today-list span, .today-list small { color: var(--text-muted); font-size: 0.78rem; }
.visit-badge { background: #fef3c7; color: #b45309; font-size: 0.7rem; font-weight: 700; border-radius: 9999px; padding: 0.3rem 0.75rem; }
</style>
