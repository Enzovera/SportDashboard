// Pure view logic of the "Propozycje" screen: labels, diff values, layout of boxes, week before/after.
import type { Layout } from '@/composables/useLayout'
import type {
  DayPlan,
  DiffRow,
  Proposal,
  ProposalDecision,
  ProposalKind,
  ProposalOrigin,
  WeekDiff,
} from '@/api/proposals'
import { dayMonth, dayShort, localDate, num, time, weekday } from '@/lib/format'

export const KIND_LABEL: Record<ProposalKind, string> = {
  nutrition: 'Żywienie',
  strength: 'Siła',
  run: 'Bieg',
}

export const ORIGIN_LABEL: Record<ProposalOrigin, string> = {
  daily_run: 'run dzienny',
  weekly_run: 'run tygodniowy',
  chat: 'czat',
}

export const DECISION_LABEL: Record<ProposalDecision['decision'], string> = {
  accepted: 'przyjęta',
  rejected: 'odrzucona',
}

/** "dziś 06:31", "wczoraj 06:30", "23.10 06:30". `today` is the response date. */
export function when(iso: string, today: string): string {
  const day = localDate(iso)
  const yesterday = localDate(new Date(new Date(today + 'T12:00:00Z').getTime() - 86_400_000))
  const prefix = day === today ? 'dziś' : day === yesterday ? 'wczoraj' : dayMonth(day)
  return `${prefix} ${time(iso)}`
}

/** "dziś 06:31 · run dzienny". */
export function proposedMeta(p: { proposedAt: string; origin: ProposalOrigin }, today: string): string {
  return `${when(p.proposedAt, today)} · ${ORIGIN_LABEL[p.origin]}`
}

/** "1 propozycja czeka", "2 propozycje czekają", "5 propozycji czeka"; null when nothing waits. */
export function pendingText(n: number): string | null {
  if (n <= 0) return null
  const d = n % 10
  const dd = n % 100
  if (n === 1) return '1 propozycja czeka'
  if (d >= 2 && d <= 4 && (dd < 12 || dd > 14)) return `${n} propozycje czekają`
  return `${n} propozycji czeka`
}

/** One side of a diff row: numbers formatted, unit appended when asked; null -> "–". */
export function diffValue(v: number | string | null, unit: string | null, withUnit = true): string {
  if (v === null) return '–'
  const text = typeof v === 'number' ? num(v, Number.isInteger(v) ? 0 : 1) : v
  return withUnit && unit && typeof v === 'number' ? `${text} ${unit}` : text
}

export function rowChanged(r: DiffRow): boolean {
  return r.before !== r.after
}

/** Compact (fold/cover) diff: only what changes, changed values first as in the coach's title. */
export function changedRows(p: Proposal): DiffRow[] {
  return p.diff.rows.filter(rowChanged)
}

/** Desktop diff table heading: "Nd 01.11 · cel dnia". */
export function diffHeading(p: Proposal): string {
  return p.effectiveFrom ? `${dayShort(p.effectiveFrom)} · ${p.diff.heading}` : p.diff.heading
}

/** History, newest decision first, local decisions merged in (dedup by id, server copy wins). */
export function mergeHistory(server: ProposalDecision[], local: ProposalDecision[]): ProposalDecision[] {
  const byId = new Map<string, ProposalDecision>()
  for (const d of local) byId.set(d.id, d)
  for (const d of server) byId.set(d.id, d)
  return [...byId.values()].sort(
    (a, b) => new Date(b.decidedAt).getTime() - new Date(a.decidedAt).getTime() || a.id.localeCompare(b.id),
  )
}

export type FoldItem = { kind: 'card'; index: number } | { kind: 'empty' } | { kind: 'blocked' } | { kind: 'history' }

/**
 * Boxes per column. Fold (FoldPropozycje): cards alternate left/right, then "Zablokowane" left and the history
 * right. Cover: one column, cards, blocked, history.
 */
export function columns(
  layout: Exclude<Layout, 'desktop'>,
  pending: number,
  hasBlocked: boolean,
  hasHistory: boolean,
): FoldItem[][] {
  const cards: FoldItem[] = pending > 0 ? [...Array(pending).keys()].map((index) => ({ kind: 'card', index })) : [{ kind: 'empty' }]
  if (layout === 'cover') {
    return [[...cards, ...(hasBlocked ? [{ kind: 'blocked' } as const] : []), ...(hasHistory ? [{ kind: 'history' } as const] : [])]]
  }
  const left: FoldItem[] = cards.filter((_, i) => i % 2 === 0)
  const right: FoldItem[] = cards.filter((_, i) => i % 2 === 1)
  if (hasBlocked) left.push({ kind: 'blocked' })
  if (hasHistory) right.push({ kind: 'history' })
  return right.length ? [left, right] : [left]
}

export interface WeekRow {
  date: string
  /** "Pn 26". */
  label: string
  now: DayPlan
  after: DayPlan
  /** A still pending proposal changes this day: the red "needs a decision" row. */
  decide: boolean
  sessionChanged: boolean
  numbersChanged: boolean
}

/** Week before/after, applying only changes of proposals that still wait for a decision. */
export function weekRows(week: WeekDiff, pendingIds: ReadonlySet<string>): WeekRow[] {
  return week.days.map((d) => {
    const live = d.changes.filter((c) => pendingIds.has(c.proposalId))
    const after = live.length ? live[live.length - 1]!.after : d.now
    return {
      date: d.date,
      label: `${weekday(d.date)} ${Number(dayMonth(d.date).slice(0, 2))}`,
      now: d.now,
      after,
      decide: live.length > 0,
      sessionChanged: after.session !== d.now.session,
      numbersChanged: after.kcal !== d.now.kcal || after.carbsG !== d.now.carbsG,
    }
  })
}

/** Planned run km of the week (sum of plan values from the backend). */
export function weekKm(rows: WeekRow[], side: 'now' | 'after'): number {
  return rows.reduce((sum, r) => sum + (r[side].runKm ?? 0), 0)
}

/** "2 750 kcal · W 370". */
export function dayNumbers(p: DayPlan): string {
  return `${num(p.kcal)} kcal · W ${num(p.carbsG)}`
}
