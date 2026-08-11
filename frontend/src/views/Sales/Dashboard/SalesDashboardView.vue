<script setup lang="ts">
import { computed, onMounted, onBeforeUnmount, ref } from 'vue'
import Message from 'primevue/message'
import { useAuthStore } from '../../../stores/auth'
import { useCrmStore } from '../../../stores/crm'
import type { Prospect } from '../../../types/crm'
import { isActiveProspectStatus } from '../../../utils/prospectPipeline'

const auth = useAuthStore()
const crm = useCrmStore()
const error = ref('')

const currentTime = ref(new Date())
let clockTimer: ReturnType<typeof setInterval> | undefined

onMounted(() => {
  clockTimer = setInterval(() => { currentTime.value = new Date() }, 60_000)
})
onBeforeUnmount(() => { if (clockTimer) clearInterval(clockTimer) })

const greeting = computed(() => {
  const hour = currentTime.value.getHours()
  if (hour >= 5 && hour < 12) return 'Good morning'
  if (hour >= 12 && hour < 18) return 'Good afternoon'
  return 'Good evening'
})

const firstName = computed(() => {
  const name = auth.user?.fullName
  if (!name) return 'there'
  return name.split(' ')[0] || 'there'
})

const formattedDate = computed(() =>
  new Intl.DateTimeFormat(undefined, {
    weekday: 'long',
    month: 'short',
    day: 'numeric',
  }).format(currentTime.value),
)

const activeProspects = computed(() =>
  crm.myProspects.filter((p) => isActiveProspectStatus(p.status)),
)

