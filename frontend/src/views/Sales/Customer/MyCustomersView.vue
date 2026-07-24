<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import Button from 'primevue/button'
import Dialog from 'primevue/dialog'
import Message from 'primevue/message'
import { useAuthStore } from '../../../stores/auth'
import { useCrmStore } from '../../../stores/crm'
import type { CustomerSite } from '../../../types/crm'

const auth = useAuthStore()
const crm = useCrmStore()
const error = ref('')
const searchQuery = ref('')
const activeTab = ref<'all' | 'today' | 'nearby'>('all')
const sortBy = ref<'distance' | 'name-asc' | 'name-desc'>('distance')
const userCoords = ref<{ lat: number; lng: number } | null>(null)
const gpsDenied = ref(false)
const showSortMenu = ref(false)
const showFilterPanel = ref(false)
const filterSegment = ref('')
const filterCategory = ref('')

type SortOption = { label: string; value: typeof sortBy.value }
const sortOptions: SortOption[] = [
  { label: 'Nearest first', value: 'distance' },
  { label: 'Name A\u2013Z', value: 'name-asc' },
  { label: 'Name Z\u2013A', value: 'name-desc' },
]

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

function getDistance(c: CustomerSite): number | null {
  if (c.address?.latitude == null || c.address?.longitude == null || !userCoords.value) return null
  return haversineKm(userCoords.value.lat, userCoords.value.lng, c.address.latitude, c.address.longitude)
}

function initials(name: string): string {
  return name.split(/\s+/).slice(0, 2).map((w) => w.charAt(0).toUpperCase()).join('')
}

