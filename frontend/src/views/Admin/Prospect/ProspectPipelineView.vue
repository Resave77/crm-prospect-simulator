<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import Button from 'primevue/button'
import Message from 'primevue/message'
import Select from 'primevue/select'
import Tag from 'primevue/tag'
import { getSalesExecutives } from '../../../api/crm'
import { BOARD_STATUSES, filterProspects } from '../../../domain/pipeline'
import { useCrmStore } from '../../../stores/crm'
import type { ProspectStatus, SalesExecutiveOption } from '../../../types/crm'

const crm = useCrmStore()
const router = useRouter()
const route = useRoute()
const sales = ref<SalesExecutiveOption[]>([])
const salesFilter = ref('')
const industryFilter = ref('')
const statusFilter = ref('')
const error = ref('')

const industries = ['N&B / Kuliner', 'Retail', 'Hospitality', 'Health & Beauty', 'Services', 'Other']

const industryOptions = [
  { value: '', label: 'All Business Segments' },
  ...industries.map((value) => ({ value, label: value })),
]

const statusOptions = [
  { value: '', label: 'All Pipeline Statuses' },
  ...BOARD_STATUSES.map((value) => ({
    value,
    label: value.replaceAll('_', ' '),
  })),
]

const filtered = computed(() =>
  filterProspects(
    crm.pipeline.filter((item) => item.status !== 'CONVERTED'),
    {
      salesExecutiveId: salesFilter.value,
      industryGroup: industryFilter.value,
      status: statusFilter.value,
      search: String(route.query.search ?? ''),
    },
  ),
)

const visibleStages = computed(() =>
  statusFilter.value
    ? BOARD_STATUSES.filter((stage) => stage === statusFilter.value)
    : BOARD_STATUSES,
)

const byStage = (stage: ProspectStatus) =>
  filtered.value.filter((item) => item.status === stage)

const activeSalesCount = computed(
  () =>
    new Set(
      filtered.value
        .map((item) => item.assignedSalesExecutive)
        .filter(Boolean),
    ).size,
)

const activeIndustryCount = computed(
  () =>
    new Set(
      filtered.value
        .map((item) => item.industryGroup)
        .filter(Boolean),
    ).size,
)

function openTicketing(id: string) {
  router.push(`/admin/prospects/${id}/review`)
}

function resetFilters() {
  salesFilter.value = ''
  industryFilter.value = ''
  statusFilter.value = ''
}

onMounted(async () => {
  try {
    const [result] = await Promise.all([
      getSalesExecutives(),
      crm.loadPipeline(),
    ])
    sales.value = result
  } catch (caught) {
    error.value = crm.errorMessage(caught)
  }
})
</script>

