import assert from 'node:assert/strict'
import { readFile } from 'node:fs/promises'
import test from 'node:test'

test('post-login redirect falls back to the existing admin dashboard route', async () => {
  const router = await readFile(new URL('../src/router/index.ts', import.meta.url), 'utf8')
  const navigation = await readFile(new URL('../src/utils/navigation.ts', import.meta.url), 'utf8')
  const login = await readFile(new URL('../src/views/Login/LoginView.vue', import.meta.url), 'utf8')

  assert.match(router, /path: 'dashboard', name: 'AdminDashboard'/)
  assert.match(navigation, /ADMIN_DASHBOARD_ROUTE = '\/admin\/dashboard'/)
  assert.match(navigation, /routePermitted\(router, intended, role\)/)
  assert.match(login, /resolvePostLoginRoute\(router, user\.role, intended\)/)
})

test('SUPER_ADMIN passes administrator route guards', async () => {
  const navigation = await readFile(new URL('../src/utils/navigation.ts', import.meta.url), 'utf8')
  const router = await readFile(new URL('../src/router/index.ts', import.meta.url), 'utf8')

  assert.match(navigation, /required === 'ADMINISTRATOR'/)
  assert.match(navigation, /actual === 'SUPER_ADMIN' \|\| actual === 'ADMINISTRATOR'/)
  assert.match(router, /meta: \{ role: 'ADMINISTRATOR' \}/)
})
