import type { ProspectStatus } from '../types/crm'

export type ProspectTab = 'assigned' | 'visited' | 'followup' | 'won'

export const TAB_STAGE_MAP: Record<ProspectTab, ProspectStatus[]> = {
  assigned: ['NEW_LEAD'],
  visited: ['CONTACTED'],
  followup: ['INTERESTED', 'QUALIFIED', 'PROPOSAL_SENT', 'NEGOTIATION'],
  won: ['WON'],
}

export const TAB_LABELS: Record<ProspectTab, string> = {
  assigned: 'Assigned',
  visited: 'Visited',
  followup: 'Follow Up',
  won: 'Won',
}

export const TAB_ORDER: ProspectTab[] = ['assigned', 'visited', 'followup', 'won']

export function stageForTab(tab: ProspectTab): ProspectStatus[] {
  return TAB_STAGE_MAP[tab]
}

export function tabForStatus(status: ProspectStatus): ProspectTab | null {
  for (const tab of TAB_ORDER) {
    if (TAB_STAGE_MAP[tab].includes(status)) return tab
  }
  return null
}

export function pipelineStatusLabel(status: ProspectStatus): string {
  return status.replaceAll('_', ' ')
}

export function isActiveProspectStatus(status: ProspectStatus): boolean {
  return status !== 'LOST' && status !== 'CONVERTED'
}

export function isTerminalProspectStatus(status: ProspectStatus): boolean {
  return status === 'LOST' || status === 'CONVERTED'
}

export function firstNonEmptyTab(counts: Record<ProspectTab, number>): ProspectTab {
  for (const tab of TAB_ORDER) {
    if (counts[tab] > 0) return tab
  }
  return 'assigned'
}

export const EMPTY_MESSAGES: Record<ProspectTab, { title: string; subtitle: string }> = {
  assigned: {
    title: 'No assigned prospects',
    subtitle: 'New prospects assigned by Admin will appear here.',
  },
  visited: {
    title: 'No visited prospects',
    subtitle: 'Completed visit check-outs will appear here.',
  },
  followup: {
    title: 'No prospects awaiting follow-up',
    subtitle: 'Prospects that need further action will appear here.',
  },
  won: {
    title: 'No won prospects',
    subtitle: 'Successfully closed prospects will appear here.',
  },
}

export const PIPELINE_ROUTE_NAME = 'SalesPipeline'
