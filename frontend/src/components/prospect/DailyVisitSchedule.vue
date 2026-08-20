<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import type { Prospect } from '../../types/crm'
import { useVisitLocation } from '../../composables/sales/useVisitLocation'
import { haversineKm, formatDistance } from '../../utils/maps'
import { WORKING_DAYS, planWeeklyVisits, routeDay } from '../../utils/visitPlanning'

const props = defineProps<{ loading?: boolean; prospects?: Prospect[] }>()
const router = useRouter()
const { state: gps, refreshOnce, distanceTo, distanceFormatted } = useVisitLocation()

const open = ref(false)
const days = [...WORKING_DAYS]
const dayPlans = ref<Record<string, Prospect[]>>({})
const ROUTE_STORAGE_KEY = 'crm-sales-weekly-route'
const selectedDay = ref('Monday')
const selectedProspectId = ref('')
const moveTargets = ref<Record<string, string>>({})
const expandedItems = ref<Record<string, boolean>>({})
const actionItem = ref<Prospect | null>(null)
const actionMoveDay = ref('')
const overflow = ref<Prospect[]>([])

onMounted(() => {
  refreshOnce().catch(() => {})
  try {
    const saved = localStorage.getItem(ROUTE_STORAGE_KEY)
    dayPlans.value = saved ? JSON.parse(saved) : Object.fromEntries(days.map((day) => [day, []]))
  } catch {
    dayPlans.value = Object.fromEntries(days.map((day) => [day, []]))
  }
})

watch(dayPlans, (value) => {
  try { localStorage.setItem(ROUTE_STORAGE_KEY, JSON.stringify(value)) } catch { /* storage may be unavailable */ }
}, { deep: true })

const activeProspects = computed(() =>
  (props.prospects ?? []).filter(p => p.status !== 'LOST' && p.status !== 'CONVERTED' && p.status !== 'WON')
)

const todayVisits = computed(() => {
  const dow = new Date().getDay()
  if (dow === 0 || dow === 6) return []
  const dayName = days[dow - 1]
  return dayPlans.value[dayName] ?? []
})

const sortedActive = computed(() => {
  const list = [...activeProspects.value]
  if (!gps.value.coords) {
    list.sort((a, b) => a.placeName.localeCompare(b.placeName))
    return list
  }
  return list.sort((a, b) => {
    const dA = a.latitude != null && a.longitude != null ? (distanceTo(a.latitude, a.longitude) ?? Infinity) : Infinity
    const dB = b.latitude != null && b.longitude != null ? (distanceTo(b.latitude, b.longitude) ?? Infinity) : Infinity
    return dA - dB
  })
})

function buildPreview() {
  const result = planWeeklyVisits(activeProspects.value, gps.value.coords ? { latitude: gps.value.coords.latitude, longitude: gps.value.coords.longitude } : null)
  dayPlans.value = result.plan
  overflow.value = result.overflow
}

function openPreview() {
  if (!gps.value.coords && !gps.value.loading) refreshOnce().catch(() => {})
  buildPreview()
  open.value = true
}

function autoRouteToday() {
  const dow = new Date().getDay()
  if (dow === 0 || dow === 6) return
  const day = days[dow - 1]
  dayPlans.value[day] = routeDay(dayPlans.value[day] ?? [], gps.value.coords ? { latitude: gps.value.coords.latitude, longitude: gps.value.coords.longitude } : null)
  selectedDay.value = day
  open.value = true
}

function move(item: Prospect, from: string, to: string) {
  if (from === to) return
  dayPlans.value[from] = dayPlans.value[from].filter((c) => c.id !== item.id)
  dayPlans.value[to] = [...dayPlans.value[to], item]
}

function openDetail(item: Prospect) {
  open.value = false
  router.push(`/sales/my-prospects/${item.id}`)
}

function checkIn(item: Prospect) {
  open.value = false
  router.push(`/sales/my-prospects/${item.id}/check-in`)
}

function prospectDistance(item: Prospect): string {
  if (item.latitude == null || item.longitude == null || !gps.value.coords) return ''
  return formatDistance(haversineKm(gps.value.coords.latitude, gps.value.coords.longitude, item.latitude, item.longitude))
}

function statusDot(s: string) {
  const map: Record<string, string> = {
    NEW_LEAD: 'dot--red', CONTACTED: 'dot--orange', INTERESTED: 'dot--amber',
    QUALIFIED: 'dot--blue', PROPOSAL_SENT: 'dot--indigo', NEGOTIATION: 'dot--purple',
  }
  return map[s] ?? 'dot--gray'
}

