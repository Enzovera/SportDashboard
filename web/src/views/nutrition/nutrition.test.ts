import { describe, expect, it } from 'vitest'
import { nutritionFull, nutritionStage1 } from '@/api/mocks/nutrition'
import { supplementsFull, supplementsStage1 } from '@/api/mocks/supplements'
import type { NutritionDay } from '@/api/nutrition'
import type { SupplementItem } from '@/api/supplements'
import { kcalOption, proteinOption, proteinRange, type NutritionPalette } from './nutritionCharts'
import {
  balanceCells,
  headerMeta,
  isKeyDay,
  kcalAriaLabel,
  kcalHeadline,
  mealMacros,
  pct,
  proposalValue,
  proposalWhere,
  proteinAriaLabel,
  proteinHeadline,
  remaining,
  supplementsTabLabel,
  targetMeta,
  todayTitle,
} from './nutritionData'
import {
  counts,
  doseLabel,
  itemSub,
  nextStatus,
  regularityDays,
  replaceItem,
  slotGroups,
  stockLasts,
  stockLeft,
  supplementsMeta,
  whenLabel,
  withStatus,
} from './supplementsData'

// eslint-disable-next-line @typescript-eslint/no-explicit-any
type Any = any

const palette: NutritionPalette = {
  chart: 'INK',
  chartMid: 'MID',
  chartSoft: 'SOFT',
  grid: 'GRID',
  axis: 'AXIS',
  ink2: 'INK2',
  ink3: 'INK3',
  signal: 'RED',
  font: 'Archivo',
  ink: 'INK',
  surface: 'SURF',
  good: 'GOOD',
  warn: 'WARN',
}

/** Expected strings use plain spaces for the NBSP thousands separator of lib/format. */
const sp = (s: string | null | undefined) => (s ?? '').replace(/ /g, ' ')

const balance = nutritionFull.balance!
const days = nutritionFull.days

describe('nutrition labels', () => {
  it('builds header meta from Fitatu sync and range', () => {
    expect(headerMeta(nutritionFull)).toBe('Fitatu 08:15 · 12–25.10')
    expect(headerMeta(nutritionStage1)).toBe('12–25.10')
  })

  it('formats the balance cells like the mockup', () => {
    const cells = balanceCells(balance)
    expect(cells.map((c) => c.value)).toEqual(['−5%', '1,73', '10', '4,8'])
    expect(sp(cells[0]!.sub)).toBe('śr. 2 528 z 2 658')
    expect(cells[1]!.sub).toBe('cel 1,9 · min 1,6')
    expect(cells[2]).toMatchObject({ unit: '/ 13', sub: '3 dni poniżej 1,6', subTone: 'warn' })
    expect(cells[3]).toMatchObject({ label: 'Węgle · długie 18.10', sub: 'cel 5,4 g/kg', subTone: 'warn' })
  })

  it('drops the carbs cell without a long run and says when all days are fine', () => {
    const cells = balanceCells({ ...balance, longRunCarbs: null, proteinDaysAtMin: 13 })
    expect(cells).toHaveLength(3)
    expect(cells[2]).toMatchObject({ sub: 'wszystkie', subTone: 'good' })
    expect(balanceCells({ ...balance, proteinDaysAtMin: 12 })[2]!.sub).toBe('1 dzień poniżej 1,6')
  })

  it('names the day type with the largest deficit', () => {
    expect(kcalHeadline(days)).toBe('Największy niedobór w dni długiego wybiegania')
    const fine = days.map((d) => ({ ...d, kcalVsTargetPct: d.complete ? -1 : null }))
    expect(kcalHeadline(fine)).toBe('Kalorie blisko celu')
    expect(kcalHeadline([])).toBe('Kalorie dzień po dniu')
  })

  it('describes how often protein reaches the target', () => {
    expect(proteinHeadline(days, balance)).toBe('Białko rzadko dobija do celu')
    const high = days.map((d) => ({ ...d, proteinGPerKg: d.complete ? 2 : null }))
    expect(proteinHeadline(high, balance)).toBe('Białko zwykle na celu')
    const mid = days.map((d) => ({ ...d, proteinGPerKg: d.complete ? 1.7 : null }))
    expect(proteinHeadline(mid, balance)).toBe('Białko powyżej minimum')
  })

  it('formats today, proposals and meals', () => {
    expect(todayTitle('long_run')).toBe('Dziś · długie wybieganie')
    expect(todayTitle(null)).toBe('Dziś')
    expect(targetMeta('long_run')).toBe('cel dnia długiego')
    expect(remaining({ eaten: 820, target: 2950 })).toBe(2130)
    expect(remaining({ eaten: 3000, target: 2950 })).toBe(-50)
    const p = nutritionFull.targetProposals[0]!
    expect(`Propozycja: ${proposalValue(p)} ${proposalWhere(p)}`).toBe('Propozycja: W 460 g w dniu długiego')
    expect(mealMacros(nutritionFull.meals[0]!)).toBe('820 kcal · B 38 · W 112 · T 24')
    expect(pct(-5)).toBe('−5%')
    expect(pct(1)).toBe('+1%')
    expect(supplementsTabLabel({ taken: 2, total: 7, next: null })).toBe('Suplementy 2/7')
    expect(supplementsTabLabel(null)).toBe('Suplementy')
  })

  it('marks only the long-run deficit day as key', () => {
    expect(days.filter(isKeyDay).map((d) => d.date)).toEqual(['2026-10-18'])
  })

  it('writes chart aria labels', () => {
    expect(sp(kcalAriaLabel(days))).toBe(
      'Kalorie zjedzone w 13 dni na tle celu dnia; 18.10, długie wybieganie: zjedzone 2 610 przy celu 2 950 kcal; dziś 820 kcal w toku',
    )
    expect(proteinAriaLabel(days, balance)).toBe(
      'Białko w g na kg od 1,53 do 1,95; cel 1,9 osiągnięty w 2 z 13 dni, 3 dni poniżej minimum 1,6',
    )
  })
})

