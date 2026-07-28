export function priceLevelLabel(level: string): string {
  const map: Record<string, string> = {
    PRICE_LEVEL_FREE: 'Free',
    PRICE_LEVEL_INEXPENSIVE: 'Inexpensive',
    PRICE_LEVEL_MODERATE: 'Moderate',
    PRICE_LEVEL_EXPENSIVE: 'Expensive',
    PRICE_LEVEL_VERY_EXPENSIVE: 'Very Expensive',
  }
  return map[level] ?? level
}

export function priceLevelSeverity(level: string): string {
  const map: Record<string, string> = {
    PRICE_LEVEL_FREE: 'success',
    PRICE_LEVEL_INEXPENSIVE: 'success',
    PRICE_LEVEL_MODERATE: 'info',
    PRICE_LEVEL_EXPENSIVE: 'warn',
    PRICE_LEVEL_VERY_EXPENSIVE: 'danger',
  }
  return map[level] ?? 'secondary'
}

export function businessStatusLabel(status: string): string {
  if (status === 'OPERATIONAL') return 'Open'
  if (status === 'CLOSED_TEMPORARILY') return 'Temporarily Closed'
  if (status === 'CLOSED_PERMANENTLY') return 'Permanently Closed'
  return status || 'Unknown'
}

export function businessStatusSeverity(status: string): string {
  if (status === 'OPERATIONAL') return 'success'
  if (status === 'CLOSED_TEMPORARILY') return 'warn'
  if (status === 'CLOSED_PERMANENTLY') return 'danger'
  return 'secondary'
}

export function stars(rating: number): string[] {
  const full = Math.floor(rating)
  const half = rating - full >= 0.5 ? 1 : 0
  const empty = 5 - full - half
  return [
    ...Array(full).fill('pi-star-fill'),
    ...Array(half).fill('pi-star-half-fill'),
    ...Array(empty).fill('pi-star'),
  ]
}
