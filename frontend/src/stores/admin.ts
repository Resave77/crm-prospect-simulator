import { reactive, ref } from 'vue'
import { defineStore } from 'pinia'
import axios from 'axios'
import * as adminApi from '../api/admin'
import type { ApiErrorEnvelope } from '../types/auth'
import type {
  AdminCreateUserInput,
  AdminManagerOption,
  AdminResetPasswordPayload,
  AdminUpdateUserInput,
  AdminUserDetail,
  AdminUserListItem,
  AdminUserListParams,
  AdminUserStatus,
  CreateSalesAssignmentPayload,
  CreateSalesRolePayload,
  SalesRole,
  SalesStructureItem,
  UpdateSalesRolePayload,
} from '../types/admin'

const defaultParams: AdminUserListParams = {
  page: 1,
  limit: 20,
  search: '',
  role: '',
  status: '',
}

export const useAdminStore = defineStore('admin', () => {
  const users = ref<AdminUserListItem[]>([])
  const total = ref(0)
  const page = ref(1)
  const limit = ref(20)
  const pages = ref(0)
  const loading = ref(false)
  const managerOptions = ref<AdminManagerOption[]>([])
  const params = reactive<AdminUserListParams>({ ...defaultParams })
  const selectedUser = ref<AdminUserDetail | null>(null)
  const detailLoading = ref(false)
  const saving = ref(false)
  const resettingPassword = ref(false)
  const salesRoles = ref<SalesRole[]>([])
  const salesRolesLoading = ref(false)
  const salesStructure = ref<SalesStructureItem[]>([])
  const salesStructureLoading = ref(false)
  const selectedEffectiveMonth = ref(new Date().toISOString().slice(0, 7))
  const savingSalesRole = ref(false)
  const savingSalesAssignment = ref(false)

  async function fetchUsers() {
    loading.value = true
    try {
      const result = await adminApi.getUsers(params)
      users.value = result.data
      total.value = result.total
      page.value = result.page
      limit.value = result.limit
      pages.value = result.pages
    } finally {
      loading.value = false
    }
  }

  async function fetchManagers() {
    if (managerOptions.value.length) return
    managerOptions.value = await adminApi.getManagerOptions()
  }

  async function createUser(input: AdminCreateUserInput) {
    const user = await adminApi.createUser(input)
    void fetchUsers().catch(() => undefined)
    return user
  }

  async function updateStatus(id: string, status: AdminUserStatus) {
    const user = await adminApi.updateStatus(id, status)
    if (selectedUser.value?.id === id) selectedUser.value = user
    await fetchUsers()
    return user
  }

  async function fetchUserById(id: string) {
    detailLoading.value = true
    try {
      const user = await adminApi.getUser(id)
      selectedUser.value = user
      return user
    } finally {
      detailLoading.value = false
    }
  }

  async function updateUser(id: string, payload: AdminUpdateUserInput) {
    saving.value = true
    try {
      const user = await adminApi.updateUser(id, payload)
      if (selectedUser.value?.id === id) selectedUser.value = user
      return user
    } finally {
      saving.value = false
    }
  }

  function clearSelectedUser() {
    selectedUser.value = null
  }

  async function resetPassword(userId: string, payload: AdminResetPasswordPayload) {
    if (resettingPassword.value) return undefined
    resettingPassword.value = true
    try {
      return await adminApi.resetUserPassword(userId, payload)
    } finally {
      resettingPassword.value = false
    }
  }

  function setParam<K extends keyof AdminUserListParams>(key: K, value: AdminUserListParams[K]) {
    params[key] = value
  }

  function setPage(p: number) {
    params.page = p
    page.value = p
  }

  function resetFilters() {
    params.page = 1
    params.search = ''
    params.role = ''
    params.status = ''
    page.value = 1
  }


  async function fetchSalesRoles() {
    salesRolesLoading.value = true
    try {
      salesRoles.value = await adminApi.getSalesRoles()
    } finally {
      salesRolesLoading.value = false
    }
  }

  async function createSalesRole(payload: CreateSalesRolePayload) {
    savingSalesRole.value = true
    try {
      const role = await adminApi.createSalesRole(payload)
      await fetchSalesRoles()
      return role
    } finally {
      savingSalesRole.value = false
    }
  }

  async function updateSalesRole(id: string, payload: UpdateSalesRolePayload) {
    savingSalesRole.value = true
    try {
      const role = await adminApi.updateSalesRole(id, payload)
      await fetchSalesRoles()
      return role
    } finally {
      savingSalesRole.value = false
    }
  }

  async function setSalesRoleStatus(id: string, isActive: boolean) {
    savingSalesRole.value = true
    try {
      const role = await adminApi.updateSalesRoleStatus(id, isActive)
      await fetchSalesRoles()
      return role
    } finally {
      savingSalesRole.value = false
    }
  }

  async function deleteSalesRole(id: string) {
    savingSalesRole.value = true
    try {
      await adminApi.deleteSalesRole(id)
      await fetchSalesRoles()
    } finally {
      savingSalesRole.value = false
    }
  }

  async function fetchSalesStructure(effectiveDate: string) {
    salesStructureLoading.value = true
    try {
      salesStructure.value = await adminApi.getSalesStructure(effectiveDate)
    } finally {
      salesStructureLoading.value = false
    }
  }

  async function createSalesAssignment(payload: CreateSalesAssignmentPayload) {
    savingSalesAssignment.value = true
    try {
      const assignment = await adminApi.createSalesAssignment(payload)
      await fetchSalesStructure(payload.effectiveFrom)
      return assignment
    } finally {
      savingSalesAssignment.value = false
    }
  }
  function errorMessage(error: unknown) {
    if (axios.isAxiosError<ApiErrorEnvelope>(error)) {
      return error.response?.data?.error?.message ?? 'Account service is unavailable.'
    }
    return error instanceof Error ? error.message : 'An unexpected error occurred.'
  }

  return {
    users, total, page, limit, pages, loading, managerOptions, params,
    selectedUser, detailLoading, saving, resettingPassword,
    salesRoles, salesRolesLoading, salesStructure, salesStructureLoading, selectedEffectiveMonth, savingSalesRole, savingSalesAssignment,
    fetchUsers, fetchManagers, createUser, updateStatus,
    fetchUserById, updateUser, clearSelectedUser, resetPassword,
    setParam, setPage, resetFilters,
    fetchSalesRoles, createSalesRole, updateSalesRole, setSalesRoleStatus, deleteSalesRole,
    fetchSalesStructure, createSalesAssignment, errorMessage,
  }
})

