// Contract and client of the Nutrition view (GET /api/nutrition). Conventions as in ./types.ts:
// - dates: ISO `YYYY-MM-DD` (local, Europe/Warsaw); timestamps: ISO 8601 with offset
// - all numbers (averages, %, g/kg, tones) are computed by the backend; the UI only formats them
// - `null` = not available (Fitatu not integrated yet = stage 2, no plan targets = stage 4, no data that day)
// - enums are English keys; Polish labels live in src/views/nutrition/nutritionData.ts
import type { NutritionRow, SupplementsToday, SyncStatus, Tone } from './types'
import { mockMode } from './client'
import { nutritionFull, nutritionStage1 } from './mocks/nutrition'
import { supplementsSummary } from './mocks/supplements'

/** Type of training day; drives the nutrition target of the day (plan_session, stage 4). */
export type DayType = 'rest' | 'strength' | 'run' | 'strength_run' | 'long_run'

export type NutritionRange = '14d'

/** Fitatu meal slots. */
export type MealSlot = 'breakfast' | 'second_breakfast' | 'lunch' | 'snack' | 'dinner' | 'supper'

// ---------- GET /api/nutrition?range=14d ----------

export interface NutritionResponse {
  range: NutritionRange
  /** First and last day of the range; `to` is today. */
  from: string
  to: string
  sync: SyncStatus[]
  /** Today's intake vs the target of today's day type; null without Fitatu (stage 2). */
  today: NutritionTodayDetail | null
  /** Averages over the complete days of the range (today excluded); null without data. */
  balance: NutritionBalance | null
  /** Every day of the range, oldest first; today last with `complete: false`. Empty without Fitatu. */
  days: NutritionDay[]
  /** Current targets per day type (plan, stage 4); empty without a plan. */
  targets: DayTarget[]
  /** Pending plan change proposals that touch a nutrition target (stage 4). */
  targetProposals: TargetProposal[]
  /** Today's meals from Fitatu, in time order. */
  meals: Meal[]
  /** Latest coach note on nutrition for today (stage 3). */
  coachNote: { at: string; text: string } | null
  /** Supplements checklist counter for the Jedzenie | Suplementy switch (stage 2). */
  supplements: SupplementsToday | null
}

export interface NutritionTodayDetail {
  dayType: DayType
  kcal: NutritionRow
  proteinG: NutritionRow
  carbsG: NutritionRow
  fatG: NutritionRow
}

export interface NutritionBalance {
  from: string
  to: string
  kcalAvg: number
  kcalTargetAvg: number
  /** (kcalAvg − kcalTargetAvg) / kcalTargetAvg in %, rounded; e.g. −5. */
  kcalVsTargetPct: number
  proteinAvgGPerKg: number
  proteinTargetGPerKg: number
  /** Hard floor from the validators (1,6 g/kg). */
  proteinMinGPerKg: number
  /** Days with protein ≥ proteinMinGPerKg, out of `days`. */
  proteinDaysAtMin: number
  days: number
  /** Carbs on the latest complete long-run day of the range; null when there was none. */
  longRunCarbs: { date: string; gPerKg: number; targetGPerKg: number; tone: Tone } | null
}

export interface NutritionDay {
  date: string
  /** null = no plan for that day and no activity to infer it from. */
  dayType: DayType | null
  /** false = today, still being logged ("w toku"). */
  complete: boolean
  kcal: number | null
  kcalTarget: number | null
  kcalVsTargetPct: number | null
  /** warn = deficit that matters (e.g. any deficit on a long-run day). */
  kcalTone: Tone
  proteinGPerKg: number | null
  /** warn = below the 1,6 g/kg floor. */
  proteinTone: Tone
}

export interface DayTarget {
  dayType: DayType
  kcal: number
  proteinG: number
  carbsG: number
  fatG: number
}

export interface TargetProposal {
  /** plan_event proposal id; decided in /propozycje. */
  id: string
  dayType: DayType
  field: 'kcal' | 'proteinG' | 'carbsG' | 'fatG'
  value: number
}

export interface Meal {
  id: string
  slot: MealSlot
  /** Time logged in Fitatu; null when Fitatu has no time for the meal. */
  at: string | null
  /** Products joined by Fitatu, e.g. "Owsianka z bananem i odżywką białkową". */
  title: string
  kcal: number
  proteinG: number
  carbsG: number
  fatG: number
}

// ---------- client ----------

async function get<T>(path: string): Promise<T> {
  const resp = await fetch(path, { headers: { accept: 'application/json' } })
  if (!resp.ok) throw new Error(`${path}: HTTP ${resp.status}`)
  return (await resp.json()) as T
}

const delay = <T>(value: T) => new Promise<T>((resolve) => setTimeout(() => resolve(structuredClone(value)), 150))

export function getNutrition(range: NutritionRange = '14d'): Promise<NutritionResponse> {
  switch (mockMode()) {
    case 'full':
      // The switch counter follows the in-memory supplements log, so ticking an item shows up here.
      return delay({ ...nutritionFull, range, supplements: supplementsSummary() })
    case 'stage1':
      return delay({ ...nutritionStage1, range })
    default:
      return get(`/api/nutrition?range=${range}`)
  }
}
