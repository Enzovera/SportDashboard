// Contract and client of the body view ("Ciało", docs/design.md "Metryki liczone w kodzie": Waga).
// Same conventions as ./types.ts: ISO dates (Europe/Warsaw), timestamps with offset, numbers computed by the
// backend (the UI only formats them), `null` = not available.
//
// Endpoint (services/api, Go):
//   GET /api/body -> BodyResponse   (fixed range: the last 12 ISO weeks including the current one)
// Source: body_composition projection (weigh-ins via Garmin; composition only when the scale sends it).

import type { SyncStatus } from './types'
import { mockMode } from './client'
import { bodyFull, bodyStage1 } from './mocks/body'

export interface BodyResponse {
  /** Today (local). */
  date: string
  /** First day (Monday) of the oldest week and last day (Sunday) of the current week. */
  from: string
  to: string
  /** Number of ISO weeks in the range (12). */
  rangeWeeks: number
  /** Scale brand the weigh-ins come from through Garmin, e.g. "Xiaomi"; null = unknown / entered in Garmin. */
  device: string | null
  sync: SyncStatus[]
  composition: BodyComposition
  /** Weekly trend of the 7-day average, kg per week (negative = losing); null with too few weigh-ins. */
  trendKgPerWeek: number | null
  /** One entry per ISO week of the range, oldest first; weeks without weigh-ins have null values. */
  weeks: BodyWeek[]
  /** Weigh-ins, newest first, at most the last 7 days or the last 7 weigh-ins, whichever reaches further. */
  weighIns: WeighIn[]
  /** Weight goal from the athlete profile (stage 4); null = not set. */
  goal: { targetKg: number; by: string | null } | null
  /** Latest coach note about body/weight (stage 3). `rule` = standing rule, not tied to a day. */
  coachNote: { date: string; kind: 'rule' | 'daily' | 'weekly'; text: string } | null
}

export interface BodyComposition {
  /** Timestamp of the latest weigh-in; null = never weighed. */
  measuredAt: string | null
  /** Latest single weigh-in, kg. */
  lastKg: number | null
  /** 7-day moving average, kg; null when there are too few weigh-ins in the last 7 days. */
  avg7dKg: number | null
  /** Change of the 7-day average across the range, kg. */
  avg7dChangeKg: number | null
  /** Latest body fat estimate from the scale, %. */
  fatPct: number | null
  /** Latest muscle mass estimate, kg. */
  muscleKg: number | null
  /** Muscle mass trend over the last 4 weeks; null without composition data. */
  muscleTrend: 'up' | 'stable' | 'down' | null
  /** Days with a weigh-in in the current ISO week, and days of the week elapsed so far (incl. today). */
  weekCount: number
  weekDays: number
  /** Elapsed days of the current week without a weigh-in. */
  weekMissing: string[]
  /** Weigh-ins in the whole range. */
  rangeCount: number
}

export interface BodyWeek {
  /** Monday of the ISO week. */
  weekStart: string
  /** Average of the week's weigh-ins; null = no weigh-in that week. */
  avgKg: number | null
  minKg: number | null
  maxKg: number | null
  count: number
}

export interface WeighIn {
  measuredAt: string
  weightKg: number
  fatPct: number | null
  muscleKg: number | null
  waterPct: number | null
}

async function get<T>(path: string): Promise<T> {
  const resp = await fetch(path, { headers: { accept: 'application/json' } })
  if (!resp.ok) throw new Error(`${path}: HTTP ${resp.status}`)
  return (await resp.json()) as T
}

const delay = <T>(value: T) => new Promise<T>((resolve) => setTimeout(() => resolve(structuredClone(value)), 150))

export function getBody(): Promise<BodyResponse> {
  switch (mockMode()) {
    case 'full':
      return delay(bodyFull)
    case 'stage1':
      return delay(bodyStage1)
    default:
      return get('/api/body')
  }
}
