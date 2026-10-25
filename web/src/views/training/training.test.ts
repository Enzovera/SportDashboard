import { describe, expect, it } from 'vitest'
import { trainingFull, trainingStage1 } from '@/api/mocks/training'
import { labelIndices, loadOption, niceRange, weeklyKmMax, weeklyKmOption, type ChartPalette } from './chartOptions'
import {
  atlSub,
  barShares,
  ctlSub,
  kmSub,
  loadAriaLabel,
  pctChange,
  rangeWeeks,
  weeklyKmAriaLabel,
  weeksWord,
} from './trainingData'

const palette: ChartPalette = {
  chart: 'INK',
  chartMid: 'MID',
  chartSoft: 'SOFT',
  grid: 'GRID',
  axis: 'AXIS',
  ink2: 'INK2',
  ink3: 'INK3',
  signal: 'RED',
  font: 'Archivo',
}

// Loose access to the composed option in assertions.
// eslint-disable-next-line @typescript-eslint/no-explicit-any
type Any = any

describe('trainingData', () => {
  it.each([
    [46, 43, 6.98],
    [40, 50, -20],
    [10, null, null],
    [10, 0, null],
  ])('pctChange(%s, %s)', (cur, prev, want) => {
    const got = pctChange(cur, prev)
    if (want === null) expect(got).toBeNull()
    else expect(got).toBeCloseTo(want, 2)
  })

  it('formats metric subtitles like the mockup', () => {
    expect(kmSub(trainingFull.load)).toBe('+7% vs 43 km')
    expect(kmSub({ ...trainingFull.load, prevWeekKm: null })).toBeNull()
    expect(ctlSub(trainingFull.load, 12)).toBe('+11 w 12 tyg.')
    expect(ctlSub({ ...trainingFull.load, ctlDeltaRange: null }, 12)).toBeNull()
    expect(atlSub(trainingFull.load)).toBe('−5 vs poprz. tydz.')
  })

  it.each([
    ['2026-08-03', '2026-10-25', 12],
    ['2026-09-28', '2026-10-25', 4],
    ['2026-10-19', '2026-10-25', 1],
  ])('rangeWeeks(%s, %s) = %i', (from, to, want) => {
    expect(rangeWeeks(from, to)).toBe(want)
  })

  it.each([
    [1, 'tydzień'],
    [2, 'tygodnie'],
    [4, 'tygodnie'],
    [5, 'tygodni'],
    [12, 'tygodni'],
    [14, 'tygodni'],
    [22, 'tygodnie'],
    [25, 'tygodni'],
  ])('weeksWord(%i) = %s', (n, want) => {
    expect(weeksWord(n)).toBe(want)
  })

  it('scales bars to the max', () => {
    expect(barShares([12, 6, 3])).toEqual([1, 0.5, 0.25])
    expect(barShares([0, 0])).toEqual([0, 0])
    expect(barShares([])).toEqual([])
  })

  it('describes the weekly km chart', () => {
    expect(weeklyKmAriaLabel(trainingFull.weeklyKm, trainingFull.todayPlannedKm)).toBe(
      'Kilometraż tygodniowy od 30 do 43 km w 12 tygodni; odciążenie w tygodniach 24.08 i 21.09; ' +
        'plan na ostatni tydzień 46 km; dziś planowane 16 km',
    )
    expect(weeklyKmAriaLabel(trainingStage1.weeklyKm, null)).toBe(
      'Kilometraż tygodniowy od 30 do 43 km w 12 tygodni; odciążenie w tygodniach 24.08 i 21.09',
    )
    expect(weeklyKmAriaLabel([], null)).toBe('Kilometraż tygodniowy')
  })

  it('describes the load chart', () => {
    expect(loadAriaLabel(trainingFull.loadSeries)).toBe('CTL z 30 do 41; ATL od 29 do 43, na koniec 37')
  })
})

describe('chart scales', () => {
  it.each([
    [0, 46, { min: 0, max: 50, interval: 10 }],
    [0, 30, { min: 0, max: 30, interval: 10 }],
    [0, 0, { min: 0, max: 10, interval: 10 }],
    [0, 8, { min: 0, max: 10, interval: 10 }],
    [0, 95, { min: 0, max: 100, interval: 20 }],
    [0, 120, { min: 0, max: 125, interval: 25 }],
    [29, 43, { min: 20, max: 50, interval: 10 }],
  ])('niceRange(%s, %s)', (min, max, want) => {
    expect(niceRange(min, max)).toEqual(want)
  })

  it.each([
    [12, [0, 4, 8, 11]],
    [4, [0, 1, 2, 3]],
    [1, [0]],
    [0, []],
    [10, [0, 3, 6, 9]],
    [9, [0, 3, 6, 8]],
    [2, [0, 1]],
    [26, [0, 9, 18, 25]],
  ])('labelIndices(%i)', (n, want) => {
    expect([...labelIndices(n)].sort((a, b) => a - b)).toEqual(want)
  })

  it('includes plan and today in the km axis max', () => {
    expect(weeklyKmMax(trainingFull.weeklyKm, 16)).toBe(46)
    expect(weeklyKmMax(trainingStage1.weeklyKm, null)).toBe(43)
    expect(weeklyKmMax([], null)).toBe(0)
  })
})

