import assert from 'node:assert/strict'
import test from 'node:test'
const selectTodayCandidates = (candidates, existingCount, capacity, order) => order(candidates).slice(0, Math.max(0, capacity - existingCount))
const moveToToday = (plan, todayKey, selected) => { const ids = new Set(selected.map(item => item.id)); const next = Object.fromEntries(Object.entries(plan).map(([key, items]) => [key, key === todayKey ? items : items.filter(item => !ids.has(item.id))])); next[todayKey] = [...(next[todayKey] ?? []), ...selected]; return next }
const browseWeekBounds = (currentWeek, planned, before = 4, after = 8) => { const start = new Date(currentWeek.getFullYear(), currentWeek.getMonth(), currentWeek.getDate(), 12); start.setDate(start.getDate() - ((start.getDay() || 7) - 1)); const min = new Date(start); min.setDate(min.getDate() - before * 7); const max = new Date(start); max.setDate(max.getDate() + after * 7); return { minimum: min.getTime(), maximum: Math.max(max.getTime(), ...planned.map(date => { const d = new Date(date); d.setDate(d.getDate() - ((d.getDay() || 7) - 1)); return d.getTime() })) } }

const order = items => [...items].sort((a, b) => a.distance - b.distance)
const item = (id, distance = 0) => ({ id, distance })

test('auto capacity selects exactly two', () => assert.deepEqual(selectTodayCandidates([item('A', 1), item('B', 2), item('C', 3)], 0, 2, order).map(x => x.id), ['A', 'B']))
test('manual capacity selects three', () => assert.equal(selectTodayCandidates([item('A'), item('B'), item('C'), item('D')], 0, 3, order).length, 3))
test('existing pending visits consume available slots', () => assert.equal(selectTodayCandidates([item('B'), item('C'), item('D')], 1, 3, order).length, 2))
test('completed historical rows do not consume active capacity', () => assert.equal(selectTodayCandidates([item('B'), item('C')], 0, 2, order).length, 2))
test('nearest candidates are selected before farther candidates', () => assert.deepEqual(selectTodayCandidates([item('D', 4), item('B', 2), item('A', 1)], 0, 2, order).map(x => x.id), ['A', 'B']))
test('future candidate is moved instead of copied', () => {
  const result = moveToToday({ '2026-09-02': [], '2026-09-07': [item('B'), item('C')] }, '2026-09-02', [item('B')])
  assert.deepEqual(result['2026-09-02'].map(x => x.id), ['B'])
  assert.deepEqual(result['2026-09-07'].map(x => x.id), ['C'])
})
test('move helper is atomic for empty selection', () => {
  const plan = { '2026-09-07': [item('B')] }
  assert.deepEqual(moveToToday(plan, '2026-09-02', []), { '2026-09-07': [item('B')], '2026-09-02': [] })
})
test('browse bounds include four weeks before and eight after current', () => {
  const bounds = browseWeekBounds(new Date(2026, 8, 2), [])
  assert.equal(new Date(bounds.minimum).getDate(), 3)
  assert.equal(new Date(bounds.maximum).getDate(), 26)
})
test('browse bounds extend to far planned week', () => {
  const bounds = browseWeekBounds(new Date(2026, 8, 2), [new Date(2026, 11, 2)])
  assert.equal(new Date(bounds.maximum).getMonth(), 10)
})
test('move helper preserves future assignments not selected', () => {
  const result = moveToToday({ '2026-09-07': [item('B'), item('C'), item('D')] }, '2026-09-02', [item('B')])
  assert.deepEqual(result['2026-09-07'].map(x => x.id), ['C', 'D'])
})
test('selection never exceeds capacity', () => assert.equal(selectTodayCandidates([item('A'), item('B'), item('C')], 3, 2, order).length, 0))
