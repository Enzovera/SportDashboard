import { describe, expect, it } from 'vitest'
import { dateRange, dayMonth, dayShort, hm, minutesWords, num, pace, paceRange, signed, time } from './format'
import { syncProblem } from './labels'

const NB = ' '

describe('num', () => {
  it.each([
    [2950, 0, `2${NB}950`],
    [11950, 0, `11${NB}950`],
    [820, 0, '820'],
    [8, 1, '8,0'],
    [76.4, 1, '76,4'],
    [-0.3, 1, '−0,3'],
    [-0.04, 1, '0,0'],
    [1234567, 0, `1${NB}234${NB}567`],
  ])('num(%s, %s) = %s', (v, d, want) => expect(num(v, d)).toBe(want))
})

describe('signed', () => {
  it.each([
    [4, '+4'],
    [-5, '−5'],
    [0, '0'],
  ])('signed(%s) = %s', (v, want) => expect(signed(v)).toBe(want))
})

describe('time and pace', () => {
  it.each([
    [298, '4:58'],
    [362, '6:02'],
    [359.6, '6:00'],
  ])('pace(%s) = %s', (v, want) => expect(pace(v)).toBe(want))
  it('paceRange', () => expect(paceRange([350, 370])).toBe('5:50–6:10'))
  it('hm', () => expect(hm(27600)).toBe('7:40'))
  it.each([
    [96, '1 h 36 min'],
    [45, '45 min'],
    [120, '2 h'],
  ])('minutesWords(%s) = %s', (v, want) => expect(minutesWords(v)).toBe(want))
})

describe('dates', () => {
  it('dayShort', () => expect(dayShort('2026-10-25')).toBe('Nd 25.10'))
  it('dayMonth', () => expect(dayMonth('2026-08-03')).toBe('03.08'))
  it.each([
    ['2026-10-19', '2026-10-25', '19–25.10'],
    ['2026-09-28', '2026-10-04', '28.09–04.10'],
  ])('dateRange(%s, %s) = %s', (a, b, want) => expect(dateRange(a, b)).toBe(want))
  it('time is Warsaw local', () => expect(time('2026-10-25T05:12:00Z')).toBe('06:12'))
})

describe('syncProblem', () => {
  const now = new Date('2026-10-25T09:00:00Z')
  it.each([
    [{ status: 'ok', lastSuccessAt: null }, null],
    [{ status: 'error', lastSuccessAt: null }, 'błąd sync'],
    [{ status: 'stale', lastSuccessAt: '2026-10-24T16:02:00Z' }, 'sync: wcz. 18:02'],
    [{ status: 'stale', lastSuccessAt: '2026-10-25T04:00:00Z' }, 'sync: 05:00'],
    [{ status: 'stale', lastSuccessAt: '2026-10-20T10:00:00Z' }, 'sync: 20.10'],
  ] as const)('%o -> %s', (s, want) => expect(syncProblem({ source: 'hevy', ...s }, now)).toBe(want))
})
