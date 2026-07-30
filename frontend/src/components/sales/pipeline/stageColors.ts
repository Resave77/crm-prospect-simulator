import type { ProspectStatus } from '../../../types/crm'

interface StageTone {
  bg: string
  fg: string
  border: string
}

const TONES: Record<ProspectStatus, StageTone> = {
  NEW_LEAD: { bg: '#e8edf3', fg: '#475569', border: '#cbd5e1' },
  CONTACTED: { bg: '#e0f7fa', fg: '#0e7490', border: '#a5f3fc' },
  INTERESTED: { bg: '#e8eaf6', fg: '#3949ab', border: '#c5cae9' },
  QUALIFIED: { bg: '#e3f2fd', fg: '#1565c0', border: '#90caf9' },
  PROPOSAL_SENT: { bg: '#ede7f6', fg: '#6a1b9a', border: '#d1c4e9' },
  NEGOTIATION: { bg: '#fff8e1', fg: '#e65100', border: '#ffe082' },
  WON: { bg: '#e8f5e9', fg: '#2e7d32', border: '#a5d6a7' },
  LOST: { bg: '#fce4ec', fg: '#c62828', border: '#ef9a9a' },
  CONVERTED: { bg: '#f3e5f5', fg: '#7b1fa2', border: '#ce93d8' },
}

export function stageTone(status: ProspectStatus): StageTone {
  return TONES[status] ?? TONES.NEW_LEAD
}

export function stageLabel(status: ProspectStatus): string {
  return status.replaceAll('_', ' ').toLowerCase().replace(/\b\w/g, c => c.toUpperCase())
}

export function stageInitial(status: ProspectStatus): string {
  const map: Partial<Record<ProspectStatus, string>> = {
    NEW_LEAD: 'NL',
    CONTACTED: 'C',
    INTERESTED: 'I',
    QUALIFIED: 'Q',
    PROPOSAL_SENT: 'PS',
    NEGOTIATION: 'N',
    WON: 'W',
    LOST: 'L',
    CONVERTED: 'CV',
  }
  return map[status] ?? status[0]
}
