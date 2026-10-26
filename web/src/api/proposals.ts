// Contract and client of the plan change proposals ("Propozycje", docs/design.md "Stan trenera").
// Same conventions as ./types.ts: ISO dates (Europe/Warsaw), timestamps with offset, English enum keys,
// numbers computed by the backend (the UI only formats them), `null` = not available.
//
// Endpoints (services/api, Go):
//   GET  /api/proposals                 -> ProposalsResponse
//   POST /api/proposals/{id}/accept     body DecisionRequest -> 200 ProposalDecision   (PlanChangeAccepted)
//   POST /api/proposals/{id}/reject     body DecisionRequest -> 200 ProposalDecision   (PlanChangeRejected)
//     404 unknown id, 409 already decided (events are final), 422 comment too long (> 500 chars).
// Only a human decides, only here. The UI holds a decision for a short undo window client-side and sends it
// once; there is no "undo" endpoint.

import { shallowRef } from 'vue'
import { mockMode } from './client'
import { proposalsFull, proposalsStage1 } from './mocks/proposals'

// ---------- GET /api/proposals ----------

export interface ProposalsResponse {
  /** Today (local), the reference for "dziś / wczoraj". */
  date: string
  /** Waiting for a decision (PlanChangeProposed that passed the validators), newest first. */
  pending: Proposal[]
  /** Rejected by a validator; never queued, the coach got the reason. Recent ones (e.g. 7 days), newest first. */
  blocked: BlockedProposal[]
  /** Decided proposals (PlanChangeAccepted / PlanChangeRejected), newest decision first, e.g. last 30 days. */
  history: ProposalDecision[]
  /** Planned week touched by pending proposals, before and after; null without a plan or with nothing pending. */
  week: WeekDiff | null
}

/** Area of the plan a proposal changes; drives the tag ("Żywienie", "Siła", "Bieg"). */
export type ProposalKind = 'nutrition' | 'strength' | 'run'

/** Which AI run produced it. */
export type ProposalOrigin = 'daily_run' | 'weekly_run' | 'chat'

export interface Proposal {
  id: string
  kind: ProposalKind
  origin: ProposalOrigin
  proposedAt: string
  /** One line with before -> after, written by the coach, e.g. "Śr 28.10, Legs: 16 → 12 serii". */
  title: string
  /** Coach rationale (LLM prose). */
  rationale: string
  /** First day the change applies, when it is not already in the title; null otherwise. */
  effectiveFrom: string | null
  diff: ProposalDiff
  /** Validator results; only `pass` and `warn` here (a `fail` blocks the proposal). */
  checks: ValidatorCheck[]
}

export interface ProposalDiff {
  /** What the rows describe, e.g. "cel dnia", "Serie robocze". */
  heading: string
  /** Every row of the changed target; unchanged ones have before == after. */
  rows: DiffRow[]
  /** Sum row (e.g. total sets); null when a sum makes no sense. */
  total: DiffRow | null
}

export interface DiffRow {
  label: string
  /** Value now; null = not set yet (e.g. a new pre-run meal). */
  before: number | string | null
  /** Value after accepting; null = removed. */
  after: number | string | null
  /** Unit appended to numbers ("g", "kcal"); null for counts and text values. */
  unit: string | null
  /** Small text under/after the new value, e.g. "6,0 g/kg", "2–3 h przed". */
  note: string | null
}

export interface ValidatorCheck {
  /** Validator key (docs/design.md "Walidatory"), e.g. km_increase, long_run_share, rest_day, taper, kcal_floor, long_run_deficit, protein_min. */
  rule: string
  /** pass = ✓, warn = ! (allowed, but look), fail = ✕ (blocks) */
  status: 'pass' | 'warn' | 'fail'
  /** Result in words, produced by the validator, e.g. "przyrost 46 → 52 km (+13%, limit 10%)". */
  text: string
}

export interface BlockedProposal {
  id: string
  kind: ProposalKind
  origin: ProposalOrigin
  proposedAt: string
  title: string
  /** At least one `fail`. */
  checks: ValidatorCheck[]
}

export interface ProposalDecision {
  id: string
  kind: ProposalKind
  origin: ProposalOrigin
  title: string
  proposedAt: string
  decidedAt: string
  decision: 'accepted' | 'rejected'
  /** Optional comment for the coach, given with the decision. */
  comment: string | null
}

