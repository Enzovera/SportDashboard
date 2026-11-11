import { describe, expect, it } from 'vitest'
import { planFull } from '@/api/mocks/plan'
import type { PlanDay } from '@/api/plan'
import {
  dayLines,
  decideCount,
  doneCell,
  fullDate,
  orderWeeks,
  raceTime,
  timeline,
  weekMeta,
  weekTitle,
} from './planData'

const plan = planFull.plan!
const cur = plan.currentWeek!
const next = plan.nextWeek!
const TODAY = '2026-10-25'

describe('format', () => {
  it('race time and date', () => {
    expect(raceTime(14400)).toBe('4:00:00')
    expect(raceTime(5999)).toBe('1:39:59')
    expect(fullDate('2027-03-14')).toBe('14.03.2027')
  })
})

describe('dayLines', () => {
  const [push, intervals, legs, rest] = next.days as [PlanDay, PlanDay, PlanDay, PlanDay]
  it('narrow layouts carry km in the detail', () => {
    expect(dayLines(push, true)).toEqual({ title: 'Push', detail: '+ bieg łatwy 6 km', rest: false })
    expect(dayLines(intervals, true).detail).toBe('10 km · 6×800 m')
    expect(dayLines(next.days[6]!, true).detail).toBe('16 km · strefa 2')
    expect(dayLines(legs, true).detail).toBe('16 serii roboczych')
  })
  it('wide layouts have a km column', () => {
    expect(dayLines(push, false).detail).toBe('+ bieg łatwy')
    expect(dayLines(intervals, false).detail).toBe('6×800 m')
    expect(dayLines(next.days[5]!, false).detail).toBeNull()
  })
  it('rest day', () => expect(dayLines(rest, true)).toEqual({ title: 'Odpoczynek', detail: null, rest: true }))
})

describe('doneCell', () => {
  it('current week', () => {
    expect(cur.days.map((d) => doneCell(d, TODAY).text)).toEqual([
      '6,0 km · 18 serii',
      '10,2 km',
      '16 serii',
      '–',
      '6,1 km · 17 serii',
      '8,0 km',
      'dziś',
    ])
    expect(doneCell(cur.days[6]!, TODAY).tone).toBe('signal')
  })
  it('missed session is a warning, future is a dash', () => {
    const missed: PlanDay = { ...cur.days[1]!, done: { runKm: null, strengthSets: null } }
    expect(doneCell(missed, TODAY)).toEqual({ text: 'pominięte', tone: 'warn' })
    expect(doneCell(next.days[0]!, TODAY)).toEqual({ text: '–', tone: 'muted' })
  })
})

describe('weeks', () => {
  it('titles and meta', () => {
    expect(weekTitle(cur, TODAY)).toBe('Ten tydzień · 19–25.10')
    expect(weekTitle(next, TODAY)).toBe('Tydzień 26.10–01.11')
    expect(weekMeta(cur)).toBe('bieg 30 / 46 km · siła 3 / 3')
    expect(weekMeta(next)).toBe('Baza · tydz. 1 · 46 km')
  })
  it('do decyzji', () => {
    expect(decideCount(next)).toBe(2)
    expect(decideCount(cur)).toBe(0)
  })
  it('order: next first on the last day of the week', () => {
    expect(orderWeeks(cur, next, TODAY).map((w) => w.from)).toEqual(['2026-10-26', '2026-10-19'])
    expect(orderWeeks(cur, next, '2026-10-22').map((w) => w.from)).toEqual(['2026-10-19', '2026-10-26'])
    expect(orderWeeks(null, next, TODAY)).toEqual([next])
  })
})

describe('timeline', () => {
  it('before the start: next block is current, no red week', () => {
    const tl = timeline(plan.blocks, TODAY)
    expect(tl.now).toBe('Dziś · 25.10 · Baza od jutra')
    expect(tl.beforeStart).toBe(true)
    expect(tl.meta).toBe('20 tyg. · 26.10–14.03')
    expect(tl.blocks.map((b) => b.current)).toEqual([true, false, false, false])
    expect(tl.blocks[0]!.segments).toEqual(Array(6).fill('block'))
    expect(tl.blocks[3]!.dates).toBe('3 tyg. · 22.02–14.03')
    expect(tl.blocks[3]!.range).toBe('22.02–14.03')
  })
  it('inside a block: past weeks, the red week, rest of the block', () => {
    const tl = timeline(plan.blocks, '2026-11-12')
    expect(tl.now).toBe('Dziś · 12.11 · Baza, tydz. 3 z 6')
    expect(tl.beforeStart).toBe(false)
    expect(tl.blocks[0]!.segments).toEqual(['past', 'past', 'now', 'block', 'block', 'block'])
    expect(tl.blocks[1]!.segments).toEqual(Array(6).fill('future'))
  })
  it('later start date', () => {
    expect(timeline(plan.blocks, '2026-10-20').now).toBe('Dziś · 20.10 · Baza od 26.10')
  })
})
