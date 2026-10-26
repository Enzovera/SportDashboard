// Undo window for plan decisions. PlanChangeAccepted / PlanChangeRejected are final events, so "Cofnij" cannot
// be a server call: the decision waits client-side for `graceMs` and is sent only when the window closes.
// Framework-free so it can be unit-tested with a fake clock.

import type { Decision } from '@/api/proposals'

export const GRACE_MS = 5000

export type QueueStatus = 'waiting' | 'sending' | 'failed'

export interface QueuedDecision {
  id: string
  decision: Decision
  comment: string | null
  /** Clock time when the request goes out. */
  deadline: number
  status: QueueStatus
  error: string | null
}

export interface Clock {
  now(): number
  setTimeout(fn: () => void, ms: number): unknown
  clearTimeout(handle: unknown): void
}

export const realClock: Clock = {
  now: () => Date.now(),
  setTimeout: (fn, ms) => setTimeout(fn, ms),
  clearTimeout: (h) => clearTimeout(h as ReturnType<typeof setTimeout>),
}

export interface DecisionQueueOptions<R> {
  graceMs?: number
  send: (id: string, decision: Decision, comment: string | null) => Promise<R>
  /** The server accepted the decision. */
  onSent?: (item: QueuedDecision, result: R) => void
  /** Any state change; receives a fresh snapshot. */
  onChange?: (items: QueuedDecision[]) => void
  clock?: Clock
}

export function createDecisionQueue<R>(opts: DecisionQueueOptions<R>) {
  const graceMs = opts.graceMs ?? GRACE_MS
  const clock = opts.clock ?? realClock
  const items = new Map<string, QueuedDecision>()
  const timers = new Map<string, unknown>()

  const snapshot = () => [...items.values()].map((i) => ({ ...i }))
  const changed = () => opts.onChange?.(snapshot())

  function stopTimer(id: string) {
    if (timers.has(id)) clock.clearTimeout(timers.get(id))
    timers.delete(id)
  }

  async function fire(id: string) {
    stopTimer(id)
    const item = items.get(id)
    if (!item || item.status !== 'waiting') return
    item.status = 'sending'
    changed()
    try {
      const result = await opts.send(item.id, item.decision, item.comment)
      items.delete(id)
      changed()
      opts.onSent?.({ ...item }, result)
    } catch (e) {
      item.status = 'failed'
      item.error = e instanceof Error ? e.message : String(e)
      changed()
    }
  }

  return {
    /** Starts the undo window. Ignored while a decision for this id is already waiting or sending. */
    schedule(id: string, decision: Decision, comment: string | null): boolean {
      const existing = items.get(id)
      if (existing && existing.status !== 'failed') return false
      const trimmed = comment?.trim() ? comment.trim() : null
      items.set(id, { id, decision, comment: trimmed, deadline: clock.now() + graceMs, status: 'waiting', error: null })
      timers.set(id, clock.setTimeout(() => void fire(id), graceMs))
      changed()
      return true
    },

    /** "Cofnij": drops a decision that has not been sent yet. */
    undo(id: string): boolean {
      const item = items.get(id)
      if (!item || item.status === 'sending') return false
      stopTimer(id)
      items.delete(id)
      changed()
      return true
    },

    /** Sends every waiting decision now (leaving the view or closing the page: the intent was clear). */
    flushAll(): Promise<void[]> {
      const waiting = [...items.values()].filter((i) => i.status === 'waiting').map((i) => fire(i.id))
      return Promise.all(waiting)
    },

    get(id: string): QueuedDecision | undefined {
      const item = items.get(id)
      return item ? { ...item } : undefined
    },

    items: snapshot,
  }
}

export type DecisionQueue = ReturnType<typeof createDecisionQueue>

/** Whole seconds left in the window, for the countdown ("4 s"). */
export function secondsLeft(deadline: number, now: number): number {
  return Math.max(0, Math.ceil((deadline - now) / 1000))
}
