import assert from 'node:assert/strict'
import test from 'node:test'
import { filterProspects } from '../src/domain/pipeline.ts'

const base = { placeName: 'Cafe A', formattedAddress: 'Jakarta', assignedSalesExecutiveId: 'sales-a', industryGroup: 'N&B / Kuliner', status: 'NEGOTIATION' }
const items = [base, { ...base, placeName: 'Retail B', assignedSalesExecutiveId: 'sales-b', industryGroup: 'Retail', status: 'CONTACTED' }]

test('pipeline filters combine sales executive and industry group', () => {
  assert.deepEqual(filterProspects(items, { salesExecutiveId: 'sales-b', industryGroup: 'Retail', status: '', search: '' }).map((item) => item.placeName), ['Retail B'])
})

test('pipeline status and text search are case-insensitive', () => {
  assert.deepEqual(filterProspects(items, { salesExecutiveId: '', industryGroup: '', status: 'NEGOTIATION', search: 'cafe' }).map((item) => item.placeName), ['Cafe A'])
})
