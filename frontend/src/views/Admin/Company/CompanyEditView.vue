<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import Button from 'primevue/button'
import InputText from 'primevue/inputtext'
import Textarea from 'primevue/textarea'
import Message from 'primevue/message'
import { getParentCompany, updateParentCompany } from '../../../api/crm'

const route = useRoute()
const router = useRouter()
const id = computed(() => String(route.params.id ?? ''))
const loading = ref(true)
const saving = ref(false)
const error = ref('')
const saved = ref(false)
const form = reactive({ name: '', npwpName: '', npwpAddress: '', npwpNumber: '', termOfPayment: '', notes: '' })
const valid = computed(() => Boolean(form.name.trim()))

function errorMessage(caught: unknown, fallback: string) {
  const response = (caught as { response?: { data?: { error?: { message?: string }; message?: string } } }).response
  return response?.data?.error?.message || response?.data?.message || (caught instanceof Error ? caught.message : fallback)
}

async function load() {
  loading.value = true
  try {
    const company = await getParentCompany(id.value)
    form.name = company.name || ''
    form.npwpName = company.npwpName || ''
    form.npwpAddress = company.npwpAddress || ''
    form.npwpNumber = company.npwpNumber || ''
    form.termOfPayment = company.termOfPayment || ''
    form.notes = company.notes || ''
  } catch (caught) {
    error.value = errorMessage(caught, 'Failed to load company data.')
  } finally { loading.value = false }
}

async function save() {
  if (!valid.value) return
  saving.value = true
  error.value = ''
  try {
    await updateParentCompany(id.value, { name: form.name.trim(), npwpName: form.npwpName.trim(), npwpAddress: form.npwpAddress.trim(), npwpNumber: form.npwpNumber.trim(), termOfPayment: form.termOfPayment.trim(), notes: form.notes.trim() })
    saved.value = true
    setTimeout(() => router.push(`/admin/companies/${id.value}`), 500)
  } catch (caught) {
    error.value = errorMessage(caught, 'Failed to update company.')
  } finally { saving.value = false }
}

onMounted(load)
</script>

<template>
  <section class="company-edit-page">
    <header class="company-edit-header">
      <Button label="Back to Company" icon="pi pi-arrow-left" text @click="router.push(`/admin/companies/${id}`)" />
      <div><h1>Edit Company</h1><p>Customer Management &gt; Company &gt; Edit</p></div>
      <div class="header-actions"><Button label="Cancel" severity="secondary" outlined @click="router.push(`/admin/companies/${id}`)" /><Button label="Save Changes" icon="pi pi-pencil" :loading="saving" :disabled="!valid" @click="save" /></div>
    </header>
    <main class="company-edit-content">
      <Message v-if="error" severity="error" :closable="false">{{ error }}</Message>
      <Message v-if="saved" severity="success" :closable="false">Company updated successfully.</Message>
      <section v-if="!loading && !error" class="company-edit-card">
        <div class="card-heading"><h2>Company Information</h2><p>Update the parent company information used by its customer sites.</p></div>
        <div class="form-grid">
          <label>Company Name *<InputText v-model="form.name" /></label>
          <label>Company NPWP Name<InputText v-model="form.npwpName" /></label>
          <label class="wide">Company NPWP Address<Textarea v-model="form.npwpAddress" rows="3" /></label>
          <label>Company NPWP Number<InputText v-model="form.npwpNumber" /></label>
          <label>Term of Payment<InputText v-model="form.termOfPayment" /></label>
          <label class="wide">Notes<Textarea v-model="form.notes" rows="3" /></label>
        </div>
      </section>
      <section v-else-if="!loading && error" class="load-error-card">
        <i class="pi pi-exclamation-triangle" />
        <strong>Company data could not be loaded</strong>
        <p>{{ error }}</p>
        <Button label="Try Again" icon="pi pi-refresh" outlined size="small" @click="load" />
      </section>
      <div v-else class="loading-state"><i class="pi pi-spin pi-spinner" /> Loading company data...</div>
    </main>
  </section>
</template>

<style scoped>
.company-edit-page{min-height:100%;background:#f8fafc;font-family:Inter,ui-sans-serif,system-ui,sans-serif;color:#0f172a}.company-edit-header{position:sticky;top:0;z-index:30;display:flex;min-height:64px;align-items:center;gap:12px;padding:8px 32px;border-bottom:1px solid #e2e8f0;background:rgba(255,255,255,.97);backdrop-filter:blur(10px)}.company-edit-header :deep(.p-button){font:500 12px Inter,ui-sans-serif,system-ui,sans-serif}.company-edit-header> :first-child{margin-right:0;padding-right:12px;border-right:1px solid #e2e8f0;border-radius:0;white-space:nowrap}.company-edit-header h1{margin:0;font:700 17px/22px Inter,ui-sans-serif,system-ui,sans-serif}.company-edit-header p{margin:2px 0 0;color:#94a3b8;font:400 11px/16px Inter,ui-sans-serif,system-ui,sans-serif}.header-actions{display:flex;gap:8px;margin-left:auto}.header-actions :deep(.p-button:last-child){background:#dc2626;border-color:#dc2626;color:#fff;box-shadow:0 5px 12px rgba(220,38,38,.14)}.company-edit-content{width:min(100%,1100px);margin:24px auto;padding:0 24px}.company-edit-card{overflow:hidden;border:1px solid #e2e8f0;border-radius:18px;background:#fff;box-shadow:0 8px 22px rgba(15,23,42,.05)}.card-heading{padding:20px;border-bottom:1px solid #eef2f7}.card-heading h2{margin:0;font:700 17px/22px Inter,ui-sans-serif,system-ui,sans-serif}.card-heading p{margin:4px 0 0;color:#64748b;font:400 12px/18px Inter,ui-sans-serif,system-ui,sans-serif}.form-grid{display:grid;grid-template-columns:repeat(2,minmax(0,1fr));gap:16px;padding:20px}.form-grid label{display:grid;gap:7px;color:#334155;font:600 12px/16px Inter,ui-sans-serif,system-ui,sans-serif}.form-grid .wide{grid-column:1/-1}.form-grid :deep(.p-inputtext),.form-grid :deep(.p-textarea){width:100%;border:1px solid #d9e2ec;border-radius:10px;font:500 13px/20px Inter,ui-sans-serif,system-ui,sans-serif}.loading-state{display:flex;justify-content:center;gap:8px;padding:48px;color:#64748b;font:400 13px Inter,ui-sans-serif,system-ui,sans-serif}@media(max-width:640px){.company-edit-header{padding:8px 16px;flex-wrap:wrap}.company-edit-header> :nth-child(2){min-width:0}.header-actions{width:100%;justify-content:flex-end}.company-edit-content{margin:16px auto;padding:0 12px}.form-grid{grid-template-columns:1fr}.form-grid .wide{grid-column:auto}}
</style>
