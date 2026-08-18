<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import Button from 'primevue/button'
import Message from 'primevue/message'
import { useAuthStore } from '../../stores/auth'
import { useCrmStore } from '../../stores/crm'
import { getMyProspect } from '../../api/crm'
import type { ProspectReview } from '../../types/crm'

const auth = useAuthStore()
const crm = useCrmStore()
const router = useRouter()

const loading = ref(true)
const error = ref('')
const prospectReviews = ref<ProspectReview[]>([])

const initials = computed(() => {
  const name = auth.user?.fullName ?? ''
  return name.split(/\s+/).slice(0, 2).map((w) => w.charAt(0).toUpperCase()).join('')
})

const roleLabel = computed(() => {
  if (auth.user?.salesRole?.name) return auth.user.salesRole.name
  if (auth.user?.role === 'ADMINISTRATOR') return 'Administrator'
  if (auth.user?.role === 'SALES_MANAGER') return 'Sales Manager'
  if (auth.user?.role === 'SALES_EXECUTIVE') return 'Sales Executive'
  return auth.user?.role ?? '—'
})

const allVisits = computed(() => {
  const visits = prospectReviews.value.flatMap((r) => r.visits ?? [])
  const seen = new Set<string>()
  return visits.filter((v) => {
    if (seen.has(v.id)) return false
    seen.add(v.id)
    return true
  })
})

const totalVisits = computed(() => allVisits.value.length)

const wonProspects = computed(() =>
  crm.myProspects.filter((p) => p.status === 'WON').length,
)

const completedVisits = computed(() =>
  allVisits.value.filter((v) => Boolean(v.checkOutAt)).length,
)

async function loadStats() {
  loading.value = true
  error.value = ''
  try {
    await crm.loadMyProspects()
    const reviews = await Promise.all(
      crm.myProspects.map((p) => getMyProspect(p.id).catch(() => null)),
    )
    prospectReviews.value = reviews.filter((r): r is ProspectReview => r !== null)
  } catch (caught) {
    error.value = crm.errorMessage(caught)
  } finally {
    loading.value = false
  }
}

async function logout() {
  await auth.logout()
  await router.replace('/login')
}

onMounted(loadStats)
</script>

