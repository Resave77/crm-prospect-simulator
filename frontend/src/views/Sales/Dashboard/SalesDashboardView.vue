<script setup lang="ts">
import { computed, onMounted, onBeforeUnmount, provide, ref } from 'vue'
import Message from 'primevue/message'
import { getMyVisits, getTeamDashboard } from '../../../api/crm'
import { useAuthStore } from '../../../stores/auth'
import YummySocialLinks from '../../../components/layout/YummySocialLinks.vue'
import { useCrmStore } from '../../../stores/crm'
import type { Prospect, TeamDashboard, VisitMonitoringItem } from '../../../types/crm'
import { isActiveProspectStatus } from '../../../utils/prospectPipeline'
import DailyVisitSchedule from '../../../components/prospect/DailyVisitSchedule.vue'
import ProspectHealthSummary from '../../../components/prospect/ProspectHealthSummary.vue'

const auth = useAuthStore()
const crm = useCrmStore()
const error = ref('')
const teamDashboard = ref<TeamDashboard | null>(null)
const recentVisits = ref<VisitMonitoringItem[]>([])
provide('salesRecentVisits', recentVisits)
const activityLoading = ref(false)

const currentTime = ref(new Date())
let clockTimer: ReturnType<typeof setInterval> | undefined

onMounted(() => {
  clockTimer = setInterval(() => { currentTime.value = new Date() }, 60_000)
})
onBeforeUnmount(() => { if (clockTimer) clearInterval(clockTimer) })

const greeting = computed(() => {
  const hour = currentTime.value.getHours()
  if (hour >= 5 && hour < 12) return 'Selamat pagi'
  if (hour >= 12 && hour < 18) return 'Selamat siang'
  return 'Selamat malam'
})

const firstName = computed(() => {
  const name = auth.user?.fullName
  if (!name) return 'there'
  return name.split(' ')[0] || 'there'
})

const permissionKeys = computed(() => auth.user?.salesRole?.permissionKeys ?? [])
const canViewTeamDashboard = computed(() => permissionKeys.value.includes('view_team_dashboard'))
const canViewSalesHistory = computed(() => permissionKeys.value.includes('view_sales_history'))
const canViewMyCustomers = computed(() => permissionKeys.value.includes('view_my_customers'))
const canViewSalesPipeline = computed(() => permissionKeys.value.includes('menu_sales_pipeline'))
const showTeamDashboard = computed(() => canViewTeamDashboard.value && teamDashboard.value?.hasTeam)

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

const recentActivities = computed(() => [...recentVisits.value].sort((a, b) => {
  const aTime = new Date(a.checkOutAt || a.checkInAt).getTime()
  const bTime = new Date(b.checkOutAt || b.checkInAt).getTime()
  return (Number.isFinite(bTime) ? bTime : 0) - (Number.isFinite(aTime) ? aTime : 0)
}))

const teamPipelineEntries = computed(() => {
  const counts = teamDashboard.value?.pipelineCounts ?? {}
  return Object.entries(counts)
    .filter(([, count]) => Number(count) > 0)
    .sort((a, b) => Number(b[1]) - Number(a[1]))
})

const teamPipelineMax = computed(() =>
  teamPipelineEntries.value.reduce((max, [, count]) => Math.max(max, Number(count)), 0),
)

function formatPipelineStage(status: string): string {
  return status
    .split('_')
    .map((part) => part.charAt(0) + part.slice(1).toLowerCase())
    .join(' ')
}

function pipelineBarWidth(count: number): string {
  if (!teamPipelineMax.value) return '0%'
  return `${Math.max(8, Math.round((Number(count) / teamPipelineMax.value) * 100))}%`
}

function formatVisitTime(p: Prospect): string {
  if (!p.updatedAt) return 'Time not scheduled'
  const d = new Date(p.updatedAt)
  return d.toLocaleTimeString(undefined, { hour: '2-digit', minute: '2-digit', hour12: false })
}

