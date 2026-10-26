// Pure helpers of the Nutrition view: labels, header meta, metric cells, headlines. No DOM, unit-tested.
import type { DayType, Meal, MealSlot, NutritionBalance, NutritionDay, NutritionResponse, TargetProposal } from '@/api/nutrition'
import type { NutritionRow, SupplementsToday, Tone } from '@/api/types'
import { dateRange, dayMonth, num, signed, time } from '@/lib/format'

export const DAY_TYPE_LABEL: Record<DayType, string> = {
  rest: 'Odpoczynek',
  strength: 'Siłowy',
  run: 'Bieg',
  strength_run: 'Siłowy + bieg',
  long_run: 'Długie wybieganie',
}

/** One-letter code under the calorie chart: "S siłowy · B bieg · D długie · O wolne". */
export const DAY_TYPE_SHORT: Record<DayType, string> = {
  rest: 'O',
  strength: 'S',
  run: 'B',
  strength_run: 'SB',
  long_run: 'D',
}

export const DAY_TYPE_LEGEND = 'S siłowy · B bieg · D długie · O wolne'

/** "cel dnia długiego", "w dniu długiego". */
const DAY_TYPE_GENITIVE: Record<DayType, string> = {
  rest: 'wolnego',
  strength: 'siłowego',
  run: 'biegowego',
  strength_run: 'siłowego z biegiem',
  long_run: 'długiego',
}

/** "w dni długiego wybiegania". */
const DAY_TYPE_PLURAL: Record<DayType, string> = {
  rest: 'wolne',
  strength: 'siłowe',
  run: 'biegowe',
  strength_run: 'siłowe z biegiem',
  long_run: 'długiego wybiegania',
}

export const MEAL_LABEL: Record<MealSlot, string> = {
  breakfast: 'Śniadanie',
  second_breakfast: 'II śniadanie',
  lunch: 'Obiad',
  snack: 'Przekąska',
  dinner: 'Kolacja',
  supper: 'Późna kolacja',
}

/** "Dziś · długie wybieganie". */
export function todayTitle(dayType: DayType | null): string {
  return dayType ? `Dziś · ${DAY_TYPE_LABEL[dayType].toLowerCase()}` : 'Dziś'
}

/** "cel dnia długiego". */
export function targetMeta(dayType: DayType): string {
  return `cel dnia ${DAY_TYPE_GENITIVE[dayType]}`
}

/** Fitatu sync time when it is fine (problems are shown by SyncStatus) + "12–25.10". */
export function headerMeta(data: NutritionResponse): string {
  const range = dateRange(data.from, data.to)
  const fitatu = data.sync.find((s) => s.source === 'fitatu')
  return fitatu?.status === 'ok' && fitatu.lastSuccessAt ? `Fitatu ${time(fitatu.lastSuccessAt)} · ${range}` : range
}

/** "Fitatu 08:15", or null when the sync is not fine. */
export function fitatuMeta(data: NutritionResponse): string | null {
  const fitatu = data.sync.find((s) => s.source === 'fitatu')
  return fitatu?.status === 'ok' && fitatu.lastSuccessAt ? `Fitatu ${time(fitatu.lastSuccessAt)}` : null
}

/** "Suplementy 2/7" on the switch; plain "Suplementy" without data. */
export function supplementsTabLabel(s: SupplementsToday | null): string {
  return s && s.total > 0 ? `Suplementy ${s.taken}/${s.total}` : 'Suplementy'
}

/** −5 -> "−5%", 1 -> "+1%". */
export function pct(value: number): string {
  return `${signed(value)}%`
}

export interface MetricCellData {
  label: string
  value: string
  unit?: string
  sub: string | null
  subTone?: Tone
}

/** The four cells of the "Bilans" box. */
export function balanceCells(b: NutritionBalance): MetricCellData[] {
  const below = b.days - b.proteinDaysAtMin
  const cells: MetricCellData[] = [
    { label: 'kcal vs cel', value: pct(b.kcalVsTargetPct), sub: `śr. ${num(b.kcalAvg)} z ${num(b.kcalTargetAvg)}` },
    {
      label: 'Białko · średnio',
      value: num(b.proteinAvgGPerKg, 2),
      unit: 'g/kg',
      sub: `cel ${num(b.proteinTargetGPerKg, 1)} · min ${num(b.proteinMinGPerKg, 1)}`,
    },
    {
      label: 'Dni ≥ min. białka',
      value: num(b.proteinDaysAtMin),
      unit: `/ ${num(b.days)}`,
      sub: below > 0 ? `${daysWord(below)} poniżej ${num(b.proteinMinGPerKg, 1)}` : 'wszystkie',
      subTone: below > 0 ? 'warn' : 'good',
    },
  ]
  if (b.longRunCarbs) {
    cells.push({
      label: `Węgle · długie ${dayMonth(b.longRunCarbs.date)}`,
      value: num(b.longRunCarbs.gPerKg, 1),
      unit: 'g/kg',
      sub: `cel ${num(b.longRunCarbs.targetGPerKg, 1)} g/kg`,
      subTone: b.longRunCarbs.tone,
    })
  }
  return cells
}

