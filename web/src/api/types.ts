// Contract of the dashboard API (services/api, Go). The Go handlers must serialize exactly these shapes.
// Conventions:
// - dates: ISO `YYYY-MM-DD` (local, Europe/Warsaw); timestamps: ISO 8601 with offset
// - all numbers are computed by the backend; the UI only formats them
// - `null` = not available yet (stage not built, or no data that day); the UI hides or dashes it
// - enums are English keys; Polish labels live in src/lib/labels.ts

export type Source = 'garmin' | 'hevy' | 'fitatu' | 'hevy_garmin'

export interface SyncStatus {
  source: Source
  /** ok: last run fine; stale: no successful run for > 3 h; error/partial: last run failed (partly) */
  status: 'ok' | 'partial' | 'error' | 'stale'
  lastSuccessAt: string | null
}

/** Tone of a short status phrase; always rendered together with words, never color only. */
export type Tone = 'good' | 'warn' | 'neutral'

/** Training stress balance bucket, computed from TSB by the backend. */
export type FormState = 'fresh' | 'neutral' | 'tired' | 'overreached'

// ---------- GET /api/today ----------

export interface TodayResponse {
  date: string
  /** Weeks to the target race; null without a plan (stage 4). */
  raceWeeksLeft: number | null
  sync: SyncStatus[]
  /** Daily coach run (stage 3). */
  coach: CoachSummary | null
  /** Planned session for today (stage 4). */
  session: TodaySession | null
  week: WeekGrid
  /** Next day's plan (stage 4). */
  tomorrow: { title: string; detail: string | null } | null
  readiness: Readiness
  /** Nutrition vs target (stage 2). */
  nutrition: NutritionToday | null
  /** Supplements checklist (stage 2). */
  supplements: SupplementsToday | null
  /** Pending plan change proposals (stage 4). */
  proposalsPending: number | null
  /** 7-day weight average; null when there are too few weigh-ins. */
  weight: { avg7dKg: number; deltaPerWeekKg: number | null } | null
}

export interface CoachSummary {
  at: string
  text: string
  flags: { tone: Tone; text: string }[]
}

export interface TodaySession {
  title: string
  /** keep = as planned, modify = changed by coach, rest = day off recommended */
  action: 'keep' | 'modify' | 'rest'
  distanceKm: number | null
  zone: string | null
  paceSPerKm: [number, number] | null
  durationMin: number | null
}

export interface WeekGrid {
  from: string
  to: string
  /** Seven days Mon..Sun. */
  days: WeekDay[]
  runKmDone: number
  runKmPlanned: number | null
  strengthDone: number
  strengthPlanned: number | null
}

export interface WeekDay {
  date: string
  /** Done run km, or planned km for today/future when a plan exists; null = nothing. */
  runKm: number | null
  runPlanned: boolean
  /** Strength session label, e.g. "Push"; null = none. */
  strength: string | null
}

export interface Readiness {
  measuredAt: string | null
  hrv7dMs: number | null
  hrvBaselineMs: number | null
  hrvTone: Tone
  hrvLabel: string | null
  sleepS: number | null
  sleepAvg7dS: number | null
  restingHr: number | null
  restingHrDelta7d: number | null
  tsb: number | null
  form: FormState | null
}

export interface NutritionRow {
  eaten: number
  target: number
}

export interface NutritionToday {
  targetLabel: string
  kcal: NutritionRow
  proteinG: NutritionRow
  carbsG: NutritionRow
  fatG: NutritionRow
  yesterday: { kcal: NutritionRow; proteinG: NutritionRow; proteinTone: Tone } | null
}

export interface SupplementsToday {
  taken: number
  total: number
  next: string | null
}

// ---------- GET /api/training?range=4w|12w|block ----------

export type TrainingRange = '4w' | '12w' | 'block'

export interface TrainingResponse {
  range: TrainingRange
  from: string
  to: string
  sync: SyncStatus[]
  load: LoadSummary
  weeklyKm: WeeklyKm[]
  /** Planned km for today, drawn as the dashed "dziś" box (stage 4). */
  todayPlannedKm: number | null
  loadSeries: LoadPoint[]
  /** Latest coach note on training (stage 3). */
  coachNote: { date: string; text: string } | null
  recentRuns: RunRow[]
  setsPerMuscle: { weekFrom: string; weekTo: string; items: MuscleSets[] }
  strengthSessions: StrengthRow[]
}

export interface LoadSummary {
  weekFrom: string
  weekTo: string
  weekKm: number
  prevWeekKm: number | null
  ctl: number | null
  ctlDeltaRange: number | null
  atl: number | null
  atlDeltaWeek: number | null
  tsb: number | null
  form: FormState | null
}

export interface WeeklyKm {
  /** Monday of the ISO week. */
  weekStart: string
  km: number
  plannedKm: number | null
  deload: boolean
}

export interface LoadPoint {
  date: string
  ctl: number
  atl: number
}

export interface RunRow {
  id: string
  date: string
  title: string
  distanceKm: number
  paceSPerKm: number | null
  avgHr: number | null
  load: number | null
}

export interface MuscleSets {
  /** Hevy muscle group key, e.g. "chest", "upper_back"; labels in src/lib/labels.ts */
  muscle: string
  sets: number
}

export interface StrengthRow {
  id: string
  date: string
  title: string
  durationMin: number
  sets: number
  tonnageKg: number
}