onMounted(async () => {
  try {
    if (canViewTeamDashboard.value) {
      teamDashboard.value = await getTeamDashboard()
      if (teamDashboard.value.hasTeam) return
    }
    await Promise.all([crm.loadMyProspects(), crm.loadMyCustomers()])
    activityLoading.value = true
    const now = new Date(currentTime.value)
    const from = new Date(now); const weekday = from.getDay() || 7
    from.setDate(now.getDate() - weekday + 1); from.setHours(0, 0, 0, 0)
    const to = new Date(from); to.setDate(from.getDate() + 5); to.setHours(23, 59, 59, 999)
    recentVisits.value = await getMyVisits({ dateFrom: from.toISOString(), dateTo: to.toISOString() })
    activityLoading.value = false
  } catch (e: unknown) {
    error.value = crm.errorMessage(e)
    activityLoading.value = false
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
      <div class="sales-dashboard-actions">
        <YummySocialLinks />
      </div>
    </header>

    <Message v-if="error" severity="error">{{ error }}</Message>

    <template v-if="showTeamDashboard && teamDashboard">
      <div class="team-dashboard-shell">
      <section class="team-overview-header">
        <div class="team-overview-copy">
          <small>{{ greeting }}, {{ auth.user?.fullName || firstName }}</small>
          <strong>Team Sales Dashboard</strong>
          <span>
            {{ teamDashboard.lead.roleName || 'Team Lead' }} · {{ formattedDate }} · Team Overview
          </span>
          <p>{{ teamDashboard.totalDescendantCount }} team member{{ teamDashboard.totalDescendantCount !== 1 ? 's' : '' }} in your reporting line.</p>
        </div>
        <div class="team-hero-actions">
          <RouterLink v-if="canViewSalesHistory" to="/sales/history"><i class="pi pi-clock" /> Riwayat</RouterLink>
          <RouterLink v-if="canViewSalesPipeline" to="/sales/pipeline"><i class="pi pi-chart-line" /> Pipeline</RouterLink>
        </div>
      </section>

      <div class="team-kpi-grid">
        <div class="team-kpi-card">
          <span class="team-kpi-icon"><i class="pi pi-users" /></span>
          <small>Team Members</small>
          <strong>{{ teamDashboard.totalDescendantCount }}</strong>
          <span>{{ teamDashboard.directMemberCount }} direct</span>
        </div>
        <div class="team-kpi-card">
          <span class="team-kpi-icon"><i class="pi pi-briefcase" /></span>
          <small>Active Prospects</small>
          <strong>{{ teamDashboard.activeProspects }}</strong>
          <span>Open pipeline work</span>
        </div>
        <div class="team-kpi-card">
          <span class="team-kpi-icon"><i class="pi pi-building" /></span>
          <small>Customers</small>
          <strong>{{ teamDashboard.customers }}</strong>
          <span>Assigned to team</span>
        </div>
        <div class="team-kpi-card">
          <span class="team-kpi-icon team-kpi-icon-green"><i class="pi pi-calendar" /></span>
          <small>Visits Today</small>
          <strong>{{ teamDashboard.visitsToday }}</strong>
          <span>{{ teamDashboard.pendingVisits }} pending</span>
        </div>
      </div>

      <div class="team-main-grid">
      <section class="team-card team-card-members">
        <div class="team-card-header">
          <div>
            <small>Team Performance</small>
            <strong>Member Summary</strong>
          </div>
          <span>{{ teamDashboard.members.length }} visible</span>
        </div>
        <div class="team-member-list">
          <article v-for="member in teamDashboard.members" :key="member.userId" class="team-member-row">
            <div class="team-member-main">
              <div class="team-member-name-row"><strong>{{ member.fullName }}</strong><a class="member-ig-icon" href="https://www.instagram.com/yummydairy/" target="_blank" rel="noopener noreferrer" title="Instagram Yummy Dairy" aria-label="Visit Instagram Yummy Dairy"><svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="2" y="2" width="20" height="20" rx="5" /><circle cx="12" cy="12" r="5" /><circle cx="17.5" cy="6.5" r="1.2" fill="currentColor" stroke="none" /></svg></a></div>
              <small>{{ member.roleName }}</small>
            </div>
            <div class="team-member-metrics">
              <span><b>{{ member.activeProspects }}</b> prospects</span>
              <span><b>{{ member.customers }}</b> customers</span>
              <span><b>{{ member.visitsToday }}</b> visits</span>
            </div>
            <div class="team-status-strip">
              <span class="team-status-positive">{{ member.completedVisits }} done</span>
              <span>{{ member.pendingVisits }} pending</span>
            </div>
          </article>
          <div v-if="!teamDashboard.members.length" class="empty-state team-empty-state">
            <strong>No active subordinates</strong>
            <span>Team members will appear here when active assignments are available.</span>
          </div>
        </div>
      </section>

      <section class="team-card">
        <div class="team-card-header">
          <div>
            <small>Pipeline Distribution</small>
            <strong>Stage Summary</strong>
          </div>
        </div>
        <div class="pipeline-bars">
          <div v-for="[status, count] in teamPipelineEntries" :key="status" class="pipeline-row">
            <div>
              <span>{{ formatPipelineStage(String(status)) }}</span>
              <strong>{{ count }}</strong>
            </div>
            <i :style="{ width: pipelineBarWidth(Number(count)) }" />
          </div>
          <div v-if="!teamPipelineEntries.length" class="empty-state team-empty-state">
            <strong>No pipeline data</strong>
            <span>No team prospects are currently assigned.</span>
          </div>
        </div>
      </section>
      </div>

      <div class="team-secondary-grid">
        <section class="team-card">
          <div class="team-card-header">
            <div>
              <small>Recent Team Activity</small>
              <strong>Visits Today</strong>
            </div>
          </div>
          <div class="visit-summary-grid">
            <div>
              <small>Completed</small>
              <strong>{{ teamDashboard.completedVisits }}</strong>
            </div>
            <div>
              <small>Pending</small>
              <strong>{{ teamDashboard.pendingVisits }}</strong>
            </div>
          </div>
        </section>

        <section class="team-card">
          <div class="team-card-header">
            <div>
              <small>Subordinate Team</small>
              <strong>Reporting Scope</strong>
            </div>
          </div>
          <div class="team-scope-row">
            <span><b>{{ teamDashboard.directMemberCount }}</b> direct reports</span>
            <span><b>{{ teamDashboard.totalDescendantCount }}</b> total members</span>
          </div>
        </section>
      </div>
      </div>
    </template>

    <template v-else>
      <section class="sales-command-header">
        <div><span class="sales-eyebrow">Sales working dashboard</span><h1>{{ greeting }}, {{ firstName }} <span aria-hidden="true">👋</span></h1><p>Berikut aktivitas dan prioritas Anda hari ini.</p></div>
        <time>{{ formattedDate }}</time>
      </section>

      <section class="sales-kpi-grid" aria-label="Ringkasan hari ini">
        <RouterLink v-if="canViewSalesPipeline" to="/sales/pipeline" class="sales-kpi kpi-attention"><span><i class="pi pi-exclamation-circle" /> Needs attention</span><strong>{{ crm.myProspects.filter(p => p.status === 'NEW_LEAD').length }}</strong><small>Prospek perlu ditindaklanjuti</small></RouterLink>
        <RouterLink v-if="canViewSalesPipeline" to="/sales/pipeline" class="sales-kpi kpi-today"><span><i class="pi pi-calendar" /> Visit today</span><strong>{{ todayVisits.length }}</strong><small>Kunjungan terjadwal hari ini</small></RouterLink>
        <RouterLink v-if="canViewSalesPipeline" to="/sales/pipeline" class="sales-kpi kpi-progress"><span><i class="pi pi-chart-line" /> In progress</span><strong>{{ activeProspects.length }}</strong><small>Prospek aktif</small></RouterLink>
        <RouterLink v-if="canViewSalesPipeline" to="/sales/pipeline" class="sales-kpi kpi-won"><span><i class="pi pi-check-circle" /> Won / converted</span><strong>{{ crm.myProspects.filter(p => ['WON', 'CONVERTED'].includes(p.status)).length }}</strong><small>Berhasil dikonversi</small></RouterLink>
      </section>

      <div class="sales-dashboard-columns"><div class="sales-dashboard-left"><DailyVisitSchedule :loading="crm.loading" :prospects="activeProspects" /></div><div class="sales-dashboard-right"><ProspectHealthSummary :prospects="crm.myProspects" /><section class="sales-mini-card activity-card"><header><h2>Aktivitas Terakhir</h2><RouterLink v-if="canViewSalesHistory" to="/sales/history">Lihat semua <i class="pi pi-arrow-right" /></RouterLink></header><div v-if="activityLoading" class="activity-empty">Memuat aktivitas...</div><div v-else-if="!recentActivities.length" class="activity-empty">Belum ada aktivitas terbaru.</div><RouterLink v-for="visit in recentActivities.slice(0, 3)" v-else :key="visit.id" class="activity-row" :to="visit.prospectId ? `/sales/my-prospects/${visit.prospectId}` : '/sales/history'"><i :class="visit.checkOutAt ? 'pi pi-check-circle activity-done' : 'pi pi-circle-fill activity-open'" /><span><b>{{ visit.checkOutAt ? 'Check-out' : 'Check-in' }} · {{ visit.customerName }}</b><small>{{ new Date(visit.checkOutAt || visit.checkInAt).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' }) }}</small></span></RouterLink></section></div></div>
    </template>
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
.sales-dash-header > :deep(.yummy-social-links) { margin-left:auto; }
.sales-dashboard-actions { display:flex; align-items:center; justify-content:flex-end; gap:.7rem; min-width:0; }
.sales-dashboard-actions :deep(.yummy-social-links) { flex-wrap:nowrap; }
.sales-identity {
  display: flex; gap: 0.75rem; align-items: center; color: #0f172a; text-decoration: none;
}
.sales-avatar {
  width: 54px; height: 40px; display: grid; place-items: center;
  padding: 0.3rem;
  border-radius: 10px; background: #fff7f7;
  border: 1px solid #ffd9dc;
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
.sales-settings-btn:hover { color: #e63946; border-color: #ffd9dc; background: #fff0f1; }
.sales-header-actions {
  display: flex;
  align-items: center;
  gap: 0.45rem;
  flex-shrink: 0;
}
.sales-ig-btn {
  width: 38px;
  height: 38px;
  display: grid;
  place-items: center;
  border-radius: 10px;
  color: #64748b;
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  text-decoration: none;
  cursor: pointer;
  transition: all .2s ease;
}
.sales-ig-btn:hover {
  background: linear-gradient(135deg, #fce18a, #ff5c87, #d942ff);
  color: #fff;
  border-color: transparent;
  box-shadow: 0 4px 14px rgba(225, 55, 90, .3);
}

/* ── Ready card ────────────────────────────────────────────── */
.ready-card {
  min-height: 124px;
  padding: 1.35rem 1.5rem; display: flex; align-items: center; justify-content: space-between;
  color: #ffffff; background: linear-gradient(120deg, #e63946 0%, #d62839 58%, #d62839 100%);
  border: 1px solid rgba(214, 40, 57, 0.22);
  border-radius: 16px; text-decoration: none;
  box-shadow: 0 12px 28px -16px rgba(230, 57, 70, 0.7);
  transition: transform 0.25s ease, box-shadow 0.25s ease;
  position: relative; overflow: hidden;
}
.ready-card::before {
  content: ''; position: absolute; top: -40%; right: -20%; width: 200px; height: 200px;
  background: radial-gradient(circle, rgba(255,255,255,0.12) 0%, transparent 70%);
  border-radius: 50%; pointer-events: none;
}
.ready-card:hover { transform: translateY(-1px); box-shadow: 0 16px 30px -16px rgba(230, 57, 70, 0.78); }
.ready-card div { display: flex; flex-direction: column; gap: 0.3rem; }
.ready-card strong { font-size: 1.25rem; font-weight: 800; color: #ffffff; line-height: 1.1; }
.ready-card span { font-size: 0.88rem; color: #ffd9dc; font-weight: 650; }
.ready-card > i {
  width: 44px; height: 44px; display: grid; place-items: center; color: #e63946;
  background: #ffffff; border-radius: 12px; font-size: 0.9rem; flex-shrink: 0;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1); transition: transform 0.2s ease;
}
.ready-card:hover > i { transform: translateX(2px); }

/* ── Section titles ────────────────────────────────────────── */
.section-title { display: flex; justify-content: space-between; align-items: center; margin-top: 0.1rem; }
.section-title strong { font-size: 0.92rem; font-weight: 800; color: #0f172a; letter-spacing: -0.01em; }
.section-title a { color: #e63946; font-size: 0.8rem; text-decoration: none; font-weight: 800; transition: opacity 0.15s ease; }
.section-title a:hover { opacity: 0.75; }

.team-dashboard-shell {
  display: grid;
  gap: 0.9rem;
}

.team-overview-header,
.team-card,
.team-kpi-card {
  border: 1px solid #dbeafe;
  background: #ffffff;
  box-shadow: 0 10px 24px rgba(15, 23, 42, 0.04);
}

.team-overview-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 1rem;
  padding: 1rem 1.05rem;
  border-radius: 16px;
}

.team-overview-copy { display: grid; gap: 0.18rem; min-width: 0; }
.team-overview-copy small,
.team-card-header small,
.team-kpi-card small,
.visit-summary-grid small {
  color: #64748b;
  font-size: 0.68rem;
  font-weight: 800;
  letter-spacing: 0.04em;
  text-transform: uppercase;
}
.team-overview-copy strong {
  color: #0f172a;
  font-size: 1.3rem;
  font-weight: 850;
  line-height: 1.15;
}
.team-overview-copy span,
.team-overview-copy p {
  color: #475569;
  font-size: 0.8rem;
  font-weight: 650;
}
.team-overview-copy p { margin: 0.28rem 0 0; }

.team-hero-actions {
  display: flex;
  gap: 0.5rem;
  flex-wrap: wrap;
  justify-content: flex-end;
}
.team-hero-actions a {
  min-height: 40px;
  display: inline-flex;
  align-items: center;
  gap: 0.4rem;
  padding: 0.5rem 0.75rem;
  border: 1px solid #bfdbfe;
  border-radius: 12px;
  color: #1d4ed8;
  background: #eff6ff;
  text-decoration: none;
  font-size: 0.75rem;
  font-weight: 800;
}

.team-kpi-grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 0.75rem;
}
.team-kpi-card {
  min-height: 96px;
  display: grid;
  grid-template-columns: 1fr auto;
  align-content: center;
  gap: 0.24rem 0.55rem;
  padding: 0.95rem;
  border-radius: 14px;
}
.team-kpi-card small,
.team-kpi-card strong,
.team-kpi-card > span:not(.team-kpi-icon) { grid-column: 1; }
.team-kpi-card strong {
  color: #0f172a;
  font-size: 1.8rem;
  line-height: 1;
}
.team-kpi-card > span:not(.team-kpi-icon) {
  color: #64748b;
  font-size: 0.75rem;
  font-weight: 650;
}
.team-kpi-icon {
  grid-column: 2;
  grid-row: 1 / span 3;
  width: 38px;
  height: 38px;
  display: grid;
  place-items: center;
  align-self: start;
  border-radius: 12px;
  color: #1d4ed8;
  background: #eff6ff;
}
.team-kpi-icon-green { color: #16a34a; background: #f0fdf4; }

.team-main-grid,
.team-secondary-grid {
  display: grid;
  grid-template-columns: minmax(0, 1.25fr) minmax(280px, 0.75fr);
  gap: 0.9rem;
  align-items: start;
}

.team-card {
  min-width: 0;
  padding: 0.9rem;
  border-radius: 16px;
}
.team-card-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 0.75rem;
  padding-bottom: 0.7rem;
  border-bottom: 1px solid #dbeafe;
}
.team-card-header > div { display: grid; gap: 0.08rem; }
.team-card-header strong {
  color: #0f172a;
  font-size: 0.95rem;
  font-weight: 850;
}
.team-card-header > span {
  color: #2563eb;
  background: #eff6ff;
  border: 1px solid #bfdbfe;
  border-radius: 999px;
  padding: 0.26rem 0.55rem;
  font-size: 0.68rem;
  font-weight: 800;
}

.team-member-list,
.pipeline-bars { display: grid; gap: 0; }
.team-member-row {
  display: grid;
  grid-template-columns: minmax(180px, 1.2fr) repeat(5, minmax(76px, auto));
  gap: 0.65rem;
  align-items: center;
  padding: 0.72rem 0.15rem;
  border-bottom: 1px solid #eaf2ff;
}
.team-member-row:last-child { border-bottom: 0; }
.team-member-main { display: grid; gap: 0.12rem; min-width: 0; }
.team-member-name-row {
  display: flex;
  align-items: center;
  gap: .35rem;
  min-width: 0;
}
.team-member-name-row strong {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.member-ig-icon {
  flex-shrink: 0;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 22px;
  height: 22px;
  border-radius: 6px;
  color: #94a3b8;
  transition: color .15s ease, background .15s ease;
}
.member-ig-icon:hover {
  background: linear-gradient(135deg, #fce18a, #ff5c87, #d942ff);
  color: #fff;
}
.team-member-main strong {
  overflow: hidden;
  color: #0f172a;
  font-size: 0.86rem;
  font-weight: 850;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.team-member-main small,
.team-member-metrics span,
.team-status-strip span {
  color: #64748b;
  font-size: 0.72rem;
  font-weight: 700;
}
.team-member-metrics {
  display: contents;
}
.team-member-metrics span {
  display: grid;
  gap: 0.05rem;
  white-space: nowrap;
}
.team-member-metrics b,
.team-scope-row b { color: #0f172a; font-size: 0.9rem; }
.team-status-strip {
  display: contents;
}
.team-status-strip span {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-height: 28px;
  padding: 0.25rem 0.48rem;
  border: 1px solid #dbeafe;
  border-radius: 999px;
  background: #f8fbff;
  white-space: nowrap;
}
.team-status-strip .team-status-positive {
  color: #15803d;
  border-color: #bbf7d0;
  background: #f0fdf4;
}

.pipeline-row {
  display: grid;
  gap: 0.4rem;
  padding: 0.72rem 0.1rem;
  border-bottom: 1px solid #eaf2ff;
}
.pipeline-row:last-child { border-bottom: 0; }
.pipeline-row div {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
}
.pipeline-row span { color: #475569; font-size: 0.78rem; font-weight: 750; }
.pipeline-row strong { color: #0f172a; font-size: 0.86rem; }
.pipeline-row i {
  height: 8px;
  border-radius: 999px;
  background: linear-gradient(90deg, #2563eb, #60a5fa);
}

.visit-summary-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0.65rem;
  padding-top: 0.75rem;
}
.visit-summary-grid div,
.team-scope-row span {
  min-height: 64px;
  display: grid;
  align-content: center;
  gap: 0.18rem;
  padding: 0.75rem;
  border: 1px solid #dbeafe;
  border-radius: 14px;
  background: #f8fbff;
}
.visit-summary-grid strong {
  color: #0f172a;
  font-size: 1.4rem;
  line-height: 1;
}
.visit-summary-grid div:first-child strong { color: #15803d; }
.team-scope-row {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0.65rem;
  padding-top: 0.75rem;
}
.team-scope-row span {
  color: #64748b;
  font-size: 0.75rem;
  font-weight: 700;
}
.team-empty-state {
  min-height: 84px;
  border-radius: 12px;
  background: #f8fbff;
}

.pipeline-link { display: inline-flex; align-items: center; gap: 0.3rem; padding: 0.3rem 0.65rem; border-radius: 9999px; background: #fff0f1; color: #e63946; font-size: 0.72rem; font-weight: 700; text-decoration: none; transition: all 0.15s ease; }
.pipeline-link:hover { background: #ffd9dc; opacity: 1; }
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
.quick-stats > a:hover { border-color: #f4b3ba; box-shadow: 0 6px 16px rgba(73, 34, 41, 0.07); transform: translateY(-1px); }
.quick-stats-info { display: flex; flex-direction: column; gap: 0.6rem; }
.quick-stats-info small { color: #64748b; font-size: 0.73rem; font-weight: 600; }
.quick-stats strong { font-size: 1.75rem; font-weight: 800; color: #0f172a; line-height: 1; letter-spacing: -0.04em; }

.stat-icon {
  width: 36px; height: 36px; display: grid; place-items: center; border-radius: 10px;
  font-size: 0.85rem; font-weight: 800; flex-shrink: 0;
}
.blue-dot { color: #e63946; background: #fff0f1; }
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
  color: #e63946;
  box-shadow: 0 5px 14px rgba(135, 28, 40, 0.16);
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
.stats-panel-header a { display: inline-flex; align-items: center; color: #e63946; font-size: 0.7rem; font-weight: 750; text-decoration: none; }

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
.stats-panel .quick-stats > a:hover { background: #fff7f7; box-shadow: none; transform: none; }
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
  background: linear-gradient(135deg, #e63946 0%, #d62839 100%); color: #ffffff;
  border-color: transparent; box-shadow: 0 6px 20px -3px rgba(230, 57, 70, 0.4);
}
.quick-actions > a.action-primary:hover { box-shadow: 0 8px 24px -3px rgba(230, 57, 70, 0.5); transform: translateY(-1px); }

.action-icon { width: 42px; height: 42px; border-radius: 16px; display: grid; place-items: center; font-size: 0.95rem; transition: transform 0.2s ease; }
.action-icon-primary { background: #ffffff; color: #e63946; box-shadow: 0 2px 8px rgba(255, 255, 255, 0.25); }
.action-icon-mint { background: #ecfdf5; color: #059669; }
.action-icon-indigo { background: #fff5f5; color: #4f46e5; }
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
.dot-blue { background: #e63946; box-shadow: 0 0 0 3px rgba(230, 57, 70, 0.15); }

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

@media (min-width: 768px) and (max-width: 1023px) {
  .team-kpi-grid { grid-template-columns: repeat(2, minmax(0, 1fr)); }
  .team-main-grid,
  .team-secondary-grid { grid-template-columns: 1fr; }
}

@media (min-width: 1024px) {
  .sales-dash-header { padding: 0.9rem 1.1rem; }
  .sales-dashboard-actions { gap:.55rem; }
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
  .team-dashboard-shell { gap: 0.7rem; }
  .team-overview-header { flex-direction: column; padding: 0.85rem; border-radius: 14px; }
  .team-overview-copy strong { font-size: 1.12rem; }
  .team-hero-actions { width: 100%; justify-content: flex-start; }
  .team-hero-actions a { flex: 1 1 120px; justify-content: center; }
  .team-kpi-grid { grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 0.6rem; }
  .team-kpi-card { min-height: 104px; padding: 0.78rem; }
  .team-kpi-card strong { font-size: 1.45rem; }
  .team-kpi-icon { width: 34px; height: 34px; border-radius: 10px; }
  .team-main-grid,
  .team-secondary-grid { grid-template-columns: 1fr; gap: 0.7rem; }
  .team-card { padding: 0.78rem; border-radius: 14px; }
  .team-member-row { grid-template-columns: 1fr; gap: 0.55rem; padding: 0.75rem 0.05rem; }
  .team-member-main strong { white-space: normal; }
  .team-member-metrics { display: grid; grid-template-columns: repeat(3, minmax(0, 1fr)); gap: 0.35rem; }
  .team-member-metrics span { min-height: 40px; padding: 0.38rem; border: 1px solid #e2e8f0; border-radius: 12px; background: #f8fafc; }
  .team-status-strip { display: flex; justify-content: flex-start; gap: 0.35rem; }
  .sales-dash-header { flex-direction:column; align-items:stretch; padding: 0.8rem; border-radius: 17px; box-shadow: 0 6px 18px rgba(73,34,41,.06); gap:.8rem; }
  .sales-dashboard-actions { width:100%; flex-direction:column; align-items:stretch; gap:0; }
  .sales-dashboard-actions :deep(.yummy-social-links) { width:100%; display:grid; grid-template-columns:repeat(3,minmax(0,1fr)); gap:.5rem; }
  .sales-dashboard-actions :deep(.yummy-social-link) { width:auto; min-height:46px; justify-content:center; padding:.45rem .35rem; border-radius:12px; }
  .sales-dashboard-actions :deep(.yummy-social-label) { display:inline; font-size:.64rem; }
  .sales-dashboard-actions :deep(.yummy-social-link > i) { display:none; }
  .sales-dashboard-actions :deep(.yummy-social-icon) { width:20px; height:20px; flex-shrink:0; }
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
  .team-kpi-grid,
  .team-member-metrics,
  .visit-summary-grid,
  .team-scope-row { grid-template-columns: 1fr; }
  .ready-card { min-height: 112px; padding: 1.1rem; }
  .ready-card strong { font-size: 1.1rem; }
  .quick-stats > a { min-height: 104px; padding: 0.85rem; }
  .quick-actions { gap: 0.55rem; }
  .quick-actions > a { min-height: 84px; font-size: 0.64rem; }
}
/* Final sales-dashboard polish: keep the visual language focused on the red brand. */
.sales-home { gap: 1.1rem; }
.sales-home { width:100%; max-width:1400px; padding-inline:clamp(1rem,2vw,1.75rem); box-sizing:border-box; }
.sales-dash-header {
  border-color: #f4d5d8;
  background: linear-gradient(135deg, #fff 0%, #fff8f8 100%);
  box-shadow: 0 8px 24px rgba(166, 28, 45, .07);
}
.ready-card { box-shadow: 0 14px 30px -16px rgba(166, 28, 45, .7); }
.ready-card { min-height: 148px; padding: 1.5rem 1.65rem; }
.ready-card > .ready-content { min-width: 0; flex: 1; }
.ready-card > .ready-open { flex: 0 0 auto; margin-left: 1rem; }
.ready-content { position: relative; z-index: 1; }
.ready-eyebrow { display: inline-flex; align-items: center; gap: .4rem; font-size: .68rem !important; letter-spacing: .07em; text-transform: uppercase; }
.ready-eyebrow i { font-size: .72rem; }
.ready-summary { font-size: .78rem !important; font-weight: 550 !important; }
.ready-open { position: relative; z-index: 1; display: inline-flex; align-items: center; gap: .55rem; padding: .65rem .8rem; border: 1px solid rgba(255,255,255,.45); border-radius: 10px; background: rgba(255,255,255,.14); color: #fff; font-size: .68rem; transition: background .2s ease, transform .2s ease; }
.ready-open i { font-size: .7rem; }
.ready-card:hover .ready-open { background: rgba(255,255,255,.24); transform: translateX(2px); }
.ready-card > i { width: auto; height: auto; background: transparent; box-shadow: none; }
.ready-card > i { color: inherit; }
.stats-panel, .quick-actions, .today-list, .team-card, .team-kpi-card, .team-overview-header {
  border-color: #f0dfe1;
  box-shadow: 0 6px 20px rgba(73, 34, 41, .05);
}
.stats-panel { overflow: hidden; }
.quick-actions > a { min-height: 94px; justify-content: center; align-items: center; gap: .7rem; }
.quick-actions > a > span:last-child { display: grid; gap: .2rem; text-align: center; }
.quick-actions > a > span:last-child b { font-size: .76rem; color: inherit; }
.quick-actions > a > span:last-child small { color: var(--text-muted); font-size: .61rem; line-height: 1.35; }
.quick-actions > a.action-primary > span:last-child small { color: #ffd9dc; }
.stat-icon.blue-dot, .stat-icon.amber-dot, .stat-icon.green-dot, .stat-icon.red-dot,
.action-icon-primary, .action-icon-mint, .action-icon-indigo, .action-icon-amber {
  color: #b4232d;
  background: #fff0f1;
}
.team-overview-header { border-left: 4px solid #e63946; }
.team-hero-actions a { color: #b4232d; border-color: #f4b3ba; background: #fff0f1; }
.team-kpi-card { border-top: 3px solid #e63946; }
.team-kpi-icon, .team-kpi-icon-green { color: #b4232d; background: #fff0f1; }
.pipeline-row i { background: linear-gradient(90deg, #e63946, #b4232d) !important; }
.visit-badge { color: #a51e2d; background: #fff0f1; border-color: #f4d5d8; }
@media (max-width: 640px) {
  .sales-home { gap: .75rem; }
  .sales-dash-header { padding: .7rem; }
  .ready-card { min-height: 116px; padding: 1rem; }
  .ready-card { min-height: 132px; }
  .ready-open { padding: .55rem .6rem; }
  .ready-open b { font-size: .58rem; }
  .ready-card > .ready-open { margin-left: .5rem; }
  .sales-identity { gap: .55rem; }
  .sales-avatar { width: 48px; height: 38px; }
  .sales-identity-text strong { font-size: .86rem; }
  .sales-identity-text small { font-size: .66rem; }
  .sales-dashboard-actions :deep(.yummy-social-links) { gap: .25rem; }
  .section-title { margin-top: .15rem; }
  .section-title strong { font-size: .84rem; }
  .section-title a { font-size: .68rem; }
  .quick-actions > a,
  .stats-panel .quick-stats > a,
  .today-list > a { min-height: 44px; }
  .quick-actions > a { padding: .7rem .55rem; }
  .action-icon { width: 36px; height: 36px; border-radius: 12px; }
  .quick-actions > a { min-height: 92px; gap: .45rem; }
  .quick-actions > a > span:last-child b { font-size: .65rem; }
  .quick-actions > a > span:last-child small { font-size: .52rem; }
  .today-list > a { grid-template-columns: 48px 8px minmax(0, 1fr) auto; gap: .55rem; padding: .75rem .7rem; }
  .today-list strong { font-size: .76rem; }
  .today-list span, .today-list small { font-size: .66rem; }
  .visit-badge { padding: .22rem .38rem; font-size: .58rem; }
}

/* Premium red visual system */
.sales-home { color: #321c21; }
.sales-dash-header { border-color: #ead1d5; background: #fff; }
.sales-avatar { background: #fff1f2; border-color: #f5c7cc; }
.sales-identity-text strong, .section-title strong { color: #321c21; }
.ready-card { background: linear-gradient(120deg, #a51e2d 0%, #c12334 52%, #e63946 100%); border-color: #a51e2d; }
.dashboard-overview { gap: 1rem; }
.stats-panel, .quick-actions, .today-list { border-color: #ead1d5; background: #fff; }
.stats-panel-header { padding-bottom: .8rem; border-bottom: 1px solid #f3e3e5; }
.stats-panel .quick-stats > a { background: linear-gradient(135deg, #fff 0%, #fffafa 100%); }
.stats-panel .quick-stats > a:hover { background: #fff1f2; }
.stat-icon { border: 1px solid #f3d4d8; }
.quick-actions > a { border-color: #ead1d5; background: #fff; box-shadow: 0 4px 12px rgba(110, 28, 42, .04); }
.quick-actions > a:hover { border-color: #e8aeb4; background: #fff7f8; box-shadow: 0 8px 18px rgba(110, 28, 42, .09); }
.quick-actions > a.action-primary { border-color: #a51e2d; background: linear-gradient(135deg, #a51e2d, #d62839); color: #fff; box-shadow: 0 9px 18px rgba(165, 30, 45, .2); }
.quick-actions > a.action-primary:hover { background: linear-gradient(135deg, #8f1d2a, #c12334); }
.action-icon-mint, .action-icon-indigo, .action-icon-amber { color: #a51e2d; background: #fff0f1; }
.today-list > a { border-bottom-color: #f3e3e5; }
.today-list > a:hover { background: #fff7f8; }
.visit-dot.dot-amber, .visit-dot.dot-blue { background: #e63946; box-shadow: 0 0 0 4px #fff0f1; }
.visit-badge { color: #a51e2d; background: #fff0f1; border-color: #f3cbd0; }
.team-card-header, .team-overview-header { border-color: #f0d9dc; }
.team-card-header > span { color: #a51e2d; background: #fff0f1; border-color: #f3cbd0; }

/* Compact sales command center */
.sales-command-header { display:flex; align-items:flex-end; justify-content:space-between; gap:1rem; padding:.35rem 0 .2rem; }
.sales-command-header h1 { margin:.2rem 0 .25rem; color:#172033; font-size:clamp(1.45rem,2.4vw,2rem); letter-spacing:-.035em; }
.sales-command-header p { margin:0; color:#64748b; font-size:.86rem; }
.sales-command-header time { color:#64748b; font-size:.78rem; font-weight:700; white-space:nowrap; }
.sales-eyebrow { color:#e63946; font-size:.62rem; font-weight:800; letter-spacing:.11em; text-transform:uppercase; }
.sales-kpi-grid { display:grid; grid-template-columns:repeat(4,minmax(0,1fr)); gap:.8rem; }
.sales-kpi { min-width:0; padding:1rem 1.05rem; border:1px solid #e5eaf0; border-radius:14px; background:#fff; color:#172033; text-decoration:none; box-shadow:0 4px 14px rgba(15,23,42,.04); transition:.2s ease; }
.sales-kpi:hover { transform:translateY(-2px); border-color:#f2b5bb; box-shadow:0 8px 20px rgba(15,23,42,.08); }
.sales-kpi span { display:flex; align-items:center; gap:.4rem; color:#64748b; font-size:.68rem; font-weight:800; text-transform:uppercase; letter-spacing:.04em; }
.sales-kpi span i { color:var(--accent,#e63946); }.sales-kpi strong { display:block; margin:.35rem 0 .15rem; font-size:1.8rem; letter-spacing:-.04em; }.sales-kpi small { color:#64748b; font-size:.72rem; }.kpi-today { --accent:#f59e0b; }.kpi-progress { --accent:#3b82f6; }.kpi-won { --accent:#16a34a; }
.sales-dashboard-grid { display:grid; grid-template-columns:minmax(0,1.55fr) minmax(280px,1fr); gap:1rem; align-items:start; }.sales-dashboard-grid > .schedule-card { display:contents; }.sales-dashboard-grid :deep(.schedule-heading), .sales-dashboard-grid :deep(.schedule-state), .sales-dashboard-grid :deep(.schedule-list) { grid-column:1; grid-row:1; }.sales-dashboard-grid :deep(.inline-weekly-route) { grid-column:1 / -1; grid-row:2; }.sales-dashboard-grid > .health-card { grid-column:2; grid-row:1; }.activity-card { grid-column:1 / -1; grid-row:3; }
.sales-dashboard-grid :deep(.schedule-heading), .sales-dashboard-grid :deep(.schedule-list), .sales-dashboard-grid :deep(.inline-weekly-route), .sales-dashboard-grid > .health-card, .activity-card { border:1px solid #e5eaf0; border-radius:14px; background:#fff; box-shadow:0 4px 14px rgba(15,23,42,.04); }
.sales-dashboard-grid :deep(.schedule-heading), .sales-dashboard-grid :deep(.schedule-list) { padding:1rem 1.1rem; }
.sales-dashboard-grid :deep(.inline-weekly-route) { padding:1.1rem; }
.sales-dashboard-grid :deep(.inline-weekly-route) { width:100%; }
.sales-dashboard-grid :deep(.inline-route-heading .route-header-actions) { width:auto; }
.sales-dashboard-columns { display:grid; grid-template-columns:minmax(0,2fr) minmax(300px,1fr); gap:1rem; align-items:start; }.sales-dashboard-left,.sales-dashboard-right { display:flex; min-width:0; flex-direction:column; gap:1rem; }
.sales-dashboard-grid > .activity-card { grid-column:1 / -1; }
.sales-dashboard-main > .schedule-card { min-width:0; }.sales-quick-links { display:flex; align-items:center; justify-content:space-between; gap:1rem; padding:1rem 1.1rem; border:1px solid #e5eaf0; border-radius:14px; background:#fff; }.sales-quick-links h2 { margin:.2rem 0 0; font-size:1.05rem; }.sales-quick-links nav { display:flex; flex-wrap:wrap; gap:.5rem; }.sales-quick-links a { display:inline-flex; align-items:center; gap:.4rem; padding:.55rem .7rem; border:1px solid #e5eaf0; border-radius:9px; color:#334155; font-size:.76rem; font-weight:700; text-decoration:none; }.sales-quick-links a:hover { color:#e63946; border-color:#f2b5bb; background:#fff7f7; }
.sales-secondary-grid { display:grid; grid-template-columns:repeat(2,minmax(0,1fr)); gap:1rem; }.sales-mini-card { padding:1rem 1.1rem; border:1px solid #e5eaf0; border-radius:14px; background:#fff; box-shadow:0 4px 14px rgba(15,23,42,.04); }.sales-mini-card header { display:flex; align-items:center; justify-content:space-between; gap:.75rem; }.sales-mini-card h2 { margin:0; color:#172033; font-size:1rem; }.sales-mini-card header a { color:#e63946; font-size:.7rem; font-weight:700; text-decoration:none; }.summary-metrics { display:grid; grid-template-columns:repeat(4,1fr); gap:.5rem; margin:1rem 0; }.summary-metrics span { display:grid; gap:.15rem; color:#64748b; font-size:.63rem; }.summary-metrics b { color:#172033; font-size:1.25rem; }.completion-row { display:flex; justify-content:space-between; color:#64748b; font-size:.68rem; }.completion-row strong { color:#172033; }.completion-track { height:7px; margin-top:.4rem; overflow:hidden; border-radius:99px; background:#eef2f7; }.completion-track i { display:block; height:100%; border-radius:inherit; background:#e63946; }.activity-row { display:flex; align-items:center; gap:.65rem; padding:.6rem 0; border-bottom:1px solid #f1f5f9; color:#172033; text-decoration:none; }.activity-row:last-child { border-bottom:0; }.activity-row span { display:grid; min-width:0; gap:.15rem; }.activity-row b { overflow:hidden; font-size:.74rem; text-overflow:ellipsis; white-space:nowrap; }.activity-row small { color:#64748b; font-size:.65rem; }.activity-done { color:#16a34a; }.activity-open { color:#f59e0b; font-size:.55rem; }.activity-empty { padding:1rem 0; color:#64748b; font-size:.72rem; }
@media (max-width: 900px) { .sales-kpi-grid { grid-template-columns:repeat(2,minmax(0,1fr)); }.sales-dashboard-columns { grid-template-columns:1fr; }.sales-dashboard-grid { grid-template-columns:1fr; }.sales-dashboard-grid :deep(.schedule-heading), .sales-dashboard-grid :deep(.schedule-state), .sales-dashboard-grid :deep(.schedule-list), .sales-dashboard-grid :deep(.inline-weekly-route), .sales-dashboard-grid > .health-card,.activity-card { grid-column:1; grid-row:auto; } }
@media (max-width: 560px) { .sales-home { padding-inline:.75rem; }.sales-command-header { align-items:flex-start; flex-direction:column; gap:.3rem; }.sales-kpi { padding:.8rem .75rem; }.sales-kpi strong { font-size:1.5rem; }.sales-kpi small { font-size:.65rem; }.sales-dashboard-grid :deep(.schedule-heading), .sales-dashboard-grid :deep(.schedule-list), .sales-dashboard-grid :deep(.inline-weekly-route), .sales-dashboard-grid > .health-card, .activity-card { border-radius:12px; }.sales-dashboard-grid :deep(.schedule-heading), .sales-dashboard-grid :deep(.schedule-list), .sales-dashboard-grid :deep(.inline-weekly-route) { padding:.85rem; } }
</style>
