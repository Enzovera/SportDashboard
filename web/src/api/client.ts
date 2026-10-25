import type { TodayResponse, TrainingRange, TrainingResponse } from './types'
import { todayFull, todayStage1 } from './mocks/today'
import { trainingFull, trainingStage1 } from './mocks/training'

/**
 * Data source, chosen by `?mock=` in the URL or VITE_API_MOCK:
 *   full   - every section filled (matches design/screens)
 *   stage1 - only what Garmin + Hevy can serve today
 *   off    - real API under /api
 * Defaults to `full` in dev until services/api exists.
 */
export type MockMode = 'full' | 'stage1' | 'off'

export function mockMode(): MockMode {
  const fromUrl = new URLSearchParams(window.location.search).get('mock')
  const mode = fromUrl ?? import.meta.env.VITE_API_MOCK ?? (import.meta.env.DEV ? 'full' : 'off')
  return mode === 'stage1' || mode === 'off' ? mode : 'full'
}

async function get<T>(path: string): Promise<T> {
  const resp = await fetch(path, { headers: { accept: 'application/json' } })
  if (!resp.ok) throw new Error(`${path}: HTTP ${resp.status}`)
  return (await resp.json()) as T
}

const delay = <T>(value: T) => new Promise<T>((resolve) => setTimeout(() => resolve(structuredClone(value)), 150))

export function getToday(): Promise<TodayResponse> {
  switch (mockMode()) {
    case 'full':
      return delay(todayFull)
    case 'stage1':
      return delay(todayStage1)
    default:
      return get('/api/today')
  }
}

export function getTraining(range: TrainingRange): Promise<TrainingResponse> {
  switch (mockMode()) {
    case 'full':
      return delay({ ...trainingFull, range })
    case 'stage1':
      return delay({ ...trainingStage1, range })
    default:
      return get(`/api/training?range=${range}`)
  }
}
