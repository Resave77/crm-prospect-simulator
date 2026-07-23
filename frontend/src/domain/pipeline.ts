import type { Prospect, ProspectStatus } from '../types/crm'

export const ACTIVE_PIPELINE_STAGES: ProspectStatus[] = ['NEW_LEAD', 'CONTACTED', 'INTERESTED', 'QUALIFIED', 'PROPOSAL_SENT', 'NEGOTIATION']
export const PIPELINE_STAGES: ProspectStatus[] = [...ACTIVE_PIPELINE_STAGES, 'WON']
export const BOARD_STATUSES: ProspectStatus[] = [...PIPELINE_STAGES, 'LOST']

export function nextStage(status: ProspectStatus) {
  const index = PIPELINE_STAGES.indexOf(status)
  return index >= 0 && index < PIPELINE_STAGES.length - 1 ? PIPELINE_STAGES[index + 1] : null
}

export function previousStage(status: ProspectStatus) {
  const index = PIPELINE_STAGES.indexOf(status)
  return index > 0 ? PIPELINE_STAGES[index - 1] : null
}

export interface PipelineFilters {
  salesExecutiveId: string
  industryGroup: string
  status: string
  search: string
}

export function filterProspects(items: Prospect[], filters: PipelineFilters) {
  const search = filters.search.trim().toLowerCase()
  return items.filter((item) =>
    (!filters.salesExecutiveId || item.assignedSalesExecutiveId === filters.salesExecutiveId) &&
    (!filters.industryGroup || item.industryGroup === filters.industryGroup) &&
    (!filters.status || item.status === filters.status) &&
    (!search || `${item.placeName} ${item.formattedAddress}`.toLowerCase().includes(search)),
  )
}
