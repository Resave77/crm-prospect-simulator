<script setup lang="ts">
import { ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import Button from 'primevue/button'
import InputText from 'primevue/inputtext'
import Select from 'primevue/select'
import Message from 'primevue/message'
import { useCustomerListStore } from '../../../stores/customerList'

const router = useRouter()
const route = useRoute()
const store = useCustomerListStore()
const error = ref('')
const success = ref('')

const form = ref({
  companyCode: '',
  name: '',
  registeredLocation: '',
  kam: ''
})

async function loadCompany() {
  try {
    if (!store.allCustomers.length) {
      await store.fetchCustomers()
    }
    const id = route.params.id as string
    const representative = store.allCustomers.find((c: any) => c.parentCode === id)
    if (!representative) {
      error.value = 'Company not found.'
      return
    }
    form.value = {
      companyCode: id,
      name: representative.parentCompanyName || 'Unknown Company',
      registeredLocation: representative.region || 'Unknown',
      kam: representative.salesExecutiveName || ''
    }
  } catch (e) {
    error.value = store.errorMessage(e)
  }
}

function saveChanges() {
  success.value = 'Company update simulated. Backend save is not yet implemented.'
}

loadCompany()
</script>

<template>
  <section class="admin-page">
    <div class="page-heading">
      <div>
        <p class="eyebrow">Edit Company</p>
        <h1>Edit Parent Company</h1>
        <p class="muted">Update parent company metadata and save changes.</p>
      </div>
      <div style="display:flex;gap:0.5rem;align-items:center;flex-wrap:wrap;">
        <Button label="Back to Companies" icon="pi pi-arrow-left" text size="small" @click="router.push('/admin/companies')" />
      </div>
    </div>

    <Message v-if="error" severity="error" :closable="false">{{ error }}</Message>
    <Message v-if="success" severity="success" :closable="false">{{ success }}</Message>

    <div class="section-card" style="max-width:720px;">
      <div class="form-row">
        <label>Company Code</label>
        <InputText v-model="form.companyCode" disabled />
      </div>
      <div class="form-row">
        <label>Legal Name</label>
        <InputText v-model="form.name" />
      </div>
      <div class="form-row">
        <label>Registered Location</label>
        <InputText v-model="form.registeredLocation" />
      </div>
      <div class="form-row">
        <label>Key Account Manager</label>
        <Select v-model="form.kam" :options="[{ label: 'Unassigned', value: '' }, { label: 'Kam R', value: 'Kam R' }, { label: 'Sales Team', value: 'Sales Team' }]" optionLabel="label" optionValue="value" />
      </div>
      <div class="form-actions" style="display:flex;gap:0.75rem;flex-wrap:wrap;">
        <Button label="Save Changes" icon="pi pi-save" severity="success" @click="saveChanges" />
        <Button label="Cancel" icon="pi pi-times" text @click="router.push('/admin/companies')" />
      </div>
    </div>
  </section>
</template>