/** 1 -> "1 dzień", 3 -> "3 dni". */
export function daysWord(n: number): string {
  return n === 1 ? '1 dzień' : `${num(n)} dni`
}

/** Deficit (≤ −5% on average) worth naming in the calorie chart title. */
const DEFICIT_PCT = -5

/**
 * Title of the calorie chart: names the day type with the largest average deficit,
 * e.g. "Największy niedobór w dni długiego wybiegania"; neutral otherwise.
 */
export function kcalHeadline(days: NutritionDay[]): string {
  const sums = new Map<DayType, { sum: number; n: number }>()
  for (const d of days) {
    if (!d.complete || d.dayType === null || d.kcalVsTargetPct === null) continue
    const s = sums.get(d.dayType) ?? { sum: 0, n: 0 }
    s.sum += d.kcalVsTargetPct
    s.n += 1
    sums.set(d.dayType, s)
  }
  let worst: { type: DayType; avg: number } | null = null
  for (const [type, s] of sums) {
    const avg = s.sum / s.n
    if (!worst || avg < worst.avg) worst = { type, avg }
  }
  if (!worst) return 'Kalorie dzień po dniu'
  if (worst.avg <= DEFICIT_PCT) return `Największy niedobór w dni ${DAY_TYPE_PLURAL[worst.type]}`
  return 'Kalorie blisko celu'
}

/** Title of the protein chart, from how often the target was reached. */
export function proteinHeadline(days: NutritionDay[], b: NutritionBalance): string {
  const values = days.filter((d) => d.complete && d.proteinGPerKg !== null).map((d) => d.proteinGPerKg as number)
  if (!values.length) return 'Białko na kilogram'
  const atTarget = values.filter((v) => v >= b.proteinTargetGPerKg).length / values.length
  const belowMin = values.some((v) => v < b.proteinMinGPerKg)
  if (atTarget >= 0.7) return 'Białko zwykle na celu'
  if (atTarget < 0.34 && belowMin) return 'Białko rzadko dobija do celu'
  if (!belowMin) return 'Białko powyżej minimum'
  return 'Białko na kilogram'
}

/** Target − eaten; negative when over the target. */
export function remaining(row: NutritionRow): number {
  return row.target - row.eaten
}

/** "W 460 g", "kcal 3 100". */
export function proposalValue(p: TargetProposal): string {
  switch (p.field) {
    case 'kcal':
      return `kcal ${num(p.value)}`
    case 'proteinG':
      return `B ${num(p.value)} g`
    case 'carbsG':
      return `W ${num(p.value)} g`
    case 'fatG':
      return `T ${num(p.value)} g`
  }
}

/** "w dniu długiego". */
export function proposalWhere(p: TargetProposal): string {
  return `w dniu ${DAY_TYPE_GENITIVE[p.dayType]}`
}

/** "820 kcal · B 38 · W 112 · T 24". */
export function mealMacros(m: Pick<Meal, 'kcal' | 'proteinG' | 'carbsG' | 'fatG'>): string {
  return `${num(m.kcal)} kcal · B ${num(m.proteinG)} · W ${num(m.carbsG)} · T ${num(m.fatG)}`
}

/** "07:40" or "" when Fitatu has no time. */
export function mealTime(m: Meal): string {
  return m.at ? time(m.at) : ''
}

/** Days that highlight in the calorie chart and the day table (e.g. a deficit on a long run). */
export function isKeyDay(d: NutritionDay): boolean {
  return d.complete && d.kcalTone === 'warn'
}

export function kcalAriaLabel(days: NutritionDay[]): string {
  const complete = days.filter((d) => d.complete)
  const parts = [`Kalorie zjedzone w ${complete.length} dni na tle celu dnia`]
  for (const d of complete.filter(isKeyDay)) {
    if (d.kcal === null || d.kcalTarget === null) continue
    const type = d.dayType ? `, ${DAY_TYPE_LABEL[d.dayType].toLowerCase()}` : ''
    parts.push(`${dayMonth(d.date)}${type}: zjedzone ${num(d.kcal)} przy celu ${num(d.kcalTarget)} kcal`)
  }
  const today = days.find((d) => !d.complete)
  if (today?.kcal != null) parts.push(`dziś ${num(today.kcal)} kcal w toku`)
  return parts.join('; ')
}

export function proteinAriaLabel(days: NutritionDay[], b: NutritionBalance): string {
  const values = days.filter((d) => d.complete && d.proteinGPerKg !== null).map((d) => d.proteinGPerKg as number)
  if (!values.length) return 'Białko w g na kg: brak danych'
  const atTarget = values.filter((v) => v >= b.proteinTargetGPerKg).length
  const belowMin = values.filter((v) => v < b.proteinMinGPerKg).length
  return (
    `Białko w g na kg od ${num(Math.min(...values), 2)} do ${num(Math.max(...values), 2)}; ` +
    `cel ${num(b.proteinTargetGPerKg, 1)} osiągnięty w ${atTarget} z ${values.length} dni, ` +
    `${daysWord(belowMin)} poniżej minimum ${num(b.proteinMinGPerKg, 1)}`
  )
}
