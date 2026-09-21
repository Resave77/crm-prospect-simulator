import assert from 'node:assert/strict'
import { readFile } from 'node:fs/promises'
import test from 'node:test'

const capabilities = await readFile(new URL('../src/config/capabilities.ts', import.meta.url), 'utf8')
const prospects = await readFile(new URL('../src/views/Admin/Prospect/ProspectListView.vue', import.meta.url), 'utf8')
const customers = await readFile(new URL('../src/views/Admin/Customer/CustomerListView.vue', import.meta.url), 'utf8')
const crm = await readFile(new URL('../src/api/crm.ts', import.meta.url), 'utf8')

test('DEV compatibility defaults unsupported trash capabilities to false when unset', () => {
  assert.match(capabilities, /VITE_ENABLE_PROSPECT_TRASH === 'true'/)
  assert.match(capabilities, /VITE_ENABLE_CUSTOMER_TRASH === 'true'/)
  assert.match(capabilities, /prospectTrash:/)
  assert.match(capabilities, /customerTrash:/)
})

test('unsupported prospect and customer trash controls are capability-gated', () => {
  assert.match(prospects, /v-if="prospectTrashEnabled"/)
  assert.match(customers, /v-if="customerTrashEnabled"/)
  assert.match(prospects, /if \(!prospectTrashEnabled\) return/)
  assert.match(customers, /if \(!customerTrashEnabled\) return/)
})

test('supported deletion-request controls remain present', () => {
  assert.match(prospects, /approveProspectDeletion/)
  assert.match(prospects, /rejectProspectDeletion/)
})

test('Prospect Finder uses the backend path-parameter contract', () => {
  assert.match(crm, /place-details\/\$\{encodeURIComponent\(googlePlaceId\)\}/)
  assert.doesNotMatch(crm, /place-details', \{ params: \{ googlePlaceId \} \}/)
})
