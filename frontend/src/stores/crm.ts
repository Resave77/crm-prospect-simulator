import { ref } from 'vue'
import { defineStore } from 'pinia'
import axios from 'axios'
import * as crmApi from '../api/crm'
import type { ApiErrorEnvelope } from '../types/auth'
import type { CustomerSite, CustomerDetail, Prospect, ProspectStatus } from '../types/crm'

export const useCrmStore = defineStore('crm', () => {
  const myProspects = ref<Prospect[]>([])
  const wonProspects = ref<Prospect[]>([])
  const adminCustomers = ref<CustomerSite[]>([])
  const myCustomers = ref<CustomerSite[]>([])
  const pipeline = ref<Prospect[]>([])
  const loading = ref(false)

  async function run<T>(operation: () => Promise<T>) {
    loading.value = true
    try {
      return await operation()
    } finally {
      loading.value = false
    }
  }

  async function loadMyProspects() {
    myProspects.value = await run(crmApi.getMyProspects)
  }

  async function transition(id: string, status: ProspectStatus, notes = '') {
    const index = myProspects.value.findIndex((item) => item.id === id)
    const previous = index >= 0 ? myProspects.value[index] : null
    if (previous) myProspects.value[index] = { ...previous, status }
    try {
      const result = await run(() => crmApi.transitionProspect(id, status, notes))
      myProspects.value = myProspects.value.map((item) => item.id === id ? result : item)
      return result
    } catch (error) {
      if (previous) myProspects.value[index] = previous
      throw error
    }
  }

  async function loadPipeline() { pipeline.value = await run(crmApi.getPipeline) }

  async function loadWonProspects() {
    wonProspects.value = await run(crmApi.getWonProspects)
  }

  async function loadAdminCustomers() {
    adminCustomers.value = await run(crmApi.getAdminCustomers)
  }

  async function loadMyCustomers() {
    myCustomers.value = await run(crmApi.getMyCustomers)
  }

  async function loadAdminCustomer(id: string) {
    return await run(() => crmApi.getAdminCustomer(id))
  }

  function errorMessage(error: unknown) {
    if (axios.isAxiosError<ApiErrorEnvelope>(error)) {
      return error.response?.data?.error?.message ?? 'CRM service is unavailable.'
    }
    return error instanceof Error ? error.message : 'An unexpected CRM error occurred.'
  }

  return { myProspects, wonProspects, adminCustomers, myCustomers, pipeline, loading, loadMyProspects, transition, loadPipeline, loadWonProspects, loadAdminCustomers, loadMyCustomers, loadAdminCustomer, errorMessage }
})
