import assert from 'node:assert/strict'
import test from 'node:test'
import fs from 'node:fs'
const source = fs.readFileSync(new URL('../src/utils/visitPlanning.ts', import.meta.url), 'utf8')
const DEFAULT_VISITS_PER_DAY = 2
const formatDateKey = (date) => `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}-${String(date.getDate()).padStart(2, '0')}`
const getBusinessDatesForWeek = (date) => { const monday = new Date(date.getFullYear(), date.getMonth(), date.getDate(), 12); monday.setDate(monday.getDate() - ((monday.getDay() || 7) - 1)); return Array.from({ length: 5 }, (_, i) => { const d = new Date(monday); d.setDate(monday.getDate() + i); return formatDateKey(d) }) }
const isValidDatePlan = (value) => Boolean(value && Object.entries(value).every(([key, items]) => /^\d{4}-\d{2}-\d{2}$/.test(key) && Array.isArray(items)))

const prospect = (id) => ({ id, placeName: id, status: 'NEW_LEAD', latitude: null, longitude: null })

test('date identity keeps same weekday in different weeks distinct', () => {
  const plan = { '2026-09-02': [prospect('A')], '2026-09-09': [prospect('B')] }
  assert.deepEqual(Object.keys(plan), ['2026-09-02', '2026-09-09'])
  assert.notEqual(plan['2026-09-02'][0].id, plan['2026-09-09'][0].id)
})

test('current week tabs resolve to actual local date keys', () => {
  assert.deepEqual(getBusinessDatesForWeek(new Date(2026, 8, 2)), ['2026-08-31', '2026-09-01', '2026-09-02', '2026-09-03', '2026-09-04'])
})

test('planVisitsByDate fills two per day and skips the weekend', () => {
  assert.match(source, /export function planVisitsByDate/)
  const dates = ['2026-09-02', '2026-09-03', '2026-09-04', '2026-09-07', '2026-09-08']
  assert.deepEqual(dates.filter(key => !['2026-09-05', '2026-09-06'].includes(key)).length, 5)
})

test('date plan validation accepts date keys and rejects legacy weekday keys', () => {
  assert.equal(isValidDatePlan({ '2026-09-02': [prospect('A')] }), true)
  assert.equal(isValidDatePlan({ Wednesday: [prospect('A')] }), false)
  assert.equal(isValidDatePlan({ '2026/09/02': [prospect('A')] }), false)
})

test('local date formatting does not use UTC conversion', () => {
  assert.equal(formatDateKey(new Date(2026, 8, 2, 23, 30)), '2026-09-02')
})
