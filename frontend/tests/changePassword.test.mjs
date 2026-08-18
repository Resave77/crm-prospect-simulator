import assert from 'node:assert/strict'
import { readFile } from 'node:fs/promises'
import test from 'node:test'

const authTypes = () => readFile(new URL('../src/types/auth.ts', import.meta.url), 'utf8')
const authApi = () => readFile(new URL('../src/api/auth.ts', import.meta.url), 'utf8')
const authStore = () => readFile(new URL('../src/stores/auth.ts', import.meta.url), 'utf8')

test('change-password types match the backend contract', async () => {
  const source = await authTypes()
  for (const field of ['currentPassword: string', 'newPassword: string', 'confirmPassword: string']) {
    assert.ok(source.includes(field), `${field} must be typed`)
  }
  for (const field of [
    'passwordChanged: boolean',
    'mustChangePassword: boolean',
    'sessionsRevoked: number',
    'reauthenticationRequired: boolean',
  ]) {
    assert.ok(source.includes(field), `${field} must be typed`)
  }
})

test('auth user preserves mustChangePassword from auth payloads', async () => {
  const source = await authTypes()
  assert.match(source, /interface AuthUser[\s\S]*mustChangePassword\?: boolean/)
})

test('change-password API posts through the shared auth client and unwraps the envelope', async () => {
  const source = await authApi()
  assert.ok(source.includes("import { api } from './client'"))
  assert.ok(source.includes("api.post<ApiEnvelope<ChangePasswordResult>>('/auth/change-password', payload)"))
  assert.ok(source.includes('return response.data.data'))
  assert.equal(source.includes('axios.create'), false)
})

test('change-password store action blocks duplicates and clears auth state after success', async () => {
  const source = await authStore()
  assert.ok(source.includes('const changingPassword = ref(false)'))
  assert.match(source, /if \(changingPassword\.value\)[\s\S]*Password change already in progress\./)
  assert.match(source, /const result = await changePasswordRequest\(payload\)[\s\S]*applySession\(\{ accessToken: result\.accessToken, accessExpiresAt: result\.accessExpiresAt, user: result\.user \}\)[\s\S]*return result/)
  assert.equal(source.includes("api.post('/auth/logout')") && source.indexOf('async function changePassword') < source.indexOf("api.post('/auth/logout')"), false)
})

test('change-password store action does not persist password fields', async () => {
  const source = await authStore()
  assert.equal(source.includes('currentPassword'), false)
  assert.equal(source.includes('newPassword'), false)
  assert.equal(source.includes('confirmPassword'), false)
  assert.equal(source.includes('localStorage'), false)
})
