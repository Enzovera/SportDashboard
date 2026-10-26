// Pure presentation logic of the body view. Every number comes from the API; this only picks words,
// rows and chart geometry.
import type { BodyComposition, BodyWeek, WeighIn } from '@/api/body'
import { dayMonth, localDate, num, weekday } from '@/lib/format'

const WEEKDAY_FULL: Record<string, string> = {
  Pn: 'poniedziałek',
  Wt: 'wtorek',
  Śr: 'środa',
  Cz: 'czwartek',
  Pt: 'piątek',
  Sb: 'sobota',
  Nd: 'niedziela',
}

/** "2026-10-25" + 1 -> "2026-10-26" (calendar arithmetic on plain dates). */
export function addDays(iso: string, days: number): string {
  const d = new Date(iso + 'T12:00:00Z')
  d.setUTCDate(d.getUTCDate() + days)
  return d.toISOString().slice(0, 10)
}

/** 0.15 -> "0,15", 0.2 -> "0,2". */
function kgSmall(value: number): string {
  return num(value, 2).replace(/(,\d)0$/, '$1')
}

/** Title of the weekly chart, e.g. "Waga spada o ok. 0,15 kg tygodniowo". */
export function trendTitle(kgPerWeek: number | null): string {
  if (kgPerWeek === null) return 'Waga tydzień po tygodniu'
  if (Math.abs(kgPerWeek) < 0.05) return 'Waga stabilna'
  const verb = kgPerWeek < 0 ? 'spada' : 'rośnie'
  return `Waga ${verb} o ok. ${kgSmall(Math.abs(kgPerWeek))} kg tygodniowo`
}

export const MUSCLE_TREND_LABEL: Record<NonNullable<BodyComposition['muscleTrend']>, string> = {
  up: 'rośnie',
  stable: 'stabilna',
  down: 'spada',
}

/** Missing weigh-ins this week: "brak: środa", "brak: środa, piątek", "brak: 3 dni"; null when complete. */
export function missingLabel(dates: string[]): string | null {
  if (dates.length === 0) return null
  if (dates.length > 2) return `brak: ${dates.length} dni`
  return 'brak: ' + dates.map((d) => WEEKDAY_FULL[weekday(d)] ?? d).join(', ')
}

/** "−1,7 kg w 12 tyg." */
export function changeLabel(changeKg: number | null, weeks: number): string | null {
  if (changeKg === null) return null
  const sign = changeKg > 0 ? '+' : ''
  return `${sign}${num(changeKg, 1)} kg w ${weeks} tyg.`
}

export interface MetricCellData {
  label: string
  value: string
  unit: string
  sub: string | null
}

/**
 * Cells of the composition box. Without a 7-day average the latest weigh-in stands in; fat and muscle only
 * when the scale sends them; with no weigh-in this week the count covers the whole range.
 */
export function compositionCells(c: BodyComposition, rangeWeeks: number): MetricCellData[] {
  const cells: MetricCellData[] = []
  if (c.avg7dKg !== null) {
    cells.push({ label: 'Waga · śr. 7 dni', value: num(c.avg7dKg, 1), unit: 'kg', sub: changeLabel(c.avg7dChangeKg, rangeWeeks) })
  } else if (c.lastKg !== null && c.measuredAt !== null) {
    cells.push({ label: 'Ostatnie ważenie', value: num(c.lastKg, 1), unit: 'kg', sub: dayMonth(localDate(c.measuredAt)) })
  }
  if (c.fatPct !== null) cells.push({ label: 'Tkanka tłuszczowa', value: num(c.fatPct, 1), unit: '%', sub: 'szacunek wagi' })
  if (c.muscleKg !== null) {
    const sub = c.muscleTrend ? MUSCLE_TREND_LABEL[c.muscleTrend] : null
    cells.push({ label: 'Masa mięśniowa', value: num(c.muscleKg, 1), unit: 'kg', sub })
  }
  if (c.weekCount > 0) {
    cells.push({ label: 'Ważenia · tydzień', value: String(c.weekCount), unit: `/ ${c.weekDays}`, sub: missingLabel(c.weekMissing) })
  } else {
    cells.push({ label: `Ważenia · ${rangeWeeks} tyg.`, value: String(c.rangeCount), unit: '', sub: null })
  }
  return cells
}

export interface WeighInRow {
  date: string
  weighIn: WeighIn | null
}

/**
 * Rows of the weigh-ins table, newest first.
 * Regular weighing (>= `regular` weigh-ins in the last 7 days): the 7 calendar days, gaps shown as "brak"
 * (today without a weigh-in yet is skipped). Sparse weighing: just the weigh-ins, up to 7.
 */
export function weighInRows(today: string, weighIns: WeighIn[], regular = 4): WeighInRow[] {
  const byDate = new Map<string, WeighIn>()
  for (const w of weighIns) {
    const d = localDate(w.measuredAt)
    if (!byDate.has(d)) byDate.set(d, w)
  }
  const windowStart = addDays(today, -6)
  const inWindow = [...byDate.keys()].filter((d) => d >= windowStart && d <= today).length
  if (inWindow < regular) {
    return [...byDate.entries()]
      .sort(([a], [b]) => (a < b ? 1 : -1))
      .slice(0, 7)
      .map(([date, weighIn]) => ({ date, weighIn }))
  }
  const rows: WeighInRow[] = []
  for (let i = 0; i < 7; i++) {
    const date = addDays(today, -i)
    const weighIn = byDate.get(date) ?? null
    if (i === 0 && !weighIn) continue
    rows.push({ date, weighIn })
  }
  return rows
}

