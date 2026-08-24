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

/**
 * Deterministic weekly planner. Prospects are first ordered by their distance
 * from the sales starting point, then distributed across the five workdays.
 * This keeps the weekly sequence nearest-to-farthest and balances the load.
 */
export function planWeeklyVisits(items: Prospect[], start: Coordinates | null, maxPerDay = MAX_VISITS_PER_DAY) {
  const eligible = stableSort(items.filter((item) => !['LOST', 'CONVERTED', 'WON'].includes(item.status)))
  const located = eligible.filter((item) => item.latitude != null && item.longitude != null)
  const withoutLocation = eligible.filter((item) => item.latitude == null || item.longitude == null)
  const plan: WeeklyPlan = Object.fromEntries(WORKING_DAYS.map((day) => [day, []]))
  const ordered = [...located].sort((a, b) => {
    const delta = distance(start, a) - distance(start, b)
    return delta || a.placeName.localeCompare(b.placeName) || a.id.localeCompare(b.id)
  })

  // Fill all weekdays as evenly as possible: 10 prospects => 2 per day.
  const dayCount = WORKING_DAYS.length
  const capacity = Math.min(maxPerDay, Math.ceil(ordered.length / dayCount))
  const remaining = [...ordered]
  for (let dayIndex = 0; dayIndex < dayCount && remaining.length; dayIndex += 1) {
    plan[WORKING_DAYS[dayIndex]] = remaining.splice(0, capacity)
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
