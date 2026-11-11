// Pure presentation logic of the plan view. The plan comes from the plan_session projection; this only
// picks words, rows and timeline states.
import type { BlockKind, DayType, PlanBlock, PlanDay, PlanWeek } from '@/api/plan'
import { dateRange, dayMonth, num } from '@/lib/format'

export const BLOCK_LABEL: Record<BlockKind, string> = {
  base: 'Baza',
  build: 'Budowa',
  peak: 'Szczyt',
  taper: 'Taper',
}

export const DAY_TYPE_LABEL: Record<DayType, string> = {
  strength_run: 'Siłowy + bieg',
  run: 'Bieg',
  long_run: 'Długie wybieganie',
  strength: 'Siłowy',
  rest: 'Odpoczynek',
}

function addDays(iso: string, days: number): string {
  const d = new Date(iso + 'T12:00:00Z')
  d.setUTCDate(d.getUTCDate() + days)
  return d.toISOString().slice(0, 10)
}

function daysBetween(a: string, b: string): number {
  return Math.round((Date.parse(b + 'T12:00:00Z') - Date.parse(a + 'T12:00:00Z')) / 86_400_000)
}

/** 14400 -> "4:00:00". */
export function raceTime(seconds: number): string {
  const s = Math.round(seconds)
  const h = Math.floor(s / 3600)
  const m = Math.floor((s % 3600) / 60)
  return `${h}:${String(m).padStart(2, '0')}:${String(s % 60).padStart(2, '0')}`
}

/** "2027-03-14" -> "14.03.2027". */
export function fullDate(iso: string): string {
  return `${iso.slice(8, 10)}.${iso.slice(5, 7)}.${iso.slice(0, 4)}`
}

/** "2026-11-01" -> "1" (day of month for the week tables). */
export function dayNumber(iso: string): string {
  return String(Number(iso.slice(8, 10)))
}

/**
 * Title and one detail line of a day.
 * withKm (narrow layouts without a km column): "16 km · strefa 2", "+ bieg łatwy 6 km".
 */
export function dayLines(day: PlanDay, withKm: boolean): { title: string; detail: string | null; rest: boolean } {
  const [main, ...rest] = day.sessions
  if (!main) return { title: 'Odpoczynek', detail: null, rest: true }
  const parts: string[] = []
  if (main.kind === 'run' && withKm && main.distanceKm !== null) parts.push(`${num(main.distanceKm)} km`)
  if (main.detail) parts.push(main.detail)
  const extras = rest.map((s) => {
    const km = withKm && s.distanceKm !== null ? ` ${num(s.distanceKm)} km` : ''
    return `+ ${s.title}${km}`
  })
  const detail = [parts.join(' · '), ...extras].filter(Boolean).join(' ')
  return { title: main.title, detail: detail || null, rest: false }
}

export type DoneTone = 'signal' | 'warn' | 'dim' | 'muted'

/** "Wykonane" cell of the current week: "6,0 km · 18 serii", "dziś", "–", "pominięte". */
export function doneCell(day: PlanDay, today: string): { text: string; tone: DoneTone } {
  if (day.date === today) return { text: 'dziś', tone: 'signal' }
  if (day.date > today || !day.done) return { text: '–', tone: 'muted' }
  const parts: string[] = []
  if (day.done.runKm !== null) parts.push(`${num(day.done.runKm, 1)} km`)
  if (day.done.strengthSets !== null) parts.push(`${day.done.strengthSets} serii`)
  if (parts.length) return { text: parts.join(' · '), tone: 'dim' }
  return day.sessions.length ? { text: 'pominięte', tone: 'warn' } : { text: '–', tone: 'muted' }
}

/** Days with pending proposals ("do decyzji"). */
export function decideCount(week: PlanWeek): number {
  return week.days.filter((d) => d.proposals.length > 0).length
}

export function isCurrentWeek(week: PlanWeek, today: string): boolean {
  return week.from <= today && today <= week.to
}