const todayDateKey = computed(() => {
  const d = currentTime.value
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`
})

const todayVisits = computed(() => {
  const key = todayDateKey.value
  return activeProspects.value.filter((p) => {
    if (!p.updatedAt) return false
    const d = new Date(p.updatedAt)
    const visitKey = `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`
    return visitKey === key
  })
})

const daySummaryText = computed(() => {
  const count = activeProspects.value.length
  if (count === 0) return 'No visits planned for today'
  return `${count} active visit${count !== 1 ? 's' : ''}`
})

const completed = computed(() =>
  crm.myProspects.filter((v) => ['WON', 'LOST', 'CONVERTED'].includes(v.status)).length,
)

const pendingCount = computed(() =>
  activeProspects.value.filter(
    (v) => v.status === 'NEGOTIATION' || v.status === 'CONTACTED',
  ).length,
)

function formatVisitTime(p: Prospect): string {
  if (!p.updatedAt) return 'Time not scheduled'
  const d = new Date(p.updatedAt)
  return d.toLocaleTimeString(undefined, { hour: '2-digit', minute: '2-digit', hour12: false })
}

onMounted(async () => {
  try {
    await Promise.all([crm.loadMyProspects(), crm.loadMyCustomers()])
  } catch (e: unknown) {
    error.value = crm.errorMessage(e)
  }
})
</script>

<template>
  <section class="sales-home">
    <!-- Greeting -->
    <header class="sales-dash-header">
      <RouterLink class="sales-identity" to="/sales/profile">
        <span class="sales-avatar">
          <img src="/yummy-logo.png" alt="Yummy Dairy" />
        </span>
        <div class="sales-identity-text">
          <strong>{{ greeting }}, {{ firstName }}</strong>
          <small>{{ formattedDate }}</small>
        </div>
      </RouterLink>
      <RouterLink class="sales-settings-btn" to="/sales/profile" aria-label="Open settings">
        <i class="pi pi-sliders-h" />
      </RouterLink>
    </header>

    <Message v-if="error" severity="error">{{ error }}</Message>

    <div class="dashboard-overview">
      <!-- Your day is ready -->
      <RouterLink class="ready-card" to="/sales/my-prospects">
        <div class="ready-content">
          <span class="ready-eyebrow"><i class="pi pi-calendar" /> Today's overview</span>
          <strong>{{ activeProspects.length > 0 ? 'Your day is ready' : 'Your schedule is clear' }}</strong>
          <span class="ready-summary">{{ daySummaryText }}</span>
          <div class="ready-meta">
            <span><b>{{ todayVisits.length }}</b> today</span>
            <span><b>{{ pendingCount }}</b> pending</span>
          </div>
        </div>
        <span class="ready-open"><i class="pi pi-arrow-up-right" /></span>
      </RouterLink>

      <!-- Quick statistics -->
      <section class="stats-panel">
        <div class="stats-panel-header">
          <div>
            <small>Performance</small>
            <strong>Quick statistics</strong>
          </div>
          <RouterLink to="/sales/history">View report <i class="pi pi-angle-right" /></RouterLink>
        </div>

        <div class="quick-stats">
          <RouterLink to="/sales/my-customers">
            <span class="stat-icon blue-dot"><i class="pi pi-users" /></span>
            <div class="quick-stats-info"><small>Total customers</small><strong>{{ crm.myCustomers.length }}</strong></div>
          </RouterLink>

          <RouterLink to="/sales/my-prospects">
            <span class="stat-icon amber-dot"><i class="pi pi-briefcase" /></span>
            <div class="quick-stats-info"><small>Today's prospects</small><strong>{{ todayVisits.length }}</strong></div>
          </RouterLink>

          <RouterLink to="/sales/history">
            <span class="stat-icon green-dot"><i class="pi pi-check" /></span>
            <div class="quick-stats-info"><small>Completed visits</small><strong>{{ completed }}</strong></div>
          </RouterLink>

          <RouterLink to="/sales/my-prospects">
            <span class="stat-icon red-dot"><i class="pi pi-clock" /></span>
            <div class="quick-stats-info"><small>Pending visits</small><strong>{{ pendingCount }}</strong></div>
          </RouterLink>
        </div>
      </section>
    </div>

    <!-- Quick actions -->
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

    <!-- Today's visits -->
    <div class="section-title">
      <strong>Today's visits</strong>
      <RouterLink to="/sales/my-prospects">See route</RouterLink>
    </div>

    <div class="today-list">
      <RouterLink
        v-for="(item, index) in todayVisits.slice(0, 3)"
        :key="item.id"
        :to="`/sales/my-prospects/${item.id}`"
      >
        <time>
          {{ formatVisitTime(item) }}
          <small>Today</small>
        </time>
        <i class="visit-dot" :class="index % 2 === 0 ? 'dot-amber' : 'dot-blue'" />
        <div>
          <strong>{{ item.placeName }}</strong>
          <small>{{ item.placeCategory || item.industryGroup || 'Uncategorized' }}</small>
        </div>
        <span class="visit-badge">Pending</span>
      </RouterLink>

      <div v-if="!todayVisits.length" class="empty-state">
        <strong>No visits today</strong>
        <span>No prospects updated today yet.</span>
      </div>
    </div>
  </section>
</template>

<style scoped>
.sales-home {
  width: min(100%, 1180px);
  margin: 0 auto;
  display: flex;
  flex-direction: column;
  gap: 1.35rem;
}

/* ── Greeting Header ───────────────────────────────────────── */
.sales-dash-header {
  display: flex; align-items: center; justify-content: space-between;
  padding: 0.85rem 1rem;
  background: #ffffff;
  border: 1px solid var(--border-light);
  border-radius: 14px;
  box-shadow: var(--shadow-xs);
}
.sales-identity {
  display: flex; gap: 0.75rem; align-items: center; color: #0f172a; text-decoration: none;
}
.sales-avatar {
  width: 54px; height: 40px; display: grid; place-items: center;
  padding: 0.3rem;
  border-radius: 10px; background: #fff8f8;
  border: 1px solid #ffd9dd;
  flex-shrink: 0;
  box-shadow: none;
}

.sales-avatar img {
  width: 100%;
  height: 100%;
  object-fit: contain;
}
.sales-identity-text { display: flex; flex-direction: column; }
.sales-identity-text strong { font-size: 1rem; font-weight: 800; color: #0f172a; line-height: 1.2; }
.sales-identity-text small { color: #64748b; font-size: 0.75rem; font-weight: 600; margin-top: 3px; }
.sales-settings-btn {
  width: 38px; height: 38px; display: grid; place-items: center;
  border-radius: 10px; color: #64748b; background: #f8fafc;
  border: 1px solid #e2e8f0; text-decoration: none; font-size: 0.9rem;
  box-shadow: none;
  transition: all 0.2s ease;
}
.sales-settings-btn:hover { color: #d14350; border-color: #ffd9dd; background: #fff1f2; }

/* ── Ready card ────────────────────────────────────────────── */
.ready-card {
  min-height: 124px;
  padding: 1.35rem 1.5rem; display: flex; align-items: center; justify-content: space-between;
  color: #ffffff; background: linear-gradient(120deg, #d14350 0%, #c63d4b 58%, #bb3342 100%);
  border: 1px solid rgba(173, 48, 64, 0.22);
  border-radius: 16px; text-decoration: none;
  box-shadow: 0 12px 28px -16px rgba(209, 67, 80, 0.7);
  transition: transform 0.25s ease, box-shadow 0.25s ease;
  position: relative; overflow: hidden;
}
.ready-card::before {
  content: ''; position: absolute; top: -40%; right: -20%; width: 200px; height: 200px;
  background: radial-gradient(circle, rgba(255,255,255,0.12) 0%, transparent 70%);
  border-radius: 50%; pointer-events: none;
}
.ready-card:hover { transform: translateY(-1px); box-shadow: 0 16px 30px -16px rgba(209, 67, 80, 0.78); }
.ready-card div { display: flex; flex-direction: column; gap: 0.3rem; }
.ready-card strong { font-size: 1.25rem; font-weight: 800; color: #ffffff; line-height: 1.1; }
.ready-card span { font-size: 0.88rem; color: #ffd9dd; font-weight: 650; }
.ready-card > i {
  width: 44px; height: 44px; display: grid; place-items: center; color: #d14350;
  background: #ffffff; border-radius: 12px; font-size: 0.9rem; flex-shrink: 0;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1); transition: transform 0.2s ease;
}
.ready-card:hover > i { transform: translateX(2px); }

/* ── Section titles ────────────────────────────────────────── */
.section-title { display: flex; justify-content: space-between; align-items: center; margin-top: 0.1rem; }
.section-title strong { font-size: 0.92rem; font-weight: 800; color: #0f172a; letter-spacing: -0.01em; }
.section-title a { color: #d14350; font-size: 0.8rem; text-decoration: none; font-weight: 800; transition: opacity 0.15s ease; }
.section-title a:hover { opacity: 0.75; }

.pipeline-link { display: inline-flex; align-items: center; gap: 0.3rem; padding: 0.3rem 0.65rem; border-radius: 9999px; background: #fff1f2; color: #d14350; font-size: 0.72rem; font-weight: 700; text-decoration: none; transition: all 0.15s ease; }
.pipeline-link:hover { background: #ffd9dd; opacity: 1; }
.pipeline-link i { font-size: 0.75rem; }

/* ── Quick stats ───────────────────────────────────────────── */
.quick-stats { display: grid; grid-template-columns: 1fr 1fr; gap: 0.75rem; }
.quick-stats > a {
  min-height: 108px;
  padding: 1rem; display: flex; align-items: flex-start; justify-content: space-between;
  color: #0f172a; background: #ffffff; border: 1px solid var(--border-light); border-radius: 14px;
  box-shadow: var(--shadow-xs); text-decoration: none;
  transition: all 0.2s ease;
}
.quick-stats > a:hover { border-color: #f3b9c0; box-shadow: 0 6px 16px rgba(73, 34, 41, 0.07); transform: translateY(-1px); }
.quick-stats-info { display: flex; flex-direction: column; gap: 0.6rem; }
.quick-stats-info small { color: #64748b; font-size: 0.73rem; font-weight: 600; }
.quick-stats strong { font-size: 1.75rem; font-weight: 800; color: #0f172a; line-height: 1; letter-spacing: -0.04em; }

.stat-icon {
  width: 36px; height: 36px; display: grid; place-items: center; border-radius: 10px;
  font-size: 0.85rem; font-weight: 800; flex-shrink: 0;
}
.blue-dot { color: #d14350; background: #fff1f2; }
.amber-dot { color: #d97706; background: #fffbeb; }
.green-dot { color: #16a34a; background: #f0fdf4; }
.red-dot { color: #dc2626; background: #fef2f2; }

/* ── Dashboard overview ─────────────────────────────────────────── */
.dashboard-overview {
  display: grid;
  gap: 0.9rem;
}

.ready-card {
  min-height: 250px;
  align-items: stretch;
  padding: 1.45rem;
}

.ready-card .ready-content {
  position: relative;
  z-index: 1;
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  justify-content: flex-end;
  gap: 0;
}

.ready-card .ready-eyebrow {
  width: fit-content;
  display: inline-flex;
  align-items: center;
  gap: 0.4rem;
  margin-bottom: auto;
  padding: 0.42rem 0.65rem;
  border: 1px solid rgba(255, 255, 255, 0.3);
  border-radius: 999px;
  background: rgba(255, 255, 255, 0.14);
  color: #ffffff;
  font-size: 0.66rem;
  font-weight: 800;
  letter-spacing: 0.05em;
  text-transform: uppercase;
}

.ready-card .ready-eyebrow i { font-size: 0.7rem; }
.ready-card .ready-summary { margin-top: 0.45rem; color: #ffe9eb; }

.ready-meta {
  display: flex !important;
  flex-direction: row !important;
  gap: 0.5rem !important;
  margin-top: 1rem;
}

.ready-meta span {
  padding: 0.38rem 0.58rem;
  border-radius: 8px;
  background: rgba(255, 255, 255, 0.13);
  color: #fff;
  font-size: 0.7rem;
}

.ready-meta b { font-size: 0.78rem; }

.ready-card .ready-open {
  position: absolute;
  z-index: 1;
  top: 1.35rem;
  right: 1.35rem;
  width: 38px;
  height: 38px;
  display: grid;
  place-items: center;
  border-radius: 10px;
  background: #fff;
  color: #d14350;
  box-shadow: 0 5px 14px rgba(80, 20, 30, 0.16);
}

.stats-panel {
  padding: 1rem;
  border: 1px solid var(--border-light);
  border-radius: 16px;
  background: #fff;
  box-shadow: var(--shadow-xs);
}

.stats-panel-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  padding: 0.1rem 0.15rem 0.85rem;
  border-bottom: 1px solid #f1e8ea;
}

.stats-panel-header > div { display: grid; gap: 0.08rem; }
.stats-panel-header small { color: #a39397; font-size: 0.62rem; font-weight: 750; text-transform: uppercase; letter-spacing: 0.08em; }
.stats-panel-header strong { color: #2b2022; font-size: 0.92rem; }
.stats-panel-header a { display: inline-flex; align-items: center; color: #d14350; font-size: 0.7rem; font-weight: 750; text-decoration: none; }

.stats-panel .quick-stats {
  grid-template-columns: 1fr 1fr;
  gap: 0;
}

.stats-panel .quick-stats > a {
  min-height: 88px;
  display: grid;
  grid-template-columns: 36px 1fr;
  align-items: center;
  justify-content: initial;
  gap: 0.7rem;
  padding: 0.8rem 0.65rem;
  border: 0;
  border-bottom: 1px solid #f1e8ea;
  border-radius: 0;
  box-shadow: none;
}

.stats-panel .quick-stats > a:nth-child(odd) { border-right: 1px solid #f1e8ea; }
.stats-panel .quick-stats > a:nth-last-child(-n + 2) { border-bottom: 0; }
.stats-panel .quick-stats > a:hover { background: #fff8f8; box-shadow: none; transform: none; }
.stats-panel .quick-stats-info { display: flex; flex-direction: column; gap: 0.2rem; }
.stats-panel .quick-stats-info small { order: 2; font-size: 0.65rem; }
.stats-panel .quick-stats-info strong { order: 1; font-size: 1.35rem; }

/* ── Quick actions ─────────────────────────────────────────── */
.quick-actions { display: grid; grid-template-columns: repeat(4, minmax(0, 1fr)); gap: 0.7rem; }
.quick-actions > a {
  min-height: 92px;
  padding: 0.75rem 0.25rem; display: flex; flex-direction: column; align-items: center;
  justify-content: center; gap: 0.5rem; color: #0f172a; background: #ffffff;
  border: 1px solid #e8eef7; border-radius: 22px; text-decoration: none; font-size: 0.7rem;
  font-weight: 700; text-align: center; transition: all 0.2s ease;
  box-shadow: 0 10px 24px rgba(15, 23, 42, 0.04);
}
.quick-actions > a:not(.action-primary):hover { border-color: #d0d8e4; background: #f8fafc; transform: translateY(-1px); box-shadow: 0 4px 12px rgba(0, 0, 0, 0.05); }

.quick-actions > a.action-primary {
  background: linear-gradient(135deg, #d14350 0%, #bb3342 100%); color: #ffffff;
  border-color: transparent; box-shadow: 0 6px 20px -3px rgba(209, 67, 80, 0.4);
}
.quick-actions > a.action-primary:hover { box-shadow: 0 8px 24px -3px rgba(209, 67, 80, 0.5); transform: translateY(-1px); }

.action-icon { width: 42px; height: 42px; border-radius: 16px; display: grid; place-items: center; font-size: 0.95rem; transition: transform 0.2s ease; }
.action-icon-primary { background: #ffffff; color: #d14350; box-shadow: 0 2px 8px rgba(255, 255, 255, 0.25); }
.action-icon-mint { background: #ecfdf5; color: #059669; }
.action-icon-indigo { background: #fff5f6; color: #4f46e5; }
.action-icon-amber { background: #fffbeb; color: #d97706; }

/* ── Today's visits ────────────────────────────────────────── */
.today-list {
  display: flex; flex-direction: column; gap: 0;
  overflow: hidden;
  background: #ffffff;
  border: 1px solid var(--border-light);
  border-radius: 14px;
  box-shadow: var(--shadow-xs);
}
.today-list > a {
  display: grid; grid-template-columns: auto auto 1fr auto; gap: 0.85rem; align-items: center;
  padding: 0.9rem 1rem; color: #0f172a; background: #ffffff; border: 0;
  border-bottom: 1px solid #edf1f6;
  border-radius: 0; text-decoration: none; box-shadow: none;
  transition: all 0.2s ease;
}
.today-list > a:last-of-type { border-bottom: 0; }
.today-list > a:hover { background: #fffbfb; box-shadow: none; }
.today-list time { display: flex; flex-direction: column; font-weight: 800; font-size: 0.9rem; color: #0f172a; line-height: 1.2; }
.today-list time small { color: #94a3b8; font-size: 0.7rem; font-weight: 600; }

.visit-dot { width: 8px; height: 8px; border-radius: 50%; flex-shrink: 0; }
.dot-amber { background: #f59e0b; box-shadow: 0 0 0 3px rgba(245, 158, 11, 0.15); }
.dot-blue { background: #d14350; box-shadow: 0 0 0 3px rgba(209, 67, 80, 0.15); }

.today-list div { display: grid; gap: 0.15rem; }
.today-list strong { font-size: 0.88rem; font-weight: 700; }
.today-list span, .today-list small { color: var(--text-muted); font-size: 0.78rem; }
.visit-badge { background: #fef3c7; color: #b45309; font-size: 0.7rem; font-weight: 700; border-radius: 9999px; padding: 0.3rem 0.75rem; }

.empty-state {
  min-height: 112px;
  display: grid;
  place-content: center;
  padding: 1rem; text-align: center; background: #fff; border: 0; border-radius: 0;
  box-shadow: none;
}
.empty-state strong { display: block; color: #0f172a; font-size: 0.88rem; margin-bottom: 0.25rem; }
.empty-state span { color: #94a3b8; font-size: 0.75rem; }

/* ── Desktop responsive ──────────────────────────────────────── */
@media (min-width: 768px) {
  .quick-stats { grid-template-columns: repeat(4, 1fr); }
  .dashboard-overview { grid-template-columns: minmax(270px, 0.8fr) minmax(420px, 1.2fr); }
}

@media (min-width: 1024px) {
  .sales-dash-header { padding: 0.9rem 1.1rem; }
  .sales-avatar { width: 64px; height: 46px; }
  .sales-identity-text strong { font-size: 1.25rem; }
  .sales-identity-text small { font-size: 0.85rem; }

  .ready-card { min-height: 258px; padding: 1.5rem 1.6rem; }
  .ready-card strong { font-size: 1.45rem; }
  .ready-card span { font-size: 0.95rem; }

  .section-title strong { font-size: 1.05rem; }

  .quick-stats > a { min-height: 118px; padding: 1.1rem; }
  .quick-stats strong { font-size: 2rem; }
  .quick-stats-info small { font-size: 0.78rem; }

  .quick-actions > a { min-height: 100px; font-size: 0.75rem; }
  .quick-actions { gap: 0.85rem; }
}

@media (max-width: 767px) {
  .sales-home { gap: 0.9rem; }
  .sales-dash-header { padding: 0.8rem; border-radius: 17px; box-shadow: 0 6px 18px rgba(73,34,41,.06); }
  .dashboard-overview { gap: 0.7rem; }
  .ready-card { min-height: 190px; padding: 1.1rem; border-radius: 19px; }
  .ready-card strong { font-size: 1.3rem; }
  .stats-panel { padding: 0.75rem; border-radius: 18px; }
  .stats-panel .quick-stats > a { min-height: 78px; }
  .quick-actions > a { border-radius: 17px; }
  .today-list { border-radius: 17px; }
}

@media (max-width: 380px) {
  .sales-home { gap: 1rem; }
  .ready-card { min-height: 112px; padding: 1.1rem; }
  .ready-card strong { font-size: 1.1rem; }
  .quick-stats > a { min-height: 104px; padding: 0.85rem; }
  .quick-actions { gap: 0.55rem; }
  .quick-actions > a { min-height: 84px; font-size: 0.64rem; }
}
</style>
