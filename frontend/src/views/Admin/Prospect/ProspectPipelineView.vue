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