/** "Ten tydzień · 19–25.10" / "Tydzień 26.10–01.11". */
export function weekTitle(week: PlanWeek, today: string): string {
  const range = dateRange(week.from, week.to)
  return isCurrentWeek(week, today) ? `Ten tydzień · ${range}` : `Tydzień ${range}`
}

/** "bieg 30 / 46 km · siła 3 / 3" for a running week, "Baza · tydz. 1 · 46 km" for an upcoming one. */
export function weekMeta(week: PlanWeek): string {
  if (week.doneKm !== null) {
    const strength = week.plannedStrength > 0 ? ` · siła ${week.doneStrength ?? 0} / ${week.plannedStrength}` : ''
    return `bieg ${num(week.doneKm)} / ${num(week.plannedKm)} km${strength}`
  }
  const block = week.block ? `${BLOCK_LABEL[week.block]}${week.weekInBlock ? ` · tydz. ${week.weekInBlock}` : ''} · ` : ''
  return `${block}${num(week.plannedKm)} km`
}

/** Narrow layouts show the week that still has days ahead first; on its last day the next week leads. */
export function orderWeeks(current: PlanWeek | null, next: PlanWeek | null, today: string): PlanWeek[] {
  const weeks = [current, next].filter((w): w is PlanWeek => w !== null)
  if (current && next && today >= current.to) return [next, current]
  return weeks
}

export type SegmentState = 'past' | 'now' | 'block' | 'future'

export interface TimelineBlock {
  block: PlanBlock
  label: string
  /** "3 tyg. · 22.02–14.03" */
  dates: string
  /** "22.02–14.03": the horizontal axis shows the weeks next to the name, so a narrow block never wraps mid-date. */
  range: string
  current: boolean
  /** One entry per week of the block. */
  segments: SegmentState[]
}

export interface Timeline {
  /** "Dziś · 25.10 · Baza od jutra", "Dziś · 12.11 · Baza, tydz. 3 z 6". */
  now: string
  /** Today falls before the first block (the "now" mark sits at the start). */
  beforeStart: boolean
  blocks: TimelineBlock[]
  totalWeeks: number
  meta: string
}

/** Macrocycle timeline: the block holding today (or the next one before the start) is the current one. */
export function timeline(blocks: PlanBlock[], today: string): Timeline {
  const first = blocks[0]
  const lastBlock = blocks[blocks.length - 1]
  let currentIdx = blocks.findIndex((b) => b.from <= today && today <= b.to)
  const beforeStart = !!first && today < first.from
  if (beforeStart) currentIdx = 0

  let now = `Dziś · ${dayMonth(today)}`
  const cur = blocks[currentIdx]
  if (cur && beforeStart) {
    now += ` · ${BLOCK_LABEL[cur.kind]} ${cur.from === addDays(today, 1) ? 'od jutra' : `od ${dayMonth(cur.from)}`}`
  } else if (cur) {
    const week = Math.floor(daysBetween(cur.from, today) / 7) + 1
    now += ` · ${BLOCK_LABEL[cur.kind]}, tydz. ${week} z ${cur.weeks}`
  }

  const items = blocks.map((b, i): TimelineBlock => {
    const segments: SegmentState[] = []
    for (let w = 0; w < b.weeks; w++) {
      const wFrom = addDays(b.from, w * 7)
      const wTo = addDays(wFrom, 6)
      if (wTo < today) segments.push('past')
      else if (wFrom <= today && today <= wTo) segments.push('now')
      else segments.push(i === currentIdx ? 'block' : 'future')
    }
    return {
      block: b,
      label: BLOCK_LABEL[b.kind],
      dates: `${b.weeks} tyg. · ${dateRange(b.from, b.to)}`,
      range: dateRange(b.from, b.to),
      current: i === currentIdx,
      segments,
    }
  })

  const totalWeeks = blocks.reduce((s, b) => s + b.weeks, 0)
  const meta = first && lastBlock ? `${totalWeeks} tyg. · ${dayMonth(first.from)}–${dayMonth(lastBlock.to)}` : ''
  return { now, beforeStart, blocks: items, totalWeeks, meta }
}