<template>
  <section class="pipeline-page">
    <header class="workspace-header">
      <div class="workspace-heading">
        <Button
          icon="pi pi-arrow-left"
          severity="secondary"
          text
          rounded
          class="back-button"
          @click="router.back()"
          title="Back"
        />

        <div class="title-block">
          <span class="eyebrow">Prospect Management</span>
          <h1>Prospect Pipeline</h1>
          <p>Monitor funnel progression and open ticketing for each prospect.</p>
        </div>
      </div>

      <div class="summary-strip">
        <div class="summary-item">
          <span>Total Records</span>
          <strong>{{ filtered.length }}</strong>
        </div>
        <div class="summary-item">
          <span>Visible Stages</span>
          <strong>{{ visibleStages.length }}</strong>
        </div>
        <div class="summary-item">
          <span>Sales</span>
          <strong>{{ activeSalesCount }}</strong>
        </div>
        <div class="summary-item">
          <span>Segments</span>
          <strong>{{ activeIndustryCount }}</strong>
        </div>
      </div>

      <Tag
        :value="`${filtered.length} records`"
        severity="secondary"
        class="record-tag"
      />
    </header>

    <Message v-if="error" severity="error" class="page-message">
      {{ error }}
    </Message>

    <section class="control-panel">
      <div class="filter-field">
        <label>Sales Executive</label>
        <Select
          v-model="salesFilter"
          :options="[{ id: '', fullName: 'All Sales Executives' }, ...sales]"
          option-label="fullName"
          option-value="id"
          placeholder="All Sales Executives"
        />
      </div>

      <div class="filter-field">
        <label>Industry Group</label>
        <Select
          v-model="industryFilter"
          :options="industryOptions"
          option-label="label"
          option-value="value"
          placeholder="All Business Segments"
        />
      </div>

      <div class="filter-field">
        <label>Pipeline Status</label>
        <Select
          v-model="statusFilter"
          :options="statusOptions"
          option-label="label"
          option-value="value"
          placeholder="All Pipeline Statuses"
        />
      </div>

      <Button
        label="Reset"
        icon="pi pi-refresh"
        severity="secondary"
        outlined
        size="small"
        class="reset-button"
        @click="resetFilters"
      />
    </section>

    <section class="board-shell">
      <div class="board-heading">
        <div>
          <strong>Sales Funnel Board</strong>
          <span>Click any prospect card to open ticketing and comments.</span>
        </div>
        <span class="board-count">{{ filtered.length }} prospects</span>
      </div>

      <div class="pipeline-board admin-pipeline-board">
        <section
          v-for="stage in visibleStages"
          :key="stage"
          class="pipeline-column"
        >
          <header class="column-header">
            <div>
              <span class="stage-dot" />
              <strong>{{ stage.replaceAll('_', ' ') }}</strong>
            </div>
            <span class="stage-count">{{ byStage(stage).length }}</span>
          </header>

          <div class="pipeline-column-body">
            <article
              v-for="item in byStage(stage)"
              :key="item.id"
              class="kanban-card"
              role="button"
              tabindex="0"
              @click="openTicketing(item.id)"
              @keydown.enter="openTicketing(item.id)"
            >
              <div class="card-top-row">
                <span class="industry-pill">{{ item.industryGroup }}</span>
                <i class="pi pi-chevron-right card-chevron" />
              </div>

              <h2>{{ item.placeName }}</h2>

              <p class="address-row">
                <i class="pi pi-map-marker" />
                <span>{{ item.formattedAddress }}</span>
              </p>

              <dl>
                <div>
                  <dt>Sales</dt>
                  <dd>{{ item.assignedSalesExecutive || 'Unassigned' }}</dd>
                </div>
                <div>
                  <dt>Stage</dt>
                  <dd>{{ item.status.replaceAll('_', ' ') }}</dd>
                </div>
              </dl>

              <div class="kanban-card-footer">
                <span class="kanban-ticketing-link">
                  <i class="pi pi-comments" />
                  Open ticketing
                </span>
              </div>
            </article>

            <div v-if="!byStage(stage).length" class="pipeline-empty">
              <i class="pi pi-inbox" />
              <strong>No prospects</strong>
              <span>No records match this stage.</span>
            </div>
          </div>
        </section>
      </div>
    </section>
  </section>
</template>

<style scoped>
.pipeline-page {
  box-sizing: border-box;
  display: flex;
  min-width: 0;
  min-height: calc(100dvh - 4rem);
  flex-direction: column;
  gap: 0.75rem;
  padding: 0.85rem 1rem 1rem;
  overflow-x: hidden;
  background: #f8fafc;
}

.workspace-header {
  display: grid;
  grid-template-columns: minmax(280px, 1fr) auto auto;
  align-items: center;
  gap: 0.9rem;
  padding: 0.72rem 0.85rem;
  border: 1px solid #e5eaf0;
  border-radius: 12px;
  background: #fff;
  box-shadow: 0 1px 2px rgba(15, 23, 42, 0.03);
}

.workspace-heading {
  display: flex;
  min-width: 0;
  align-items: center;
  gap: 0.65rem;
}

.back-button {
  flex: 0 0 auto;
}

.title-block {
  min-width: 0;
  display: grid;
  gap: 0.05rem;
}

