import { ref } from 'vue'
import { defineStore } from 'pinia'
import axios from 'axios'
import * as crmApi from '../api/crm'
import type { ApiErrorEnvelope } from '../types/auth'
import type { CustomerSite, CustomerDetail, Prospect, ProspectStatus } from '../types/crm'

export const useCrmStore = defineStore('crm', () => {
  const myProspects = ref<Prospect[]>([])
  const adminCustomers = ref<CustomerSite[]>([])
  const myCustomers = ref<CustomerSite[]>([])
  const pipeline = ref<Prospect[]>([])
  const loading = ref(false)

  // Keep loading accurate even when multiple CRM requests run at the same time.
  let activeRequests = 0

  // Protect myProspects from stale responses:
  // - request sequence ignores older list responses
  // - epoch invalidates any list request that started before/during a transition
  let prospectsRequestSequence = 0
  let prospectsEpoch = 0

  async function run<T>(operation: () => Promise<T>) {
    activeRequests += 1
    loading.value = true

    try {
      return await operation()
    } finally {
      activeRequests = Math.max(0, activeRequests - 1)
      loading.value = activeRequests > 0
    }
  }

  async function loadMyProspects() {
    const requestSequence = ++prospectsRequestSequence
    const requestEpoch = prospectsEpoch

    const result = await run(crmApi.getMyProspects)

    // Apply only the newest request and only if no prospect mutation
    // started/completed while this request was in flight.
    if (
      requestSequence === prospectsRequestSequence &&
      requestEpoch === prospectsEpoch
    ) {
      myProspects.value = result
    }

    return result
  }

  async function transition(id: string, status: ProspectStatus, notes = '') {
    const index = myProspects.value.findIndex((item) => item.id === id)
    const previous = index >= 0 ? myProspects.value[index] : null

    // Invalidate list requests that started before this transition.
    prospectsEpoch += 1

    if (previous) {
      myProspects.value[index] = { ...previous, status }
    }

    try {
      const result = await run(() => crmApi.transitionProspect(id, status, notes))

      myProspects.value = myProspects.value.map((item) =>
        item.id === id ? result : item,
      )

      // Invalidate list requests that may have started while the transition
      // request was still in flight and could contain pre-transition data.
      prospectsEpoch += 1

      return result
    } catch (error) {
      if (previous) {
        const currentIndex = myProspects.value.findIndex((item) => item.id === id)

        if (currentIndex >= 0) {
          myProspects.value[currentIndex] = previous
        } else {
          myProspects.value = [...myProspects.value, previous]
        }
      }

      // Also invalidate requests started during the failed mutation.
      prospectsEpoch += 1
      throw error
    }
  }

  async function loadPipeline() {
    pipeline.value = await run(crmApi.getPipeline)
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

    return error instanceof Error
      ? error.message
      : 'An unexpected CRM error occurred.'
  }

  return {
    myProspects,
    adminCustomers,
    myCustomers,
    pipeline,
    loading,
    loadMyProspects,
    transition,
    loadPipeline,
    loadAdminCustomers,
    loadMyCustomers,
    loadAdminCustomer,
    errorMessage,
  }
})