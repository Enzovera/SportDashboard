import type { TodayResponse, TrainingRange, TrainingResponse } from './types'
import { todayFull, todayStage1 } from './mocks/today'
import { trainingFull, trainingStage1 } from './mocks/training'

/**
 * Data source, chosen by the `mock` cookie:
 *   mock=1              - mocks with every section filled (matches design/screens)
 *   mock=0 or no cookie - real API under /api
 * `?mock=1` / `?mock=0` in the URL only sets the cookie (see applyMockParam).
 * A build with `VITE_MOCK=1` always uses the mocks: a preview without a backend, where cookies may be blocked.
 */
export type MockMode = 'full' | 'stage1' | 'off'

export const FORCE_MOCK = import.meta.env.VITE_MOCK === '1'

export function mockMode(): MockMode {
  if (FORCE_MOCK) return 'full'
  return /(?:^|;\s*)mock=1(?:;|$)/.test(document.cookie) ? 'full' : 'off'
}

/** Stores `?mock=1|0` from the URL in the cookie; the router guard then drops it from the address. */
export function applyMockParam(value: unknown): void {
  if (value === '1' || value === '0') {
    document.cookie = `mock=${value}; path=/; max-age=31536000; SameSite=Lax`
  }
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
