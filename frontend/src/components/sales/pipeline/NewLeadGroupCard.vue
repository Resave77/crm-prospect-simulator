<script setup lang="ts">
import { computed, nextTick, ref, watch } from 'vue'
import type { Prospect, ProspectStatus } from '../../../types/crm'
import { stageLabel } from './stageColors'

const props = withDefaults(defineProps<{
  items: Prospect[]
  action?: 'sales' | 'admin'
  stage?: ProspectStatus
}>(), { action: 'sales', stage: 'NEW_LEAD' })

const emit = defineEmits<{
  viewDetail: [item: Prospect]
  moveNext: [item: Prospect]
  viewFeedback: [item: Prospect]
  deleteLost: [item: Prospect]
}>()

type NewLeadView = 'group' | 'sales' | 'pipeline'
const view = ref<NewLeadView>('group')
const selectedSales = ref<string | null>(null)
const cardRef = ref<HTMLElement | null>(null)
const stageName = computed(() => stageLabel(props.stage))

const salesGroups = computed(() => {
  const groups = new Map<string, Prospect[]>()
  for (const item of props.items) {
    const name = item.assignedSalesExecutive?.trim() || 'Unassigned'
    const existing = groups.get(name) ?? []
    existing.push(item)
    groups.set(name, existing)
  }
  return [...groups.entries()].map(([name, items]) => ({ name, items }))
})

const selectedSalesGroup = computed(() =>
  salesGroups.value.find((group) => group.name === selectedSales.value) ?? null,
)

watch(salesGroups, (groups) => {
  if (view.value === 'pipeline' && !groups.some((group) => group.name === selectedSales.value)) {
    selectedSales.value = null
    view.value = 'sales'
  }
})

function openSales() {
  view.value = 'sales'
  focusCard('nearest')
}

function openSalesPipeline(name: string) {
  selectedSales.value = name
  view.value = 'pipeline'
  focusCard('start')
}

function goBack() {
  if (view.value === 'pipeline') {
    selectedSales.value = null
    view.value = 'sales'
    focusCard('nearest')
    return
  }
  view.value = 'group'
  focusCard('nearest')
}

function focusCard(block: ScrollLogicalPosition) {
  void nextTick(() => {
    const card = cardRef.value
    if (!card) return

    const behavior: ScrollBehavior = window.matchMedia?.('(prefers-reduced-motion: reduce)').matches
      ? 'auto'
      : 'smooth'
    card.scrollIntoView({ behavior, block, inline: 'nearest' })
  })
}

function initials(item: Prospect) {
  const names = item.placeName.trim().split(/\s+/).filter(Boolean)
  return names.slice(0, 2).map((name) => name.charAt(0).toUpperCase()).join('') || '?'
}

function salesNumber(name: string) {
  const index = salesGroups.value.findIndex((group) => group.name === name)
  return index >= 0 ? String(index + 1) : '1'
}

</script>

