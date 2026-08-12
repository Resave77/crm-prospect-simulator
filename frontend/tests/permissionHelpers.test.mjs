import assert from 'node:assert/strict'
import { readFile } from 'node:fs/promises'
import test from 'node:test'

test('auth store exposes permissionKeys and hasPermission helpers', async () => {
  const auth = await readFile(new URL('../src/stores/auth.ts', import.meta.url), 'utf8')
  assert.match(auth, /const permissionKeys = computed/)
  assert.match(auth, /function hasPermission\(key: string\)/)
  assert.match(auth, /userHasPermission\(user\.value, key\)/)
  assert.match(auth, /permissionKeys,/)
  assert.match(auth, /hasPermission,/)
})

test('permission helper keeps SUPER_ADMIN unrestricted and sales roles permission-driven', async () => {
  const navigation = await readFile(new URL('../src/utils/navigation.ts', import.meta.url), 'utf8')
  assert.match(navigation, /export function permissionKeysFor/)
  assert.match(navigation, /user\.salesRole\?\.permissionKeys \?\? \[\]/)
  assert.match(navigation, /user\.role === 'SUPER_ADMIN'/)
  assert.match(navigation, /permissionKeysFor\(user\)\.includes\(key\)/)
})