function statusLabel(s: string) {
  return s.replace(/_/g, ' ').toLowerCase().replace(/\b\w/g, c => c.toUpperCase())
}

const totalPreview = computed(() => Object.values(dayPlans.value).reduce((sum, items) => sum + items.length, 0))
const selectedDayVisits = computed(() => dayPlans.value[selectedDay.value] ?? [])
const unscheduledProspects = computed(() => activeProspects.value.filter((item) => !Object.values(dayPlans.value).some((items) => items.some((candidate) => candidate.id === item.id))))
function addToRoute() {
  const item = unscheduledProspects.value.find((candidate) => candidate.id === selectedProspectId.value)
  if (!item) return
  for (const day of days) dayPlans.value[day] = (dayPlans.value[day] ?? []).filter((candidate) => candidate.id !== item.id)
  dayPlans.value[selectedDay.value] = [...(dayPlans.value[selectedDay.value] ?? []), item]
  selectedProspectId.value = ''
}
function removeFromRoute(item: Prospect) {
  for (const day of days) dayPlans.value[day] = (dayPlans.value[day] ?? []).filter((candidate) => candidate.id !== item.id)
}
function toggleItem(item: Prospect) { expandedItems.value[item.id] = !expandedItems.value[item.id] }
function openActionModal(item: Prospect) { actionItem.value = item; actionMoveDay.value = '' }
function closeActionModal() { actionItem.value = null }
function moveFromModal() {
  if (!actionItem.value || !actionMoveDay.value) return
  const from = days.find((day) => dayPlans.value[day]?.some((candidate) => candidate.id === actionItem.value?.id))
  if (from) move(actionItem.value, from, actionMoveDay.value)
  selectedDay.value = actionMoveDay.value
  closeActionModal()
}
function confirmRemove(item: Prospect) {
  if (window.confirm(`Remove ${item.placeName} from this weekly route?`)) removeFromRoute(item)
}
function moveFromInline(item: Prospect, from: string) {
  const to = moveTargets.value[item.id]
  if (!to || to === from) return
  move(item, from, to)
  moveTargets.value[item.id] = ''
  selectedDay.value = to
}
</script>

