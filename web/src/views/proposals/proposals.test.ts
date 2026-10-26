import { describe, expect, it, vi } from 'vitest'
import { applyMockDecision, type ProposalDecision } from '@/api/proposals'
import { proposalsFull, proposalsStage1 } from '@/api/mocks/proposals'
import { type Clock, createDecisionQueue, secondsLeft } from './decisionQueue'
import {
  changedRows,
  columns,
  dayNumbers,
  diffHeading,
  diffValue,
  mergeHistory,
  pendingText,
  proposedMeta,
  weekKm,
  weekRows,
} from './proposals'

const NB = ' '

/** Deterministic clock: timers fire only when `advance` passes their time. */
function fakeClock() {
  let t = 0
  let seq = 0
  const timers = new Map<number, { at: number; fn: () => void }>()
  const clock: Clock = {
    now: () => t,
    setTimeout: (fn, ms) => {
      const id = ++seq
      timers.set(id, { at: t + ms, fn })
      return id
    },
    clearTimeout: (h) => void timers.delete(h as number),
  }
  async function advance(ms: number) {
    t += ms
    for (const [id, timer] of [...timers]) {
      if (timer.at <= t) {
        timers.delete(id)
        timer.fn()
      }
    }
    await Promise.resolve()
    await Promise.resolve()
  }
  return { clock, advance, pending: () => timers.size }
}

function setup(send = vi.fn(async () => 'ok')) {
  const c = fakeClock()
  const sent: string[] = []
  const q = createDecisionQueue({ graceMs: 5000, send, clock: c.clock, onSent: (i) => sent.push(i.id) })
  return { ...c, q, send, sent }
}

describe('decision queue (undo window)', () => {
  it('sends only after the window closes', async () => {
    const { q, send, advance, sent } = setup()
    q.schedule('a', 'accept', '  przed biegiem jem lekko ')
    expect(q.get('a')).toMatchObject({ status: 'waiting', deadline: 5000, comment: 'przed biegiem jem lekko' })
    await advance(4999)
    expect(send).not.toHaveBeenCalled()
    await advance(1)
    expect(send).toHaveBeenCalledWith('a', 'accept', 'przed biegiem jem lekko')
    expect(sent).toEqual(['a'])
    expect(q.items()).toEqual([])
  })

  it('undo inside the window sends nothing', async () => {
    const { q, send, advance, pending } = setup()
    q.schedule('a', 'reject', '')
    await advance(3000)
    expect(q.undo('a')).toBe(true)
    expect(pending()).toBe(0)
    await advance(10_000)
    expect(send).not.toHaveBeenCalled()
    expect(q.items()).toEqual([])
  })

  it('empty comment becomes null; a second click while waiting is ignored', async () => {
    const { q, send, advance } = setup()
    expect(q.schedule('a', 'reject', '   ')).toBe(true)
    expect(q.schedule('a', 'accept', null)).toBe(false)
    await advance(5000)
    expect(send).toHaveBeenCalledTimes(1)
    expect(send).toHaveBeenCalledWith('a', 'reject', null)
  })

  it('cannot undo while sending', async () => {
    let resolve!: (v: string) => void
    const send = vi.fn(() => new Promise<string>((r) => (resolve = r)))
    const { q, advance, sent } = setup(send)
    q.schedule('a', 'accept', null)
    await advance(5000)
    expect(q.get('a')?.status).toBe('sending')
    expect(q.undo('a')).toBe(false)
    resolve('ok')
    await advance(0)
    expect(sent).toEqual(['a'])
  })

  it('a failed send keeps the item with the error and can be retried', async () => {
    const send = vi.fn().mockRejectedValueOnce(new Error('HTTP 500')).mockResolvedValueOnce('ok')
    const { q, advance, sent } = setup(send)
    q.schedule('a', 'accept', null)
    await advance(5000)
    expect(q.get('a')).toMatchObject({ status: 'failed', error: 'HTTP 500' })
    expect(q.schedule('a', 'accept', null)).toBe(true)
    await advance(5000)
    expect(sent).toEqual(['a'])
  })

  it('flushAll sends waiting decisions at once', async () => {
    const { q, send, advance, pending } = setup()
    q.schedule('a', 'accept', null)
    q.schedule('b', 'reject', 'nie')
    await q.flushAll()
    expect(send).toHaveBeenCalledTimes(2)
    expect(pending()).toBe(0)
    await advance(10_000)
    expect(send).toHaveBeenCalledTimes(2)
  })

  it.each([
    [5000, 0, 5],
    [5000, 1, 5],
    [5000, 1000, 4],
    [5000, 4999, 1],
    [5000, 5000, 0],
    [5000, 6000, 0],
  ])('secondsLeft(%s, %s) = %s', (deadline, now, want) => expect(secondsLeft(deadline, now)).toBe(want))
})