// ---------- weekly weight chart (plain SVG, drawn at the box's pixel width) ----------

export interface ChartPad {
  left: number
  right: number
  top: number
  bottom: number
}

export interface WeightChart {
  width: number
  height: number
  plotLeft: number
  plotRight: number
  baseline: number
  yTicks: { y: number; label: string }[]
  xLabels: { x: number; label: string; anchor: 'start' | 'middle' | 'end' }[]
  whiskers: { x: number; yMin: number; yMax: number }[]
  /** Polylines of consecutive weeks with data; gaps are not bridged. */
  segments: string[]
  points: { x: number; y: number; current: boolean }[]
  /** Value label next to the latest point. */
  last: { x: number; y: number; label: string; current: boolean } | null
}

export const DEFAULT_PAD: ChartPad = { left: 30, right: 44, top: 10, bottom: 26 }

function inWeek(weekStart: string, date: string): boolean {
  return date >= weekStart && date < addDays(weekStart, 7)
}

/** Geometry of the weekly chart: average as squares + line, min–max of the week as whiskers. */
export function weightChart(
  weeks: BodyWeek[],
  today: string,
  width: number,
  height: number,
  pad: ChartPad = DEFAULT_PAD,
  minLabelGap = 52,
): WeightChart | null {
  const withData = weeks.filter((w) => w.avgKg !== null)
  if (weeks.length === 0 || withData.length === 0 || width <= pad.left + pad.right) return null

  const lows = withData.map((w) => w.minKg ?? w.avgKg!)
  const highs = withData.map((w) => w.maxKg ?? w.avgKg!)
  let lo = Math.floor(Math.min(...lows))
  let hi = Math.ceil(Math.max(...highs))
  if (hi - lo < 2) {
    const mid = (Math.min(...lows) + Math.max(...highs)) / 2
    lo = Math.floor(mid - 1)
    hi = lo + 2
    if (Math.max(...highs) > hi) hi += 1
  }

  const plotLeft = pad.left
  const plotRight = width - pad.right
  const plotTop = pad.top
  const baseline = height - pad.bottom
  const plotBottom = baseline - 8
  const n = weeks.length
  const x = (i: number) => (n === 1 ? (plotLeft + plotRight) / 2 : plotLeft + (i * (plotRight - plotLeft)) / (n - 1))
  const y = (v: number) => plotTop + ((hi - v) / (hi - lo)) * (plotBottom - plotTop)
  const r1 = (v: number) => Math.round(v * 10) / 10

  const step = hi - lo > 6 ? 2 : 1
  const yTicks: WeightChart['yTicks'] = []
  for (let v = lo; v <= hi; v += step) yTicks.push({ y: r1(y(v)), label: String(v) })

  const k = Math.max(1, Math.ceil((minLabelGap * (n - 1)) / Math.max(1, plotRight - plotLeft)))
  const xLabels: WeightChart['xLabels'] = []
  weeks.forEach((w, i) => {
    if ((n - 1 - i) % k !== 0) return
    const px = x(i)
    const anchor = px - minLabelGap / 2 < 0 ? 'start' : 'middle'
    xLabels.push({ x: r1(px), label: dayMonth(w.weekStart), anchor })
  })

  const whiskers: WeightChart['whiskers'] = []
  const points: WeightChart['points'] = []
  const segments: string[] = []
  let current: string[] = []
  weeks.forEach((w, i) => {
    if (w.avgKg === null) {
      if (current.length > 1) segments.push(current.join(' '))
      current = []
      return
    }
    const px = r1(x(i))
    if (w.minKg !== null && w.maxKg !== null && w.maxKg > w.minKg) {
      whiskers.push({ x: px, yMin: r1(y(w.minKg)), yMax: r1(y(w.maxKg)) })
    }
    const py = r1(y(w.avgKg))
    points.push({ x: px, y: py, current: inWeek(w.weekStart, today) })
    current.push(`${px},${py}`)
  })
  if (current.length > 1) segments.push(current.join(' '))

  const lastIdx = weeks.map((w) => w.avgKg !== null).lastIndexOf(true)
  const lastWeek = weeks[lastIdx]!
  const lastPoint = points[points.length - 1]!
  const last = { x: lastPoint.x, y: lastPoint.y, label: num(lastWeek.avgKg!, 1), current: lastPoint.current }

  return { width, height, plotLeft, plotRight, baseline, yTicks, xLabels, whiskers, segments, points, last }
}

/** Screen reader summary: "Średnia tygodniowa wagi z 78,1 do 76,4 kg w 12 tygodni; zakres ważeń w tygodniu do 1,2 kg". */
export function weightChartLabel(weeks: BodyWeek[]): string {
  const data = weeks.filter((w) => w.avgKg !== null)
  if (data.length === 0) return 'Brak ważeń'
  if (data.length === 1) return `Jedno ważenie: ${num(data[0]!.avgKg!, 1)} kg`
  const first = data[0]!
  const last = data[data.length - 1]!
  const spread = Math.max(...data.map((w) => (w.maxKg ?? 0) - (w.minKg ?? 0)))
  const base = `Średnia tygodniowa wagi z ${num(first.avgKg!, 1)} do ${num(last.avgKg!, 1)} kg w ${weeks.length} tygodni`
  return spread > 0 ? `${base}; zakres ważeń w tygodniu do ${num(spread, 1)} kg` : base
}
