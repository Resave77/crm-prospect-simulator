import { computed, ref } from 'vue'
import { defineStore } from 'pinia'
import axios, { AxiosError } from 'axios'
import { changePassword as changePasswordRequest, getCurrentUser } from '../api/auth'
import { api, observeSession, refreshSession, setAccessToken } from '../api/client'
import type {
  ApiEnvelope,
  ApiErrorEnvelope,
  AuthPayload,
  AuthUser,
  ChangePasswordPayload,
  ChangePasswordResult,
  UserRole,
} from '../types/auth'
import { hasPermission as userHasPermission, permissionKeysFor } from '../utils/navigation'

export const useAuthStore = defineStore('auth', () => {
  const user = ref<AuthUser | null>(null)
  const accessExpiresAt = ref<string | null>(null)
  const bootstrapped = ref(false)
  const loading = ref(false)
  const changingPassword = ref(false)

  const authenticated = computed(() => user.value !== null)
  const role = computed<UserRole | null>(() => user.value?.role ?? null)
  const permissionKeys = computed(() => user.value ? permissionKeysFor(user.value) : [])

  function hasPermission(key: string) {
    return user.value ? userHasPermission(user.value, key) : false
  }

  function applySession(payload: AuthPayload | null) {
    user.value = payload?.user ?? null
    accessExpiresAt.value = payload?.accessExpiresAt ?? null
    setAccessToken(payload?.accessToken ?? null)
  }

  observeSession(applySession)

  async function bootstrap() {
    if (bootstrapped.value) return
    try {
      applySession(await refreshSession())
    } catch {
      applySession(null)
    } finally {
      bootstrapped.value = true
    }
  }

  async function login(email: string, password: string) {
    loading.value = true
    try {
      const response = await api.post<ApiEnvelope<AuthPayload>>('/auth/login', { email, password })
      applySession(response.data.data)
      bootstrapped.value = true
      return response.data.data.user
    } finally {
      loading.value = false
    }
  }

  async function refreshUser() {
    const currentUser = await getCurrentUser()
    user.value = currentUser
    return currentUser
  }

  async function logout() {
    try {
      await api.post('/auth/logout')
    } finally {
      applySession(null)
      bootstrapped.value = true
    }
  }

  async function changePassword(payload: ChangePasswordPayload): Promise<ChangePasswordResult> {
    if (changingPassword.value) {
      throw new Error('Password change already in progress.')
    }

    changingPassword.value = true
    try {
      const result = await changePasswordRequest(payload)
      applySession(null)
      bootstrapped.value = true
      return result
    } finally {
      changingPassword.value = false
    }
  }

  function errorMessage(error: unknown) {
    if (axios.isAxiosError<ApiErrorEnvelope>(error)) {
      return error.response?.data?.error?.message ?? 'Unable to connect to the CRM service.'
    }
    return error instanceof Error ? error.message : 'An unexpected error occurred.'
  }

  return {
    user,
    accessExpiresAt,
    bootstrapped,
    loading,
    changingPassword,
    authenticated,
    role,
    permissionKeys,
    hasPermission,
    bootstrap,
    login,
    refreshUser,
    logout,
    changePassword,
    errorMessage,
  }
})



