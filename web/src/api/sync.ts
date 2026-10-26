// On-demand sync of all sources ("Odśwież", docs/design.md, "Integracje").
// POST /api/sync queues a request (one active at a time, a second click joins it); the ingest scheduler runs
// Hevy, Garmin and Fitatu and the API rebuilds projections. GET /api/sync/{id} turns `done` only after that,
// so bumping `syncGeneration` then makes every useApi() view reload fresh data.
import { reactive, ref } from 'vue'
import type { Source } from './types'
import { mockMode } from './client'

export interface SyncRequest {
  id: number
  status: 'pending' | 'running' | 'done'
  requestedAt: string
  finishedAt: string | null
  /** Latest ingest run per source within the request; empty while pending. */
  sources: { source: Source; status: 'running' | 'ok' | 'partial' | 'error' }[]
}

/** Bumped after every finished sync; useApi() reloads on change. */
export const syncGeneration = ref(0)

export const syncState = reactive({
  busy: false,
  /** Short UI text when the sync could not run, e.g. within the cooldown. */
  problem: null as string | null,
})

const POLL_MS = 2000
const TIMEOUT_MS = 10 * 60 * 1000

const wait = (ms: number) => new Promise((resolve) => setTimeout(resolve, ms))

async function send(path: string, method: 'GET' | 'POST'): Promise<Response> {
  return fetch(path, { method, headers: { accept: 'application/json' } })
}

function showProblem(text: string, ms: number) {
  syncState.problem = text
  setTimeout(() => {
    if (syncState.problem === text) syncState.problem = null
  }, ms)
}

export async function syncNow(): Promise<void> {
  if (syncState.busy) return
  syncState.busy = true
  syncState.problem = null
  try {
    if (mockMode() !== 'off') {
      await wait(1500)
    } else {
      const resp = await send('/api/sync', 'POST')
      if (resp.status === 429) {
        const retry = Number(resp.headers.get('retry-after')) || 60
        showProblem('sync przed chwilą', retry * 1000)
        return
      }
      if (!resp.ok) throw new Error(`HTTP ${resp.status}`)
      let req = (await resp.json()) as SyncRequest
      const deadline = Date.now() + TIMEOUT_MS
      while (req.status !== 'done') {
        if (Date.now() > deadline) throw new Error('timeout')
        await wait(POLL_MS)
        const r = await send(`/api/sync/${req.id}`, 'GET')
        if (!r.ok) throw new Error(`HTTP ${r.status}`)
        req = (await r.json()) as SyncRequest
      }
    }
    // Per-source errors show up in SyncStatus after the reload (ingest_run), no separate message here.
    syncGeneration.value++
  } catch {
    showProblem('błąd odświeżania', 10_000)
  } finally {
    syncState.busy = false
  }
}
