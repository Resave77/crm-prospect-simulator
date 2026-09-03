import type { Prospect } from '../types/crm'
import { haversineKm } from './maps'
import { ref } from 'vue'

export const WORKING_DAYS = ['Monday', 'Tuesday', 'Wednesday', 'Thursday', 'Friday'] as const
export const MAX_VISITS_PER_DAY = 5
export const DEFAULT_VISITS_PER_DAY = 2

export type Coordinates = { latitude: number; longitude: number }
export type PlannedVisitsByDate = Record<string, Prospect[]>
export type WeeklyPlan = PlannedVisitsByDate
export type DatePlan = PlannedVisitsByDate
export const sharedWeeklyPlan = ref<PlannedVisitsByDate>({})

export function formatDateKey(date: Date): string {
  return `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}-${String(date.getDate()).padStart(2, '0')}`
}

export function parseDateKey(key: string): Date | null {
  const match = /^(\d{4})-(\d{2})-(\d{2})$/.exec(key)
  if (!match) return null
  const date = new Date(Number(match[1]), Number(match[2]) - 1, Number(match[3]), 12)
  return date.getFullYear() === Number(match[1]) && date.getMonth() === Number(match[2]) - 1 && date.getDate() === Number(match[3]) ? date : null
}

export function startOfWorkWeek(date: Date): Date {
  const result = new Date(date.getFullYear(), date.getMonth(), date.getDate(), 12)
  result.setDate(result.getDate() - ((result.getDay() || 7) - 1))
  return result
}

export function getBusinessDatesForWeek(date: Date): string[] {
  const monday = startOfWorkWeek(date)
  return WORKING_DAYS.map((_, index) => { const day = new Date(monday); day.setDate(monday.getDate() + index); return formatDateKey(day) })
}

export function addCalendarDays(date: Date, days: number): Date {
  const result = new Date(date.getFullYear(), date.getMonth(), date.getDate(), 12)
  result.setDate(result.getDate() + days)
  return result
}

export function selectTodayCandidates<T extends { id: string }>(candidates: T[], existingCount: number, capacity: number, order: (items: T[]) => T[]): T[] {
  return order(candidates).slice(0, Math.max(0, capacity - existingCount))
}

export function moveToToday<T extends { id: string }>(plan: Record<string, T[]>, todayKey: string, selected: T[]): Record<string, T[]> {
  const selectedIds = new Set(selected.map(item => item.id))
  const next = Object.fromEntries(Object.entries(plan).map(([key, items]) => [key, key === todayKey ? items : items.filter(item => !selectedIds.has(item.id))])) as Record<string, T[]>
  next[todayKey] = [...(next[todayKey] ?? []), ...selected.filter(item => !(next[todayKey] ?? []).some(existing => existing.id === item.id))]
  return next
}

export function browseWeekBounds(currentWeek: Date, plannedWeekStarts: Date[], weeksBefore = 4, weeksAfter = 8) {
  const current = startOfWorkWeek(currentWeek)
  const minimum = addCalendarDays(current, -weeksBefore * 7).getTime()
  const maximum = Math.max(addCalendarDays(current, weeksAfter * 7).getTime(), ...plannedWeekStarts.map(start => startOfWorkWeek(start).getTime()))
  return { minimum, maximum }
}

export function formatWeekRange(date: Date): string {
  const dates = getBusinessDatesForWeek(date).map(key => parseDateKey(key)!)
  const first = dates[0]
  const last = dates[4]
  const fmt = new Intl.DateTimeFormat('id-ID', { day: 'numeric', month: 'short' })
  const firstText = fmt.format(first)
  const lastText = fmt.format(last)
  return first.getMonth() === last.getMonth() ? `${first.getDate()} – ${lastText}` : `${firstText} – ${lastText}`
}

export function isValidDatePlan(value: unknown): value is PlannedVisitsByDate {
  return Boolean(value && typeof value === 'object' && !Array.isArray(value) && Object.entries(value).every(([key, items]) =>
    Boolean(parseDateKey(key)) && Array.isArray(items) && items.every(item => Boolean(item && typeof item === 'object' && 'id' in item && typeof item.id === 'string'))))
}

