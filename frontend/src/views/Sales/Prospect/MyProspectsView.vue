<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import Button from 'primevue/button'
import Dialog from 'primevue/dialog'
import Message from 'primevue/message'
import Tag from 'primevue/tag'
import { useAuthStore } from '../../../stores/auth'
import { useCrmStore } from '../../../stores/crm'
import type { Prospect, ProspectStatus } from '../../../types/crm'

const auth = useAuthStore()
const crm = useCrmStore()
const error = ref('')
const searchQuery = ref('')
const activeTab = ref<'assigned' | 'visited' | 'followup' | 'won'>('assigned')
const sortBy = ref<'distance' | 'name-asc' | 'name-desc' | 'recently-assigned' | 'last-updated'>('distance')
const userCoords = ref<{ lat: number; lng: number } | null>(null)
const gpsDenied = ref(false)
const showSortMenu = ref(false)
const showFilterPanel = ref(false)
const filterIndustry = ref('')
const filterCategory = ref('')
const filterHasCoords = ref(false)

type SortOption = { label: string; value: typeof sortBy.value }
const sortOptions: SortOption[] = [
  { label: 'Nearest first', value: 'distance' },
  { label: 'Name A\u2013Z', value: 'name-asc' },
  { label: 'Name Z\u2013A', value: 'name-desc' },
  { label: 'Recently assigned', value: 'recently-assigned' },
  { label: 'Last updated', value: 'last-updated' },
]

const TAB_STATUSES: Record<typeof activeTab.value, ProspectStatus[]> = {
  assigned: ['NEW_LEAD', 'INTERESTED', 'PROPOSAL_SENT'],
  visited: ['CONTACTED'],
  followup: ['QUALIFIED', 'NEGOTIATION'],
  won: ['WON'],
}

function haversineKm(lat1: number, lng1: number, lat2: number, lng2: number): number {
  const R = 6371
  const dLat = ((lat2 - lat1) * Math.PI) / 180
  const dLng = ((lng2 - lng1) * Math.PI) / 180
  const a = Math.sin(dLat / 2) ** 2 + Math.cos((lat1 * Math.PI) / 180) * Math.cos((lat2 * Math.PI) / 180) * Math.sin(dLng / 2) ** 2
  return R * 2 * Math.atan2(Math.sqrt(a), Math.sqrt(1 - a))
}

function formatDistance(km: number): string {
  if (km < 1) return `${Math.round(km * 1000)} m`
  return `${km.toFixed(1)} km`
}

function getDistance(p: Prospect): number | null {
  if (p.latitude == null || p.longitude == null || !userCoords.value) return null
  return haversineKm(userCoords.value.lat, userCoords.value.lng, p.latitude, p.longitude)
}

function initials(name: string): string {
  return name.split(/\s+/).slice(0, 2).map((w) => w.charAt(0).toUpperCase()).join('')
}

function statusSeverity(status: ProspectStatus): 'info' | 'warn' | 'success' | 'danger' | 'secondary' {
  switch (status) {
    case 'NEW_LEAD': return 'info'
    case 'CONTACTED': return 'info'
    case 'INTERESTED': return 'info'
    case 'QUALIFIED': return 'warn'
    case 'PROPOSAL_SENT': return 'info'
    case 'NEGOTIATION': return 'warn'
    case 'WON': return 'success'
    case 'LOST': return 'danger'
    default: return 'secondary'
  }
}

function statusLabel(status: ProspectStatus): string {
  return status.replaceAll('_', ' ')
}

function tabForStatus(status: ProspectStatus): typeof activeTab.value {
  if (TAB_STATUSES.assigned.includes(status)) return 'assigned'
  if (TAB_STATUSES.visited.includes(status)) return 'visited'
  if (TAB_STATUSES.followup.includes(status)) return 'followup'
  if (TAB_STATUSES.won.includes(status)) return 'won'
  return 'assigned'
}

function hasCoordinates(p: Prospect): boolean {
  return p.latitude != null && p.longitude != null
}

