import assert from 'node:assert/strict'
import test from 'node:test'
import { completedVisitsByProspectId, pendingRoute } from '../src/utils/visitCompletion.ts'

test('completion matches prospectId and selects latest checkout', () => {
  const visits = [{ prospectId: 'A', checkOutAt: '2026-09-02T10:00:00' }, { prospectId: 'A', checkOutAt: '2026-09-02T10:42:00' }]
  assert.equal(completedVisitsByProspectId(visits).get('A').checkOutAt, '2026-09-02T10:42:00')
})
test('null checkout remains pending and same-name IDs do not match', () => {
  const completed = completedVisitsByProspectId([{ prospectId: 'B', checkOutAt: '2026-09-02T10:00:00' }, { prospectId: 'A', checkOutAt: null }])
  assert.equal(completed.has('A'), false)
  assert.deepEqual(pendingRoute([{ id: 'A' }, { id: 'B' }, { id: 'C' }], new Set(['A'])).map(x => x.id), ['B', 'C'])
})
test('replan excludes completed while preserving historical item separately', () => {
  const historical = [{ id: 'A' }]
  const future = pendingRoute([{ id: 'A' }, { id: 'B' }, { id: 'C' }], new Set(['A']))
  assert.deepEqual(historical.map(x => x.id), ['A'])
  assert.deepEqual(future.map(x => x.id), ['B', 'C'])
})