<template>
  <section class="profile-page">
    <Button class="profile-back" icon="pi pi-arrow-left" severity="secondary" text rounded @click="router.back()" title="Back" />
    <!-- Header -->
    <div class="profile-header">
      <div class="profile-header-left">
        <div class="profile-header-avatar" aria-hidden="true">{{ initials }}</div>
        <div class="profile-header-text">
          <h1>Profile</h1>
          <p>Account & preferences</p>
        </div>
      </div>
      <button class="profile-header-action" aria-label="Sign out" @click="logout">
        <i class="pi pi-sign-out" />
      </button>
    </div>

    <!-- Loading skeleton -->
    <template v-if="loading">
      <div class="profile-card skeleton-card">
        <div class="sk-circle" />
        <div class="sk-line w50" />
        <div class="sk-line w30" />
        <div class="sk-strip">
          <div class="sk-strip-item"><div class="sk-label" /><div class="sk-value" /></div>
          <div class="sk-strip-item"><div class="sk-label" /><div class="sk-value" /></div>
          <div class="sk-strip-item"><div class="sk-label" /><div class="sk-value" /></div>
        </div>
      </div>
      <div class="perf-section skeleton-perf">
        <div class="sk-section-title" />
        <div class="sk-stats">
          <div class="sk-stat" />
          <div class="sk-stat" />
          <div class="sk-stat" />
        </div>
      </div>
    </template>

    <!-- Profile + Performance (loaded) -->
    <template v-else>
      <!-- Profile Identity Card -->
      <div class="profile-card">
        <div class="profile-card-top">
          <span class="profile-card-kicker">Sales account</span>
          <span class="profile-status"><i class="pi pi-circle-fill" /> Aktif</span>
        </div>
        <div class="profile-card-identity">
          <div class="profile-card-avatar" aria-hidden="true">{{ initials }}</div>
          <div class="profile-card-copy">
            <h2 class="profile-card-name">{{ auth.user?.fullName ?? '—' }}</h2>
            <p class="profile-card-role">{{ roleLabel }}</p>
          </div>
        </div>

        <div class="profile-card-strip">
          <div class="strip-item">
            <span class="strip-icon"><i class="pi pi-id-card" /></span>
            <span class="strip-copy"><span class="strip-label">Employee ID</span><span class="strip-value" :title="auth.user?.employeeId ?? ''">{{ auth.user?.employeeId ?? '—' }}</span></span>
          </div>
          <div class="strip-item">
            <span class="strip-icon"><i class="pi pi-envelope" /></span>
            <span class="strip-copy"><span class="strip-label">Email address</span><span class="strip-value" :title="auth.user?.email ?? ''">{{ auth.user?.email ?? '—' }}</span></span>
          </div>
          <div class="strip-item">
            <span class="strip-icon"><i class="pi pi-phone" /></span>
            <span class="strip-copy"><span class="strip-label">Phone number</span><span class="strip-value">{{ auth.user?.phone ?? '—' }}</span></span>
          </div>
        </div>

        <div class="account-security">
          <div class="security-copy">
            <span class="security-kicker">Account security</span>
            <strong>Password</strong>
            <span class="security-mask" aria-label="Password is hidden">••••••••</span>
          </div>
          <Button label="Change Password" icon="pi pi-key" outlined size="small" @click="router.push('/change-password')" />
        </div>
        <div class="profile-card-strip"><div class="strip-item"><span class="strip-icon"><i class="pi pi-map-marker" /></span><span class="strip-copy"><span class="strip-label">Location</span><span class="strip-value">{{ [auth.user?.city, auth.user?.province, auth.user?.district].filter(Boolean).join(', ') || '—' }}</span></span></div><div class="strip-item"><span class="strip-icon"><i class="pi pi-briefcase" /></span><span class="strip-copy"><span class="strip-label">Job</span><span class="strip-value">{{ [auth.user?.jobTitle, auth.user?.positionGrade, auth.user?.subDepartment].filter(Boolean).join(' · ') || '—' }}</span></span></div><div class="strip-item"><span class="strip-icon"><i class="pi pi-calendar" /></span><span class="strip-copy"><span class="strip-label">Join date</span><span class="strip-value">{{ auth.user?.joinDate?.slice(0, 10) || '—' }}</span></span></div></div>
      </div>

      <!-- Error in stats -->
      <Message v-if="error" class="profile-error" severity="error" closable @close="error = ''">
        {{ error }}
        <Button label="Retry" size="small" text class="retry-btn" @click="loadStats" />
      </Message>

      <!-- Performance Section -->
      <div class="perf-section">
        <div class="perf-heading"><div><span>Overview</span><h2 class="perf-title">Performance</h2></div><i class="pi pi-chart-line" /></div>
        <div class="perf-stats">
          <div class="stat-card stat-blue">
            <span class="stat-icon"><i class="pi pi-map-marker" /></span>
            <span class="stat-value">{{ totalVisits }}</span>
            <span class="stat-label">Total Visits</span>
          </div>
          <div class="stat-card stat-green">
            <span class="stat-icon"><i class="pi pi-trophy" /></span>
            <span class="stat-value">{{ wonProspects }}</span>
            <span class="stat-label">Won Prospects</span>
          </div>
          <div class="stat-card stat-purple">
            <span class="stat-icon"><i class="pi pi-check-circle" /></span>
            <span class="stat-value">{{ completedVisits }}</span>
            <span class="stat-label">Completed Visits</span>
          </div>
        </div>
      </div>
    </template>
  </section>
</template>

<style scoped>
.profile-page {
  display: grid; gap: 0.85rem; width: 100%;
  padding-bottom: calc(68px + env(safe-area-inset-bottom) + 1rem);
}