function localDateKey(value: string | Date): string {
  const d = new Date(value)
  const year = d.getFullYear()
  const month = String(d.getMonth() + 1).padStart(2, '0')
  const day = String(d.getDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
}

function hasCoordinates(c: CustomerSite): boolean {
  return c.address?.latitude != null && c.address?.longitude != null
}

function customerPhone(c: CustomerSite): string {
  return c.contacts?.[0]?.phone ?? ''
}

function openGoogleMaps(c: CustomerSite) {
  if (c.address?.latitude != null && c.address?.longitude != null) {
    window.open(`https://www.google.com/maps/dir/?api=1&destination=${c.address.latitude},${c.address.longitude}`, '_blank', 'noopener')
  } else if (c.address?.previewAddress) {
    window.open(`https://www.google.com/maps/dir/?api=1&destination=${encodeURIComponent(c.address.previewAddress)}`, '_blank', 'noopener')
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

function applyFilters(list: CustomerSite[]): CustomerSite[] {
  let result = list
  if (filterSegment.value) result = result.filter((c) => c.segment === filterSegment.value)
  if (filterCategory.value) result = result.filter((c) => c.category === filterCategory.value)
  return result
}

function activeFilterCount(): number {
  let n = 0
  if (filterSegment.value) n++
  if (filterCategory.value) n++
  return n
}

function resetFilters() {
  filterSegment.value = ''
  filterCategory.value = ''
}

const allCustomers = computed(() => crm.myCustomers)

const availableSegments = computed(() =>
  [...new Set(allCustomers.value.map((c) => c.segment).filter(Boolean))].sort(),
)

const availableCategories = computed(() =>
  [...new Set(allCustomers.value.map((c) => c.category).filter(Boolean))].sort(),
)

const filteredCustomers = computed(() => {
  let list = [...allCustomers.value]
  const q = searchQuery.value.trim().toLowerCase()
  if (q) {
    list = list.filter((c) => {
      const phone = customerPhone(c)
      const hay = `${c.name} ${c.address?.previewAddress} ${c.segment} ${c.category} ${phone} ${c.salesExecutiveName} ${c.parentCompanyName} ${c.region}`.toLowerCase()
      return hay.includes(q)
    })
  }
  list = applyFilters(list)
  return list
})

const todayCustomers = computed(() => {
  const today = localDateKey(new Date())
  return filteredCustomers.value.filter((c) => localDateKey(c.updatedAt) === today)
})

const nearbyCustomers = computed(() => {
  if (!userCoords.value) return []
  return filteredCustomers.value
    .map((c) => ({ c, dist: getDistance(c) }))
    .filter((x): x is { c: CustomerSite; dist: number } => x.dist !== null)
    .sort((a, b) => a.dist - b.dist)
    .map((x) => x.c)
})

const tabCountAll = computed(() => filteredCustomers.value.length)
const tabCountToday = computed(() => todayCustomers.value.length)
const tabCountNearby = computed(() => nearbyCustomers.value.length)

const displayedCustomers = computed(() => {
  let list = activeTab.value === 'today' ? todayCustomers.value : activeTab.value === 'nearby' ? nearbyCustomers.value : filteredCustomers.value
  if (sortBy.value === 'name-asc') list = [...list].sort((a, b) => a.name.localeCompare(b.name))
  else if (sortBy.value === 'name-desc') list = [...list].sort((a, b) => b.name.localeCompare(a.name))
  else if (sortBy.value === 'distance') {
    list = [...list].sort((a, b) => {
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
  try { await crm.loadMyCustomers() } catch (e: unknown) { error.value = crm.errorMessage(e) }
})
</script>

<template>
  <section class="mc-page">
    <div class="mc-header">
      <div class="mc-header-left">
        <span class="mc-avatar">{{ auth.user?.fullName?.slice(0, 1) }}</span>
        <div class="mc-header-text">
          <strong>My Customers</strong>
          <small>{{ allCustomers.length }} customers &middot; {{ todayCustomers.length }} updated today</small>
        </div>
      </div>
      <button class="mc-header-action" @click="showFilterPanel = true" aria-label="Open filters">
        <i class="pi pi-sliders-h" />
        <span v-if="activeFilterCount()" class="mc-notif-dot">{{ activeFilterCount() }}</span>
      </button>
    </div>

    <div class="mc-search">
      <i class="pi pi-search" />
      <input v-model="searchQuery" placeholder="Search customers, company or area" aria-label="Search customers" />
      <button v-if="searchQuery" class="mc-search-clear" aria-label="Clear search" @click="searchQuery = ''">
        <i class="pi pi-times" />
      </button>
    </div>

    <div class="mc-tabs">
      <button class="mc-tab" :class="{ active: activeTab === 'all' }" @click="activeTab = 'all'">
        All <span class="mc-tab-badge">{{ tabCountAll }}</span>
      </button>
      <button class="mc-tab" :class="{ active: activeTab === 'today' }" @click="activeTab = 'today'">
        Today <span class="mc-tab-badge">{{ tabCountToday }}</span>
      </button>
      <button class="mc-tab" :class="{ active: activeTab === 'nearby' }" @click="activeTab = 'nearby'">
        Nearby <span class="mc-tab-badge">{{ tabCountNearby }}</span>
      </button>
    </div>

    <Message v-if="error" severity="error" closable @close="error = ''">{{ error }}</Message>
    <Message v-if="gpsDenied && activeTab === 'nearby'" severity="warn" :closable="false">
      Location access was denied. Nearby tab requires GPS permission.
    </Message>

    <div v-if="crm.loading" class="mc-skeleton-list">
      <div v-for="n in 4" :key="n" class="mc-skeleton-card">
        <div class="sk-row"><div class="sk-avatar" /><div class="sk-lines"><div class="sk-line w60" /><div class="sk-line w40" /></div></div>
        <div class="sk-line w80" /><div class="sk-line w50" />
        <div class="sk-row sk-actions"><div class="sk-btn" /><div class="sk-btn" /><div class="sk-btn" /></div>
      </div>
    </div>

    <div v-else-if="!displayedCustomers.length && !error" class="mc-empty">
      <div class="mc-empty-icon"><i class="pi pi-map-marker" /></div>
      <strong>No customers found</strong>
      <span v-if="searchQuery || activeFilterCount()">Try adjusting your search or filters.</span>
      <span v-else-if="activeTab === 'today'">No customers updated today.</span>
      <span v-else-if="activeTab === 'nearby'">No customers with valid coordinates.</span>
      <span v-else>You have no existing customers yet.</span>
      <button v-if="searchQuery || activeFilterCount()" class="mc-reset-btn" @click="searchQuery = ''; resetFilters()">
        <i class="pi pi-refresh" /> Reset filters
      </button>
    </div>

    <template v-else>
      <div class="mc-section-header">
        <strong>Customer list</strong>
        <button class="mc-sort-trigger" @click="showSortMenu = !showSortMenu">
          <i class="pi pi-sort-alt" /> {{ sortOptions.find((s) => s.value === sortBy)?.label }}
          <i class="pi pi-chevron-down" />
        </button>
      </div>

      <div v-if="showSortMenu" class="mc-sort-menu">
        <button v-for="opt in sortOptions" :key="opt.value" :class="{ active: sortBy === opt.value }" @click="sortBy = opt.value; showSortMenu = false">
          {{ opt.label }}
        </button>
      </div>

      <div class="mc-card-list">
        <article v-for="customer in displayedCustomers" :key="customer.id" class="mc-card">
          <div class="mc-card-top">
            <span class="mc-card-avatar" :style="{ background: hasCoordinates(customer) ? '#059669' : '#94a3b8' }">{{ initials(customer.name || 'Unnamed') }}</span>
            <div class="mc-card-identity">
              <strong>{{ customer.name || 'Unnamed customer' }}</strong>
              <small>{{ customer.segment || customer.category || 'Uncategorized' }}</small>
            </div>
            <span v-if="getDistance(customer) !== null" class="mc-distance-pill">
              <i class="pi pi-map-marker" /> {{ formatDistance(getDistance(customer)!) }}
            </span>
          </div>

          <div class="mc-card-middle">
            <div class="mc-card-address">
              <i class="pi pi-map-marker" />
              <span>{{ customer.address?.previewAddress || 'No address' }}</span>
            </div>
            <div class="mc-card-meta">
              <span v-if="customer.customerCode" class="mc-meta-item"><i class="pi pi-id-card" /> {{ customer.customerCode }}</span>
              <span v-if="customer.parentCompanyName" class="mc-meta-item"><i class="pi pi-building" /> {{ customer.parentCompanyName }}</span>
              <span v-if="customerPhone(customer)" class="mc-meta-item"><i class="pi pi-phone" /> {{ customerPhone(customer) }}</span>
            </div>
            <div class="mc-card-tags">
              <span v-if="customer.segment" class="mc-tag mc-tag-segment">{{ customer.segment }}</span>
              <span v-if="customer.category" class="mc-tag mc-tag-category">{{ customer.category }}</span>
            </div>
          </div>

          <div class="mc-card-actions">
            <button class="mc-action-btn mc-action-navigate" :disabled="!hasCoordinates(customer) && !customer.address?.previewAddress" @click="openGoogleMaps(customer)" :title="hasCoordinates(customer) || customer.address?.previewAddress ? 'Navigate with Google Maps' : 'No location data available'">
              <i class="pi pi-directions" /> Navigate
            </button>
            <RouterLink class="mc-action-btn mc-action-detail" :to="`/sales/my-customers/${customer.id}`">
              <i class="pi pi-eye" /> View detail
            </RouterLink>
            <RouterLink class="mc-action-btn mc-action-checkin" :to="`/sales/my-customers/${customer.id}`">
              <i class="pi pi-sign-in" /> Check in
            </RouterLink>
          </div>
        </article>
      </div>
    </template>

    <button class="mc-fab" @click="showFilterPanel = true" aria-label="Open filter panel">
      <i class="pi pi-filter" />
      <span v-if="activeFilterCount()" class="mc-fab-badge">{{ activeFilterCount() }}</span>
    </button>

    <Dialog v-model:visible="showFilterPanel" modal header="Filter customers" :style="{ width: 'min(92vw, 400px)' }" :dt="{ background: 'var(--surface-card)', borderRadius: 'var(--radius-xl)' }">
      <div class="mc-filter-body">
        <label class="mc-filter-field">
          <span>Segment</span>
          <select v-model="filterSegment">
            <option value="">All segments</option>
            <option v-for="s in availableSegments" :key="s" :value="s">{{ s }}</option>
          </select>
        </label>
        <label class="mc-filter-field">
          <span>Category</span>
          <select v-model="filterCategory">
            <option value="">All categories</option>
            <option v-for="c in availableCategories" :key="c" :value="c">{{ c }}</option>
          </select>
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
.mc-page { display: flex; flex-direction: column; gap: 0.85rem; padding-bottom: 1.5rem; }

.mc-header {
  display: flex; align-items: center; justify-content: space-between;
  padding: 0.15rem 0;
}
.mc-header-left { display: flex; align-items: center; gap: 0.7rem; }
.mc-avatar {
  width: 38px; height: 38px; display: grid; place-items: center;
  border-radius: 50%; background: linear-gradient(135deg, #059669, #047857);
  color: #fff; font-weight: 800; font-size: 0.85rem; flex-shrink: 0;
  box-shadow: 0 2px 8px rgba(5, 150, 105, 0.25);
}
.mc-header-text { display: flex; flex-direction: column; gap: 0.05rem; }
.mc-header-text strong { font-size: 1rem; font-weight: 800; color: #0f172a; }
.mc-header-text small { color: #64748b; font-size: 0.7rem; font-weight: 500; }
.mc-header-action {
  position: relative; width: 36px; height: 36px; display: grid; place-items: center;
  border-radius: 50%; border: 1px solid #e2e8f0; background: #fff; color: #64748b;
  cursor: pointer; font-size: 0.9rem; transition: all 0.15s ease;
}
.mc-header-action:hover { color: #059669; border-color: #d1d5db; background: #ecfdf5; }
.mc-notif-dot {
  position: absolute; top: -3px; right: -3px; min-width: 16px; height: 16px;
  display: grid; place-items: center; border-radius: 9999px; background: #dc2626;
  color: #fff; font-size: 0.55rem; font-weight: 700; padding: 0 4px;
  border: 2px solid #fff;
}

.mc-search {
  display: flex; align-items: center; gap: 0.5rem;
  padding: 0.55rem 0.8rem; background: #fff; border: 1px solid #e2e8f0;
  border-radius: 14px; transition: border-color 0.15s ease, box-shadow 0.15s ease;
}
.mc-search:focus-within { border-color: #059669; box-shadow: 0 0 0 3px rgba(5, 150, 105, 0.08); }
.mc-search i { color: #94a3b8; font-size: 0.85rem; flex-shrink: 0; }
.mc-search input {
  flex: 1; border: 0; outline: 0; background: transparent; color: #0f172a;
  font-size: 0.8rem; font-weight: 500;
}
.mc-search input::placeholder { color: #94a3b8; }
.mc-search-clear {
  width: 22px; height: 22px; display: grid; place-items: center;
  border-radius: 50%; border: 0; background: #f1f5f9; color: #64748b;
  cursor: pointer; font-size: 0.65rem; transition: background 0.15s ease;
}
.mc-search-clear:hover { background: #e2e8f0; }

.mc-tabs {
  display: grid; grid-template-columns: repeat(3, 1fr); gap: 0.4rem;
}
.mc-tab {
  display: flex; align-items: center; justify-content: center; gap: 0.35rem;
  padding: 0.55rem 0; border-radius: 12px; border: 1px solid #e2e8f0;
  background: #fff; color: #64748b; font-size: 0.75rem; font-weight: 600;
  cursor: pointer; transition: all 0.15s ease;
}
.mc-tab.active { background: #ecfdf5; color: #059669; border-color: #a7f3d0; font-weight: 700; }
.mc-tab:not(.active):hover { background: #f8fafc; border-color: #cbd5e1; }
.mc-tab-badge {
  min-width: 18px; height: 18px; display: inline-grid; place-items: center;
  border-radius: 9999px; background: #f1f5f9; font-size: 0.6rem; font-weight: 700;
  padding: 0 4px;
}
.mc-tab.active .mc-tab-badge { background: #d1fae5; color: #059669; }

.mc-section-header {
  display: flex; align-items: center; justify-content: space-between;
}
.mc-section-header strong { font-size: 0.82rem; font-weight: 800; color: #0f172a; }
.mc-sort-trigger {
  display: flex; align-items: center; gap: 0.3rem; padding: 0.3rem 0.6rem;
  border-radius: 10px; border: 1px solid #e2e8f0; background: #fff;
  color: #64748b; font-size: 0.68rem; font-weight: 600; cursor: pointer;
  transition: all 0.15s ease;
}
.mc-sort-trigger:hover { border-color: #cbd5e1; background: #f8fafc; }
.mc-sort-trigger i { font-size: 0.6rem; }
.mc-sort-menu {
  display: flex; gap: 0.35rem; flex-wrap: wrap;
}
.mc-sort-menu button {
  padding: 0.4rem 0.7rem; border-radius: 10px; border: 1px solid #e2e8f0;
  background: #fff; color: #64748b; font-size: 0.68rem; font-weight: 600;
  cursor: pointer; transition: all 0.15s ease;
}
.mc-sort-menu button.active { background: #059669; color: #fff; border-color: #059669; }
.mc-sort-menu button:not(.active):hover { border-color: #cbd5e1; background: #f8fafc; }

.mc-card-list { display: flex; flex-direction: column; gap: 0.65rem; }
.mc-card {
  padding: 0.9rem 1rem; background: #fff; border: 1px solid #eef1f6;
  border-radius: 16px; box-shadow: 0 1px 3px rgba(0, 0, 0, 0.03);
  display: flex; flex-direction: column; gap: 0.6rem;
  transition: border-color 0.15s ease, box-shadow 0.15s ease;
}
.mc-card:hover { border-color: #d6dce6; box-shadow: 0 4px 12px rgba(0, 0, 0, 0.05); }

.mc-card-top { display: flex; align-items: center; gap: 0.65rem; }
.mc-card-avatar {
  width: 38px; height: 38px; display: grid; place-items: center;
  border-radius: 12px; color: #fff; font-weight: 800; font-size: 0.72rem;
  flex-shrink: 0;
}
.mc-card-identity { flex: 1; min-width: 0; display: flex; flex-direction: column; gap: 0.05rem; }
.mc-card-identity strong { font-size: 0.82rem; font-weight: 700; color: #0f172a; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.mc-card-identity small { color: #64748b; font-size: 0.65rem; font-weight: 500; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.mc-distance-pill {
  display: flex; align-items: center; gap: 0.25rem; padding: 0.2rem 0.5rem;
  border-radius: 9999px; background: #ecfdf5; color: #059669;
  font-size: 0.6rem; font-weight: 700; white-space: nowrap; flex-shrink: 0;
}
.mc-distance-pill i { font-size: 0.55rem; }

.mc-card-middle { display: flex; flex-direction: column; gap: 0.35rem; }
.mc-card-address { display: flex; align-items: flex-start; gap: 0.4rem; color: #64748b; font-size: 0.72rem; line-height: 1.4; }
.mc-card-address i { margin-top: 0.1rem; font-size: 0.68rem; flex-shrink: 0; color: #94a3b8; }
.mc-card-meta { display: flex; flex-wrap: wrap; gap: 0.4rem 0.7rem; }
.mc-meta-item { display: flex; align-items: center; gap: 0.25rem; color: #475569; font-size: 0.65rem; font-weight: 500; }
.mc-meta-item i { font-size: 0.55rem; color: #94a3b8; }
.mc-card-tags { display: flex; flex-wrap: wrap; gap: 0.3rem; }
.mc-tag {
  display: inline-block; padding: 0.15rem 0.5rem; border-radius: 6px;
  font-size: 0.58rem; font-weight: 600; line-height: 1.5;
}
.mc-tag-segment { background: #dbeafe; color: #1d4ed8; }
.mc-tag-category { background: #f1f5f9; color: #475569; }

.mc-card-actions {
  display: grid; grid-template-columns: 1fr 1fr 1fr; gap: 0.4rem;
  padding-top: 0.45rem; border-top: 1px solid #f1f5f9;
}
.mc-action-btn {
  display: flex; align-items: center; justify-content: center; gap: 0.3rem;
  padding: 0.45rem 0; border-radius: 10px; border: 1px solid #e2e8f0;
  background: #fff; font-size: 0.65rem; font-weight: 600; cursor: pointer;
  text-decoration: none; transition: all 0.15s ease;
}
.mc-action-btn:disabled { opacity: 0.4; cursor: not-allowed; }
.mc-action-navigate { color: #2563eb; }
.mc-action-navigate:not(:disabled):hover { background: #eff6ff; border-color: #bfdbfe; }
.mc-action-detail { color: #0f172a; }
.mc-action-detail:hover { background: #f8fafc; border-color: #cbd5e1; }
.mc-action-checkin { color: #059669; }
.mc-action-checkin:hover { background: #ecfdf5; border-color: #a7f3d0; }

.mc-empty {
  display: flex; flex-direction: column; align-items: center; gap: 0.5rem;
  padding: 2.5rem 1rem; text-align: center;
}
.mc-empty-icon {
  width: 52px; height: 52px; display: grid; place-items: center;
  border-radius: 16px; background: #ecfdf5; color: #059669; font-size: 1.4rem;
}
.mc-empty strong { color: #0f172a; font-size: 0.88rem; }
.mc-empty span { color: #64748b; font-size: 0.75rem; line-height: 1.5; }
.mc-reset-btn {
  display: flex; align-items: center; gap: 0.3rem; margin-top: 0.3rem;
  padding: 0.45rem 0.85rem; border-radius: 10px; border: 1px solid #e2e8f0;
  background: #fff; color: #059669; font-size: 0.72rem; font-weight: 600;
  cursor: pointer; transition: all 0.15s ease;
}
.mc-reset-btn:hover { background: #ecfdf5; border-color: #a7f3d0; }

.mc-skeleton-list { display: flex; flex-direction: column; gap: 0.65rem; }
.mc-skeleton-card {
  padding: 1rem; background: #fff; border: 1px solid #eef1f6; border-radius: 16px;
  display: flex; flex-direction: column; gap: 0.6rem;
  animation: mc-pulse 1.5s ease-in-out infinite;
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
@keyframes mc-pulse { 0%, 100% { opacity: 1; } 50% { opacity: 0.5; } }

.mc-fab {
  position: fixed; bottom: 88px; right: calc(50% - 190px);
  width: 48px; height: 48px; border-radius: 50%; border: 0;
  background: linear-gradient(135deg, #059669, #047857); color: #fff;
  font-size: 1.1rem; cursor: pointer; z-index: 50;
  box-shadow: 0 6px 20px rgba(5, 150, 105, 0.35);
  display: grid; place-items: center; transition: transform 0.2s ease, box-shadow 0.2s ease;
}
.mc-fab:hover { transform: scale(1.08); box-shadow: 0 8px 24px rgba(5, 150, 105, 0.45); }
.mc-fab-badge {
  position: absolute; top: -2px; right: -2px; min-width: 18px; height: 18px;
  display: grid; place-items: center; border-radius: 9999px; background: #dc2626;
  color: #fff; font-size: 0.55rem; font-weight: 700; padding: 0 4px;
  border: 2px solid #fff;
}

.mc-filter-body { display: flex; flex-direction: column; gap: 0.85rem; }
.mc-filter-field { display: flex; flex-direction: column; gap: 0.3rem; }
.mc-filter-field span { color: #64748b; font-size: 0.72rem; font-weight: 700; text-transform: uppercase; letter-spacing: 0.04em; }
.mc-filter-field select {
  padding: 0.55rem 0.7rem; border: 1px solid #e2e8f0; border-radius: 10px;
  background: #f8fafc; color: #0f172a; font-size: 0.8rem; font-weight: 500;
  cursor: pointer; transition: border-color 0.15s ease;
}
.mc-filter-field select:focus { outline: 0; border-color: #059669; }

@media (max-width: 480px) {
  .mc-fab { right: 1rem; }
}
</style>
