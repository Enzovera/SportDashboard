import { describe, expect, it } from 'vitest'
import { todayFull, todayStage1 } from '@/api/mocks/today'
import {
  headerMeta,
  nutritionLines,
  pluralPl,
  proposalsText,
  sectionColumns,
  desktopCells,
  sessionMeta,
  weekKm,
  weekTable,
  weekTotals,
} from './today'

const NB = ' '

describe('sectionColumns', () => {
  it('fold, full data: mockup order, supplements inside nutrition', () => {
    expect(sectionColumns(todayFull, 'fold')).toEqual([
      ['coach', 'session', 'week', 'tomorrow'],
      ['readiness', 'nutrition', 'proposals', 'weight'],
    ])
  })

  it('cover, full data: one column, supplements standalone', () => {
    expect(sectionColumns(todayFull, 'cover')).toEqual([
      ['coach', 'session', 'readiness', 'nutrition', 'supplements', 'proposals', 'week', 'weight'],
    ])
  })

  it('stage 1 keeps only week and readiness', () => {
    expect(sectionColumns(todayStage1, 'fold')).toEqual([['week'], ['readiness']])
    expect(sectionColumns(todayStage1, 'cover')).toEqual([['readiness', 'week']])
  })

  it('fold shows supplements standalone when nutrition is missing', () => {
    const d = { ...todayFull, nutrition: null }
    expect(sectionColumns(d, 'fold')[1]).toEqual(['readiness', 'supplements', 'proposals', 'weight'])
  })

  it('hides proposals when none are pending', () => {
    const d = { ...todayFull, proposalsPending: 0 }
    expect(sectionColumns(d, 'cover')[0]).not.toContain('proposals')
  })
})

describe('headerMeta', () => {
  it.each([
    [20, 'Nd 25.10 · 20 tyg. do maratonu'],
    [null, 'Nd 25.10'],
  ])('weeks %s -> %s', (weeks, want) => expect(headerMeta({ date: '2026-10-25', raceWeeksLeft: weeks })).toBe(want))
})

describe('sessionMeta', () => {
  it.each([
    ['keep', false, 'bez zmian w planie'],
    ['keep', true, 'bez zmian'],
    ['modify', false, 'zmienione przez trenera'],
    ['rest', true, 'dzień wolny'],
  ] as const)('%s compact=%s -> %s', (action, compact, want) => expect(sessionMeta(action, compact)).toBe(want))
})

describe('week totals', () => {
  it('with plan', () => {
    expect(weekKm(todayFull.week)).toBe('30 / 46 km')
    expect(weekTotals(todayFull.week)).toBe('Bieg 30 / 46 km · siła 3 / 3')
  })
  it('without plan', () => {
    expect(weekKm(todayStage1.week)).toBe('30 km')
    expect(weekTotals(todayStage1.week)).toBe('Bieg 30 km · siła 3')
  })
})

describe('weekTable', () => {
  it('marks today and the red cell only when today has a run', () => {
    const t = weekTable(todayFull.week, todayFull.date)
    expect(t.heads.map((h) => h.label)).toEqual(['Pn', 'Wt', 'Śr', 'Cz', 'Pt', 'Sb', 'Nd'])
    expect(t.heads.map((h) => h.today)).toEqual([false, false, false, false, false, false, true])
    expect(t.run.map((c) => c.text)).toEqual(['6', '10', '–', '–', '6', '8', '16'])
    expect(t.run.map((c) => c.signal)).toEqual([false, false, false, false, false, false, true])
    expect(t.run[2]?.empty).toBe(true)
    expect(t.strength.map((c) => c.text)).toEqual(['Push', '–', 'Legs', '–', 'Pull', '–', '–'])
  })

  it('no signal cell when today has no run (stage 1)', () => {
    const t = weekTable(todayStage1.week, todayStage1.date)
    expect(t.run[6]).toEqual({ text: '–', empty: true, signal: false, planned: false })
    expect(t.heads[6]?.today).toBe(true)
  })

  it('future planned runs are flagged as planned, not signal', () => {
    const week = {
      ...todayFull.week,
      days: todayFull.week.days.map((d) => (d.date === '2026-10-24' ? { ...d, runPlanned: true } : d)),
    }
    const t = weekTable(week, '2026-10-23')
    expect(t.run[5]).toEqual({ text: '8', empty: false, signal: false, planned: true })
    expect(t.run[4]?.signal).toBe(true)
  })
})

describe('nutritionLines', () => {
  it('computes what is left', () => {
    const rows = nutritionLines(todayFull.nutrition!)
    expect(rows.map((r) => [r.label, r.eaten, r.target, r.left])).toEqual([
      ['Kalorie', '820', `2${NB}950`, `2${NB}130`],
      ['Białko', '38', '145', '107'],
      ['Węglowodany', '112', '412', '300'],
      ['Tłuszcz', '24', '80', '56'],
    ])
  })
})

describe('proposalsText', () => {
  it.each([
    [1, 'one'],
    [2, 'few'],
    [4, 'few'],
    [5, 'many'],
    [12, 'many'],
    [22, 'few'],
  ] as const)('pluralPl(%s) = %s', (n, want) => expect(pluralPl(n)).toBe(want))

  it.each([
    [1, false, 'propozycja zmiany czeka'],
    [2, false, 'propozycje zmian czekają'],
    [5, false, 'propozycji zmian czeka'],
    [2, true, 'propozycje zmian'],
  ] as const)('%s compact=%s -> %s', (n, compact, want) => expect(proposalsText(n, compact)).toBe(want))
})

describe('desktopCells', () => {
  const shape = (d: Parameters<typeof desktopCells>[0]) => desktopCells(d).map((c) => [c.keys.join('+'), c.span])

  it('full: hero and coach, then nutrition / supplements / readiness, then week and link rows', () => {
    expect(shape(todayFull)).toEqual([
      ['session', 6],
      ['coach', 6],
      ['nutrition', 4],
      ['supplements', 4],
      ['readiness', 4],
      ['week', 8],
      ['tomorrow+proposals+weight', 4],
    ])
  })

  it('stage1: readiness and week share one row', () => {
    expect(shape(todayStage1)).toEqual([
      ['readiness', 4],
      ['week', 8],
    ])
  })

  it('a lone hero takes the whole row', () => {
    expect(shape({ ...todayFull, coach: null }).slice(0, 2)).toEqual([
      ['session', 12],
      ['nutrition', 4],
    ])
  })
})
