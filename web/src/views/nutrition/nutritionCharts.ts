// ECharts options for the Nutrition view. Pure: colors come in as a palette read from CSS variables elsewhere.
import type { NutritionBalance, NutritionDay } from '@/api/nutrition'
import { dayMonth, num, signed } from '@/lib/format'
import { labelIndices, niceRange, type ChartOption, type ChartPalette } from '@/views/training/chartOptions'
import { DAY_TYPE_SHORT, isKeyDay } from './nutritionData'

export interface NutritionPalette extends ChartPalette {
  ink: string
  surface: string
  good: string
  warn: string
}

interface Rect {
  x: number
  y: number
  width: number
  height: number
}

function slotWidth(size: number | number[] | undefined): number {
  return Array.isArray(size) ? (size[0] ?? 0) : (size ?? 0)
}

const label = (p: NutritionPalette, size: number) => ({ color: p.ink3, fontSize: size, fontFamily: p.font })

function valueAxis(p: NutritionPalette, range: { min: number; max: number; interval: number }, size: number, fmt: (v: number) => string) {
  return {
    type: 'value' as const,
    ...range,
    axisLine: { show: false },
    axisTick: { show: false },
    axisLabel: { ...label(p, size), margin: 6, formatter: fmt },
    splitLine: { lineStyle: { color: p.grid, width: 1 } },
  }
}

/** Day number + day type letter; today red, key days (deficit on a long run) bold ink. */
function dayAxis(days: NutritionDay[], p: NutritionPalette, size: number, withType: boolean) {
  const style = { fontSize: size, fontFamily: p.font, lineHeight: size + 5 }
  return {
    type: 'category' as const,
    data: days.map((d) => d.date),
    boundaryGap: true,
    axisLine: { lineStyle: { color: p.axis, width: 1 } },
    axisTick: { show: false },
    axisLabel: {
      margin: 8,
      interval: 0,
      formatter: (value: string, i: number) => {
        const d = days[i]
        const key = !d ? 'n' : !d.complete ? 's' : isKeyDay(d) ? 'k' : 'n'
        const day = String(Number(value.slice(8, 10)))
        const type = withType && d?.dayType ? `\n{${key}|${DAY_TYPE_SHORT[d.dayType]}}` : ''
        return `{${key}|${day}}${type}`
      },
      rich: {
        n: { ...style, color: p.ink3 },
        k: { ...style, color: p.chart, fontWeight: 800 },
        s: { ...style, color: p.signal, fontWeight: 800 },
      },
    },
  }
}

/**
 * Calories day by day: eaten bars (key days in ink), the day's target as a short ink tick,
 * today's intake so far as a dashed red outline, the deficit of key days above them in amber.
 */
export function kcalOption(days: NutritionDay[], p: NutritionPalette, wide = false): ChartOption {
  const size = wide ? 12 : 10.5
  const max = Math.max(0, ...days.flatMap((d) => [d.kcal ?? 0, d.kcalTarget ?? 0]))
  const range = niceRange(0, max)

  const series: ChartOption['series'] = [
    {
      type: 'bar',
      silent: true,
      barWidth: wide ? '46%' : '54%',
      emphasis: { disabled: true },
      data: days.map((d) =>
        d.complete && d.kcal !== null
          ? { value: d.kcal, itemStyle: { color: isKeyDay(d) ? p.chart : p.chartMid } }
          : '-',
      ),
    },
    {
      type: 'custom',
      silent: true,
      clip: false,
      z: 3,
      data: days.flatMap((d, i) => (d.kcalTarget === null ? [] : [[i, d.kcalTarget]])),
      renderItem: (_params, api) => {
        const [x = 0, y = 0] = api.coord([api.value(0), api.value(1)])
        const half = slotWidth(api.size?.([1, 0])) * 0.36
        return {
          type: 'line',
          shape: { x1: x - half, y1: y, x2: x + half, y2: y },
          style: { stroke: p.chart, lineWidth: wide ? 2.5 : 2 },
        }
      },
    },
  ]

  const todayIdx = days.findIndex((d) => !d.complete)
  const today = days[todayIdx]
  if (today && today.kcal !== null && today.kcal > 0) {
    series.push({
      type: 'custom',
      silent: true,
      clip: false,
      z: 4,
      data: [[todayIdx, 0, today.kcal]],
      renderItem: (_params, api) => {
        const [x = 0, bottom = 0] = api.coord([api.value(0), api.value(1)])
        const [, top = 0] = api.coord([api.value(0), api.value(2)])
        const w = slotWidth(api.size?.([1, 0])) * (wide ? 0.44 : 0.48)
        return {
          type: 'rect',
          shape: { x: x - w / 2, y: top, width: w, height: bottom - top - 0.5 },
          style: { fill: 'none', stroke: p.signal, lineWidth: 1.5, lineDash: [3, 2] },
        }
      },
    })
  }

  const deficits = days.flatMap((d, i) =>
    isKeyDay(d) && d.kcal !== null && d.kcalTarget !== null ? [[i, Math.max(d.kcal, d.kcalTarget), d.kcal - d.kcalTarget]] : [],
  )
  if (deficits.length) {
    series.push({
      type: 'custom',
      silent: true,
      clip: false,
      z: 5,
      data: deficits,
      renderItem: (_params, api) => {
        const [x = 0, y = 0] = api.coord([api.value(0), api.value(1)])
        return {
          type: 'text',
          style: {
            text: signed(Number(api.value(2))),
            x,
            y: y - 6,
            align: 'center',
            verticalAlign: 'bottom',
            fill: p.warn,
            font: `700 ${wide ? 12.5 : 11}px ${p.font}`,
          },
        }
      },
    })
  }

  return {
    animation: false,
    textStyle: { fontFamily: p.font },
    grid: { left: wide ? 40 : 30, right: 4, top: 22, bottom: wide ? 44 : 40, containLabel: false },
    xAxis: dayAxis(days, p, size, true),
    yAxis: valueAxis(p, range, size, (v) => String(v)),
    series,
  }
}