/* ── Header ──────────────────────────────────────────────── */
.profile-header {
  display: flex; align-items: center; justify-content: space-between;
}
.profile-header-left { display: flex; align-items: center; gap: 0.75rem; }
.profile-header-avatar {
  width: 40px; height: 40px; border-radius: 50%; display: grid; place-items: center;
  background: linear-gradient(135deg, #e63946, #d62839); color: #fff;
  font-size: 0.82rem; font-weight: 800; flex-shrink: 0;
  box-shadow: 0 2px 8px rgba(230, 57, 70, 0.25);
}
.profile-header-text h1 { margin: 0; font-size: 1.15rem; font-weight: 800; letter-spacing: -0.02em; }
.profile-header-text p { margin: 0; font-size: 0.72rem; color: var(--text-muted); }
.profile-header-action {
  width: 40px; height: 40px; border-radius: 50%; border: 1px solid var(--border-light);
  background: #fff; color: var(--text-muted); display: grid; place-items: center;
  cursor: pointer; font-size: 0.85rem; transition: all 0.15s ease;
}
.profile-header-action:hover { color: #dc2626; border-color: #fecaca; background: #fef2f2; }

/* ── Profile Identity Card ───────────────────────────────── */
.profile-card {
  background: linear-gradient(135deg, #1e3a5f 0%, #0f172a 100%);
  border-radius: 20px; padding: 1.5rem 1.25rem;
  display: flex; flex-direction: column; align-items: center;
  color: #fff; box-shadow: 0 8px 30px rgba(15, 23, 42, 0.2);
}
.profile-card-avatar {
  width: 64px; height: 64px; border-radius: 50%;
  background: rgba(255, 255, 255, 0.15); border: 2px solid rgba(255, 255, 255, 0.2);
  display: grid; place-items: center; font-size: 1.35rem; font-weight: 800;
  margin-bottom: 0.75rem;
}
.profile-card-name { margin: 0; font-size: 1.15rem; font-weight: 800; letter-spacing: -0.01em; }
.profile-card-role { margin: 0.2rem 0 0; font-size: 0.78rem; opacity: 0.8; }

.profile-card-strip {
  width: 100%; margin-top: 1.25rem; padding-top: 1rem;
  border-top: 1px solid rgba(255, 255, 255, 0.12);
  display: grid; grid-template-columns: repeat(3, 1fr); gap: 0.5rem;
}
.account-security { display:flex; align-items:center; justify-content:space-between; gap:1rem; margin:0 1.1rem 1.1rem; padding:1rem; border:1px solid #f0e6e8; border-radius:12px; background:#fff8f9; }
.security-copy { display:grid; gap:.18rem; min-width:0; }
.security-kicker { color:#b42332; font-size:.62rem; font-weight:800; letter-spacing:.08em; text-transform:uppercase; }
.security-copy strong { font-size:.85rem; }
.security-mask { color:var(--text-muted); letter-spacing:.18em; font-size:.75rem; }
@media (max-width: 767px) { .account-security { align-items:stretch; flex-direction:column; margin:0 1rem 1rem; } .account-security :deep(button) { width:100%; } }
.strip-item { text-align: center; min-width: 0; }
.strip-label {
  display: block; font-size: 0.52rem; font-weight: 700; text-transform: uppercase;
  letter-spacing: 0.06em; opacity: 0.65; margin-bottom: 0.2rem;
}
.strip-value {
  display: block; font-size: 0.7rem; font-weight: 600; line-height: 1.3;
  overflow: hidden; text-overflow: ellipsis; white-space: nowrap;
}

/* ── Performance Section ─────────────────────────────────── */
.perf-section {
  background: var(--surface-card); border: 1px solid var(--border-light);
  border-radius: var(--radius-xl); padding: 1.15rem;
  box-shadow: var(--shadow-sm);
}
.perf-title {
  margin: 0 0 0.85rem; font-size: 0.68rem; font-weight: 700;
  text-transform: uppercase; letter-spacing: 0.06em; color: var(--text-muted);
}
.perf-stats { display: grid; grid-template-columns: repeat(3, 1fr); gap: 0.6rem; }

.stat-card {
  border-radius: 12px; padding: 0.85rem 0.5rem;
  display: flex; flex-direction: column; align-items: center; gap: 0.3rem;
}
.stat-blue { background: #fff0f1; }
.stat-green { background: #f0fdf4; }
.stat-purple { background: #f5f3ff; }
.stat-value { font-size: 1.5rem; font-weight: 800; line-height: 1; }
.stat-blue .stat-value { color: #e63946; }
.stat-green .stat-value { color: #16a34a; }
.stat-purple .stat-value { color: #c54b59; }
.stat-label {
  font-size: 0.55rem; font-weight: 700; text-transform: uppercase;
  letter-spacing: 0.03em; text-align: center; color: var(--text-muted);
}

/* Refined account card */
.profile-card {
  position: relative;
  overflow: hidden;
  align-items: stretch;
  padding: 0;
  color: var(--text-primary);
  background: #fff;
  border: 1px solid #eee3e5;
  box-shadow: 0 10px 26px rgba(73, 34, 41, 0.08);
}
.profile-card-kicker { font-size: 0.62rem; font-weight: 800; letter-spacing: 0.09em; text-transform: uppercase; }
.profile-status { display: inline-flex; align-items: center; gap: 0.3rem; padding: 0.3rem 0.5rem; border: 1px solid rgba(255,255,255,.28); border-radius: 999px; background: rgba(255,255,255,.13); font-size: 0.62rem; font-weight: 750; }
.profile-status i { font-size: 0.38rem; color: #bbf7d0; }
.profile-card-identity { display: flex; align-items: flex-end; gap: 0.8rem; padding: 0 1rem; margin-top: -30px; position: relative; z-index: 1; }
.profile-card-avatar { width: 68px; height: 68px; margin: 0; border-radius: 20px; border: 4px solid #fff; color: #fff; background: linear-gradient(135deg, #e63946, #d62839); box-shadow: 0 6px 16px rgba(135,28,40,.2); }
.profile-card-copy { min-width: 0; padding-bottom: 0.3rem; }
.profile-card-name { color: var(--text-primary); }
.profile-card-role { color: var(--text-muted); opacity: 1; }
.profile-card-strip { width: auto; margin: 1rem; padding: 0; border: 1px solid #f1e8ea; border-radius: 14px; display: grid; grid-template-columns: 1fr; gap: 0; background: #fcf9f9; }
.strip-item { min-width: 0; display: flex; align-items: center; gap: 0.7rem; padding: 0.72rem; text-align: left; border-bottom: 1px solid #f1e8ea; }
.strip-item:last-child { border-bottom: 0; }
.strip-icon { width: 34px; height: 34px; display: grid; place-items: center; flex: 0 0 auto; border-radius: 10px; color: #e63946; background: #fff0f1; }
.strip-icon i { font-size: 0.8rem; }
.strip-copy { min-width: 0; display: grid; gap: 0.08rem; }
.strip-label { margin: 0; color: var(--text-muted); opacity: 1; }
.strip-value { color: var(--text-primary); font-size: 0.73rem; }
.perf-heading { display: flex; align-items: center; justify-content: space-between; margin-bottom: 0.85rem; }
.perf-heading > div { display: grid; gap: 0.08rem; }
.perf-heading span { color: #e63946; font-size: 0.55rem; font-weight: 800; letter-spacing: 0.08em; text-transform: uppercase; }
.perf-heading > i { width: 34px; height: 34px; display: grid; place-items: center; border-radius: 10px; color: #e63946; background: #fff0f1; }
.perf-title { margin: 0; color: var(--text-primary); font-size: 0.92rem; text-transform: none; letter-spacing: -0.01em; }
.stat-card { position: relative; align-items: flex-start; padding: 0.75rem; gap: 0.2rem; }
.stat-icon { width: 28px; height: 28px; display: grid; place-items: center; margin-bottom: 0.35rem; border-radius: 8px; background: rgba(255,255,255,.7); }
.stat-icon i { font-size: 0.72rem; }
.stat-label { text-align: left; }

/* ── Retry ───────────────────────────────────────────────── */
.retry-btn { margin-left: 0.5rem; }

/* ── Skeleton ────────────────────────────────────────────── */
.skeleton-card {
  background: linear-gradient(135deg, #1e3a5f 0%, #0f172a 100%);
  border-radius: 20px; padding: 1.5rem 1.25rem;
  display: flex; flex-direction: column; align-items: center;
}
.sk-circle {
  width: 64px; height: 64px; border-radius: 50%; background: rgba(255,255,255,0.1);
  margin-bottom: 0.75rem; animation: sk-pulse 1.5s ease-in-out infinite;
}
.sk-line {
  height: 12px; border-radius: 6px; background: rgba(255,255,255,0.1);
  margin-bottom: 0.5rem; animation: sk-pulse 1.5s ease-in-out infinite;
}
.sk-line.w30 { width: 30%; }
.sk-line.w50 { width: 50%; }
.sk-strip {
  width: 100%; margin-top: 1rem; padding-top: 1rem;
  border-top: 1px solid rgba(255,255,255,0.08);
  display: grid; grid-template-columns: repeat(3, 1fr); gap: 0.5rem;
}
.sk-strip-item { display: flex; flex-direction: column; align-items: center; gap: 0.3rem; }
.sk-label { width: 50%; height: 8px; border-radius: 4px; background: rgba(255,255,255,0.08); animation: sk-pulse 1.5s ease-in-out infinite; }
.sk-value { width: 70%; height: 10px; border-radius: 4px; background: rgba(255,255,255,0.06); animation: sk-pulse 1.5s ease-in-out infinite; }

.skeleton-perf {
  background: var(--surface-card); border: 1px solid var(--border-light);
  border-radius: var(--radius-xl); padding: 1.15rem;
}
.sk-section-title { width: 35%; height: 10px; border-radius: 5px; background: #e2e8f0; margin-bottom: 0.85rem; animation: sk-pulse 1.5s ease-in-out infinite; }
.sk-stats { display: grid; grid-template-columns: repeat(3, 1fr); gap: 0.6rem; }
.sk-stat { height: 80px; border-radius: 12px; background: #f1f5f9; animation: sk-pulse 1.5s ease-in-out infinite; }

@keyframes sk-pulse { 0%, 100% { opacity: 1; } 50% { opacity: 0.5; } }

/* ── Responsive ──────────────────────────────────────────── */
@media (max-width: 767px) {
  .profile-page { gap: 0.8rem; }
  .profile-header { padding: 0.8rem; border: 1px solid #eee3e5; border-radius: 16px; background: #fff; box-shadow: 0 5px 16px rgba(73,34,41,.06); }
  .profile-card { padding: 0; border-radius: 20px; box-shadow: 0 12px 26px rgba(73,34,41,.08); }
  .profile-card-avatar { width: 68px; height: 68px; border-radius: 20px; }
  .profile-card-strip { gap: 0; margin: 1rem; padding: 0; border-radius: 14px; background: #fcf9f9; }
  .perf-section { padding: 1rem; border: 1px solid #eee3e5; border-radius: 18px; background: #fff; }
  .stats-grid { gap: 0.5rem; }
  .stat-card { padding: 0.8rem 0.35rem; border-radius: 13px; }
  .strip-value { font-size: 0.65rem; }
  .stat-value { font-size: 1.3rem; }
}

/* ── Desktop ─────────────────────────────────────────────── */
.profile-back,
.profile-header,
.profile-error { grid-column: 1 / -1; }

@media (min-width: 768px) {
  .profile-page {
    grid-template-columns: minmax(0, 1fr) minmax(0, 1.35fr);
    gap: 1.25rem;
    align-items: start;
    padding-bottom: 0;
  }
  .profile-back { display: none; }
  .profile-header { padding: 0.25rem 0; }
  .profile-header-avatar { width: 44px; height: 44px; }
  .profile-header-text h1 { font-size: 1.4rem; }

  .profile-card { padding: 0; }
  .profile-card-avatar { width: 76px; height: 76px; font-size: 1.5rem; }
  .profile-card-name { font-size: 1.3rem; }

  .perf-section { padding: 1.5rem; }
  .stat-card { padding: 1.1rem 0.75rem; }
  .stat-value { font-size: 1.7rem; }
}

/* Profile dashboard composition */
.profile-page {
  min-width: 0;
  box-sizing: border-box;
}
.profile-page > * { min-width: 0; max-width: 100%; box-sizing: border-box; }
.profile-header { gap: 1rem; padding-block: .2rem; }
.profile-header-left, .profile-header-text { min-width: 0; }
.profile-header-text h1 { color: var(--text-primary); font-size: clamp(1.45rem, 2vw, 1.75rem); line-height: 1.15; }
.profile-header-text p { margin-top: .18rem; font-size: .78rem; }
.profile-header-action { flex: 0 0 40px; }

.profile-card {
  display: flex;
  min-width: 0;
  overflow: hidden;
  border-radius: 18px;
  box-shadow: 0 8px 24px rgba(73, 34, 41, .07);
}
.profile-card-top {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: .75rem;
  padding: 1rem 1.1rem .75rem;
  border-bottom: 1px solid #f2e7e9;
  background: linear-gradient(90deg, #fff7f8, #fff);
}
.profile-card-kicker { color: #b42332; font-size: .72rem; font-weight: 800; letter-spacing: .08em; text-transform: uppercase; }
.profile-status { flex: 0 0 auto; border: 1px solid #bbf7d0; background: #f0fdf4; color: #166534; font-size: .7rem; }
.profile-status i { color: #22c55e; }
.profile-card-identity {
  align-items: center;
  gap: .9rem;
  margin: 0;
  padding: 1.15rem 1.1rem;
}
.profile-card-avatar {
  width: 60px;
  height: 60px;
  flex: 0 0 60px;
  border: 0;
  border-radius: 16px;
  font-size: 1.2rem;
}
.profile-card-copy { min-width: 0; padding: 0; }
.profile-card-name { overflow-wrap: anywhere; font-size: clamp(1.2rem, 1.8vw, 1.45rem); line-height: 1.2; }
.profile-card-role { margin-top: .3rem; font-size: .84rem; line-height: 1.4; }
.profile-card-strip {
  min-width: 0;
  margin: 0 1.1rem 1.1rem;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  border: 0;
  gap: .55rem;
  background: transparent;
}
.strip-item {
  min-width: 0;
  padding: .75rem;
  border: 1px solid #f0e6e8;
  border-radius: 12px;
  background: #fcfafb;
}
.strip-item:last-child { grid-column: 1 / -1; border-bottom: 1px solid #f0e6e8; }
.strip-copy { overflow: hidden; }
.strip-label { font-size: .68rem; letter-spacing: .04em; }
.strip-value { overflow: visible; text-overflow: clip; white-space: normal; overflow-wrap: anywhere; font-size: .84rem; line-height: 1.4; }

.perf-section {
  min-width: 0;
  display: flex;
  flex-direction: column;
  padding: 1.35rem;
  border-color: #eee3e5;
  border-radius: 18px;
  box-shadow: 0 8px 24px rgba(73, 34, 41, .06);
}
.perf-heading { margin-bottom: 1rem; }
.perf-heading span { font-size: .68rem; }
.perf-title { margin-top: .12rem; font-size: 1.08rem; }
.perf-stats { flex: 1; grid-template-columns: repeat(3, minmax(0, 1fr)); gap: .8rem; }
.stat-card {
  min-width: 0;
  min-height: 128px;
  justify-content: center;
  padding: 1rem;
  border: 1px solid transparent;
}
.stat-blue { border-color: #ffe0e3; }
.stat-green { border-color: #d7f5df; }
.stat-purple { border-color: #ebe5fa; }
.stat-icon { width: 34px; height: 34px; margin-bottom: .45rem; }
.stat-icon i { font-size: .85rem; }
.stat-value { font-size: clamp(1.65rem, 2.5vw, 2rem); }
.stat-label { margin-top: .15rem; font-size: .72rem; line-height: 1.35; }
.skeleton-card { align-items: stretch; background: #fff; border: 1px solid #eee3e5; }
.skeleton-card .sk-circle, .skeleton-card .sk-line, .skeleton-card .sk-label, .skeleton-card .sk-value { background: #f1e8ea; }

@media (min-width: 1100px) {
  .profile-page { grid-template-columns: minmax(320px, .9fr) minmax(0, 1.6fr); gap: 1.25rem; align-items: stretch; }
  .profile-card, .perf-section { height: 100%; }
}

@media (min-width: 768px) and (max-width: 1099px) {
  .profile-page { grid-template-columns: 1fr; }
  .profile-card, .perf-section { grid-column: 1; }
  .profile-card-strip { grid-template-columns: repeat(3, minmax(0, 1fr)); }
  .strip-item:last-child { grid-column: auto; }
}

@media (max-width: 767px) {
  .profile-header { padding: .8rem; }
  .profile-card-top { padding: .9rem 1rem .7rem; }
  .profile-card-identity { padding: 1rem; }
  .profile-card-strip { margin: 0 1rem 1rem; grid-template-columns: 1fr; }
  .strip-item:last-child { grid-column: auto; }
  .perf-stats { grid-template-columns: repeat(2, minmax(0, 1fr)); }
  .stat-card { min-height: 108px; }
  .stat-card:last-child { grid-column: 1 / -1; }
}

@media (max-width: 390px) {
  .profile-page { gap: .7rem; }
  .profile-header-avatar { width: 38px; height: 38px; }
  .profile-card-avatar { width: 54px; height: 54px; flex-basis: 54px; }
  .perf-section { padding: 1rem; }
  .perf-stats { gap: .55rem; }
  .stat-card { padding: .8rem; }
}
</style>
