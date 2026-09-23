<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import Button from 'primevue/button'
import Select from 'primevue/select'
import Message from 'primevue/message'
import { getPipeline, getSalesExecutives, assignProspect } from '../../../api/crm'
import type { Prospect, SalesExecutiveOption } from '../../../types/crm'

const prospects = ref<Prospect[]>([])
const sales = ref<SalesExecutiveOption[]>([])
const selected = ref<string[]>([])
const salesExecutiveId = ref('')
const loading = ref(true)
const saving = ref(false)
const error = ref('')
const search = ref('')
const filtered = computed(() => prospects.value.filter((p) => `${p.placeName} ${p.placeCategory} ${p.assignedSalesExecutive}`.toLowerCase().includes(search.value.toLowerCase())))
const allSelected = computed(() => filtered.value.length > 0 && filtered.value.every((p) => selected.value.includes(p.id)))
function toggle(id: string) { selected.value = selected.value.includes(id) ? selected.value.filter((x) => x !== id) : [...selected.value, id] }
function toggleAll() { selected.value = allSelected.value ? selected.value.filter((id) => !filtered.value.some((p) => p.id === id)) : [...new Set([...selected.value, ...filtered.value.map((p) => p.id)])] }
async function load() { loading.value = true; try { [prospects.value, sales.value] = await Promise.all([getPipeline(), getSalesExecutives()]) } catch (e) { error.value = e instanceof Error ? e.message : 'Failed to load assignment data.' } finally { loading.value = false } }
async function save() { if (!salesExecutiveId.value || !selected.value.length) return; saving.value = true; error.value = ''; try { const owner = sales.value.find((s) => s.id === salesExecutiveId.value); await Promise.all(selected.value.map((id) => assignProspect(id, salesExecutiveId.value))); prospects.value = prospects.value.map((p) => selected.value.includes(p.id) ? { ...p, assignedSalesExecutiveId: salesExecutiveId.value, assignedSalesExecutive: owner?.fullName || '' } : p); selected.value = [] } catch (e) { error.value = e instanceof Error ? e.message : 'Failed to assign prospects.' } finally { saving.value = false } }
onMounted(load)
</script>
<template>
  <section class="assignment-page">
    <header class="assignment-header"><div><button type="button" class="back" @click="$router.push('/admin/dashboard')">← Back to Dashboard</button><p class="eyebrow">Prospect Management</p><h1>Prospect Assignment</h1><p>Assign or reassign prospects to a Sales Executive in bulk.</p></div><Button label="Refresh" icon="pi pi-refresh" severity="secondary" outlined :loading="loading" @click="load" /></header>
    <Message v-if="error" severity="error" :closable="false">{{ error }}</Message>
    <section class="assignment-toolbar"><input v-model="search" placeholder="Search prospect name, category, sales executive..." /><Select v-model="salesExecutiveId" :options="sales" optionLabel="fullName" optionValue="id" placeholder="Select Sales Executive" filter /><Button label="Assign Selected" icon="pi pi-check" :disabled="!selected.length || !salesExecutiveId" :loading="saving" @click="save" /></section>
    <section class="assignment-table"><div v-if="loading" class="empty">Loading prospects...</div><div v-else-if="!filtered.length" class="empty">No prospects found.</div><table v-else><thead><tr><th><input type="checkbox" :checked="allSelected" @change="toggleAll" /></th><th>Prospect</th><th>Category</th><th>Current Sales Executive</th><th>Status</th></tr></thead><tbody><tr v-for="p in filtered" :key="p.id"><td><input type="checkbox" :checked="selected.includes(p.id)" @change="toggle(p.id)" /></td><td><strong>{{ p.placeName }}</strong><small>{{ p.formattedAddress }}</small></td><td>{{ p.placeCategory || '—' }}</td><td>{{ p.assignedSalesExecutive || 'Unassigned' }}</td><td>{{ p.status.replaceAll('_', ' ') }}</td></tr></tbody></table></section>
  </section>
</template>
<style scoped>
.assignment-page{min-height:100%;padding:24px;background:#f8fafc;color:#0f172a;font-family:Inter,ui-sans-serif,system-ui,sans-serif}.assignment-header{display:flex;justify-content:space-between;gap:16px;align-items:flex-start;margin-bottom:16px}.back{border:0;background:transparent;color:#64748b;cursor:pointer}.eyebrow{margin:18px 0 4px;color:#94a3b8;font-size:11px;font-weight:700;text-transform:uppercase;letter-spacing:.08em}.assignment-header h1{margin:0;font-size:24px}.assignment-header p:last-child{margin:6px 0 0;color:#64748b;font-size:13px}.assignment-toolbar{display:flex;align-items:center;gap:10px;padding:14px;border:1px solid #e2e8f0;border-radius:14px;background:#fff}.assignment-toolbar input{height:40px;flex:1;min-width:220px;padding:0 12px;border:1px solid #e2e8f0;border-radius:10px;outline:0}.assignment-toolbar :deep(.p-select){width:240px;height:40px}.assignment-table{margin-top:14px;overflow:hidden;border:1px solid #e2e8f0;border-radius:14px;background:#fff}.assignment-table table{width:100%;border-collapse:collapse;font-size:13px}.assignment-table th{padding:12px;text-align:left;background:#f8fafc;color:#64748b;font-size:11px;text-transform:uppercase}.assignment-table td{padding:12px;border-top:1px solid #eef2f7}.assignment-table td strong,.assignment-table td small{display:block}.assignment-table td small{margin-top:3px;color:#64748b;font-size:11px;overflow:hidden;text-overflow:ellipsis;white-space:nowrap}.empty{padding:48px;text-align:center;color:#64748b}@media(max-width:800px){.assignment-toolbar{flex-wrap:wrap}.assignment-toolbar input,.assignment-toolbar :deep(.p-select){width:100%;flex-basis:100%}.assignment-header{flex-direction:column}}
</style>
