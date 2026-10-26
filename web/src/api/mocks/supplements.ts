import type { SupplementsToday } from '../types'
import type {
  RegularityDay,
  SupplementItem,
  SupplementLogStatus,
  SupplementRegularity,
  SupplementsResponse,
  SupplementWhen,
} from '../supplements'

// Values from design/screens/FoldSuplementy.html and DesktopSuplementy.html (Nd 25.10, long run).
// `full` is in-memory state: logSupplement() and decideReminderProposal() change it until the page reloads.

const DAY: Record<string, RegularityDay> = { x: 'taken', o: 'missed', '-': 'na' }

/** 30 chars, oldest first: x = taken, o = missed, - = rule did not apply. */
function regularity(
  supplementId: string,
  name: string,
  when: SupplementWhen,
  strip: string,
  warn = false,
): SupplementRegularity {
  const days = [...strip].map((c) => DAY[c] ?? 'na')
  const taken = days.filter((d) => d === 'taken').length
  const applicable = days.filter((d) => d !== 'na').length
  return {
    supplementId,
    name,
    when,
    days,
    taken,
    applicable,
    pct: applicable ? Math.round((taken / applicable) * 100) : 0,
    tone: warn ? 'warn' : 'neutral',
  }
}

function initialChecklist(): SupplementItem[] {
  return [
    {
      id: 'd3k2',
      name: 'Witamina D3 + K2',
      dose: { amount: 2000, unit: 'IU' },
      note: 'z posiłkiem',
      countedInFitatu: false,
      slot: 'morning',
      remindAt: '07:30',
      status: 'taken',
      loggedAt: '2026-10-25T07:42:00+01:00',
    },
    {
      id: 'omega3',
      name: 'Omega-3',
      dose: { amount: 2, unit: 'capsule' },
      note: 'z posiłkiem',
      countedInFitatu: false,
      slot: 'morning',
      remindAt: '07:30',
      status: 'taken',
      loggedAt: '2026-10-25T07:42:00+01:00',
    },
    {
      id: 'caffeine',
      name: 'Kofeina',
      dose: { amount: 200, unit: 'mg' },
      note: '60 min przed',
      countedInFitatu: false,
      slot: 'before',
      remindAt: null,
      status: 'pending',
      loggedAt: null,
    },
    {
      id: 'electrolytes',
      name: 'Elektrolity',
      dose: { amount: 1, unit: 'sachet' },
      note: 'bieg > 90 min',
      countedInFitatu: false,
      slot: 'during',
      remindAt: null,
      status: 'pending',
      loggedAt: null,
    },
    {
      id: 'creatine',
      name: 'Kreatyna',
      dose: { amount: 5, unit: 'g' },
      note: 'pora dowolna',
      countedInFitatu: false,
      slot: 'after',
      remindAt: null,
      status: 'pending',
      loggedAt: null,
    },
    {
      id: 'whey',
      name: 'Białko serwatkowe',
      dose: { amount: 30, unit: 'g' },
      note: null,
      countedInFitatu: true,
      slot: 'after',
      remindAt: null,
      status: 'pending',
      loggedAt: null,
    },
    {
      id: 'magnesium',
      name: 'Magnez',
      dose: { amount: 300, unit: 'mg' },
      note: 'przed snem',
      countedInFitatu: false,
      slot: 'evening',
      remindAt: '20:00',
      status: 'pending',
      loggedAt: null,
    },
  ]
}

const state: SupplementsResponse = {
  date: '2026-10-25',
  dayType: 'long_run',
  checklist: initialChecklist(),
  regularity: {
    from: '2026-09-25',
    to: '2026-10-24',
    items: [
      regularity('electrolytes', 'Elektrolity', 'long_run', '--x------x------x------x------'),
      regularity('d3k2', 'Witamina D3 + K2', 'daily', 'xxxxxxxxxxxxxxxoxxxxxxxxxxxxxx'),
      regularity('creatine', 'Kreatyna', 'daily', 'xxxxxxxxoxxxxxxxxxxxxxoxxxxxxx'),
      regularity('whey', 'Białko serwatkowe', 'training_day', 'xoxxxx-xxxoxx-xxxoxx-xxxxxx-ox'),
      regularity('omega3', 'Omega-3', 'daily', 'xxxoxxxxoxxxxxxoxxxxxoxxxxoxxx'),
      regularity('caffeine', 'Kofeina', 'quality_run', '--x-x----x-o----x-x----x-o----'),
      regularity('magnesium', 'Magnez', 'daily', 'oxxxooxxxxxooxxxxxooxxxxooxxxx', true),
    ],
  },
  stock: [
    { supplementId: 'd3k2', name: 'Witamina D3 + K2', left: { amount: 12, unit: 'capsule' }, lasts: { count: 12, per: 'daily' }, reorder: true },
    { supplementId: 'electrolytes', name: 'Elektrolity', left: { amount: 6, unit: 'sachet' }, lasts: { count: 6, per: 'long_run' }, reorder: false },
    { supplementId: 'magnesium', name: 'Magnez', left: { amount: 20, unit: 'tablet' }, lasts: { count: 20, per: 'daily' }, reorder: false },
    { supplementId: 'omega3', name: 'Omega-3', left: { amount: 54, unit: 'capsule' }, lasts: { count: 27, per: 'daily' }, reorder: false },
    { supplementId: 'creatine', name: 'Kreatyna', left: { amount: 205, unit: 'g' }, lasts: { count: 41, per: 'daily' }, reorder: false },
  ],
  proposal: {
    id: 'p-magnesium-time',
    supplementId: 'magnesium',
    supplementName: 'Magnez',
    text: 'Magnez pomijasz głównie po wieczornych treningach (7 z 9 pominięć). Przesunąć przypomnienie na 21:30?',
    currentTime: '20:00',
    proposedTime: '21:30',
  },
}

export function supplementsFull(): SupplementsResponse {
  return state
}

/** Before stage 2 nothing is set up yet: empty checklist, no history, no stock. */
export const supplementsStage1: SupplementsResponse = {
  date: '2026-10-25',
  dayType: null,
  checklist: [],
  regularity: { from: '2026-09-25', to: '2026-10-24', items: [] },
  stock: [],
  proposal: null,
}

/** Counter for the Jedzenie | Suplementy switch and the Today screen. */
export function supplementsSummary(): SupplementsToday {
  const list = state.checklist
  const next = list.find((i) => i.status === 'pending')
  return { taken: list.filter((i) => i.status === 'taken').length, total: list.length, next: next?.name ?? null }
}

export function supplementsLogMock(id: string, status: SupplementLogStatus): SupplementItem {
  const item = state.checklist.find((i) => i.id === id)
  if (!item) throw new Error(`/api/supplements/${id}/log: HTTP 404`)
  item.status = status
  item.loggedAt = new Date().toISOString()
  return item
}

export function supplementsProposalDecideMock(id: string, decision: 'accept' | 'reject'): void {
  const p = state.proposal
  if (!p || p.id !== id) throw new Error(`/api/proposals/${id}/${decision}: HTTP 404`)
  if (decision === 'accept') {
    const item = state.checklist.find((i) => i.id === p.supplementId)
    if (item) item.remindAt = p.proposedTime
  }
  state.proposal = null
}