function openGoogleMaps(p: Prospect) {
  if (p.latitude != null && p.longitude != null) {
    window.open(`https://www.google.com/maps/dir/?api=1&destination=${p.latitude},${p.longitude}`, '_blank', 'noopener,noreferrer')
  } else if (p.formattedAddress) {
    window.open(`https://www.google.com/maps/dir/?api=1&destination=${encodeURIComponent(p.formattedAddress)}`, '_blank', 'noopener,noreferrer')
  }
}

function acquireGPS() {
  if (!navigator.geolocation) return
  navigator.geolocation.getCurrentPosition(
    (pos) => { userCoords.value = { lat: pos.coords.latitude, lng: pos.coords.longitude }; gpsDenied.value = false },
    () => { gpsDenied.value = true },
    { enableHighAccuracy: true, timeout: 10000 },
  )
}

function applyFilters(list: Prospect[]): Prospect[] {
  let result = list
  if (filterIndustry.value) result = result.filter((p) => p.industryGroup === filterIndustry.value)
  if (filterCategory.value) result = result.filter((p) => p.placeCategory === filterCategory.value)
  if (filterHasCoords.value) result = result.filter((p) => hasCoordinates(p))
  return result
}

function activeFilterCount(): number {
  let n = 0
  if (filterIndustry.value) n++
  if (filterCategory.value) n++
  if (filterHasCoords.value) n++
  return n
}

function resetFilters() {
  filterIndustry.value = ''
  filterCategory.value = ''
  filterHasCoords.value = false
}

const nonTerminal = computed(() =>
  crm.myProspects.filter((p) => !['LOST', 'CONVERTED'].includes(p.status)),
)

const tabCounts = computed(() => ({
  assigned: nonTerminal.value.filter((p) => TAB_STATUSES.assigned.includes(p.status)).length,
  visited: nonTerminal.value.filter((p) => TAB_STATUSES.visited.includes(p.status)).length,
  followup: nonTerminal.value.filter((p) => TAB_STATUSES.followup.includes(p.status)).length,
  won: crm.myProspects.filter((p) => p.status === 'WON').length,
}))

const availableIndustries = computed(() =>
  [...new Set(nonTerminal.value.map((p) => p.industryGroup).filter(Boolean))].sort(),
)

const availableCategories = computed(() =>
  [...new Set(nonTerminal.value.map((p) => p.placeCategory).filter(Boolean))].sort(),
)

const activeProspects = computed(() => {
  const statuses = TAB_STATUSES[activeTab.value]
  const pool = activeTab.value === 'won' ? crm.myProspects : nonTerminal.value
  return pool.filter((p) => statuses.includes(p.status))
})

const filteredProspects = computed(() => {
  let list = [...activeProspects.value]
  const q = searchQuery.value.trim().toLowerCase()
  if (q) {
    list = list.filter((p) => {
      const hay = `${p.placeName} ${p.formattedAddress} ${p.industryGroup} ${p.placeCategory} ${p.phoneNumber} ${p.assignedSalesExecutive}`.toLowerCase()
      return hay.includes(q)
    })
  }
  list = applyFilters(list)
  return list
})

const displayedProspects = computed(() => {
  let list = [...filteredProspects.value]
  if (sortBy.value === 'name-asc') list.sort((a, b) => a.placeName.localeCompare(b.placeName))
  else if (sortBy.value === 'name-desc') list.sort((a, b) => b.placeName.localeCompare(a.placeName))
  else if (sortBy.value === 'recently-assigned') list.sort((a, b) => new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime())
  else if (sortBy.value === 'last-updated') list.sort((a, b) => new Date(b.updatedAt).getTime() - new Date(a.updatedAt).getTime())
  else if (sortBy.value === 'distance') {
    list.sort((a, b) => {
      const da = getDistance(a)
      const db = getDistance(b)
      if (da === null && db === null) return 0
      if (da === null) return 1
      if (db === null) return -1
      return da - db
    })
  }
  return list
})

onMounted(async () => {
  acquireGPS()
  try { await crm.loadMyProspects() } catch (e: unknown) { error.value = crm.errorMessage(e) }
})
</script>