.eyebrow {
  color: #64748b;
  font-size: 0.56rem;
  font-weight: 800;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

.title-block h1 {
  margin: 0;
  color: #0f172a;
  font-size: 1.08rem;
  line-height: 1.2;
  letter-spacing: -0.02em;
}

.title-block p {
  margin: 0;
  overflow: hidden;
  color: #94a3b8;
  font-size: 0.67rem;
  line-height: 1.4;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.summary-strip {
  display: grid;
  grid-template-columns: repeat(4, minmax(72px, auto));
  overflow: hidden;
  border: 1px solid #e5eaf0;
  border-radius: 9px;
  background: #f8fafc;
}

.summary-item {
  display: grid;
  gap: 0.04rem;
  min-width: 74px;
  padding: 0.42rem 0.58rem;
  border-right: 1px solid #e5eaf0;
  text-align: center;
}

.summary-item:last-child {
  border-right: 0;
}

.summary-item span {
  color: #94a3b8;
  font-size: 0.5rem;
  font-weight: 800;
  text-transform: uppercase;
}

.summary-item strong {
  color: #0f172a;
  font-size: 0.8rem;
}

.record-tag {
  white-space: nowrap;
}

.page-message {
  margin: 0;
}

.control-panel {
  display: grid;
  grid-template-columns: repeat(3, minmax(170px, 1fr)) auto;
  align-items: end;
  gap: 0.65rem;
  padding: 0.65rem;
  border: 1px solid #e5eaf0;
  border-radius: 10px;
  background: #fff;
  box-shadow: 0 1px 2px rgba(15, 23, 42, 0.03);
}

.filter-field {
  display: grid;
  min-width: 0;
  gap: 0.24rem;
}

.filter-field label {
  color: #64748b;
  font-size: 0.55rem;
  font-weight: 800;
  letter-spacing: 0.05em;
  text-transform: uppercase;
}

.filter-field :deep(.p-select) {
  width: 100%;
  min-width: 0;
  height: 38px;
  border-radius: 8px;
}

.filter-field :deep(.p-select-label) {
  padding-block: 0.5rem;
  overflow: hidden;
  font-size: 0.72rem;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.reset-button {
  height: 38px;
}

.board-shell {
  min-width: 0;
  overflow: hidden;
  border: 1px solid #e5eaf0;
  border-radius: 10px;
  background: #fff;
  box-shadow: 0 1px 2px rgba(15, 23, 42, 0.03);
}

.board-heading {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  padding: 0.65rem 0.8rem;
  border-bottom: 1px solid #e5eaf0;
  background: #f8fafc;
}

.board-heading > div {
  display: grid;
  gap: 0.08rem;
}

.board-heading strong {
  color: #0f172a;
  font-size: 0.8rem;
}

.board-heading span {
  color: #94a3b8;
  font-size: 0.64rem;
}

.board-count {
  padding: 0.18rem 0.5rem;
  border-radius: 999px;
  background: #e2e8f0;
  color: #475569 !important;
  font-size: 0.58rem !important;
  font-weight: 800;
}

.pipeline-board {
  width: 100%;
  min-width: 0;
  padding: 0.65rem;
  display: grid;
  grid-auto-columns: minmax(260px, 1fr);
  grid-auto-flow: column;
  gap: 0.65rem;
  overflow-x: auto;
  scroll-snap-type: x proximity;
  scrollbar-width: thin;
  background: #f8fafc;
}

.pipeline-column {
  width: auto;
  min-width: 260px;
  min-height: 420px;
  overflow: hidden;
  border: 1px solid #e2e8f0;
  border-radius: 10px;
  background: #eef3f9;
  scroll-snap-align: start;
}

.column-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.65rem;
  padding: 0.65rem 0.72rem;
  border-bottom: 1px solid #e2e8f0;
  background: #fff;
}

.column-header > div {
  display: flex;
  align-items: center;
  gap: 0.4rem;
}

.column-header strong {
  color: #334155;
  font-size: 0.64rem;
  font-weight: 800;
  letter-spacing: 0.04em;
  text-transform: uppercase;
}

.stage-dot {
  width: 0.42rem;
  height: 0.42rem;
  border-radius: 999px;
  background: #d14350;
  box-shadow: 0 0 0 3px #ffd9dd;
}

.stage-count {
  min-width: 1.5rem;
  padding: 0.12rem 0.38rem;
  border-radius: 999px;
  background: #fff1f2;
  color: #d14350;
  font-size: 0.6rem;
  font-weight: 800;
  text-align: center;
}

.pipeline-column-body {
  min-height: 360px;
  padding: 0.55rem;
  display: grid;
  align-content: start;
  gap: 0.48rem;
}

.kanban-card {
  padding: 0.66rem;
  border: 1px solid #e5eaf0;
  border-radius: 9px;
  background: #fff;
  box-shadow: 0 1px 2px rgba(15, 23, 42, 0.04);
  cursor: pointer;
  transition:
    transform 150ms ease,
    border-color 150ms ease,
    box-shadow 150ms ease;
}

.kanban-card:hover {
  transform: translateY(-1px);
  border-color: #e9909a;
  box-shadow: 0 6px 18px -8px rgba(209, 67, 80, 0.25);
}

.kanban-card:focus-visible {
  outline: 2px solid #d14350;
  outline-offset: 2px;
}

.card-top-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.5rem;
}

.industry-pill {
  display: inline-block;
  max-width: calc(100% - 24px);
  overflow: hidden;
  padding: 0.13rem 0.4rem;
  border-radius: 5px;
  background: #fff1f2;
  color: #d14350;
  font-size: 0.5rem;
  font-weight: 700;
  letter-spacing: 0.02em;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.card-chevron {
  color: #cbd5e1;
  font-size: 0.58rem;
  transition: transform 150ms ease, color 150ms ease;
}

.kanban-card:hover .card-chevron {
  transform: translateX(2px);
  color: #d14350;
}

.kanban-card h2 {
  margin: 0.4rem 0 0.22rem;
  overflow: hidden;
  color: #0f172a;
  font-size: 0.76rem;
  font-weight: 750;
  line-height: 1.35;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.address-row {
  display: grid;
  grid-template-columns: 14px minmax(0, 1fr);
  gap: 0.28rem;
  margin: 0.24rem 0;
  color: #64748b;
  font-size: 0.58rem;
  line-height: 1.45;
}

.address-row i {
  margin-top: 0.1rem;
  color: #94a3b8;
  font-size: 0.58rem;
}

.address-row span {
  display: -webkit-box;
  overflow: hidden;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
}

.kanban-card dl {
  margin: 0.5rem 0;
  display: grid;
  gap: 0.28rem;
}

.kanban-card dl div {
  display: grid;
  grid-template-columns: 54px minmax(0, 1fr);
  gap: 0.45rem;
  font-size: 0.56rem;
}

.kanban-card dt {
  color: #94a3b8;
}

.kanban-card dd {
  margin: 0;
  overflow: hidden;
  color: #334155;
  font-weight: 750;
  text-align: right;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.kanban-card-footer {
  display: flex;
  justify-content: flex-end;
  margin-top: 0.3rem;
  padding-top: 0.4rem;
  border-top: 1px solid #edf1f5;
}

.kanban-ticketing-link {
  display: inline-flex;
  align-items: center;
  gap: 0.3rem;
  color: #d14350;
  font-size: 0.56rem;
  font-weight: 800;
}

.kanban-ticketing-link i {
  font-size: 0.6rem;
}

.pipeline-empty {
  min-height: 130px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 0.28rem;
  margin: 0;
  padding: 1rem;
  color: #94a3b8;
  text-align: center;
}

.pipeline-empty i {
  font-size: 1rem;
}

.pipeline-empty strong {
  color: #64748b;
  font-size: 0.66rem;
}

.pipeline-empty span {
  font-size: 0.56rem;
}

@media (min-width: 1500px) {
  .pipeline-board {
    grid-auto-columns: minmax(280px, 1fr);
  }
}

@media (max-width: 1100px) {
  .workspace-header {
    grid-template-columns: minmax(260px, 1fr) auto;
  }

  .summary-strip {
    grid-column: 1 / -1;
    grid-row: 2;
  }

  .control-panel {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .reset-button {
    width: 100%;
  }
}

@media (max-width: 768px) {
  .pipeline-page {
    min-height: auto;
    padding: 0.7rem;
    overflow: visible;
  }

  .workspace-header {
    grid-template-columns: 1fr;
    align-items: stretch;
  }

  .title-block p {
    white-space: normal;
  }

  .summary-strip {
    grid-column: auto;
    grid-row: auto;
    grid-template-columns: repeat(2, 1fr);
  }

  .summary-item:nth-child(2) {
    border-right: 0;
  }

  .summary-item:nth-child(-n + 2) {
    border-bottom: 1px solid #e5eaf0;
  }

  .record-tag {
    justify-self: start;
  }

  .control-panel {
    grid-template-columns: 1fr;
  }

  .board-heading {
    align-items: flex-start;
    flex-direction: column;
  }

  .pipeline-board {
    grid-auto-columns: minmax(84vw, 1fr);
    padding: 0.55rem;
  }

  .pipeline-column {
    min-width: 84vw;
  }
}

@media (max-width: 480px) {
  .pipeline-page {
    padding: 0.55rem;
  }

  .workspace-heading {
    align-items: flex-start;
  }

  .pipeline-board {
    grid-auto-columns: minmax(88vw, 1fr);
  }

  .pipeline-column {
    min-width: 88vw;
  }
}
</style>
