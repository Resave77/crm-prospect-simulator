import assert from 'node:assert/strict'
import { readFile } from 'node:fs/promises'
import test from 'node:test'

test('all visible administrator and sales navigation targets have routes', async () => {
  const router = await readFile(new URL('../src/router/index.ts', import.meta.url), 'utf8')
  const admin = await readFile(new URL('../src/layouts/AdminLayout.vue', import.meta.url), 'utf8')
  const sales = await readFile(new URL('../src/layouts/SalesLayout.vue', import.meta.url), 'utf8')
  for (const route of ['/admin/prospect-finder','/admin/prospects/pipeline','/admin/accounts','/admin/customers','/admin/visit-monitoring','/admin/reports','/sales/history','/sales/profile']) {
    assert.match(router, new RegExp(route.split('/').at(-1).replace('-', '[-]')))
    assert.ok(admin.includes(route) || sales.includes(route), `${route} must be visible and routable`)
  }
})

test('sales navigation exposes five functional destinations', async () => {
  const sales = await readFile(new URL('../src/layouts/SalesLayout.vue', import.meta.url), 'utf8')
  for (const route of ['/sales/dashboard','/sales/my-customers','/sales/pipeline','/sales/history','/sales/profile']) assert.ok(sales.includes(route))
  assert.ok(sales.includes('visibleNavItems'), 'desktop and mobile navigation should share one config')
  assert.ok(sales.includes('Prospect Pipeline'))
  assert.doesNotMatch(sales, /label: 'Prospect'/)
})

test('sales navigation items are permission-aware', async () => {
  const sales = await readFile(new URL('../src/layouts/SalesLayout.vue', import.meta.url), 'utf8')
  for (const permission of ['view_sales_dashboard', 'view_my_customers', 'menu_sales_pipeline', 'view_sales_history', 'view_own_profile']) {
    assert.ok(sales.includes(`permission: '${permission}'`), `${permission} should control a sales nav item`)
  }
  assert.match(sales, /auth\.hasPermission\(item\.permission\)/)
})

test('route permission guard allows permitted routes and denies missing permissions', async () => {
  const router = await readFile(new URL('../src/router/index.ts', import.meta.url), 'utf8')
  assert.match(router, /permission\?: string/)
  assert.match(router, /path: 'dashboard'.*permission: 'view_sales_dashboard'/)
  assert.match(router, /path: 'pipeline'.*permission: 'menu_sales_pipeline'/)
  assert.match(router, /to\.meta\.permission && !hasPermission\(user, to\.meta\.permission\)/)
  assert.match(router, /routePermitted\(router, fallback, user\)/)
  assert.match(router, /return FORBIDDEN_ROUTE/)
})
