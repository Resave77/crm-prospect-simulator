import assert from 'node:assert/strict'
import { readFile } from 'node:fs/promises'
import test from 'node:test'

const source = await readFile(new URL('../src/views/Sales/Prospect/ProspectDetailView.vue', import.meta.url), 'utf8')
const api = await readFile(new URL('../src/api/crm.ts', import.meta.url), 'utf8')

test('Prospect Detail is CRM-first and business info is explicit', () => {
  assert.match(source, /review\.value\?\.prospect\.phoneNumber \|\| placeDetails\.value\?\.phoneNumber/)
  assert.match(source, /@click="loadBusinessInfo"/)
  assert.match(source, /getProspectBusinessInfo\(review\.value\.prospect\.id, 'SALES_EXECUTIVE'\)/)
  assert.match(api, /\/business-info/)
  assert.match(source, /Reviews not loaded\./)
})