<template>
  <section class="schedule-card" aria-labelledby="daily-schedule-title">
    <div class="schedule-heading">
      <div>
        <span class="schedule-eyebrow">Field activity</span>
        <h2 id="daily-schedule-title">Daily Visit Schedule</h2>
      </div>
      <span class="schedule-date">{{ new Intl.DateTimeFormat('en-GB', { weekday: 'short', day: '2-digit', month: 'short' }).format(new Date()) }}</span>
    </div>

    <div v-if="loading" class="schedule-state">
      <i class="pi pi-spin pi-spinner" />
      <span>Loading prospects...</span>
    </div>

    <template v-else>
      <div v-if="!sortedActive.length" class="schedule-state">
        <i class="pi pi-calendar-plus" />
        <strong>No active prospects</strong>
        <span>Add prospects to see your visit route.</span>
      </div>

      <div v-else class="schedule-list">
        <div class="schedule-gps" v-if="!gps.coords">
          <i class="pi pi-map-marker" />
          <span v-if="gps.loading">Getting your location...</span>
          <span v-else-if="gps.error">Location unavailable — showing alphabetical order</span>
          <span v-else>Tap to enable GPS for distance sorting</span>
          <button v-if="!gps.loading && !gps.coords" class="gps-btn" type="button" @click="refreshOnce().catch(() => {})">Enable</button>
        </div>

        <div v-for="(item, idx) in sortedActive.slice(0, 5)" :key="item.id" class="schedule-item">
          <span class="schedule-rank">{{ idx + 1 }}</span>
          <span :class="['schedule-dot', statusDot(item.status)]" />
          <div class="schedule-item-info">
            <strong>{{ item.placeName }}</strong>
            <small>{{ item.formattedAddress || item.placeCategory }}</small>
          </div>
          <span v-if="prospectDistance(item)" class="schedule-dist">
            <i class="pi pi-map-marker" />
            {{ prospectDistance(item) }}
          </span>
        </div>

        <div v-if="sortedActive.length > 5" class="schedule-more">
          +{{ sortedActive.length - 5 }} more prospects
        </div>
      </div>
    </template>

    <div v-if="!loading && sortedActive.length" class="inline-weekly-route">
      <div class="inline-route-heading"><div><span class="schedule-eyebrow">Weekly route</span><strong>Visit plan by day</strong></div><div class="route-header-actions"><span>{{ totalPreview }} prospects</span><button type="button" class="route-primary-btn" @click="openPreview"><i class="pi pi-calendar-plus" /> Plan my week</button><button type="button" class="route-secondary-btn" @click="autoRouteToday"><i class="pi pi-directions" /> Auto route today</button></div></div>
      <div class="day-actions" role="tablist" aria-label="Weekly visit days">
        <button v-for="day in days" :key="day" type="button" class="day-action" :class="{ active: selectedDay === day }" @click="selectedDay = day"><strong>{{ day.slice(0, 3) }}</strong><span>{{ dayPlans[day]?.length ?? 0 }}</span></button>
      </div>
      <div class="add-route-control"><select v-model="selectedProspectId" aria-label="Select prospect to schedule"><option value="">Select prospect from pipeline...</option><option v-for="item in unscheduledProspects" :key="item.id" :value="item.id">{{ item.placeName }}</option></select><button type="button" :disabled="!selectedProspectId" @click="addToRoute"><i class="pi pi-plus" /> Add to {{ selectedDay.slice(0, 3) }}</button></div>
      <div class="selected-day-list">
        <div class="selected-day-title"><strong>{{ selectedDay }}</strong><span>{{ selectedDayVisits.length }} customers</span></div>
        <div v-for="(item, index) in selectedDayVisits" :key="item.id" class="inline-route-item">
          <span class="route-number">{{ index + 1 }}</span><div class="route-prospect-info route-prospect-link" role="button" tabindex="0" @click="openActionModal(item)" @keydown.enter="openActionModal(item)"><strong>{{ item.placeName }}</strong><small>{{ item.formattedAddress || 'No address' }}</small></div><span v-if="prospectDistance(item)" class="route-dist"><i class="pi pi-map-marker" /> {{ prospectDistance(item) }}</span><i class="pi pi-ellipsis-v route-expand-icon" />
          <div v-if="expandedItems[item.id]" class="inline-item-actions"><button type="button" class="route-action route-action--detail" @click="openDetail(item)"><i class="pi pi-eye" /><span>View detail</span></button><button type="button" class="route-action route-action--checkin" @click="checkIn(item)"><i class="pi pi-sign-in" /><span>Check in</span></button><select v-model="moveTargets[item.id]" class="inline-move-select" aria-label="Move prospect to another day"><option value="">Move to...</option><option v-for="day in days" :key="day" :value="day" :disabled="day === selectedDay">{{ day }}</option></select><button type="button" class="route-action route-action--move" :disabled="!moveTargets[item.id]" @click="moveFromInline(item, selectedDay)"><i class="pi pi-arrow-right-arrow-left" /><span>Move</span></button><button type="button" class="route-action route-action--remove" @click="removeFromRoute(item)"><i class="pi pi-trash" /><span>Remove</span></button></div>
        </div>
        <div v-if="!selectedDayVisits.length" class="inline-route-empty"><i class="pi pi-calendar-plus" /> No customers planned for {{ selectedDay }}.</div>
        <div v-if="overflow.length" class="schedule-overflow"><i class="pi pi-exclamation-triangle" /><span>{{ overflow.length }} customer(s) need planning next week or manual scheduling.</span></div>
      </div>
    </div>
  </section>

  <Teleport to="body">
    <div v-if="actionItem" class="action-backdrop" @click.self="closeActionModal">
      <section class="action-modal" role="dialog" aria-modal="true" aria-labelledby="action-title">
        <header class="action-modal-head"><h2 id="action-title">Customer Actions</h2><button type="button" class="route-close" aria-label="Close" @click="closeActionModal"><i class="pi pi-times" /></button></header>
        <div class="action-customer"><strong>{{ actionItem.placeName }}</strong><small>{{ actionItem.formattedAddress || 'No address' }}</small><span>{{ prospectDistance(actionItem) || 'Distance unavailable' }}</span></div>
        <div class="action-options"><button type="button" class="action-option" @click="openDetail(actionItem)"><span><strong>View details</strong><small>Open prospect detail, visit history, and notes.</small></span><i class="pi pi-chevron-right" /></button><button type="button" class="action-option action-option--primary" @click="checkIn(actionItem)"><span><strong>Check in</strong><small>Start the visit at this prospect.</small></span><i class="pi pi-sign-in" /></button><div class="action-move-box"><div><strong>Move schedule</strong><small>Choose another weekday for this visit.</small></div><div class="action-move-controls"><select v-model="actionMoveDay" aria-label="Move schedule to"><option value="">Choose day...</option><option v-for="day in days" :key="day" :value="day">{{ day }}</option></select><button type="button" :disabled="!actionMoveDay" @click="moveFromModal">Move</button></div></div><button type="button" class="action-option action-option--danger" @click="confirmRemove(actionItem); closeActionModal()"><span><strong>Remove from route</strong><small>You will be asked to confirm this action.</small></span><i class="pi pi-trash" /></button></div>
      </section>
    </div>
  </Teleport>

  <Teleport to="body">
    <div v-if="open" class="route-backdrop" @click.self="open = false">
      <section class="route-modal" role="dialog" aria-modal="true" aria-labelledby="route-title">
        <header class="route-modal-head">
          <div>
            <span class="schedule-eyebrow">Recommended planning</span>
            <h2 id="route-title">Weekly Route Preview</h2>
            <p>{{ totalPreview }} assigned prospects · drag a card to another day</p>
          </div>
          <button class="route-close" type="button" aria-label="Close" @click="open = false">
            <i class="pi pi-times" />
          </button>
        </header>

        <div class="route-notice">
          <i class="pi pi-info-circle" /> Preview only. Existing confirmed schedules and assignments are not changed.
        </div>

        <div class="route-days">
          <article
            v-for="day in days"
            :key="day"
            class="route-day"
            @dragover.prevent
            @drop="(event) => {
              const payload = (event as DragEvent).dataTransfer?.getData('text/plain')
              if (payload) {
                const [from, id] = payload.split('|')
                const item = dayPlans[from]?.find(c => c.id === id)
                if (item) move(item, from, day)
              }
            }"
          >
            <header>
              <strong>{{ day }}</strong>
              <span>{{ dayPlans[day]?.length ?? 0 }} visits</span>
            </header>
            <div v-if="!dayPlans[day]?.length" class="route-empty">Drop prospect here</div>
            <div
              v-for="(item, i) in dayPlans[day]"
              :key="item.id"
              class="route-prospect"
              draggable="true"
              @dragstart="(e) => (e as DragEvent).dataTransfer?.setData('text/plain', `${day}|${item.id}`)"
            >
              <span class="route-number">{{ i + 1 }}</span>
              <div class="route-prospect-info">
                <strong>{{ item.placeName }}</strong>
                <small>{{ item.formattedAddress || 'No address' }}</small>
                <span v-if="prospectDistance(item)" class="route-dist">
                  <i class="pi pi-map-marker" /> {{ prospectDistance(item) }}
                </span>
              </div>
              <i class="pi pi-bars route-drag" />
              <div class="route-actions">
                <button type="button" class="route-action route-action--detail" @click.stop="openDetail(item)">
                  <i class="pi pi-eye" /><span>Detail</span>
                </button>
                <button type="button" class="route-action route-action--checkin" @click.stop="checkIn(item)">
                  <i class="pi pi-sign-in" /><span>Check in</span>
                </button>
              </div>
            </div>
          </article>
        </div>

        <footer class="route-modal-foot">
          <span><i class="pi pi-sort-alt" /> Ordered by nearest distance from your location</span>
          <button type="button" class="route-done" @click="open = false">Close preview</button>
        </footer>
      </section>
    </div>
  </Teleport>