describe('nutrition charts', () => {
  it('draws eaten bars, target ticks, today outline and the key deficit', () => {
    const o = kcalOption(days, palette) as Any
    expect(o.yAxis).toMatchObject({ min: 0, max: 3000, interval: 1000 })
    const bars = o.series[0].data
    expect(bars).toHaveLength(14)
    expect(bars[6]).toEqual({ value: 2610, itemStyle: { color: 'INK' } })
    expect(bars[0].itemStyle.color).toBe('MID')
    expect(bars[13]).toBe('-')
    expect(o.series[1].data).toHaveLength(14)
    expect(o.series[2].data).toEqual([[13, 0, 820]])
    expect(o.series[3].data).toEqual([[6, 2950, -340]])
    const fmt = o.xAxis.axisLabel.formatter
    expect(fmt('2026-10-18', 6)).toBe('{k|18}\n{k|D}')
    expect(fmt('2026-10-25', 13)).toBe('{s|25}\n{s|D}')
    expect(fmt('2026-10-12', 0)).toBe('{n|12}\n{n|SB}')
  })

  it('omits today outline and deficits when there are none', () => {
    const plain: NutritionDay[] = days.slice(0, 5)
    const o = kcalOption(plain, palette) as Any
    expect(o.series).toHaveLength(2)
  })

  it('ranges protein in 0,2 steps around target and minimum', () => {
    expect(proteinRange([1.53, 1.95], 1.9, 1.6)).toEqual({ min: 1.4, max: 2, interval: 0.2 })
    expect(proteinRange([1.8], 1.9, 1.6)).toEqual({ min: 1.4, max: 2, interval: 0.2 })
  })

  it('draws protein points, hollow amber below the minimum', () => {
    const compact = proteinOption(days, balance, palette) as Any
    const line = compact.series[1].data
    expect(line).toHaveLength(13)
    expect(line[3]).toMatchObject({ value: 1.58, symbolSize: 9, itemStyle: { borderColor: 'WARN' } })
    expect(line[0]).toMatchObject({ value: 1.62, symbolSize: 7 })
    const wide = proteinOption(days, balance, palette, true) as Any
    expect(wide.series[1].data).toHaveLength(14)
    expect(wide.series[1].data[13]).toBe('-')
  })
})

describe('supplements', () => {
  const list = () => structuredClone(supplementsFull().checklist)

  it('groups the checklist by time of day', () => {
    const groups = slotGroups(list(), 'long_run')
    expect(groups.map((g) => g.label)).toEqual(['Rano', 'Bieg', 'Po biegu', 'Wieczór'])
    expect(groups.map((g) => g.items.length)).toEqual([2, 2, 2, 1])
    expect(slotGroups(list(), 'strength').map((g) => g.label)).toEqual(['Rano', 'Trening', 'Po treningu', 'Wieczór'])
    expect(slotGroups([], null)).toEqual([])
  })

  it('labels rows like the mockup', () => {
    const [d3, , caffeine, electrolytes, , whey, magnesium] = list() as SupplementItem[]
    expect(sp(itemSub(d3!))).toBe('2 000 IU · z posiłkiem')
    expect(itemSub(electrolytes!)).toBe('1 saszetka · bieg > 90 min')
    expect(itemSub(whey!)).toBe('30 g · liczy się w Fitatu')
    expect(doseLabel({ amount: 2, unit: 'capsule' })).toBe('2 kaps.')
    expect(whenLabel(d3!)).toBe('przyjęte')
    expect(whenLabel(caffeine!)).toBe('przed')
    expect(whenLabel(electrolytes!)).toBe('w trakcie')
    expect(whenLabel(magnesium!)).toBe('20:00')
    expect(whenLabel({ ...caffeine!, status: 'skipped' })).toBe('pominięte')
  })

  it('toggles taken ↔ skipped and updates optimistically', () => {
    const items = list()
    expect(counts(items)).toEqual({ taken: 2, total: 7 })
    expect(nextStatus(items[0]!)).toBe('skipped')
    expect(nextStatus(items[2]!)).toBe('taken')
    const next = withStatus(items, 'caffeine', 'taken')
    expect(counts(next).taken).toBe(3)
    expect(items[2]!.status).toBe('pending')
    const replaced = replaceItem(next, { ...items[2]!, status: 'skipped' })
    expect(replaced[2]!.status).toBe('skipped')
  })

  it('formats regularity and stock', () => {
    const reg = supplementsFull().regularity.items
    for (const r of reg) expect(r.days).toHaveLength(30)
    expect(reg.map(regularityDays)).toEqual([
      '4 / 4 długie',
      '29 / 30',
      '28 / 30',
      '22 / 26 tren.',
      '25 / 30',
      '6 / 8 jakość.',
      '21 / 30',
    ])
    expect(reg.map((r) => r.pct)).toEqual([100, 97, 93, 85, 83, 75, 70])
    const stock = supplementsFull().stock
    expect(stock.map(stockLeft)).toEqual(['12 kaps.', '6 sasz.', '20 tabl.', '54 kaps.', '205 g'])
    expect(stock.map(stockLasts)).toEqual(['~12 dni · zamów', '~6 długich', '~20 dni', '~27 dni', '~41 dni'])
    expect(supplementsMeta('2026-10-25')).toBe('Suplementy · Nd 25.10')
  })

  it('has an empty stage1', () => {
    expect(supplementsStage1.checklist).toEqual([])
    expect(supplementsStage1.proposal).toBeNull()
  })
})