export interface WeekDiff {
  from: string
  to: string
  /** Block of the macrocycle and week number in it, e.g. "Baza · tydz. 1"; null without a plan. */
  block: string | null
  /** Seven days Mon..Sun. */
  days: WeekDiffDay[]
}

export interface WeekDiffDay {
  date: string
  now: DayPlan
  /** What pending proposals change on this day; empty = unchanged. Later entries win. */
  changes: { proposalId: string; after: DayPlan }[]
}

export interface DayPlan {
  /** Session label, e.g. "Długie 16 km", "Push + bieg 6 km", "Odpoczynek". */
  session: string
  /** Planned run distance, for the week total; null = no run. */
  runKm: number | null
  kcal: number
  carbsG: number
}

// ---------- POST /api/proposals/{id}/accept|reject ----------

export interface DecisionRequest {
  comment: string | null
}

export type Decision = 'accept' | 'reject'

// ---------- client ----------

/** Pending count from the last load/decision; for the nav badge (null until known). */
export const proposalsPendingCount = shallowRef<number | null>(null)

const mockState: Record<'full' | 'stage1', ProposalsResponse | null> = { full: null, stage1: null }

function mockData(mode: 'full' | 'stage1'): ProposalsResponse {
  mockState[mode] ??= structuredClone(mode === 'full' ? proposalsFull : proposalsStage1)
  return mockState[mode]!
}

const delay = <T>(value: () => T, ms = 150) =>
  new Promise<T>((resolve, reject) =>
    setTimeout(() => {
      try {
        resolve(structuredClone(value()))
      } catch (e) {
        reject(e)
      }
    }, ms),
  )

export async function getProposals(): Promise<ProposalsResponse> {
  const mode = mockMode()
  let data: ProposalsResponse
  if (mode === 'off') {
    const resp = await fetch('/api/proposals', { headers: { accept: 'application/json' } })
    if (!resp.ok) throw new Error(`/api/proposals: HTTP ${resp.status}`)
    data = (await resp.json()) as ProposalsResponse
  } else {
    data = await delay(() => mockData(mode))
  }
  proposalsPendingCount.value = data.pending.length
  return data
}

/** Sends the final decision. `keepalive` lets it finish when the page is being closed. */
export async function decideProposal(id: string, decision: Decision, comment?: string | null): Promise<ProposalDecision> {
  const body: DecisionRequest = { comment: comment?.trim() ? comment.trim() : null }
  const mode = mockMode()
  let result: ProposalDecision
  if (mode === 'off') {
    const path = `/api/proposals/${encodeURIComponent(id)}/${decision}`
    const resp = await fetch(path, {
      method: 'POST',
      headers: { accept: 'application/json', 'content-type': 'application/json' },
      body: JSON.stringify(body),
      keepalive: true,
    })
    if (!resp.ok) throw new Error(`${path}: HTTP ${resp.status}`)
    result = (await resp.json()) as ProposalDecision
  } else {
    result = await delay(() => applyMockDecision(mockData(mode), id, decision, body.comment))
  }
  if (proposalsPendingCount.value !== null) proposalsPendingCount.value = Math.max(0, proposalsPendingCount.value - 1)
  return result
}

/** What the Go handler does, in memory: drop from the queue, append to history, update the week projection. */
export function applyMockDecision(
  state: ProposalsResponse,
  id: string,
  decision: Decision,
  comment: string | null,
  now: Date = new Date(),
): ProposalDecision {
  const idx = state.pending.findIndex((p) => p.id === id)
  if (idx < 0) throw new Error(`/api/proposals/${id}/${decision}: HTTP 409`)
  const [p] = state.pending.splice(idx, 1)
  const entry: ProposalDecision = {
    id: p!.id,
    kind: p!.kind,
    origin: p!.origin,
    title: p!.title,
    proposedAt: p!.proposedAt,
    decidedAt: now.toISOString(),
    decision: decision === 'accept' ? 'accepted' : 'rejected',
    comment,
  }
  state.history.unshift(entry)
  if (state.week) {
    for (const day of state.week.days) {
      const change = day.changes.find((c) => c.proposalId === id)
      if (change && decision === 'accept') day.now = change.after
      day.changes = day.changes.filter((c) => c.proposalId !== id)
    }
    if (state.pending.length === 0) state.week = null
  }
  return entry
}
