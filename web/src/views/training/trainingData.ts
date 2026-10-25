// Pure data shaping for the Training view: subtitles and chart descriptions. No DOM, no Vue.
import type { LoadPoint, LoadSummary, WeeklyKm } from '@/api/types'
import { dayMonth, num, signed } from '@/lib/format'

/** Relative change in percent; null when there is nothing to compare with. */
export function pctChange(current: number, previous: number | null): number | null {
  if (previous === null || previous === 0) return null
  return ((current - previous) / previous) * 100
}

/** Whole weeks covered by an inclusive date range, e.g. 03.08–25.10 -> 12. */
export function rangeWeeks(from: string, to: string): number {
  const days = (Date.parse(to + 'T00:00:00Z') - Date.parse(from + 'T00:00:00Z')) / 86_400_000 + 1
  return Math.max(1, Math.round(days / 7))
}

/** Polish plural of "tydzień" after a number: 1 tydzień, 2 tygodnie, 5 tygodni, 22 tygodnie. */
export function weeksWord(n: number): string {
  if (n === 1) return 'tydzień'
  const tens = n % 100
  const ones = n % 10
  return ones >= 2 && ones <= 4 && (tens < 12 || tens > 14) ? 'tygodnie' : 'tygodni'
}

/** "+7% vs 43 km" */
export function kmSub(load: LoadSummary): string | null {
  const pct = pctChange(load.weekKm, load.prevWeekKm)
  if (pct === null || load.prevWeekKm === null) return null
  return `${signed(pct)}% vs ${num(load.prevWeekKm)} km`
}

/** "+11 w 12 tyg." */
export function ctlSub(load: LoadSummary, weeks: number): string | null {
  return load.ctlDeltaRange === null ? null : `${signed(load.ctlDeltaRange)} w ${weeks} tyg.`
}

/** "−5 vs poprz. tydz." */
export function atlSub(load: LoadSummary): string | null {
  return load.atlDeltaWeek === null ? null : `${signed(load.atlDeltaWeek)} vs poprz. tydz.`
}

/** Bar lengths 0..1 relative to the largest value. */
export function barShares(values: number[]): number[] {
  const max = Math.max(0, ...values)
  return values.map((v) => (max > 0 ? Math.max(0, v) / max : 0))
}

function joinAnd(items: string[]): string {
  if (items.length <= 1) return items.join('')
  return items.slice(0, -1).join(', ') + ' i ' + items[items.length - 1]
}

export function weeklyKmAriaLabel(weeks: WeeklyKm[], todayPlannedKm: number | null): string {
  if (!weeks.length) return 'Kilometraż tygodniowy'
  const km = weeks.map((w) => w.km)
  const parts = [
    `Kilometraż tygodniowy od ${num(Math.min(...km))} do ${num(Math.max(...km))} km ` +
      `w ${weeks.length} ${weeksWord(weeks.length)}`,
  ]
  const deloads = weeks.filter((w) => w.deload).map((w) => dayMonth(w.weekStart))
  if (deloads.length) parts.push(`odciążenie w ${deloads.length === 1 ? 'tygodniu' : 'tygodniach'} ${joinAnd(deloads)}`)
  if (weeks.some((w) => w.plannedKm !== null)) {
    const lastPlan = [...weeks].reverse().find((w) => w.plannedKm !== null)
    if (lastPlan?.plannedKm != null) parts.push(`plan na ostatni tydzień ${num(lastPlan.plannedKm)} km`)
  }
  if (todayPlannedKm !== null) parts.push(`dziś planowane ${num(todayPlannedKm)} km`)
  return parts.join('; ')
}

export function loadAriaLabel(series: LoadPoint[]): string {
  const first = series[0]
  const last = series[series.length - 1]
  if (!first || !last) return 'Fitness i zmęczenie'
  const atl = series.map((p) => p.atl)
  return (
    `CTL z ${num(first.ctl)} do ${num(last.ctl)}; ` +
    `ATL od ${num(Math.min(...atl))} do ${num(Math.max(...atl))}, na koniec ${num(last.atl)}`
  )
}
