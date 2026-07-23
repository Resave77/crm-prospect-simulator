import { reactive, ref } from 'vue'
import { defineStore } from 'pinia'
import axios from 'axios'
import * as crmApi from '../api/crm'
import type { ApiErrorEnvelope } from '../types/auth'
import type { CustomerListParams, CustomerSite, ListFilterOptions } from '../types/crm'

const defaultParams: CustomerListParams = {
  page: 1,
  limit: 20,
  keyword: '',
  segment: '',
  category: '',
  sales: '',
  region: '',
  sort: ''
}

export const useCustomerListStore = defineStore('customerList', () => {
  const items = ref<CustomerSite[]>([])
  const allCustomers = ref<CustomerSite[]>([])
  const total = ref(0)
  const page = ref(1)
  const limit = ref(20)
  const pages = ref(0)
  const loading = ref(false)
  const filterOptions = ref<ListFilterOptions | null>(null)
  const params = reactive<CustomerListParams>({ ...defaultParams })
  const selectedIds = ref<Set<string>>(new Set())

  function deriveFilterOptions() {
    const segments = [...new Set(allCustomers.value.map(c => c.segment).filter(Boolean))].sort()
    const categories = [...new Set(allCustomers.value.map(c => c.category).filter(Boolean))].sort()
    const regions = [...new Set(allCustomers.value.map(c => c.region).filter(Boolean))].sort()
    const salesMap = new Map<string, { id: string; fullName: string }>()
    for (const c of allCustomers.value) {
      if (c.salesExecutiveId && c.salesExecutiveName && !salesMap.has(c.salesExecutiveId)) {
        salesMap.set(c.salesExecutiveId, { id: c.salesExecutiveId, fullName: c.salesExecutiveName })
      }
    }
    filterOptions.value = {
      segments,
      categories,
      regions,
      salesExecutives: [...salesMap.values()]
    }
  }

  async function fetchCustomers() {
    loading.value = true
    try {
      if (allCustomers.value.length === 0) {
        allCustomers.value = await crmApi.getAdminCustomers()
      }
      if (!filterOptions.value) deriveFilterOptions()

      let filtered = [...allCustomers.value]

      if (params.keyword) {
        const kw = params.keyword.toLowerCase()
        filtered = filtered.filter(c =>
          c.name.toLowerCase().includes(kw) ||
          c.customerCode.toLowerCase().includes(kw) ||
          c.parentCompanyName.toLowerCase().includes(kw) ||
          c.parentCode.toLowerCase().includes(kw) ||
          (c.salesExecutiveName && c.salesExecutiveName.toLowerCase().includes(kw))
        )
      }
      if (params.segment) {
        filtered = filtered.filter(c => c.segment === params.segment)
      }
      if (params.category) {
        filtered = filtered.filter(c => c.category === params.category)
      }
      if (params.region) {
        filtered = filtered.filter(c => c.region === params.region)
      }
      if (params.sales) {
        filtered = filtered.filter(c => c.salesExecutiveName === params.sales)
      }

      switch (params.sort) {
        case 'oldest':
          filtered.sort((a, b) => new Date(a.convertedAt).getTime() - new Date(b.convertedAt).getTime())
          break
        case 'name':
          filtered.sort((a, b) => a.name.localeCompare(b.name))
          break
        case 'code':
          filtered.sort((a, b) => a.customerCode.localeCompare(b.customerCode))
          break
        case 'converted':
          filtered.sort((a, b) => new Date(b.convertedAt).getTime() - new Date(a.convertedAt).getTime())
          break
        case 'updated':
          filtered.sort((a, b) => new Date(b.updatedAt).getTime() - new Date(a.updatedAt).getTime())
          break
        default:
          filtered.sort((a, b) => new Date(b.convertedAt).getTime() - new Date(a.convertedAt).getTime())
          break
      }

      total.value = filtered.length
      pages.value = Math.ceil(filtered.length / limit.value) || 1
      if (page.value > pages.value) page.value = pages.value
      const start = (page.value - 1) * limit.value
      items.value = filtered.slice(start, start + limit.value)
    } finally {
      loading.value = false
    }
  }

  async function fetchFilterOptions() {
    if (filterOptions.value) return filterOptions.value
    try {
      filterOptions.value = await crmApi.getCustomerFilterOptions()
    } catch {
      if (allCustomers.value.length > 0) deriveFilterOptions()
    }
    return filterOptions.value
  }

  function setParam<K extends keyof CustomerListParams>(key: K, value: CustomerListParams[K]) {
    params[key] = value
  }

  function setPage(p: number) {
    params.page = p
    page.value = p
  }

  function resetFilters() {
    params.page = 1
    params.keyword = ''
    params.segment = ''
    params.category = ''
    params.sales = ''
    params.region = ''
    params.sort = ''
    page.value = 1
    selectedIds.value.clear()
  }

  function toggleSelect(id: string) {
    if (selectedIds.value.has(id)) {
      selectedIds.value.delete(id)
    } else {
      selectedIds.value.add(id)
    }
    selectedIds.value = new Set(selectedIds.value)
  }

  function toggleSelectAll() {
    if (selectedIds.value.size === items.value.length) {
      selectedIds.value.clear()
    } else {
      selectedIds.value = new Set(items.value.map((i) => i.id))
    }
  }

  function isAllSelected() {
    return items.value.length > 0 && selectedIds.value.size === items.value.length
  }

  function errorMessage(error: unknown) {
    if (axios.isAxiosError<ApiErrorEnvelope>(error)) {
      return error.response?.data?.error?.message ?? 'Customer service is unavailable.'
    }
    return error instanceof Error ? error.message : 'An unexpected error occurred.'
  }

  return {
    items, allCustomers, total, page, limit, pages, loading, filterOptions, params, selectedIds,
    fetchCustomers, fetchFilterOptions, setParam, setPage, resetFilters,
    toggleSelect, toggleSelectAll, isAllSelected, errorMessage
  }
})