describe('weeklyKmOption', () => {
  it('draws bars, plan ticks and the today box with full data', () => {
    const o: Any = weeklyKmOption(trainingFull.weeklyKm, 16, palette)
    expect(o.series).toHaveLength(3)
    const [bars, plan, today] = o.series
    expect(bars.type).toBe('bar')
    expect(bars.data[0]).toEqual({ value: 31, itemStyle: { color: 'MID' } })
    expect(bars.data[3].itemStyle.color).toBe('SOFT')
    expect(plan.type).toBe('custom')
    expect(plan.data).toHaveLength(12)
    expect(today.data).toEqual([[11, 30, 46]])
    expect(o.yAxis).toMatchObject({ min: 0, max: 50, interval: 10 })
    expect(o.xAxis.data[0]).toBe('03.08')
    const shown = o.xAxis.data.filter((_: string, i: number) => o.xAxis.axisLabel.interval(i))
    expect(shown).toEqual(['03.08', '31.08', '28.09', '19.10'])
  })

  it('renders the today box in red with its value', () => {
    const o: Any = weeklyKmOption(trainingFull.weeklyKm, 16, palette)
    const api = {
      value: (d: number) => [11, 30, 46][d],
      coord: ([x, y]: number[]) => [x! * 10, 200 - y! * 2],
      size: () => [20, 0],
    }
    const el = o.series[2].renderItem({}, api)
    expect(el.children[0].style).toMatchObject({ stroke: 'RED', fill: 'none' })
    expect(el.children[0].shape).toMatchObject({ y: 108, height: 32, width: 9 })
    expect(el.children[1].style).toMatchObject({ text: '16', fill: 'RED' })
  })

  it('omits plan and today without a plan (stage 1)', () => {
    const o: Any = weeklyKmOption(trainingStage1.weeklyKm, null, palette)
    expect(o.series).toHaveLength(1)
    expect(o.series[0].type).toBe('bar')
  })

  it('skips plan ticks for weeks without a plan', () => {
    const weeks = trainingFull.weeklyKm.map((w, i) => (i < 6 ? { ...w, plannedKm: null } : w))
    const o: Any = weeklyKmOption(weeks, null, palette)
    expect(o.series).toHaveLength(2)
    expect(o.series[1].data.map((d: number[]) => d[0])).toEqual([6, 7, 8, 9, 10, 11])
  })
})

describe('loadOption', () => {
  it('draws CTL solid on top and ATL dashed with end markers and labels', () => {
    const o: Any = loadOption(trainingFull.loadSeries, palette)
    const [atl, ctl] = o.series
    expect(ctl.lineStyle).toMatchObject({ color: 'INK', width: 2.5, type: 'solid' })
    expect(atl.lineStyle).toMatchObject({ color: 'MID', type: [5, 3] })
    const ctlEnd = ctl.data[ctl.data.length - 1]
    expect(ctlEnd).toMatchObject({ value: 41, symbol: 'rect', symbolSize: 7 })
    expect(ctlEnd.label).toMatchObject({ formatter: 'CTL 41', position: 'top', fontWeight: 800 })
    expect(atl.data[atl.data.length - 1].label).toMatchObject({ formatter: 'ATL 37', position: 'bottom' })
    expect(ctl.data[0]).toBe(30)
    expect(o.yAxis).toMatchObject({ min: 20, max: 50, interval: 10 })
    const shown = o.xAxis.data.filter((_: string, i: number) => o.xAxis.axisLabel.interval(i))
    expect(shown).toEqual(['03.08', '19.10'])
  })

  it('puts ATL label on top when ATL is higher', () => {
    const o: Any = loadOption([{ date: '2026-10-19', ctl: 30, atl: 40 }], palette)
    const [atl, ctl] = o.series
    expect(atl.data[0].label.position).toBe('top')
    expect(ctl.data[0].label.position).toBe('bottom')
  })
})