<template>
  <section ref="cardRef" class="new-lead-group" :aria-label="`${stageName} prospects`">
    <div class="new-lead-group-intro">
      <button v-if="view !== 'group'" type="button" class="new-lead-back" @click="goBack">
        <i class="pi pi-arrow-left" />
      </button>
      <div class="new-lead-group-title">
        <span class="new-lead-group-kicker">{{ stageName }}</span>
        <strong>{{ view === 'group' ? 'Sales owners' : view === 'sales' ? 'Sales list' : `Sales ${salesNumber(selectedSales || '')} pipeline` }}</strong>
      </div>
      <span class="new-lead-group-count" :aria-label="`${items.length} ${stageName} prospects`">{{ items.length }}</span>
    </div>

    <template v-if="view === 'group'">
      <button type="button" class="new-lead-summary" :aria-label="`View ${stageName} prospects by sales owner`" @click="openSales">
        <div class="new-lead-summary-avatars" aria-hidden="true">
          <span class="new-lead-summary-avatar">{{ salesGroups.length }}</span>
        </div>
        <span class="new-lead-summary-copy">
          <strong>View by sales owner</strong>
          <small>Open the {{ stageName.toLowerCase() }} list</small>
        </span>
        <span class="new-lead-summary-action">View <i class="pi pi-arrow-right" /></span>
      </button>
    </template>

    <template v-else-if="view === 'sales'">
      <div class="new-lead-simple-list">
        <button
          v-for="group in salesGroups"
          :key="group.name"
          type="button"
          class="new-lead-simple-row"
          :aria-label="`Open pipeline for sales ${group.name}`"
          :title="group.name"
          @click="openSalesPipeline(group.name)"
        >
          <span class="new-lead-sales-initial">{{ salesNumber(group.name) }}</span>
          <span class="new-lead-sales-copy">
            <strong>{{ group.name }}</strong>
            <small>Assigned {{ stageName.toLowerCase() }}</small>
          </span>
          <span class="new-lead-sales-count" :aria-label="`${group.items.length} leads`">{{ group.items.length }}</span>
          <i class="pi pi-arrow-up-right" />
        </button>
      </div>
    </template>

    <template v-else-if="selectedSalesGroup">
      <div class="new-lead-pipeline">
        <div class="new-lead-pipeline-heading">
          <div>
            <span>Customer prospect pipeline</span>
            <strong>Sales {{ salesNumber(selectedSalesGroup.name) }}</strong>
          </div>
          <strong class="new-lead-pipeline-count" :aria-label="`${selectedSalesGroup.items.length} leads`">{{ selectedSalesGroup.items.length }}</strong>
        </div>
        <div class="new-lead-pipeline-list">
          <article
            v-for="item in selectedSalesGroup.items"
            :key="item.id"
            class="new-lead-pipeline-card"
            role="link"
            tabindex="0"
            @click="emit('viewDetail', item)"
            @keydown.enter.self="emit('viewDetail', item)"
            @keydown.space.self.prevent="emit('viewDetail', item)"
          >
            <div class="new-lead-pipeline-card-top">
              <span class="new-lead-pipeline-category">{{ item.placeCategory || item.industryGroup || 'Modern Trade' }}</span>
              <i class="pi pi-chevron-right" />
            </div>
            <strong class="new-lead-pipeline-name">{{ item.placeName }}</strong>
            <p class="new-lead-pipeline-address"><i class="pi pi-map-marker" /> {{ item.formattedAddress || 'Address unavailable' }}</p>
            <div class="new-lead-pipeline-meta">
              <div>
                <span>Sales</span>
                <strong>{{ item.assignedSalesExecutive || selectedSalesGroup.name }}</strong>
              </div>
              <div>
                <span>Stage</span>
                <strong>{{ stageLabel(item.status) }}</strong>
              </div>
            </div>
            <div class="new-lead-pipeline-footer">
              <span><i class="pi pi-comments" /> {{ action === 'admin' ? 'Open ticketing' : 'Open prospect' }}</span>
              <span v-if="action === 'admin' && stage === 'LOST'" class="new-lead-pipeline-actions">
                <button
                  type="button"
                  class="new-lead-action-button new-lead-feedback-button"
                  aria-label="Lihat feedback prospect"
                  title="Lihat feedback"
                  @click.stop="emit('viewFeedback', item)"
                >
                  <i class="pi pi-eye" />
                </button>
                <button
                  type="button"
                  class="new-lead-action-button new-lead-delete-button"
                  aria-label="Hapus prospect"
                  title="Hapus prospect"
                  @click.stop="emit('deleteLost', item)"
                >
                  <i class="pi pi-trash" />
                </button>
              </span>
            </div>
          </article>
        </div>
      </div>
    </template>
  </section>
</template>

<style scoped>
.new-lead-group {
  scroll-margin-block: .75rem;
  overflow: hidden;
  border: 1px solid #e5eaf0;
  border-radius: 9px;
  background: #fff;
  box-shadow: 0 1px 2px rgba(15, 23, 42, .04);
}

