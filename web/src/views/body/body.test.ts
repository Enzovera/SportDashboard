import { describe, expect, it } from 'vitest'
import { bodyFull, bodyStage1 } from '@/api/mocks/body'
import {
  addDays,
  changeLabel,
  compositionCells,
  missingLabel,
  trendTitle,
  weighInRows,
  weightChart,
  weightChartLabel,
} from './bodyData'

describe('trendTitle', () => {
  it.each([
    [-0.15, 'Waga spada o ok. 0,15 kg tygodniowo'],
    [0.2, 'Waga rośnie o ok. 0,2 kg tygodniowo'],
    [0.02, 'Waga stabilna'],
    [null, 'Waga tydzień po tygodniu'],
  ])('%s -> %s', (v, want) => expect(trendTitle(v)).toBe(want))
})

describe('labels', () => {
  it('missing days', () => {
    expect(missingLabel([])).toBeNull()
    expect(missingLabel(['2026-10-21'])).toBe('brak: środa')
    expect(missingLabel(['2026-10-21', '2026-10-23'])).toBe('brak: środa, piątek')
    expect(missingLabel(['2026-10-19', '2026-10-20', '2026-10-21'])).toBe('brak: 3 dni')
  })
  it('change over the range', () => {
    expect(changeLabel(-1.7, 12)).toBe('−1,7 kg w 12 tyg.')
    expect(changeLabel(0.4, 12)).toBe('+0,4 kg w 12 tyg.')
    expect(changeLabel(null, 12)).toBeNull()
  })
  it('addDays crosses months', () => expect(addDays('2026-10-30', 3)).toBe('2026-11-02'))
})

describe('compositionCells', () => {
  it('full: four cells as in the mockup', () => {
    const cells = compositionCells(bodyFull.composition, 12)
    expect(cells.map((c) => [c.label, c.value, c.sub])).toEqual([
      ['Waga · śr. 7 dni', '76,4', '−1,7 kg w 12 tyg.'],
      ['Tkanka tłuszczowa', '15,8', 'szacunek wagi'],
      ['Masa mięśniowa', '61,3', 'stabilna'],
      ['Ważenia · tydzień', '6', 'brak: środa'],
    ])
  })
  it('stage1: last weigh-in and the range count', () => {
    const cells = compositionCells(bodyStage1.composition, 12)
    expect(cells.map((c) => [c.label, c.value, c.sub])).toEqual([
      ['Ostatnie ważenie', '77,0', '14.10'],
      ['Ważenia · 12 tyg.', '3', null],
    ])
  })
})

describe('weighInRows', () => {
  it('regular weighing: 7 calendar days with gaps', () => {
    const rows = weighInRows('2026-10-25', bodyFull.weighIns)
    expect(rows.map((r) => [r.date, r.weighIn?.weightKg ?? null])).toEqual([
      ['2026-10-25', 76.2],
      ['2026-10-24', 76.6],
      ['2026-10-23', 76.3],
      ['2026-10-22', 76.9],
      ['2026-10-21', null],
      ['2026-10-20', 76.4],
      ['2026-10-19', 76.1],
    ])
  })
  it('skips today when not weighed yet', () => {
    const rows = weighInRows('2026-10-26', bodyFull.weighIns)
    expect(rows[0]!.date).toBe('2026-10-25')
    expect(rows).toHaveLength(6)
  })
  it('sparse weighing: only the weigh-ins', () => {
    const rows = weighInRows('2026-10-25', bodyStage1.weighIns)
    expect(rows.map((r) => r.date)).toEqual(['2026-10-14', '2026-09-16', '2026-08-21'])
  })
})

describe('weightChart', () => {
  it('full: 12 points, current week red, line through all weeks', () => {
    const c = weightChart(bodyFull.weeks, '2026-10-25', 300, 196)!
    expect(c.points).toHaveLength(12)
    expect(c.points.filter((p) => p.current)).toHaveLength(1)
    expect(c.points[11]!.current).toBe(true)
    expect(c.segments).toHaveLength(1)
    expect(c.whiskers).toHaveLength(12)
    expect(c.yTicks.map((t) => t.label)).toEqual(['76', '77', '78', '79'])
    expect(c.last).toMatchObject({ label: '76,4', current: true })
    // higher weight is drawn higher
    expect(c.points[0]!.y).toBeLessThan(c.points[11]!.y)
    // labels thinned on a narrow chart, always including the latest week
    expect(c.xLabels.length).toBeLessThan(12)
    expect(c.xLabels[c.xLabels.length - 1]!.label).toBe('19.10')
  })
  it('wide chart labels every week', () => {
    const c = weightChart(bodyFull.weeks, '2026-10-25', 840, 278)!
    expect(c.xLabels).toHaveLength(12)
  })
  it('stage1: isolated points, no lines, no whiskers, nothing red', () => {
    const c = weightChart(bodyStage1.weeks, '2026-10-25', 300, 196)!
    expect(c.points).toHaveLength(3)
    expect(c.segments).toHaveLength(0)
    expect(c.whiskers).toHaveLength(0)
    expect(c.points.some((p) => p.current)).toBe(false)
    expect(c.last).toMatchObject({ label: '77,0', current: false })
  })
  it('no data -> null', () => {
    expect(weightChart([], '2026-10-25', 300, 196)).toBeNull()
  })
  it('aria label', () => {
    expect(weightChartLabel(bodyFull.weeks)).toBe(
      'Średnia tygodniowa wagi z 78,1 do 76,4 kg w 12 tygodni; zakres ważeń w tygodniu do 1,2 kg',
    )
  })
})
