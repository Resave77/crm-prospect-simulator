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
 * Plan the week in near/far pairs. For example, six prospects become:
 * Mon [nearest, farthest], Tue [2nd nearest, 2nd farthest], Wed [3rd nearest,
 * 3rd farthest]. This keeps every day geographically balanced instead of
 * putting all nearby prospects on the first day.
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

  // Use two visits per day by default, up to five days. Respect an explicitly
  // smaller capacity by opening additional days where possible.
  const requestedCapacity = Number.isFinite(maxPerDay) ? Math.max(1, Math.floor(maxPerDay)) : 2
  const dayCount = Math.min(WORKING_DAYS.length, Math.max(Math.ceil(ordered.length / 2), Math.ceil(ordered.length / requestedCapacity)))
  const remaining = [...ordered]
  for (let dayIndex = 0; dayIndex < dayCount && remaining.length; dayIndex += 1) {
    const dayItems: Prospect[] = []
    const near = remaining.shift()
    if (near) dayItems.push(near)
    const farIndex = remaining.length - 1
    if (farIndex >= 0 && dayItems.length < requestedCapacity) dayItems.push(remaining.splice(farIndex, 1)[0])
    while (remaining.length && dayItems.length < requestedCapacity && dayIndex === dayCount - 1) dayItems.push(remaining.shift()!)
    plan[WORKING_DAYS[dayIndex]] = dayItems
  }

  // Records without coordinates remain visible as explicit unplanned items.
  return { plan, overflow: [...remaining, ...withoutLocation] }
}

/** Orders a day strictly from nearest to farthest from the sales start point. */
export function routeDay(items: Prospect[], start: Coordinates | null) {
  return stableSort(items).sort((a, b) => {
    const delta = distance(start, a) - distance(start, b)
    return delta || a.placeName.localeCompare(b.placeName) || a.id.localeCompare(b.id)
  })
}