describe('labels and values', () => {
  it.each([
    [0, null],
    [1, '1 propozycja czeka'],
    [2, '2 propozycje czekają'],
    [5, '5 propozycji czeka'],
    [12, '12 propozycji czeka'],
    [22, '22 propozycje czekają'],
  ])('pendingText(%s) = %s', (n, want) => expect(pendingText(n)).toBe(want))

  it('proposedMeta: today, yesterday, older', () => {
    const [a, b] = proposalsFull.pending
    expect(proposedMeta(a!, '2026-10-25')).toBe('dziś 06:31 · run dzienny')
    expect(proposedMeta(b!, '2026-10-25')).toBe('wczoraj 06:30 · run dzienny')
    expect(proposedMeta(b!, '2026-10-27')).toBe('24.10 06:30 · run dzienny')
  })

  it.each([
    [2950, 'kcal', false, `2${NB}950`],
    [3140, 'kcal', true, `3${NB}140 kcal`],
    [412, 'g', true, '412 g'],
    [4, null, true, '4'],
    ['80 g węgli', null, true, '80 g węgli'],
    [null, 'g', true, '–'],
  ] as const)('diffValue(%s, %s, %s) = %s', (v, unit, withUnit, want) => expect(diffValue(v, unit, withUnit)).toBe(want))

  it('changedRows drops unchanged rows, keeps the coach order', () => {
    expect(changedRows(proposalsFull.pending[0]!).map((r) => r.label)).toEqual(['Węglowodany', 'Kalorie', 'Przed biegiem'])
  })

  it('diffHeading adds the effective day when the title has none', () => {
    expect(diffHeading(proposalsFull.pending[0]!)).toBe('Nd 01.11 · cel dnia')
    expect(diffHeading(proposalsFull.pending[1]!)).toBe('Serie robocze')
  })

  it('dayNumbers', () => expect(dayNumbers({ session: 'x', runKm: null, kcal: 2750, carbsG: 370 })).toBe(`2${NB}750 kcal · W 370`))
})

describe('mergeHistory', () => {
  const entry = (id: string, decidedAt: string, decision: ProposalDecision['decision'] = 'accepted'): ProposalDecision => ({
    id,
    kind: 'run',
    origin: 'daily_run',
    title: id,
    proposedAt: decidedAt,
    decidedAt,
    decision,
    comment: null,
  })

  it('newest decision first, across offsets', () => {
    const got = mergeHistory(
      [entry('a', '2026-10-11T08:40:00+02:00'), entry('b', '2026-10-18T21:15:00+02:00')],
      [entry('c', '2026-10-25T07:00:00+01:00')],
    )
    expect(got.map((d) => d.id)).toEqual(['c', 'b', 'a'])
  })

  it('server copy wins over the local one', () => {
    const got = mergeHistory([entry('a', '2026-10-25T07:00:05+01:00', 'rejected')], [entry('a', '2026-10-25T07:00:00+01:00')])
    expect(got).toHaveLength(1)
    expect(got[0]!.decision).toBe('rejected')
  })
})

describe('columns', () => {
  it('fold, mockup: cards alternate, blocked left, history right', () => {
    expect(columns('fold', 2, true, true)).toEqual([
      [{ kind: 'card', index: 0 }, { kind: 'blocked' }],
      [{ kind: 'card', index: 1 }, { kind: 'history' }],
    ])
  })
  it('fold, one card', () => {
    expect(columns('fold', 1, true, true)).toEqual([[{ kind: 'card', index: 0 }, { kind: 'blocked' }], [{ kind: 'history' }]])
  })
  it('fold, stage1: only the empty box', () => {
    expect(columns('fold', 0, false, false)).toEqual([[{ kind: 'empty' }]])
  })
  it('cover: one column in mockup order', () => {
    expect(columns('cover', 2, true, true)).toEqual([
      [{ kind: 'card', index: 0 }, { kind: 'card', index: 1 }, { kind: 'blocked' }, { kind: 'history' }],
    ])
  })
})

describe('weekRows', () => {
  const week = proposalsFull.week!
  it('applies pending changes and marks those days for a decision', () => {
    const rows = weekRows(week, new Set(['p-2026-10-25-carbs', 'p-2026-10-24-legs']))
    expect(rows.map((r) => r.label)).toEqual(['Pn 26', 'Wt 27', 'Śr 28', 'Cz 29', 'Pt 30', 'Sb 31', 'Nd 1'])
    expect(rows.filter((r) => r.decide).map((r) => r.label)).toEqual(['Śr 28', 'Nd 1'])
    expect(rows[2]).toMatchObject({ sessionChanged: true, numbersChanged: false })
    expect(rows[6]).toMatchObject({ sessionChanged: false, numbersChanged: true })
    expect(rows[6]!.after.carbsG).toBe(460)
    expect(weekKm(rows, 'now')).toBe(46)
    expect(weekKm(rows, 'after')).toBe(46)
  })
  it('ignores changes of decided proposals', () => {
    const rows = weekRows(week, new Set(['p-2026-10-24-legs']))
    expect(rows[6]).toMatchObject({ decide: false, numbersChanged: false })
    expect(rows[6]!.after.carbsG).toBe(412)
  })
})

describe('mock API state', () => {
  it('accept moves the proposal to history and into the week plan', () => {
    const state = structuredClone(proposalsFull)
    const d = applyMockDecision(state, 'p-2026-10-25-carbs', 'accept', 'ok', new Date('2026-10-25T08:00:00Z'))
    expect(d).toMatchObject({ decision: 'accepted', comment: 'ok', decidedAt: '2026-10-25T08:00:00.000Z' })
    expect(state.pending.map((p) => p.id)).toEqual(['p-2026-10-24-legs'])
    expect(state.history[0]!.id).toBe('p-2026-10-25-carbs')
    expect(state.week!.days[6]!.now.carbsG).toBe(460)
    expect(state.week!.days[6]!.changes).toEqual([])
  })
  it('reject keeps the plan; the last decision drops the week diff; deciding twice is a conflict', () => {
    const state = structuredClone(proposalsFull)
    applyMockDecision(state, 'p-2026-10-24-legs', 'reject', null)
    expect(state.week!.days[2]!.now.session).toBe('Legs 16 serii')
    applyMockDecision(state, 'p-2026-10-25-carbs', 'reject', null)
    expect(state.week).toBeNull()
    expect(() => applyMockDecision(state, 'p-2026-10-25-carbs', 'accept', null)).toThrow('409')
  })
  it('stage1 is empty', () => {
    expect(proposalsStage1).toMatchObject({ pending: [], blocked: [], history: [], week: null })
  })
})
