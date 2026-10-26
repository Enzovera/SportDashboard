// Contract and client of the plan view ("Plan", docs/design.md "Stan trenera").
// Same conventions as ./types.ts: ISO dates (Europe/Warsaw), numbers computed by the backend (the UI only
// formats them), `null` = not available.
//
// Endpoint (services/api, Go):
//   GET /api/plan -> PlanResponse
// Read-only view of the `plan_session` projection (plus the macrocycle from PlanCreated). The plan is
// event-sourced; this view never changes it. Pending proposals (PlanChangeProposed that passed the
// validators) are only referenced, so the affected days can link to /propozycje.

import { mockMode } from './client'
import { planFull, planStage1 } from './mocks/plan'

export interface PlanResponse {
  /** Today (local). */
  date: string
  /** null = no plan yet (before stage 4). */
  plan: Plan | null
  /** Pending plan change proposals; null without a plan. */
  proposalsPending: number | null
}

export interface Plan {
  goal: RaceGoal
  /** Macrocycle blocks in order; consecutive, the last one ends on race day. */
  blocks: PlanBlock[]
  /** ISO week containing today; null when the plan has no sessions in it. */
  currentWeek: PlanWeek | null
  /** Next ISO week; null after the race. */
  nextWeek: PlanWeek | null
}

export interface RaceGoal {
  /** e.g. "Maraton" (from the profile). */
  race: string
  date: string
  /** Whole weeks from today to race day. */
  weeksLeft: number
  /** Target finish time, s. */
  targetTimeS: number | null
  /** Target pace, s/km. */
  targetPaceSPerKm: number | null
  /** Secondary goals from the athlete profile, e.g. { label: "Półmaraton", value: "< 1:40", detail: "najlepszy 1:47" }. */
  secondary: { label: string; value: string; detail: string | null }[]
}

export type BlockKind = 'base' | 'build' | 'peak' | 'taper'

export interface PlanBlock {
  kind: BlockKind
  from: string
  to: string
  weeks: number
  /** Short description from the plan (written by the coach, accepted by the user). */
  summary: string | null
}

/** Nutrition target type of a day (plan_session); labels in views/plan/planData.ts. */
export type DayType = 'strength_run' | 'run' | 'long_run' | 'strength' | 'rest'

export interface PlanWeek {
  /** Monday and Sunday. */
  from: string
  to: string
  /** Block of this week and its 1-based number inside the block; null = before the first block. */
  block: BlockKind | null
  weekInBlock: number | null
  plannedKm: number
  plannedStrength: number
  /** Done so far (Garmin + Hevy); null for a week that has not started. */
  doneKm: number | null
  doneStrength: number | null
  /** Seven days, Mon..Sun. */
  days: PlanDay[]
}

export interface PlanDay {
  date: string
  /** plan_session rows of the day ordered by slot; empty = rest day. */
  sessions: PlanSession[]
  /** Planned running km of the day; null = no run. */
  runKm: number | null
  dayType: DayType
  /** Daily nutrition target; null when not set. */
  target: { kcal: number; proteinG: number; carbsG: number; fatG: number } | null
  /** What was done (past days of the current week only; null for today and future days). */
  done: { runKm: number | null; strengthSets: number | null } | null
  /** Pending proposals that change this day; non-empty = "do decyzji". */
  proposals: { id: string; summary: string }[]
}

export interface PlanSession {
  kind: 'run' | 'strength'
  /** e.g. "Długie wybieganie", "Push", "bieg łatwy" for a secondary run. */
  title: string
  /** e.g. "strefa 2", "6×800 m", "16 serii roboczych". */
  detail: string | null
  distanceKm: number | null
}

async function get<T>(path: string): Promise<T> {
  const resp = await fetch(path, { headers: { accept: 'application/json' } })
  if (!resp.ok) throw new Error(`${path}: HTTP ${resp.status}`)
  return (await resp.json()) as T
}

const delay = <T>(value: T) => new Promise<T>((resolve) => setTimeout(() => resolve(structuredClone(value)), 150))

export function getPlan(): Promise<PlanResponse> {
  switch (mockMode()) {
    case 'full':
      return delay(planFull)
    case 'stage1':
      return delay(planStage1)
    default:
      return get('/api/plan')
  }
}
