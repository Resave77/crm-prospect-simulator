import assert from 'node:assert/strict'
import { readFile } from 'node:fs/promises'
import test from 'node:test'

const customer = await readFile(new URL('../src/views/Admin/Customer/CustomerListView.vue', import.meta.url), 'utf8')
const prospect = await readFile(new URL('../src/views/Admin/Prospect/ProspectListView.vue', import.meta.url), 'utf8')

test('customer list pagination has guarded numbered and previous/next navigation', () => {
  assert.match(customer, /@click="goToPage\(p as number\)"/)
  assert.match(customer, /@click="goToPage\(store\.page - 1\)"/)
  assert.match(customer, /@click="goToPage\(store\.page \+ 1\)"/)
  assert.match(customer, /Math\.max\(1, Math\.min\(p, store\.pages \|\| 1\)\)/)
})

test('prospect list pagination slices the complete loaded dataset and updates active page', () => {
  assert.match(prospect, /const paginatedProspects = computed\(\(\) =>/)
  assert.match(prospect, /allFiltered\.value\.slice\(start, start \+ prospectPageSize\.value\)/)
  assert.match(prospect, /@click="goToProspectPage\(page\)"/)
  assert.match(prospect, /page === prospectPage/)
  assert.match(prospect, /@click="goToProspectPage\(prospectPage - 1\)"/)
  assert.match(prospect, /@click="goToProspectPage\(prospectPage \+ 1\)"/)
})

test('customer and prospect checkbox controls have scoped square styling', () => {
  assert.match(customer, /\.customer-site-table \.customer-checkbox[\s\S]*border-radius: 4px !important/)
  assert.match(prospect, /\.prospect-page \.prospect-checkbox[\s\S]*border-radius: 4px !important/)
})
