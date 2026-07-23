import assert from 'node:assert/strict'
import { readFile } from 'node:fs/promises'
import test from 'node:test'

test('all visible administrator and sales navigation targets have routes', async () => {
  const router = await readFile(new URL('../src/router/index.ts', import.meta.url), 'utf8')
  const admin = await readFile(new URL('../src/layouts/AdminLayout.vue', import.meta.url), 'utf8')
  const sales = await readFile(new URL('../src/layouts/SalesLayout.vue', import.meta.url), 'utf8')
  for (const route of ['/admin/prospect-finder','/admin/prospects/pipeline','/admin/sales-executives','/admin/customer-assignment','/admin/visit-monitoring','/admin/prospect-assignment','/admin/reports','/sales/history','/sales/profile']) {
    assert.match(router, new RegExp(route.split('/').at(-1).replace('-', '[-]')))
    assert.ok(admin.includes(route) || sales.includes(route), `${route} must be visible and routable`)
  }
})

test('sales navigation exposes five functional destinations', async () => {
  const sales = await readFile(new URL('../src/layouts/SalesLayout.vue', import.meta.url), 'utf8')
  for (const route of ['/sales/dashboard','/sales/my-customers','/sales/my-prospects','/sales/history','/sales/profile']) assert.ok(sales.includes(route))
  assert.equal((sales.match(/<RouterLink/g) ?? []).length >= 6, true)
})