/** Axis range in steps of 0,2 g/kg covering the values, the target and the minimum. */
export function proteinRange(values: number[], target: number, min: number): { min: number; max: number; interval: number } {
  const lo = Math.min(...values, min)
  const hi = Math.max(...values, target)
  const loI = Math.floor((lo - 0.05) / 0.2 + 1e-9)
  let hiI = Math.ceil(hi / 0.2 - 1e-9)
  if (hiI <= loI) hiI = loI + 1
  return { min: Number((loI * 0.2).toFixed(1)), max: Number((hiI * 0.2).toFixed(1)), interval: 0.2 }
}

/**
 * Protein g/kg per day: ink line with square points, days below the minimum as larger hollow amber squares,
 * the target as a solid green line and the minimum as a dashed amber line, both labelled.
 * wide: shares the X axis with the calorie chart (14 days incl. today); otherwise complete days only.
 */
export function proteinOption(days: NutritionDay[], b: NutritionBalance, p: NutritionPalette, wide = false): ChartOption {
  const size = wide ? 12 : 10.5
  const shown = wide ? days : days.filter((d) => d.complete)
  const values = shown.flatMap((d) => (d.proteinGPerKg === null ? [] : [d.proteinGPerKg]))
  const range = proteinRange(values.length ? values : [b.proteinTargetGPerKg], b.proteinTargetGPerKg, b.proteinMinGPerKg)
  const labelFont = `700 ${wide ? 12.5 : 11.5}px ${p.font}`

  const xAxis = wide
    ? dayAxis(shown, p, size, false)
    : (() => {
        const ids = labelIndices(shown.length, 3)
        return {
          type: 'category' as const,
          data: shown.map((d) => dayMonth(d.date)),
          boundaryGap: false,
          axisLine: { lineStyle: { color: p.axis, width: 1 } },
          axisTick: { show: false },
          axisLabel: {
            ...label(p, size),
            margin: 8,
            interval: (i: number) => ids.has(i),
            alignMinLabel: 'left' as const,
            alignMaxLabel: 'right' as const,
          },
        }
      })()

  return {
    animation: false,
    textStyle: { fontFamily: p.font },
    grid: { left: wide ? 40 : 30, right: wide ? 4 : 8, top: 14, bottom: wide ? 26 : 24, containLabel: false },
    xAxis,
    yAxis: valueAxis(p, range, size, (v) => num(v, 1)),
    series: [
      {
        type: 'custom',
        silent: true,
        clip: false,
        z: 1,
        data: [[0]],
        renderItem: (params, api) => {
          const r = params.coordSys as unknown as Rect
          const [, yGoal = 0] = api.coord([0, b.proteinTargetGPerKg])
          const [, yMin = 0] = api.coord([0, b.proteinMinGPerKg])
          return {
            type: 'group',
            children: [
              { type: 'line', shape: { x1: r.x, y1: yGoal, x2: r.x + r.width, y2: yGoal }, style: { stroke: p.good, lineWidth: 1.5 } },
              {
                type: 'line',
                shape: { x1: r.x, y1: yMin, x2: r.x + r.width, y2: yMin },
                style: { stroke: p.warn, lineWidth: 1.5, lineDash: [5, 4] },
              },
              {
                type: 'text',
                style: {
                  text: `cel ${num(b.proteinTargetGPerKg, 1)}`,
                  x: r.x + r.width,
                  y: yGoal - 4,
                  align: 'right',
                  verticalAlign: 'bottom',
                  fill: p.good,
                  font: labelFont,
                },
              },
              {
                type: 'text',
                style: {
                  text: `minimum ${num(b.proteinMinGPerKg, 1)}`,
                  x: r.x + 8,
                  y: yMin + 5,
                  align: 'left',
                  verticalAlign: 'top',
                  fill: p.warn,
                  font: labelFont,
                },
              },
            ],
          }
        },
      },
      {
        type: 'line',
        silent: true,
        z: 3,
        symbol: 'rect',
        showAllSymbol: true,
        emphasis: { disabled: true },
        lineStyle: { color: p.chart, width: 2, join: 'round' },
        data: shown.map((d) => {
          if (d.proteinGPerKg === null) return '-'
          return d.proteinGPerKg < b.proteinMinGPerKg
            ? {
                value: d.proteinGPerKg,
                symbolSize: 9,
                itemStyle: { color: p.surface, borderColor: p.warn, borderWidth: 2.25 },
              }
            : { value: d.proteinGPerKg, symbolSize: 7, itemStyle: { color: p.chart, borderWidth: 0 } }
        }),
      },
    ],
  }
}