</template>

<style scoped>
.schedule-card {
  min-height: 190px;
  padding: 1rem 1.1rem;
  border: 1px solid #e5eaf0;
  border-radius: 16px;
  background: #fff;
  box-shadow: 0 1px 3px rgba(15,23,42,.04);
  display: flex;
  flex-direction: column;
}

.schedule-heading {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 1rem;
}

.schedule-eyebrow {
  color: #94a3b8;
  font-size: .62rem;
  font-weight: 800;
  letter-spacing: .08em;
  text-transform: uppercase;
}

h2 { margin: .15rem 0 0; color: #0f172a; font-size: 1rem; }

.schedule-date {
  padding: .25rem .55rem;
  border-radius: 999px;
  background: #eff6ff;
  color: #2563eb;
  font-size: .65rem;
  font-weight: 700;
  white-space: nowrap;
}

.schedule-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: .35rem;
  min-height: 125px;
  color: #94a3b8;
  text-align: center;
}

.schedule-state i { color: #93c5fd; font-size: 1.35rem; }
.schedule-state strong { color: #475569; font-size: .75rem; }
.schedule-state span { max-width: 280px; font-size: .65rem; line-height: 1.45; }

.schedule-list {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: .35rem;
  margin-top: .65rem;
}

.schedule-gps {
  display: flex;
  align-items: center;
  gap: .4rem;
  padding: .4rem .55rem;
  border-radius: 8px;
  background: #f0f9ff;
  color: #0369a1;
  font-size: .62rem;
  font-weight: 600;
}

.schedule-gps i { font-size: .7rem; }

.gps-btn {
  margin-left: auto;
  padding: .2rem .5rem;
  border: 1px solid #7dd3fc;
  border-radius: 6px;
  background: #fff;
  color: #0369a1;
  font: 600 .58rem inherit;
  cursor: pointer;
}

.gps-btn:hover { background: #e0f2fe; }

.schedule-item {
  display: flex;
  align-items: center;
  gap: .45rem;
  padding: .45rem .5rem;
  border: 1px solid #f1f5f9;
  border-radius: 10px;
  background: #fafbfc;
  transition: border-color .12s;
}

.schedule-item:hover { border-color: #dbeafe; }

.schedule-rank {
  display: grid;
  place-items: center;
  width: 20px;
  height: 20px;
  border-radius: 6px;
  background: #eff6ff;
  color: #2563eb;
  font-size: .6rem;
  font-weight: 800;
  flex-shrink: 0;
}

.dot--red { background: #e35d6a; }
.dot--orange { background: #f59e0b; }
.dot--amber { background: #f59e0b; }
.dot--blue { background: #3b82f6; }
.dot--indigo { background: #6366f1; }
.dot--purple { background: #8b5cf6; }
.dot--gray { background: #94a3b8; }

.schedule-dot {
  width: 7px;
  height: 7px;
  border-radius: 50%;
  flex-shrink: 0;
}

.schedule-item-info {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
}

.schedule-item-info strong {
  overflow: hidden;
  color: #1e293b;
  font-size: .7rem;
  font-weight: 700;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.schedule-item-info small {
  overflow: hidden;
  color: #94a3b8;
  font-size: .58rem;
  text-overflow: ellipsis;
  white-space: nowrap;
  max-width: 180px;
}

.schedule-dist {
  display: inline-flex;
  align-items: center;
  gap: .2rem;
  padding: .15rem .4rem;
  border-radius: 6px;
  background: #f0fdf4;
  color: #16a34a;
  font-size: .58rem;
  font-weight: 700;
  white-space: nowrap;
  flex-shrink: 0;
}

.schedule-dist i { font-size: .5rem; }

.schedule-more {
  text-align: center;
  padding: .3rem;
  color: #94a3b8;
  font-size: .6rem;
  font-weight: 600;
}

.schedule-preview-btn,
.route-done {
  display: inline-flex;
  align-items: center;
  gap: .45rem;
  border: 0;
  border-radius: 11px;
  background: linear-gradient(135deg, #2563eb, #1d4ed8);
  color: #fff;
  padding: .58rem .8rem;
  font: 700 .68rem inherit;
  cursor: pointer;
  box-shadow: 0 6px 14px rgba(37,99,235,.2);
  transition: transform .15s, box-shadow .15s;
  margin-top: auto;
}

.schedule-preview-btn:hover {
  transform: translateY(-1px);
  box-shadow: 0 9px 18px rgba(37,99,235,.27);
}

.schedule-preview-btn .pi-arrow-right { font-size: .6rem; opacity: .75; }

.route-backdrop {
  position: fixed;
  z-index: 1000;
  inset: 0;
  display: grid;
  place-items: center;
  padding: 1rem;
  background: rgba(15,23,42,.42);
  backdrop-filter: blur(3px);
}

.route-modal {
  width: min(1180px, 100%);
  max-height: min(88vh, 760px);
  overflow: auto;
  padding: 1rem;
  border-radius: 18px;
  background: #f8fafc;
  box-shadow: 0 24px 70px rgba(15,23,42,.24);
}

.route-modal-head,
.route-modal-foot {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 1rem;
}

.route-modal-head p { margin: .25rem 0 0; color: #64748b; font-size: .68rem; }

.route-close {
  width: 32px;
  height: 32px;
  border: 1px solid #e2e8f0;
  border-radius: 9px;
  background: #fff;
  color: #64748b;
  cursor: pointer;
}

.route-notice {
  margin: .85rem 0;
  padding: .6rem .75rem;
  border: 1px solid #bfdbfe;
  border-radius: 10px;
  background: #eff6ff;
  color: #1d4ed8;
  font-size: .68rem;
}

.route-days {
  display: grid;
  grid-template-columns: repeat(5, minmax(170px, 1fr));
  gap: .6rem;
  overflow: auto;
}

.route-day {
  min-height: 250px;
  padding: .6rem;
  border: 1px solid #e2e8f0;
  border-radius: 12px;
  background: #fff;
}

.route-day > header {
  display: flex;
  justify-content: space-between;
  margin-bottom: .55rem;
  color: #0f172a;
  font-size: .72rem;
}

.route-day > header span { color: #94a3b8; font-size: .6rem; }

.route-prospect {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: .4rem;
  margin: .35rem 0;
  padding: .45rem;
  border: 1px solid #e5eaf0;
  border-radius: 9px;
  cursor: grab;
  transition: border-color .12s, box-shadow .12s;
}

.route-prospect:hover {
  border-color: #bfdbfe;
  box-shadow: 0 2px 6px rgba(37,99,235,.08);
}

.route-number {
  display: grid;
  place-items: center;
  width: 20px;
  height: 20px;
  border-radius: 6px;
  background: #eff6ff;
  color: #2563eb;
  font-size: .6rem;
  font-weight: 800;
  flex-shrink: 0;
}

.route-prospect-info {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
}

.route-prospect-info strong {
  overflow: hidden;
  color: #1e293b;
  font-size: .66rem;
  font-weight: 700;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.route-prospect-info small {
  overflow: hidden;
  color: #94a3b8;
  font-size: .55rem;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.route-dist {
  display: inline-flex;
  align-items: center;
  gap: .15rem;
  margin-top: .15rem;
  color: #16a34a;
  font-size: .55rem;
  font-weight: 700;
}

.route-dist i { font-size: .45rem; }

.route-drag { color: #cbd5e1; font-size: .7rem; cursor: grab; }

.route-actions {
  display: flex;
  gap: .3rem;
  width: 100%;
}

.route-action {
  display: inline-flex;
  align-items: center;
  gap: .25rem;
  padding: .25rem .45rem;
  border: 1px solid #e2e8f0;
  border-radius: 6px;
  background: #fff;
  font: 600 .55rem inherit;
  cursor: pointer;
  transition: background .1s;
}

.route-action i { font-size: .55rem; }

.route-action--detail { color: #475569; }
.route-action--detail:hover { background: #f8fafc; }

.route-action--checkin { color: #16a34a; border-color: #bbf7d0; }
.route-action--checkin:hover { background: #f0fdf4; }

.route-empty {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 60px;
  border: 1.5px dashed #e2e8f0;
  border-radius: 8px;
  color: #cbd5e1;
  font-size: .6rem;
}

.route-modal-foot {
  margin-top: .85rem;
  padding-top: .75rem;
  border-top: 1px solid #e5eaf0;
  align-items: center;
}

.route-modal-foot span {
  display: flex;
  align-items: center;
  gap: .35rem;
  color: #94a3b8;
  font-size: .62rem;
}

.route-done {
  box-shadow: none;
  background: #475569;
}

.route-done:hover {
  background: #334155;
  transform: none;
  box-shadow: none;
}

@media (max-width: 480px) {
  .schedule-card { padding: .9rem; }
}

@media (max-width: 768px) {
  .route-backdrop {
    align-items: flex-end;
    padding: .5rem;
  }

  .route-modal {
    max-height: 92vh;
    border-radius: 18px 18px 0 0;
    padding: .85rem;
  }

  .route-days {
    grid-template-columns: repeat(5, minmax(155px, 1fr));
  }

  .route-modal-foot {
    align-items: flex-start;
    flex-direction: column;
  }

  .route-done {
    width: 100%;
    justify-content: center;
  }
}
.inline-weekly-route { margin-top:.8rem; padding-top:.75rem; border-top:1px solid #edf1f5; }
.inline-route-heading { display:flex; align-items:center; justify-content:space-between; gap:.75rem; margin-bottom:.55rem; }.inline-route-heading > div { display:grid; gap:.12rem; }.inline-route-heading strong { color:#334155; font-size:.72rem; }.route-header-actions { display:flex !important; align-items:center; gap:.35rem; }.route-header-actions > span { padding:.2rem .45rem; border-radius:999px; background:#f1f5f9; color:#64748b; font-size:.58rem; font-weight:700; }.route-primary-btn,.route-secondary-btn { border:1px solid #dbeafe; border-radius:7px; padding:.34rem .5rem; background:#eff6ff; color:#2563eb; cursor:pointer; font:inherit; font-size:.58rem; font-weight:800; }.route-secondary-btn { border-color:#e2e8f0; background:#fff; color:#475569; }.route-primary-btn:hover,.route-secondary-btn:hover { filter:brightness(.97); }.schedule-overflow { display:flex; align-items:center; gap:.4rem; margin-top:.6rem; padding:.5rem .6rem; border:1px solid #fed7aa; border-radius:8px; background:#fff7ed; color:#9a3412; font-size:.62rem; line-height:1.35; }
.day-actions { display:grid; grid-template-columns:repeat(5,1fr); gap:.35rem; }.day-action { display:grid; gap:.12rem; min-height:42px; place-items:center; border:1px solid #e2e8f0; border-radius:9px; background:#fff; color:#64748b; cursor:pointer; }.day-action strong { font-size:.62rem; }.day-action span { min-width:18px; padding:.08rem .28rem; border-radius:999px; background:#f1f5f9; color:#94a3b8; font-size:.55rem; font-weight:800; }.day-action.active { border-color:#bfdbfe; background:#eff6ff; color:#2563eb; box-shadow:0 0 0 2px rgba(37,99,235,.08); }.day-action.active span { background:#2563eb; color:#fff; }
.add-route-control { display:grid; grid-template-columns:minmax(0,1fr) 112px auto; gap:.5rem; margin-top:.65rem; padding:.45rem; border:1px solid #edf1f5; border-radius:10px; background:#fbfdff; }.add-route-control select,.add-route-control button { min-width:0; min-height:30px; padding:.28rem .45rem; border:1px solid #dbe3ee; border-radius:7px; background:#fff; color:#475569; font:600 .6rem inherit; }.add-route-control button { border-color:#2563eb; background:#2563eb; color:#fff; cursor:pointer; white-space:nowrap; }.add-route-control button:disabled { opacity:.45; cursor:not-allowed; }
.selected-day-list { margin-top:.55rem; padding:.5rem; border:1px solid #e5eaf0; border-radius:10px; background:#f8fafc; }.selected-day-title { display:flex; align-items:center; justify-content:space-between; margin:0 .2rem .35rem; color:#334155; font-size:.68rem; }.selected-day-title span { color:#94a3b8; font-size:.58rem; }.inline-route-item { display:flex; align-items:center; gap:.4rem; padding:.45rem; margin:.3rem 0; border:1px solid #e5eaf0; border-radius:9px; background:#fff; }.inline-route-item .route-prospect-info { flex:1; min-width:0; }.inline-route-item .route-action { flex:0 0 auto; width:auto; padding:.35rem .45rem; }.inline-route-empty { display:flex; align-items:center; justify-content:center; gap:.35rem; min-height:58px; color:#94a3b8; font-size:.62rem; }
@media (max-width:480px) { .inline-route-item { flex-wrap:wrap; }.inline-route-item .route-prospect-info { min-width:calc(100% - 34px); }.inline-route-item .route-dist { margin-left:1.65rem; }.inline-route-item .route-action { flex:1; min-height:34px; }.day-action { min-height:40px; } }
@media (max-width:480px) { .add-route-control { grid-template-columns:1fr; }.add-route-control button { width:100%; } }
.inline-move-select { min-height:30px; max-width:105px; padding:.25rem; border:1px solid #e2e8f0; border-radius:7px; background:#fff; color:#64748b; font:600 .57rem inherit; }.route-action--move { border-color:#c7d2fe; background:#eef2ff; color:#4f46e5; }.route-action--remove { border-color:#fecaca; background:#fff1f2; color:#dc2626; }.route-action:disabled { cursor:not-allowed; opacity:.45; }
.action-backdrop { position:fixed; z-index:1100; inset:0; display:grid; place-items:center; padding:1rem; background:rgba(15,23,42,.42); backdrop-filter:blur(3px); }.action-modal { width:min(440px,100%); padding:1rem; border-radius:16px; background:#fff; box-shadow:0 24px 70px rgba(15,23,42,.25); }.action-modal-head { display:flex; align-items:center; justify-content:space-between; margin-bottom:.8rem; }.action-modal-head h2 { margin:0; color:#334155; font-size:1rem; }.action-customer { display:grid; gap:.2rem; padding:.75rem; border:1px solid #dbe5ef; border-radius:11px; background:#f8fafc; }.action-customer strong { color:#0f172a; font-size:.78rem; }.action-customer small { color:#64748b; font-size:.62rem; line-height:1.45; }.action-customer span { width:max-content; padding:.18rem .4rem; border-radius:999px; background:#dcfce7; color:#15803d; font-size:.58rem; font-weight:800; }.action-options { display:grid; gap:.5rem; margin-top:.75rem; }.action-option { display:flex; align-items:center; justify-content:space-between; gap:.75rem; width:100%; padding:.7rem .75rem; border:1px solid #dbe5ef; border-radius:10px; background:#fff; color:#475569; text-align:left; cursor:pointer; }.action-option span { display:grid; gap:.15rem; }.action-option strong { color:#334155; font-size:.72rem; }.action-option small { color:#94a3b8; font-size:.59rem; }.action-option--primary { border-color:#bfdbfe; background:#eff6ff; }.action-option--primary strong,.action-option--primary i { color:#2563eb; }.action-option--danger { border-color:#fecaca; background:#fffafa; }.action-option--danger strong,.action-option--danger i { color:#dc2626; }
.route-prospect-link { cursor:pointer; }.route-prospect-link strong { color:#1d4ed8; }.route-prospect-link strong:hover { text-decoration:underline; }.route-action--checkin { min-height:28px; padding:.25rem .45rem; font-size:.56rem; }
.route-expand-icon { color:#94a3b8; font-size:.62rem; }.inline-item-actions { display:flex; flex-wrap:wrap; width:100%; gap:.35rem; padding-top:.4rem; border-top:1px solid #f1f5f9; }.inline-item-actions .route-action { flex:0 0 auto; }.inline-item-actions .inline-move-select { flex:0 1 115px; }
@media (max-width:480px) { .inline-route-item .route-action,.inline-move-select { flex:1; min-height:34px; max-width:none; } }
@media (max-width:768px) {
  .schedule-card { padding:.85rem; border-radius:14px; }
  .schedule-heading { align-items:center; }
  .schedule-date { font-size:.58rem; }
  .day-actions { gap:.3rem; overflow-x:auto; padding:.1rem 0 .2rem; scrollbar-width:none; }
  .day-actions::-webkit-scrollbar { display:none; }
  .day-action { min-width:52px; min-height:42px; }
  .add-route-control { grid-template-columns:minmax(0,1fr) auto; gap:.4rem; padding:.4rem; }
  .add-route-control select { min-width:0; }
  .add-route-control button { padding-inline:.55rem; }
  .inline-route-item { align-items:flex-start; }
  .inline-route-item .route-prospect-info small { -webkit-line-clamp:2; display:-webkit-box; -webkit-box-orient:vertical; white-space:normal; }
  .action-backdrop { align-items:end; padding:.45rem; }
  .action-modal { width:100%; max-height:90vh; overflow:auto; padding:.9rem; border-radius:18px 18px 10px 10px; }
  .action-option { min-height:54px; padding:.7rem; }
  .action-move-controls { flex-wrap:wrap; }
  .action-move-controls select,.action-move-controls button { min-height:38px; }
}
.action-move-box { display:grid; gap:.45rem; padding:.7rem .75rem; border:1px solid #dbe5ef; border-radius:10px; background:#f8fafc; }.action-move-box > div:first-child { display:grid; gap:.15rem; }.action-move-box strong { color:#334155; font-size:.72rem; }.action-move-box small { color:#94a3b8; font-size:.59rem; }.action-move-controls { display:flex; gap:.4rem; }.action-move-controls select { flex:1; min-width:0; padding:.45rem; border:1px solid #dbe5ef; border-radius:7px; background:#fff; color:#475569; font-size:.65rem; }.action-move-controls button { padding:.45rem .7rem; border:0; border-radius:7px; background:#4f46e5; color:#fff; font-size:.65rem; font-weight:700; cursor:pointer; }.action-move-controls button:disabled { opacity:.45; cursor:not-allowed; }
</style>
