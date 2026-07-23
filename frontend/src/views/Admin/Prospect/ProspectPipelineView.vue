<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import Message from 'primevue/message'
import Select from 'primevue/select'
import Tag from 'primevue/tag'
import { getSalesExecutives } from '../../../api/crm'
import { BOARD_STATUSES, filterProspects } from '../../../domain/pipeline'
import { useCrmStore } from '../../../stores/crm'
import type { ProspectStatus, SalesExecutiveOption } from '../../../types/crm'

const crm = useCrmStore(); const route = useRoute()
const sales = ref<SalesExecutiveOption[]>([]); const salesFilter = ref(''); const industryFilter = ref(''); const statusFilter = ref(''); const error = ref('')
const industries = ['N&B / Kuliner', 'Retail', 'Hospitality', 'Health & Beauty', 'Services', 'Other']
const industryOptions = [{ value: '', label: 'All Business Segments' }, ...industries.map((value) => ({ value, label: value }))]
const statusOptions = [{ value: '', label: 'All Pipeline Statuses' }, ...BOARD_STATUSES.map((value) => ({ value, label: value.replaceAll('_', ' ') }))]
const filtered = computed(() => filterProspects(crm.pipeline.filter((item) => item.status !== 'CONVERTED'), { salesExecutiveId: salesFilter.value, industryGroup: industryFilter.value, status: statusFilter.value, search: String(route.query.search ?? '') }))
const visibleStages = computed(() => statusFilter.value ? BOARD_STATUSES.filter((stage) => stage === statusFilter.value) : BOARD_STATUSES)
const byStage = (stage: ProspectStatus) => filtered.value.filter((item) => item.status === stage)
onMounted(async () => { try { const [result] = await Promise.all([getSalesExecutives(), crm.loadPipeline()]); sales.value = result } catch (caught) { error.value = crm.errorMessage(caught) } })
</script>

<template>
  <section class="admin-page"><div class="page-heading compact-heading"><div><p class="eyebrow">Read-only sales funnel</p><h1>Prospect Pipeline</h1><p class="muted">Administrator visibility across assigned prospects. Pipeline moves remain Sales Executive-only.</p></div><Tag :value="`${filtered.length} records`" severity="secondary" /></div>
    <Message v-if="error" severity="error">{{ error }}</Message>
    <div class="filter-bar"><label><span>Sales Executive</span><Select v-model="salesFilter" :options="[{ id: '', fullName: 'All Sales Executives' }, ...sales]" option-label="fullName" option-value="id" /></label><label><span>Industry Group</span><Select v-model="industryFilter" :options="industryOptions" option-label="label" option-value="value" /></label><label><span>Pipeline Status</span><Select v-model="statusFilter" :options="statusOptions" option-label="label" option-value="value" /></label></div>
    <div class="pipeline-board admin-pipeline-board"><section v-for="stage in visibleStages" :key="stage" class="pipeline-column"><header><strong>{{ stage.replaceAll('_', ' ') }}</strong><span>{{ byStage(stage).length }}</span></header><div class="pipeline-column-body"><article v-for="item in byStage(stage)" :key="item.id" class="kanban-card"><span class="industry-pill">{{ item.industryGroup }}</span><h2>{{ item.placeName }}</h2><p><i class="pi pi-map-marker" /> {{ item.formattedAddress }}</p><dl><div><dt>Sales</dt><dd>{{ item.assignedSalesExecutive }}</dd></div><div><dt>Stage</dt><dd>{{ item.status.replaceAll('_', ' ') }}</dd></div></dl><RouterLink v-if="item.status === 'WON'" :to="`/admin/prospects/${item.id}/review`">Review won prospect <i class="pi pi-arrow-right" /></RouterLink></article><p v-if="!byStage(stage).length" class="pipeline-empty">No prospects</p></div></section></div>
  </section>
</template>

<style scoped>
.pipeline-board { width: 100%; padding: 0.2rem 0 1rem; display: flex; gap: 0.8rem; overflow-x: auto; scroll-snap-type: x proximity; }
.pipeline-column { width: 285px; min-width: 285px; min-height: 360px; overflow: hidden; background: #eef3f9; border: 1px solid var(--border-light); border-radius: var(--radius-lg); scroll-snap-align: start; }
.pipeline-column > header { padding: 0.7rem 0.8rem; display: flex; align-items: center; justify-content: space-between; color: #26344b; background: var(--surface-card); border-bottom: 1px solid var(--border-light); font-size: 0.68rem; }
.pipeline-column > header strong { font-weight: 700; }
.pipeline-column > header span { min-width: 1.4rem; padding: 0.15rem 0.4rem; color: var(--brand-blue); background: var(--brand-blue-bg); border-radius: 1rem; text-align: center; font-size: 0.62rem; font-weight: 700; }
.pipeline-column-body { padding: 0.6rem; display: grid; align-content: start; gap: 0.5rem; }
.kanban-card { padding: 0.7rem; background: var(--surface-card); border: 1px solid var(--border-light); border-radius: var(--radius-md); box-shadow: var(--shadow-xs); transition: box-shadow var(--transition-fast), border-color var(--transition-fast); }
.kanban-card:hover { box-shadow: var(--shadow-sm); border-color: var(--border-default); }
.kanban-card > a { color: inherit; text-decoration: none; }
.kanban-card h2 { margin: 0.35rem 0 0.2rem; font-size: 0.75rem; line-height: 1.35; font-weight: 700; }
.kanban-card p { margin: 0.25rem 0; color: #68758a; font-size: 0.58rem; line-height: 1.45; }
.kanban-card small { color: var(--text-muted); font-size: 0.52rem; }
.kanban-card dl { margin: 0.55rem 0; display: grid; gap: 0.25rem; }
.kanban-card dl div { display: flex; justify-content: space-between; gap: 0.5rem; font-size: 0.55rem; }
.kanban-card dt { color: var(--text-muted); }
.kanban-card dd { margin: 0; color: #344158; font-weight: 700; text-align: right; }
.kanban-card > a:last-child { display: flex; justify-content: space-between; color: var(--brand-blue); font-size: 0.58rem; font-weight: 700; }
.kanban-controls { margin-top: 0.5rem; display: flex; align-items: center; justify-content: space-between; border-top: 1px solid #edf1f5; padding-top: 0.4rem; }
.kanban-controls :deep(.p-button) { padding: 0.3rem; font-size: 0.54rem; }
.pipeline-empty { margin: 0.75rem 0; color: #8a96a8; text-align: center; font-size: 0.6rem; }
.industry-pill { display: inline-block; padding: 0.1rem 0.4rem; color: var(--brand-blue); background: var(--brand-blue-bg); border-radius: 0.3rem; font-size: 0.48rem; font-weight: 600; letter-spacing: 0.02em; }
</style>
