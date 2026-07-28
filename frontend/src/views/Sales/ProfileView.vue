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
  if (auth.user?.role === 'ADMINISTRATOR') return 'Administrator'
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
    <Button icon="pi pi-arrow-left" severity="secondary" text rounded @click="router.back()" title="Back" />
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
        <div class="profile-card-avatar" aria-hidden="true">{{ initials }}</div>
        <h2 class="profile-card-name">{{ auth.user?.fullName ?? '—' }}</h2>
        <p class="profile-card-role">{{ roleLabel }}</p>

        <div class="profile-card-strip">
          <div class="strip-item">
            <span class="strip-label">Employee ID</span>
            <span class="strip-value" :title="auth.user?.id ?? ''">{{ auth.user?.id ?? '—' }}</span>
          </div>
          <div class="strip-item">
            <span class="strip-label">Email</span>
            <span class="strip-value" :title="auth.user?.email ?? ''">{{ auth.user?.email ?? '—' }}</span>
          </div>
          <div class="strip-item">
            <span class="strip-label">Phone</span>
            <span class="strip-value">&mdash;</span>
          </div>
        </div>
      </div>

      <!-- Error in stats -->
      <Message v-if="error" severity="error" closable @close="error = ''">
        {{ error }}
        <Button label="Retry" size="small" text class="retry-btn" @click="loadStats" />
      </Message>

      <!-- Performance Section -->
      <div class="perf-section">
        <h2 class="perf-title">Performance</h2>
        <div class="perf-stats">
          <div class="stat-card stat-blue">
            <span class="stat-value">{{ totalVisits }}</span>
            <span class="stat-label">Total Visits</span>
          </div>
          <div class="stat-card stat-green">
            <span class="stat-value">{{ wonProspects }}</span>
            <span class="stat-label">Won Prospects</span>
          </div>
          <div class="stat-card stat-purple">
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
  background: linear-gradient(135deg, #2563eb, #1d4ed8); color: #fff;
  font-size: 0.82rem; font-weight: 800; flex-shrink: 0;
  box-shadow: 0 2px 8px rgba(37, 99, 235, 0.25);
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
.stat-blue { background: #eff6ff; }
.stat-green { background: #f0fdf4; }
.stat-purple { background: #f5f3ff; }
.stat-value { font-size: 1.5rem; font-weight: 800; line-height: 1; }
.stat-blue .stat-value { color: #2563eb; }
.stat-green .stat-value { color: #16a34a; }
.stat-purple .stat-value { color: #7c3aed; }
.stat-label {
  font-size: 0.55rem; font-weight: 700; text-transform: uppercase;
  letter-spacing: 0.03em; text-align: center; color: var(--text-muted);
}

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
  .profile-page { gap: 0.7rem; }
  .profile-card { padding: 1.25rem 1rem; }
  .profile-card-strip { gap: 0.35rem; }
  .strip-value { font-size: 0.65rem; }
  .stat-value { font-size: 1.3rem; }
}
</style>