<template>
  <section class="mp-page">
    <!-- 1. Header -->
    <div class="mp-header">
      <div class="mp-header-left">
        <span class="mp-avatar">{{ auth.user?.fullName?.slice(0, 1) }}</span>
        <div class="mp-header-text">
          <strong>My prospects</strong>
          <small>{{ nonTerminal.length }} prospects &middot; {{ tabCounts[activeTab] }} shown</small>
        </div>
      </div>
      <button class="mp-header-action" @click="showFilterPanel = true" aria-label="Open filters">
        <i class="pi pi-sliders-h" />
        <span v-if="activeFilterCount()" class="mp-notif-dot">{{ activeFilterCount() }}</span>
      </button>
    </div>

    <!-- 2. Search -->
    <div class="mp-search">
      <i class="pi pi-search" />
      <input v-model="searchQuery" placeholder="Search business, category or area" aria-label="Search prospects" />
      <button v-if="searchQuery" class="mp-search-clear" aria-label="Clear search" @click="searchQuery = ''">
        <i class="pi pi-times" />
      </button>
    </div>

    <!-- 3. Tabs -->
    <div class="mp-tabs">
      <button class="mp-tab" :class="{ active: activeTab === 'assigned' }" @click="activeTab = 'assigned'">
        Assigned <span class="mp-tab-badge">{{ tabCounts.assigned }}</span>
      </button>
      <button class="mp-tab" :class="{ active: activeTab === 'visited' }" @click="activeTab = 'visited'">
        Visited <span class="mp-tab-badge">{{ tabCounts.visited }}</span>
      </button>
      <button class="mp-tab" :class="{ active: activeTab === 'followup' }" @click="activeTab = 'followup'">
        Follow Up <span class="mp-tab-badge">{{ tabCounts.followup }}</span>
      </button>
      <button class="mp-tab" :class="{ active: activeTab === 'won' }" @click="activeTab = 'won'">
        Won <span class="mp-tab-badge">{{ tabCounts.won }}</span>
      </button>
    </div>

    <!-- Messages -->
    <Message v-if="error" severity="error" closable @close="error = ''">{{ error }}</Message>
    <Message v-if="gpsDenied" severity="warn" :closable="false">
      Location access was denied. Distance sorting requires GPS permission.
    </Message>

    <!-- Loading skeleton -->
    <div v-if="crm.loading && !nonTerminal.length" class="mp-skeleton-list">
      <div v-for="n in 4" :key="n" class="mp-skeleton-card">
        <div class="sk-row"><div class="sk-avatar" /><div class="sk-lines"><div class="sk-line w60" /><div class="sk-line w40" /></div></div>
        <div class="sk-line w80" /><div class="sk-line w50" />
        <div class="sk-row sk-actions"><div class="sk-btn" /><div class="sk-btn" /><div class="sk-btn" /></div>
      </div>
    </div>

    <!-- Empty state -->
    <div v-else-if="!displayedProspects.length && !error" class="mp-empty">
      <div class="mp-empty-icon"><i class="pi pi-briefcase" /></div>
      <strong>No prospects found</strong>
      <span v-if="searchQuery || activeFilterCount()">Try changing your search or filters.</span>
      <span v-else-if="activeTab === 'won'">No won prospects yet.</span>
      <span v-else>No prospects in this category.</span>
      <button v-if="searchQuery || activeFilterCount()" class="mp-reset-btn" @click="searchQuery = ''; resetFilters()">
        <i class="pi pi-refresh" /> Reset filters
      </button>
    </div>

    <!-- Sort + Cards -->
    <template v-else>
      <div class="mp-section-header">
        <strong>Prospects</strong>
        <button class="mp-sort-trigger" @click="showSortMenu = !showSortMenu">
          <i class="pi pi-sort-alt" /> {{ sortOptions.find((s) => s.value === sortBy)?.label }}
          <i class="pi pi-chevron-down" />
        </button>
      </div>

      <div v-if="showSortMenu" class="mp-sort-menu">
        <button v-for="opt in sortOptions" :key="opt.value" :class="{ active: sortBy === opt.value }" @click="sortBy = opt.value; showSortMenu = false">
          {{ opt.label }}
        </button>
      </div>

      <div class="mp-card-list">
        <article v-for="prospect in displayedProspects" :key="prospect.id" class="mp-card">
          <div class="mp-card-top">
            <span class="mp-card-avatar">{{ initials(prospect.placeName || 'Unnamed') }}</span>
            <div class="mp-card-identity">
              <strong>{{ prospect.placeName || 'Unnamed prospect' }}</strong>
              <small>{{ prospect.industryGroup || prospect.placeCategory || 'Uncategorized' }}</small>
            </div>
            <span v-if="getDistance(prospect) !== null" class="mp-distance-pill">
              <i class="pi pi-map-marker" /> {{ formatDistance(getDistance(prospect)!) }}
            </span>
          </div>

          <div class="mp-card-middle">
            <div class="mp-card-address">
              <i class="pi pi-map-marker" />
              <span>{{ prospect.formattedAddress || 'No address' }}</span>
            </div>
            <div class="mp-card-tags">
              <Tag :value="statusLabel(prospect.status)" :severity="statusSeverity(prospect.status)" />
            </div>
          </div>

          <div class="mp-card-actions">
            <button class="mp-action-btn mp-action-navigate" :disabled="!hasCoordinates(prospect) && !prospect.formattedAddress" @click="openGoogleMaps(prospect)" :title="hasCoordinates(prospect) || prospect.formattedAddress ? 'Navigate with Google Maps' : 'No location data available'">
              <i class="pi pi-directions" /> Navigate
            </button>
            <RouterLink class="mp-action-btn mp-action-detail" :to="`/sales/my-prospects/${prospect.id}`">
              <i class="pi pi-eye" /> Detail
            </RouterLink>
            <RouterLink class="mp-action-btn mp-action-visit" :to="`/sales/my-prospects/${prospect.id}`">
              <i class="pi pi-sign-in" /> Visit
            </RouterLink>
          </div>
        </article>
      </div>
    </template>

    <!-- FAB filter -->
    <button class="mp-fab" @click="showFilterPanel = true" aria-label="Open filter panel">
      <i class="pi pi-filter" />
      <span v-if="activeFilterCount()" class="mp-fab-badge">{{ activeFilterCount() }}</span>
    </button>

    <!-- Filter dialog -->
    <Dialog v-model:visible="showFilterPanel" modal header="Filter prospects" :style="{ width: 'min(92vw, 400px)' }" :dt="{ background: 'var(--surface-card)', borderRadius: 'var(--radius-xl)' }">
      <div class="mp-filter-body">
        <label class="mp-filter-field">
          <span>Industry</span>
          <select v-model="filterIndustry">
            <option value="">All industries</option>
            <option v-for="s in availableIndustries" :key="s" :value="s">{{ s }}</option>
          </select>
        </label>
        <label class="mp-filter-field">
          <span>Category</span>
          <select v-model="filterCategory">
            <option value="">All categories</option>
            <option v-for="c in availableCategories" :key="c" :value="c">{{ c }}</option>
          </select>
        </label>
        <label class="mp-filter-check">
          <input v-model="filterHasCoords" type="checkbox" />
          <span>Only with GPS coordinates</span>
        </label>
      </div>
      <template #footer>
        <Button label="Reset" severity="secondary" text @click="resetFilters" />
        <Button label="Apply filters" @click="showFilterPanel = false" />
      </template>
    </Dialog>
  </section>
