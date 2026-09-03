export type CompletionVisit = { prospectId?: string; checkOutAt?: string | null }

export function completedVisitsByProspectId<T extends CompletionVisit>(visits: T[]): Map<string, T> {
  const result = new Map<string, T>()
  for (const visit of visits) if (visit.prospectId && visit.checkOutAt) {
    const previous = result.get(visit.prospectId)
    if (!previous || new Date(visit.checkOutAt).getTime() > new Date(previous.checkOutAt ?? 0).getTime()) result.set(visit.prospectId, visit)
  }
  return result
}

export function pendingRoute<T extends { id: string }>(items: T[], completedIds: Set<string>) {
  return items.filter(item => !completedIds.has(item.id))
}