function distance(from: Coordinates | null, item: Prospect): number {
  if (!from || item.latitude == null || item.longitude == null) return Number.POSITIVE_INFINITY
  return haversineKm(from.latitude, from.longitude, item.latitude, item.longitude)
}

function stableSort(items: Prospect[]) {
  return [...items].sort((a, b) => a.placeName.localeCompare(b.placeName) || a.id.localeCompare(b.id))
}

/** Build a deterministic route, filling each available day before moving on. */
export function planWeeklyVisits(
  items: Prospect[],
  start: Coordinates | null,
  maxPerDay = DEFAULT_VISITS_PER_DAY,
  startDayIndex = 0,
  existingPlan: WeeklyPlan = {},
) {
  const eligible = stableSort(items.filter((item) => !['LOST', 'CONVERTED', 'WON'].includes(item.status)))
  const preservedIds = new Set(Object.entries(existingPlan).filter(([_, value]) => {
    const dayIndex = WORKING_DAYS.indexOf(_ as typeof WORKING_DAYS[number])
    return dayIndex >= 0 && dayIndex < startDayIndex
  }).flatMap(([_, value]) => value.map(item => item.id)))
  const actionable = eligible.filter(item => !preservedIds.has(item.id))
  const located = actionable.filter((item) => item.latitude != null && item.longitude != null)
  const withoutLocation = actionable.filter((item) => item.latitude == null || item.longitude == null)
  const plan: WeeklyPlan = Object.fromEntries(WORKING_DAYS.map((day) => [day, []]))
  const orderedLocated = [...located].sort((a, b) => {
    const delta = distance(start, a) - distance(start, b)
    return delta || a.placeName.localeCompare(b.placeName) || a.id.localeCompare(b.id)
  })
  // Prospects without coordinates remain plannable after located prospects;
  // they are deterministic by name rather than silently becoming overflow.
  const ordered = [...orderedLocated, ...withoutLocation]

  const requestedCapacity = Number.isFinite(maxPerDay) ? Math.min(MAX_VISITS_PER_DAY, Math.max(1, Math.floor(maxPerDay))) : DEFAULT_VISITS_PER_DAY
  const firstDay = Math.max(0, Math.min(WORKING_DAYS.length - 1, Math.floor(startDayIndex)))
  const remaining = [...ordered]
  for (let dayIndex = firstDay; dayIndex < WORKING_DAYS.length && remaining.length; dayIndex += 1) {
    plan[WORKING_DAYS[dayIndex]] = remaining.splice(0, requestedCapacity)
  }

  return { plan, overflow: remaining }
}

/** Date-aware planner used when a route continues into a later work week. */
export function planVisitsByDate(items: Prospect[], start: Coordinates | null, maxPerDay = DEFAULT_VISITS_PER_DAY, startDate = new Date()) {
  const base = planWeeklyVisits(items, start, maxPerDay, 0)
  const ordered = WORKING_DAYS.flatMap(day => base.plan[day] ?? []).concat(base.overflow)
  const capacity = Number.isFinite(maxPerDay) ? Math.min(MAX_VISITS_PER_DAY, Math.max(1, Math.floor(maxPerDay))) : DEFAULT_VISITS_PER_DAY
  const plan: DatePlan = {}
  const cursor = new Date(startDate)
  cursor.setHours(12, 0, 0, 0)
  for (const item of ordered) {
    while (cursor.getDay() === 0 || cursor.getDay() === 6) cursor.setDate(cursor.getDate() + 1)
    const key = formatDateKey(cursor)
    const dayItems = plan[key] ?? []
    if (dayItems.length >= capacity) { cursor.setDate(cursor.getDate() + 1); while (cursor.getDay() === 0 || cursor.getDay() === 6) cursor.setDate(cursor.getDate() + 1) }
    const nextKey = formatDateKey(cursor)
    plan[nextKey] = [...(plan[nextKey] ?? []), item]
  }
  return { plan, overflow: [] as Prospect[] }
}

/** Orders a day strictly from nearest to farthest from the sales start point. */
export function routeDay(items: Prospect[], start: Coordinates | null) {
  return stableSort(items).sort((a, b) => {
    const delta = distance(start, a) - distance(start, b)
    return delta || a.placeName.localeCompare(b.placeName) || a.id.localeCompare(b.id)
  })
}