</template>

<style scoped>
/* ── Page ───────────────────────────────────────────────────── */
.mp-page { display: flex; flex-direction: column; gap: 0.85rem; padding-bottom: 1.5rem; }

/* ── 1. Header ─────────────────────────────────────────────── */
.mp-header { display: flex; align-items: center; justify-content: space-between; padding: 0.15rem 0; }
.mp-header-left { display: flex; align-items: center; gap: 0.7rem; }
.mp-avatar {
  width: 38px; height: 38px; display: grid; place-items: center;
  border-radius: 50%; background: linear-gradient(135deg, #2563eb, #1d4ed8);
  color: #fff; font-weight: 800; font-size: 0.85rem; flex-shrink: 0;
  box-shadow: 0 2px 8px rgba(37, 99, 235, 0.25);
}
.mp-header-text { display: flex; flex-direction: column; gap: 0.05rem; }
.mp-header-text strong { font-size: 1rem; font-weight: 800; color: #0f172a; }
.mp-header-text small { color: #64748b; font-size: 0.7rem; font-weight: 500; }
.mp-header-action {
  position: relative; width: 36px; height: 36px; display: grid; place-items: center;
  border-radius: 50%; border: 1px solid #e2e8f0; background: #fff; color: #64748b;
  cursor: pointer; font-size: 0.9rem; transition: all 0.15s ease;
}
.mp-header-action:hover { color: #2563eb; border-color: #cbd5e1; background: #eff6ff; }
.mp-notif-dot {
  position: absolute; top: -3px; right: -3px; min-width: 16px; height: 16px;
  display: grid; place-items: center; border-radius: 9999px; background: #dc2626;
  color: #fff; font-size: 0.55rem; font-weight: 700; padding: 0 4px; border: 2px solid #fff;
}

/* ── 2. Search ─────────────────────────────────────────────── */
.mp-search {
  display: flex; align-items: center; gap: 0.5rem;
  padding: 0.55rem 0.8rem; background: #fff; border: 1px solid #e2e8f0;
  border-radius: 14px; transition: border-color 0.15s ease, box-shadow 0.15s ease;
}
.mp-search:focus-within { border-color: #2563eb; box-shadow: 0 0 0 3px rgba(37, 99, 235, 0.08); }
.mp-search i { color: #94a3b8; font-size: 0.85rem; flex-shrink: 0; }
.mp-search input {
  flex: 1; border: 0; outline: 0; background: transparent; color: #0f172a;
  font-size: 0.8rem; font-weight: 500;
}
.mp-search input::placeholder { color: #94a3b8; }
.mp-search-clear {
  width: 22px; height: 22px; display: grid; place-items: center;
  border-radius: 50%; border: 0; background: #f1f5f9; color: #64748b;
  cursor: pointer; font-size: 0.65rem; transition: background 0.15s ease;
}
.mp-search-clear:hover { background: #e2e8f0; }

/* ── 3. Tabs ───────────────────────────────────────────────── */
.mp-tabs { display: grid; grid-template-columns: repeat(4, 1fr); gap: 0.35rem; }
.mp-tab {
  display: flex; align-items: center; justify-content: center; gap: 0.3rem;
  padding: 0.5rem 0; border-radius: 10px; border: 1px solid #e2e8f0;
  background: #fff; color: #64748b; font-size: 0.68rem; font-weight: 600;
  cursor: pointer; transition: all 0.15s ease;
}
.mp-tab.active { background: #eff6ff; color: #2563eb; border-color: #bfdbfe; font-weight: 700; }
.mp-tab:not(.active):hover { background: #f8fafc; border-color: #cbd5e1; }
.mp-tab-badge {
  min-width: 16px; height: 16px; display: inline-grid; place-items: center;
  border-radius: 9999px; background: #f1f5f9; font-size: 0.55rem; font-weight: 700; padding: 0 3px;
}
.mp-tab.active .mp-tab-badge { background: #dbeafe; color: #2563eb; }

/* ── 4. Section header + sort ──────────────────────────────── */
.mp-section-header { display: flex; align-items: center; justify-content: space-between; }
.mp-section-header strong { font-size: 0.82rem; font-weight: 800; color: #0f172a; }
.mp-sort-trigger {
  display: flex; align-items: center; gap: 0.3rem; padding: 0.3rem 0.6rem;
  border-radius: 10px; border: 1px solid #e2e8f0; background: #fff;
  color: #64748b; font-size: 0.68rem; font-weight: 600; cursor: pointer; transition: all 0.15s ease;
}
.mp-sort-trigger:hover { border-color: #cbd5e1; background: #f8fafc; }
.mp-sort-trigger i { font-size: 0.6rem; }
.mp-sort-menu { display: flex; gap: 0.35rem; flex-wrap: wrap; }
.mp-sort-menu button {
  padding: 0.4rem 0.7rem; border-radius: 10px; border: 1px solid #e2e8f0;
  background: #fff; color: #64748b; font-size: 0.68rem; font-weight: 600;
  cursor: pointer; transition: all 0.15s ease;
}
.mp-sort-menu button.active { background: #2563eb; color: #fff; border-color: #2563eb; }
.mp-sort-menu button:not(.active):hover { border-color: #cbd5e1; background: #f8fafc; }

/* ── 5. Card list ──────────────────────────────────────────── */
.mp-card-list { display: flex; flex-direction: column; gap: 0.65rem; }
.mp-card {
  padding: 0.9rem 1rem; background: #fff; border: 1px solid #eef1f6;
  border-radius: 16px; box-shadow: 0 1px 3px rgba(0, 0, 0, 0.03);
  display: flex; flex-direction: column; gap: 0.6rem;
  transition: border-color 0.15s ease, box-shadow 0.15s ease;
}
.mp-card:hover { border-color: #d6dce6; box-shadow: 0 4px 12px rgba(0, 0, 0, 0.05); }

/* Card top */
.mp-card-top { display: flex; align-items: center; gap: 0.65rem; }
.mp-card-avatar {
  width: 38px; height: 38px; display: grid; place-items: center;
  border-radius: 12px; background: #eff6ff; color: #2563eb;
  font-weight: 800; font-size: 0.72rem; flex-shrink: 0;
}
.mp-card-identity { flex: 1; min-width: 0; display: flex; flex-direction: column; gap: 0.05rem; }
.mp-card-identity strong { font-size: 0.82rem; font-weight: 700; color: #0f172a; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.mp-card-identity small { color: #64748b; font-size: 0.65rem; font-weight: 500; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.mp-distance-pill {
  display: flex; align-items: center; gap: 0.25rem; padding: 0.2rem 0.5rem;
  border-radius: 9999px; background: #eff6ff; color: #2563eb;
  font-size: 0.6rem; font-weight: 700; white-space: nowrap; flex-shrink: 0;
}
.mp-distance-pill i { font-size: 0.55rem; }

/* Card middle */
.mp-card-middle { display: flex; flex-direction: column; gap: 0.35rem; }
.mp-card-address { display: flex; align-items: flex-start; gap: 0.4rem; color: #64748b; font-size: 0.72rem; line-height: 1.4; }
.mp-card-address i { margin-top: 0.1rem; font-size: 0.68rem; flex-shrink: 0; color: #94a3b8; }
.mp-card-tags { display: flex; flex-wrap: wrap; gap: 0.3rem; }

/* Card actions */
.mp-card-actions {
  display: grid; grid-template-columns: 1fr 1fr 1fr; gap: 0.4rem;
  padding-top: 0.45rem; border-top: 1px solid #f1f5f9;
}
.mp-action-btn {
  display: flex; align-items: center; justify-content: center; gap: 0.3rem;
  padding: 0.45rem 0; border-radius: 10px; border: 1px solid #e2e8f0;
  background: #fff; font-size: 0.65rem; font-weight: 600; cursor: pointer;
  text-decoration: none; transition: all 0.15s ease;
}
.mp-action-btn:disabled { opacity: 0.4; cursor: not-allowed; }
.mp-action-navigate { color: #2563eb; }
.mp-action-navigate:not(:disabled):hover { background: #eff6ff; border-color: #bfdbfe; }
.mp-action-detail { color: #0f172a; }
.mp-action-detail:hover { background: #f8fafc; border-color: #cbd5e1; }
.mp-action-visit { color: #059669; }
.mp-action-visit:hover { background: #ecfdf5; border-color: #a7f3d0; }

/* ── Empty state ───────────────────────────────────────────── */
.mp-empty {
  display: flex; flex-direction: column; align-items: center; gap: 0.5rem;
  padding: 2.5rem 1rem; text-align: center;
}
.mp-empty-icon {
  width: 52px; height: 52px; display: grid; place-items: center;
  border-radius: 16px; background: #eff6ff; color: #2563eb; font-size: 1.4rem;
}
.mp-empty strong { color: #0f172a; font-size: 0.88rem; }
.mp-empty span { color: #64748b; font-size: 0.75rem; line-height: 1.5; }
.mp-reset-btn {
  display: flex; align-items: center; gap: 0.3rem; margin-top: 0.3rem;
  padding: 0.45rem 0.85rem; border-radius: 10px; border: 1px solid #e2e8f0;
  background: #fff; color: #2563eb; font-size: 0.72rem; font-weight: 600;
  cursor: pointer; transition: all 0.15s ease;
}
.mp-reset-btn:hover { background: #eff6ff; border-color: #bfdbfe; }

/* ── Skeleton ──────────────────────────────────────────────── */
.mp-skeleton-list { display: flex; flex-direction: column; gap: 0.65rem; }
.mp-skeleton-card {
  padding: 1rem; background: #fff; border: 1px solid #eef1f6; border-radius: 16px;
  display: flex; flex-direction: column; gap: 0.6rem;
  animation: mp-pulse 1.5s ease-in-out infinite;
}
.sk-row { display: flex; align-items: center; gap: 0.65rem; }
.sk-avatar { width: 38px; height: 38px; border-radius: 12px; background: #e2e8f0; flex-shrink: 0; }
.sk-lines { flex: 1; display: flex; flex-direction: column; gap: 0.3rem; }
.sk-line { height: 10px; border-radius: 6px; background: #e2e8f0; }
.sk-line.w60 { width: 60%; }
.sk-line.w40 { width: 40%; }
.sk-line.w80 { width: 80%; }
.sk-line.w50 { width: 50%; }
.sk-actions { justify-content: flex-end; }
.sk-btn { width: 70px; height: 28px; border-radius: 8px; background: #e2e8f0; }
@keyframes mp-pulse { 0%, 100% { opacity: 1; } 50% { opacity: 0.5; } }

/* ── FAB ───────────────────────────────────────────────────── */
.mp-fab {
  position: fixed; bottom: 88px; right: calc(50% - 190px);
  width: 48px; height: 48px; border-radius: 50%; border: 0;
  background: linear-gradient(135deg, #2563eb, #1d4ed8); color: #fff;
  font-size: 1.1rem; cursor: pointer; z-index: 50;
  box-shadow: 0 6px 20px rgba(37, 99, 235, 0.35);
  display: grid; place-items: center; transition: transform 0.2s ease, box-shadow 0.2s ease;
}
.mp-fab:hover { transform: scale(1.08); box-shadow: 0 8px 24px rgba(37, 99, 235, 0.45); }
.mp-fab-badge {
  position: absolute; top: -2px; right: -2px; min-width: 18px; height: 18px;
  display: grid; place-items: center; border-radius: 9999px; background: #dc2626;
  color: #fff; font-size: 0.55rem; font-weight: 700; padding: 0 4px; border: 2px solid #fff;
}

/* ── Filter dialog ─────────────────────────────────────────── */
.mp-filter-body { display: flex; flex-direction: column; gap: 0.85rem; }
.mp-filter-field { display: flex; flex-direction: column; gap: 0.3rem; }
.mp-filter-field span { color: #64748b; font-size: 0.72rem; font-weight: 700; text-transform: uppercase; letter-spacing: 0.04em; }
.mp-filter-field select {
  padding: 0.55rem 0.7rem; border: 1px solid #e2e8f0; border-radius: 10px;
  background: #f8fafc; color: #0f172a; font-size: 0.8rem; font-weight: 500;
  cursor: pointer; transition: border-color 0.15s ease;
}
.mp-filter-field select:focus { outline: 0; border-color: #2563eb; }
.mp-filter-check {
  display: flex; align-items: center; gap: 0.5rem; color: #0f172a;
  font-size: 0.8rem; font-weight: 500; cursor: pointer;
}
.mp-filter-check input[type="checkbox"] { width: 16px; height: 16px; accent-color: #2563eb; }

/* ── Responsive ────────────────────────────────────────────── */
@media (max-width: 480px) {
  .mp-fab { right: 1rem; }
}
</style>
