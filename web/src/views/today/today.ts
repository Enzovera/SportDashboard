// Pure view logic of the "Dziś" screen: which sections appear where, and table rows derived from the contract.
import type { Layout } from '@/composables/useLayout'
import type { NutritionToday, TodayResponse, TodaySession, WeekGrid } from '@/api/types'
import { WEEKDAYS, dayShort, num } from '@/lib/format'

export type SectionKey =
  | 'coach'
  | 'session'
  | 'week'
  | 'tomorrow'
  | 'readiness'
  | 'nutrition'
  | 'supplements'
  | 'proposals'
  | 'weight'

// Order per design/screens/FoldDzis.html (two columns) and CoverDzis.html (one column).
const FOLD: SectionKey[][] = [
  ['coach', 'session', 'week', 'tomorrow'],
  ['readiness', 'nutrition', 'supplements', 'proposals', 'weight'],
]
const COVER: SectionKey[][] = [
  ['coach', 'session', 'readiness', 'nutrition', 'supplements', 'proposals', 'week', 'weight'],
]

function present(key: SectionKey, d: TodayResponse, layout: Layout): boolean {
  switch (key) {
    case 'coach':
      return d.coach !== null
    case 'session':
      return d.session !== null
    case 'tomorrow':
      return d.tomorrow !== null
    case 'nutrition':
      return d.nutrition !== null
    case 'supplements':
      // On the fold screen supplements are a row inside the nutrition box; standalone only without it.
      return d.supplements !== null && (layout !== 'fold' || d.nutrition === null)
    case 'proposals':
      // Zero pending proposals means nothing to decide: no red badge.
      return d.proposalsPending !== null && d.proposalsPending > 0
    case 'weight':
      return d.weight !== null
    case 'week':
    case 'readiness':
      return true
  }
}

/** Columns of section keys for fold/cover, without sections that have no data. */
export function sectionColumns(d: TodayResponse, layout: Layout): SectionKey[][] {
  const order = layout === 'fold' ? FOLD : COVER
  return order.map((col) => col.filter((key) => present(key, d, layout)))
}

export interface DesktopCell {
  keys: SectionKey[]
  /** Columns of the 12-column grid. */
  span: number
  /** Stretch to the row height (equal-height boxes side by side). */
  stretch: boolean
}

// design/screens/DesktopDzis.html: today's decision first (hero + coach), then eat / take / readiness,
// then the week with the small link rows. Cells flow in the grid, so missing sections close the gaps.
const DESKTOP: DesktopCell[] = [
  { keys: ['session'], span: 6, stretch: true },
  { keys: ['coach'], span: 6, stretch: true },
  { keys: ['nutrition'], span: 4, stretch: true },
  { keys: ['supplements'], span: 4, stretch: true },
  { keys: ['readiness'], span: 4, stretch: true },
  { keys: ['week'], span: 8, stretch: false },
  { keys: ['tomorrow', 'proposals', 'weight'], span: 4, stretch: false },
]

/** Desktop grid cells without sections that have no data; a lone half of a 6 | 6 pair takes the row. */
export function desktopCells(d: TodayResponse): DesktopCell[] {
  const cells = DESKTOP.map((c) => ({ ...c, keys: c.keys.filter((k) => present(k, d, 'desktop')) })).filter(
    (c) => c.keys.length > 0,
  )
  const pair = cells.filter((c) => c.keys[0] === 'session' || c.keys[0] === 'coach')
  if (pair.length === 1) pair[0]!.span = 12
  return cells
}

/** Header meta: "Nd 25.10 · 20 tyg. do maratonu". */
export function headerMeta(d: Pick<TodayResponse, 'date' | 'raceWeeksLeft'>): string {
  const day = dayShort(d.date)
  return d.raceWeeksLeft === null ? day : `${day} · ${d.raceWeeksLeft} tyg. do maratonu`
}

export function sessionMeta(action: TodaySession['action'], compact: boolean): string {
  switch (action) {
    case 'keep':
      return compact ? 'bez zmian' : 'bez zmian w planie'
    case 'modify':
      return 'zmienione przez trenera'
    case 'rest':
      return 'dzień wolny'
  }
}

/** "30 / 46 km", or "30 km" without a plan. */
export function weekKm(week: Pick<WeekGrid, 'runKmDone' | 'runKmPlanned'>): string {
  return week.runKmPlanned === null
    ? `${num(week.runKmDone)} km`
    : `${num(week.runKmDone)} / ${num(week.runKmPlanned)} km`
}

/** "Bieg 30 / 46 km · siła 3 / 3". */
export function weekTotals(week: WeekGrid): string {
  const strength =
    week.strengthPlanned === null ? `${week.strengthDone}` : `${week.strengthDone} / ${week.strengthPlanned}`
  return `Bieg ${weekKm(week)} · siła ${strength}`
}

export interface WeekCell {
  text: string
  /** Nothing that day: rendered as a muted dash. */
  empty: boolean
  /** Today's run: the red signal cell. */
  signal: boolean
  /** Planned, not done yet (future days). */
  planned: boolean
}

export interface WeekTable {
  heads: { label: string; today: boolean }[]
  run: WeekCell[]
  strength: WeekCell[]
}

export function weekTable(week: WeekGrid, today: string): WeekTable {
  const days = week.days.slice(0, 7)
  return {
    heads: days.map((d, i) => ({ label: WEEKDAYS[i] ?? '', today: d.date === today })),
    run: days.map((d) => {
      const isToday = d.date === today
      if (d.runKm === null) return { text: '–', empty: true, signal: false, planned: false }
      return { text: num(d.runKm), empty: false, signal: isToday, planned: d.runPlanned && !isToday }
    }),
    strength: days.map((d) =>
      d.strength === null
        ? { text: '–', empty: true, signal: false, planned: false }
        : { text: d.strength, empty: false, signal: false, planned: false },
    ),
  }
}

export interface NutritionLine {
  label: string
  unit: string
  eaten: string
  target: string
  left: string
}

export function nutritionLines(n: NutritionToday): NutritionLine[] {
  const rows: [string, string, NutritionToday['kcal']][] = [
    ['Kalorie', 'kcal', n.kcal],
    ['Białko', 'g', n.proteinG],
    ['Węglowodany', 'g', n.carbsG],
    ['Tłuszcz', 'g', n.fatG],
  ]
  return rows.map(([label, unit, r]) => ({
    label,
    unit,
    eaten: num(r.eaten),
    target: num(r.target),
    left: num(r.target - r.eaten),
  }))
}

type PluralForm = 'one' | 'few' | 'many'

/** Polish plural category: 1 / 2–4 (not 12–14) / the rest. */
export function pluralPl(n: number): PluralForm {
  if (n === 1) return 'one'
  const d = n % 10
  const dd = n % 100
  return d >= 2 && d <= 4 && (dd < 12 || dd > 14) ? 'few' : 'many'
}

/** "propozycje zmian czekają" (fold) / "propozycje zmian" (cover), inflected by count. */
export function proposalsText(n: number, compact: boolean): string {
  const forms: Record<PluralForm, [string, string]> = {
    one: ['propozycja zmiany', 'czeka'],
    few: ['propozycje zmian', 'czekają'],
    many: ['propozycji zmian', 'czeka'],
  }
  const [noun, verb] = forms[pluralPl(n)]
  return compact ? noun : `${noun} ${verb}`
}
