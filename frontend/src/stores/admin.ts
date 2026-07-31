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

  function errorMessage(error: unknown) {
    if (axios.isAxiosError<ApiErrorEnvelope>(error)) {
      return error.response?.data?.error?.message ?? 'Account service is unavailable.'
    }
    return error instanceof Error ? error.message : 'An unexpected error occurred.'
  }

  return {
    users, total, page, limit, pages, loading, managerOptions, params,
    selectedUser, detailLoading, saving, resettingPassword,
    fetchUsers, fetchManagers, createUser, updateStatus,
    fetchUserById, updateUser, clearSelectedUser, resetPassword,
    setParam, setPage, resetFilters, errorMessage,
  }
})