.new-lead-group-intro {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: .5rem;
  padding: .68rem .72rem;
  border-bottom: 1px solid #edf0f4;
  background: linear-gradient(135deg, #fff 0%, #fff7f8 100%);
}
.new-lead-group-title { display: grid; min-width: 0; gap: .1rem; }
.new-lead-back {
  display: grid;
  place-items: center;
  width: 1.45rem;
  height: 1.45rem;
  padding: 0;
  border: 0;
  border-radius: .45rem;
  background: #f1f5f9;
  color: #64748b;
  cursor: pointer;
}
.new-lead-back:hover { background: #e2e8f0; color: #334155; }
.new-lead-group-kicker {
  color: #ef4444;
  font-size: .54rem;
  font-weight: 800;
  letter-spacing: .08em;
  text-transform: uppercase;
}
.new-lead-group-intro strong { color: #172033; font-size: .72rem; }
.new-lead-group-count {
  display: grid;
  place-items: center;
  min-width: 1.8rem;
  height: 1.8rem;
  padding: 0 .42rem;
  border: 1px solid #fecdd3;
  border-radius: .58rem;
  background: #fff1f2;
  color: #be123c;
  font-size: .72rem;
  font-weight: 800;
}
.new-lead-overview {
  display: grid;
  grid-template-columns: 1.2fr .8fr;
  gap: .5rem;
  padding: .62rem .72rem .18rem;
  background: #fff;
}
.new-lead-total,
.new-lead-owner-total {
  display: grid;
  gap: .1rem;
  min-width: 0;
  padding: .58rem .62rem;
  border: 1px solid #f1dfe2;
  border-radius: .65rem;
  background: #fff8f9;
}
.new-lead-owner-total { border-color: #e4e8f0; background: #f8fafc; }
.new-lead-total span,
.new-lead-owner-total span {
  color: #8b93a1;
  font-size: .48rem;
  font-weight: 700;
  letter-spacing: .04em;
  text-transform: uppercase;
}
.new-lead-total strong,
.new-lead-owner-total strong {
  color: #be123c;
  font-size: 1.2rem;
  line-height: 1;
}
.new-lead-owner-total strong { color: #3548a8; }
.new-lead-summary {
  display: flex;
  align-items: center;
  width: 100%;
  gap: .55rem;
  padding: .68rem .72rem .76rem;
  border: 0;
  background: #fff;
  color: inherit;
  text-align: left;
  cursor: pointer;
}
.new-lead-summary:hover { background: #fffafb; }
.new-lead-summary-avatars { display: flex; min-width: 2rem; }
.new-lead-summary-avatar {
  display: grid;
  place-items: center;
  width: 1.78rem;
  height: 1.78rem;
  margin-left: -.5rem;
  border: 2px solid #fff;
  border-radius: 999px;
  background: #3548a8;
  color: #fff;
  font-size: .58rem;
  font-weight: 800;
}
.new-lead-summary-avatar:first-child { margin-left: 0; }
.new-lead-summary-avatar:nth-child(2) { background: #ef6b73; }
.new-lead-summary-avatar:nth-child(3) { background: #67a27c; }
.new-lead-summary-avatar { width: 2rem; height: 2rem; font-size: .58rem; }
.new-lead-summary-copy { display: grid; min-width: 0; flex: 1; gap: .08rem; }
.new-lead-summary-copy strong { overflow: hidden; color: #172033; font-size: .68rem; text-overflow: ellipsis; white-space: nowrap; }
.new-lead-summary-copy small { color: #a0a7b4; font-size: .5rem; }
.new-lead-summary-action {
  display: inline-flex;
  align-items: center;
  gap: .24rem;
  flex: 0 0 auto;
  color: #e63946;
  font-size: .53rem;
  font-weight: 800;
}
.new-lead-summary-action i { font-size: .5rem; }
.new-lead-group-list { display: grid; }
.new-lead-simple-list { display: grid; background: #fff; }
.new-lead-simple-row {
  display: flex;
  align-items: center;
  gap: .62rem;
  width: 100%;
  min-height: 48px;
  padding: .58rem .75rem;
  border: 0;
  border-bottom: 1px solid #edf0f4;
  background: #fff;
  color: #64748b;
  font-family: inherit;
  font-size: .62rem;
  text-align: left;
  cursor: pointer;
  transition: background .15s ease, color .15s ease;
}
.new-lead-simple-row:last-child { border-bottom: 0; }
.new-lead-simple-row:nth-child(even) { background: #fbfcfd; }
.new-lead-simple-row:hover { background: #fff5f6; color: #334155; }
.new-lead-simple-row:focus-visible { outline: 2px solid #e63946; outline-offset: -2px; }
.new-lead-sales-initial { display: grid !important; place-items: center; width: 2rem; height: 2rem; flex: 0 0 auto; border-radius: .65rem; background: #3548a8; color: #fff; font-size: .57rem; font-weight: 800; letter-spacing: .02em; }
.new-lead-simple-row:nth-child(2n) .new-lead-sales-initial { background: #e86b73; }
.new-lead-simple-row:nth-child(3n) .new-lead-sales-initial { background: #6a9a7b; }
.new-lead-sales-copy { display: grid !important; min-width: 0; flex: 1; gap: .1rem; }
.new-lead-sales-copy strong { overflow: hidden; color: #172033; font-size: .66rem; font-weight: 800; text-overflow: ellipsis; white-space: nowrap; }
.new-lead-sales-copy small { color: #9aa3b1; font-size: .51rem; }
.new-lead-sales-count,
.new-lead-pipeline-count {
  display: grid;
  place-items: center;
  min-width: 1.8rem;
  height: 1.8rem;
  padding: 0 .4rem;
  border: 1px solid #fecdd3;
  border-radius: .55rem;
  background: #fff1f2;
  color: #be123c;
  font-size: .72rem;
  font-weight: 800;
  line-height: 1;
}
.new-lead-simple-row > i { flex: 0 0 auto; color: #a3acb9; font-size: .64rem; }
.new-lead-pipeline { background: #fff; }
.new-lead-pipeline-heading { display: flex; align-items: center; justify-content: space-between; gap: .5rem; padding: .68rem .75rem; border-bottom: 1px solid #edf0f4; background: #fffafb; }
.new-lead-pipeline-heading > div { display: grid; min-width: 0; gap: .12rem; }
.new-lead-pipeline-heading span { color: #94a3b8; font-size: .51rem; font-weight: 700; letter-spacing: .05em; text-transform: uppercase; }
.new-lead-pipeline-heading > div strong { overflow: hidden; color: #334155; font-size: .62rem; text-overflow: ellipsis; white-space: nowrap; }
.new-lead-pipeline-count { flex: 0 0 auto; background: #fff1f2 !important; color: #be123c !important; font-size: .72rem !important; }
.new-lead-pipeline-list { display: grid; gap: .48rem; padding: .55rem; background: #f8fafc; }
.new-lead-pipeline-card { position: relative; min-width: 0; padding: .68rem; border: 1px solid #e5eaf0; border-radius: .58rem; background: #fff; box-shadow: 0 1px 2px rgba(15, 23, 42, .04); cursor: pointer; transition: border-color .15s ease, box-shadow .15s ease, transform .15s ease; }
.new-lead-pipeline-card::before { position: absolute; top: -1px; right: -.5px; left: -.5px; height: 3px; border-radius: .58rem .58rem 0 0; background: #ef4444; content: ''; }
.new-lead-pipeline-card:hover { border-color: #f3a0a7; box-shadow: 0 5px 14px rgba(230, 57, 70, .1); transform: translateY(-1px); }
.new-lead-pipeline-card:focus-visible { outline: 2px solid #e63946; outline-offset: 2px; }
.new-lead-pipeline-card-top { display: flex; align-items: center; justify-content: space-between; gap: .5rem; margin-bottom: .44rem; }
.new-lead-pipeline-card-top > i { color: #c2cad4; font-size: .55rem; }
.new-lead-pipeline-category { max-width: 75%; overflow: hidden; padding: .16rem .38rem; border-radius: .28rem; background: #fff0f1; color: #e63946; font-size: .49rem; font-weight: 800; text-overflow: ellipsis; white-space: nowrap; }
.new-lead-pipeline-name { display: block; overflow: hidden; color: #172033; font-size: .72rem; font-weight: 800; line-height: 1.3; text-overflow: ellipsis; white-space: nowrap; }
.new-lead-pipeline-address { display: -webkit-box; overflow: hidden; margin: .34rem 0 .55rem; color: #7c8795; font-size: .52rem; line-height: 1.4; -webkit-box-orient: vertical; -webkit-line-clamp: 2; line-clamp: 2; }
.new-lead-pipeline-address i { margin-right: .18rem; color: #a9b2be; font-size: .48rem; }
.new-lead-pipeline-meta { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: .5rem; padding-top: .5rem; border-top: 1px solid #f0f1f4; }
.new-lead-pipeline-meta div { display: grid; min-width: 0; gap: .12rem; }
.new-lead-pipeline-meta span { color: #a0a7b4; font-size: .48rem; font-weight: 600; }
.new-lead-pipeline-meta strong { overflow: hidden; color: #334155; font-size: .54rem; font-weight: 750; text-overflow: ellipsis; white-space: nowrap; }
.new-lead-pipeline-footer { display: flex; justify-content: flex-end; margin-top: .55rem; padding-top: .46rem; border-top: 1px solid #f0f1f4; }
.new-lead-pipeline-footer span { display: inline-flex; align-items: center; gap: .24rem; color: #e63946; font-size: .51rem; font-weight: 800; }
.new-lead-pipeline-footer i { font-size: .5rem; }
.new-lead-pipeline-actions { margin-left: auto; }
.new-lead-action-button { display: grid; place-items: center; width: 1.5rem; height: 1.5rem; padding: 0; border: 1px solid #fecdd3; border-radius: .42rem; background: #fff7f8; color: #e63946; cursor: pointer; }
.new-lead-action-button:hover { background: #fff0f1; border-color: #f3a0a7; }
.new-lead-action-button:focus-visible { outline: 2px solid #e63946; outline-offset: 1px; }
.new-lead-action-button i { font-size: .56rem; }
.new-lead-delete-button { color: #b4232f; }
.new-lead-pipeline-row { display: flex; align-items: center; width: 100%; min-width: 0; gap: .5rem; padding: .65rem .7rem; border: 0; border-bottom: 1px solid #edf0f4; background: #fff; color: #334155; text-align: left; cursor: pointer; }
.new-lead-pipeline-row:last-child { border-bottom: 0; }
.new-lead-pipeline-row:hover { background: #fffafb; }
.new-lead-pipeline-row:focus-visible { outline: 2px solid #e63946; outline-offset: -2px; }
.new-lead-pipeline-copy { display: grid; min-width: 0; flex: 1; gap: .12rem; }
.new-lead-pipeline-copy strong { overflow: hidden; color: #172033; font-size: .62rem; text-overflow: ellipsis; white-space: nowrap; }
.new-lead-pipeline-copy small { overflow: hidden; color: #94a3b8; font-size: .49rem; text-overflow: ellipsis; white-space: nowrap; }
.new-lead-pipeline-stage { flex: 0 0 auto; padding: .2rem .34rem; border: 1px solid #fecdd3; border-radius: 999px; background: #fff1f2; color: #be123c; font-size: .46rem; font-weight: 800; text-transform: uppercase; }
.new-lead-pipeline-row > i { flex: 0 0 auto; color: #94a3b8; font-size: .55rem; }
.new-lead-tree { background: #fff; }
.new-lead-tree-root,
.new-lead-tree-row {
  display: flex;
  align-items: center;
  min-width: 0;
}
.new-lead-tree-root {
  gap: .38rem;
  padding: .58rem .62rem;
  border-bottom: 1px solid #edf0f4;
}
.new-lead-tree-root-index,
.new-lead-tree-index {
  width: 1.1rem;
  flex: 0 0 1.1rem;
  color: #e63946;
  font-size: .58rem;
  font-weight: 800;
  text-align: center;
}
.new-lead-tree-root-toggle,
.new-lead-tree-toggle {
  display: grid;
  place-items: center;
  width: 1.18rem;
  height: 1.18rem;
  flex: 0 0 auto;
  border: 1px solid #dbe3ec;
  border-radius: .35rem;
  background: #fff;
  color: #64748b;
  font-size: .48rem;
}
.new-lead-tree-root > div { display: grid; min-width: 0; gap: .04rem; }
.new-lead-tree-root strong { overflow: hidden; color: #e63946; font-size: .62rem; text-overflow: ellipsis; white-space: nowrap; }
.new-lead-tree-root small { color: #94a3b8; font-size: .48rem; }
.new-lead-tree-children { position: relative; padding: .18rem 0 .18rem 0; }
.new-lead-tree-children::before {
  position: absolute;
  top: 0;
  bottom: 1.25rem;
  left: 1.68rem;
  width: 1px;
  background: #dbe3ec;
  content: '';
}
.new-lead-tree-row {
  position: relative;
  width: 100%;
  gap: .38rem;
  padding: .58rem .62rem;
  border: 0;
  background: #fff;
  color: #172033;
  text-align: left;
  cursor: pointer;
}
.new-lead-tree-row::before {
  position: absolute;
  top: 50%;
  left: 1.68rem;
  width: .65rem;
  height: 1px;
  background: #dbe3ec;
  content: '';
}
.new-lead-tree-row:hover { background: #fffafb; }
.new-lead-tree-row:focus-visible { outline: 2px solid #e63946; outline-offset: -2px; }
.new-lead-tree-row .new-lead-tree-index { color: #9b5b20; }
.new-lead-tree-branch { position: relative; z-index: 1; display: grid; place-items: center; width: 1.18rem; flex: 0 0 auto; background: #fff; }
.new-lead-tree-copy { display: grid; min-width: 0; flex: 1; gap: .04rem; padding-left: .08rem; }
.new-lead-tree-copy strong { overflow: hidden; color: #334155; font-size: .62rem; font-weight: 750; text-overflow: ellipsis; white-space: nowrap; }
.new-lead-tree-copy small { color: #94a3b8; font-size: .48rem; }
.new-lead-name-row {
  display: flex;
  align-items: center;
  width: 100%;
  min-width: 0;
  gap: .52rem;
  padding: .7rem .62rem;
  border: 0;
  border-bottom: 1px solid #edf0f4;
  background: #fff;
  color: #172033;
  text-align: left;
  cursor: pointer;
}
.new-lead-name-row:last-child { border-bottom: 0; }
.new-lead-name-row:hover { background: #fffafb; }
.new-lead-name-row:focus-visible { outline: 2px solid #e63946; outline-offset: -2px; }
.new-lead-name-row strong { min-width: 0; flex: 1; overflow: hidden; font-size: .67rem; text-overflow: ellipsis; white-space: nowrap; }
.new-lead-name-row > i { flex: 0 0 auto; color: #94a3b8; font-size: .6rem; }
.new-lead-name-avatar { display: grid; place-items: center; width: 1.65rem; height: 1.65rem; flex: 0 0 auto; border-radius: 999px; background: #3548a8; color: #fff; font-size: .56rem; font-weight: 800; }
.new-lead-row {
  min-width: 0;
  padding: .62rem;
  border-bottom: 1px solid #edf0f4;
  background: #fff;
  cursor: pointer;
  transition: background .15s ease;
}
.new-lead-row:last-child { border-bottom: 0; }
.new-lead-row:hover { background: #fffafb; }
.new-lead-row:focus-visible { outline: 2px solid #e63946; outline-offset: -2px; }
.new-lead-row-expanded { background: #fffafb; }
.new-lead-row-preview { display: flex; align-items: center; min-width: 0; gap: .5rem; }
.new-lead-avatar-stack { display: flex; min-width: 2rem; }
.new-lead-avatar {
  display: grid;
  place-items: center;
  width: 1.78rem;
  height: 1.78rem;
  border: 2px solid #fff;
  border-radius: 999px;
  color: #fff;
  font-size: .58rem;
  font-weight: 800;
}
.new-lead-avatar-primary { z-index: 1; background: #3548a8; }
.new-lead-avatar-secondary { margin-left: -.55rem; background: #ef6b73; }
.new-lead-name-block { display: grid; min-width: 0; flex: 1; gap: .08rem; }
.new-lead-name { overflow: hidden; color: #172033; font-size: .7rem; font-weight: 800; text-overflow: ellipsis; white-space: nowrap; }
.new-lead-helper { color: #a0a7b4; font-size: .5rem; }
.new-lead-activity {
  display: grid;
  place-items: center;
  width: 1.5rem;
  height: 1.5rem;
  flex: 0 0 auto;
  border-radius: .5rem;
  background: #fff0f1;
  color: #e63946;
  font-size: .65rem;
}
.new-lead-menu {
  display: grid;
  place-items: center;
  width: 1.3rem;
  height: 1.4rem;
  padding: 0 0 .3rem;
  border: 0;
  border-radius: .4rem;
  background: transparent;
  color: #94a3b8;
  font-family: inherit;
  font-size: 1rem;
  line-height: 1;
  cursor: pointer;
}
.new-lead-menu:hover { background: #f1f5f9; color: #475569; }
.new-lead-row-meta {
  display: flex;
  justify-content: space-between;
  gap: .5rem;
  margin-top: .55rem;
  padding-top: .48rem;
  border-top: 1px solid #f0f1f4;
  color: #8b93a1;
  font-size: .5rem;
  font-weight: 600;
}
.new-lead-row-meta span { display: inline-flex; align-items: center; gap: .22rem; }
.new-lead-row-meta i { color: #b7beca; font-size: .5rem; }
.new-lead-row-meta time { white-space: nowrap; }
.new-lead-details { margin: .58rem -.1rem -.1rem; padding: .62rem .56rem .08rem; border-top: 1px solid #edf0f4; }
.new-lead-details-heading { display: flex; align-items: center; justify-content: space-between; gap: .5rem; margin-bottom: .55rem; }
.new-lead-details-heading > div { display: grid; min-width: 0; gap: .1rem; }
.new-lead-details-heading span { color: #94a3b8; font-size: .49rem; font-weight: 700; letter-spacing: .05em; text-transform: uppercase; }
.new-lead-details-heading strong { overflow: hidden; color: #1e293b; font-size: .62rem; text-overflow: ellipsis; white-space: nowrap; }
.new-lead-details-heading b { flex: 0 0 auto; padding: .17rem .34rem; border: 1px solid #fecdd3; border-radius: 999px; background: #fff1f2; color: #be123c; font-size: .46rem; text-transform: uppercase; }
.new-lead-detail-grid { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: .42rem; }
.new-lead-detail-grid > div { display: grid; min-width: 0; gap: .16rem; }
.new-lead-detail-wide { grid-column: 1 / -1; }
.new-lead-detail-grid span { display: flex; align-items: center; gap: .22rem; color: #9aa3b1; font-size: .48rem; font-weight: 600; }
.new-lead-detail-grid span i { font-size: .47rem; }
.new-lead-detail-grid strong { overflow: hidden; color: #475569; font-size: .55rem; font-weight: 650; line-height: 1.35; text-overflow: ellipsis; white-space: nowrap; }
.new-lead-notes { margin: .52rem 0 0; padding: .42rem .5rem; border-radius: .42rem; background: #f1f5f9; color: #64748b; font-size: .54rem; line-height: 1.4; }
.new-lead-actions { display: flex; gap: .38rem; margin-top: .58rem; }
.new-lead-actions button { flex: 1; min-width: 0; min-height: 1.75rem; padding: .3rem .45rem; border-radius: .42rem; font-family: inherit; font-size: .52rem; font-weight: 700; cursor: pointer; }
.new-lead-open { border: 1px solid #e2e8f0; background: #fff; color: #475569; }
.new-lead-open:hover { background: #f8fafc; }
.new-lead-progress { border: 1px solid #e63946; background: #e63946; color: #fff; }
.new-lead-progress:hover { background: #d62839; }
</style>
