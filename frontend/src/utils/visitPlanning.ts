import type { Prospect } from '../types/crm'
import { haversineKm } from './maps'

export const WORKING_DAYS = ['Monday', 'Tuesday', 'Wednesday', 'Thursday', 'Friday'] as const
export const MAX_VISITS_PER_DAY = 5

export type Coordinates = { latitude: number; longitude: number }
export type WeeklyPlan = Record<string, Prospect[]>

function distance(from: Coordinates | null, item: Prospect): number {
  if (!from || item.latitude == null || item.longitude == null) return Number.POSITIVE_INFINITY
  return haversineKm(from.latitude, from.longitude, item.latitude, item.longitude)
}

function stableSort(items: Prospect[]) {
  return [...items].sort((a, b) => a.placeName.localeCompare(b.placeName) || a.id.localeCompare(b.id))
}

/** Deterministic geographic planner. It never drops eligible records: overflow is returned separately. */
export function planWeeklyVisits(items: Prospect[], start: Coordinates | null, maxPerDay = MAX_VISITS_PER_DAY) {
  const eligible = stableSort(items.filter((item) => !['LOST', 'CONVERTED', 'WON'].includes(item.status)))
  const located = eligible.filter((item) => item.latitude != null && item.longitude != null)
  const withoutLocation = eligible.filter((item) => item.latitude == null || item.longitude == null)
  const remaining = [...located]
  const plan: WeeklyPlan = Object.fromEntries(WORKING_DAYS.map((day) => [day, []]))
  let anchor = start
  // Use only the number of weekdays needed by the workload. The remainder is
  // placed at the edges so 8 prospects become 3 / 2 / 3, while preserving
  // geographic progression and leaving unused weekdays empty.
  const activeDayCount = Math.min(WORKING_DAYS.length, Math.max(1, Math.ceil(remaining.length / 3)))
  const baseCapacity = Math.floor(remaining.length / activeDayCount)
  const extraCapacity = remaining.length % activeDayCount
  const capacities = Array.from({ length: activeDayCount }, (_, index) => {
    const getsExtra = index < Math.ceil(extraCapacity / 2) || index >= activeDayCount - Math.floor(extraCapacity / 2)
    return Math.min(maxPerDay, baseCapacity + (getsExtra ? 1 : 0))
  })

  for (let dayIndex = 0; dayIndex < activeDayCount; dayIndex += 1) {
    const day = WORKING_DAYS[dayIndex]
    const capacity = Math.min(capacities[dayIndex], remaining.length)
    const dayItems: Prospect[] = []
    for (let i = 0; i < capacity; i += 1) {
      const next = remaining.reduce<Prospect | null>((best, candidate) => {
        if (!best) return candidate
        const candidateDistance = distance(anchor, candidate)
        const bestDistance = distance(anchor, best)
        return candidateDistance < bestDistance || (candidateDistance === bestDistance && candidate.placeName.localeCompare(best.placeName) < 0) ? candidate : best
      }, null)
      if (!next) break
      dayItems.push(next)
      remaining.splice(remaining.indexOf(next), 1)
      anchor = { latitude: next.latitude!, longitude: next.longitude! }
    }
    plan[day] = dayItems
    if (!remaining.length) break
  }

  // Records without coordinates remain visible as explicit unplanned items.
  return { plan, overflow: [...remaining, ...withoutLocation] }
}

/** Orders a single day from the sales start point using nearest-next-stop. */
export function routeDay(items: Prospect[], start: Coordinates | null) {
  const remaining = stableSort(items)
  const ordered: Prospect[] = []
  let anchor = start
  while (remaining.length) {
    const next = remaining.reduce<Prospect | null>((best, candidate) => {
      if (!best) return candidate
      const candidateDistance = distance(anchor, candidate)
      const bestDistance = distance(anchor, best)
      return candidateDistance < bestDistance || (candidateDistance === bestDistance && candidate.placeName.localeCompare(best.placeName) < 0) ? candidate : best
    }, null)
    if (!next) break
    ordered.push(next)
    remaining.splice(remaining.indexOf(next), 1)
    anchor = next.latitude != null && next.longitude != null ? { latitude: next.latitude, longitude: next.longitude } : anchor
  }
  return ordered
}
