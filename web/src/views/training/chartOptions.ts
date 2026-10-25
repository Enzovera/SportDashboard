// ECharts options for the Training view. Pure: colors come in as a palette (read from CSS variables elsewhere).
import type { ComposeOption } from 'echarts/core'
import type { BarSeriesOption, CustomSeriesOption, LineSeriesOption } from 'echarts/charts'
import type { GridComponentOption } from 'echarts/components'
import type { LoadPoint, WeeklyKm } from '@/api/types'
import { dayMonth, num } from '@/lib/format'

export type ChartOption = ComposeOption<BarSeriesOption | LineSeriesOption | CustomSeriesOption | GridComponentOption>

/** Chart colors and font, resolved from design tokens (--chart, --chart-mid, ...). */
export interface ChartPalette {
  chart: string
  chartMid: string
  chartSoft: string
  grid: string
  axis: string
  ink2: string
  ink3: string
  signal: string
  font: string
}

const STEPS = [10, 20, 25, 50, 100, 200, 250, 500, 1000]
const MAX_INTERVALS = 5

/** Nice axis range covering [min, max] with at most 5 intervals, preferring 10-unit steps. */
export function niceRange(min: number, max: number): { min: number; max: number; interval: number } {
  for (const step of STEPS) {
    const lo = Math.floor(min / step) * step
    let hi = Math.ceil(max / step) * step
    if (hi === lo) hi = lo + step
    if ((hi - lo) / step <= MAX_INTERVALS) return { min: lo, max: hi, interval: step }
  }
  const step = STEPS[STEPS.length - 1]!
  return { min: Math.floor(min / step) * step, max: Math.ceil(max / step) * step || step, interval: step }
}

/**
 * About `target` evenly spaced category labels, always including the first and the last;
 * a regular label too close to the last one is dropped.
 */
export function labelIndices(n: number, target = 4): Set<number> {
  const out = new Set<number>()
  if (n <= 0) return out
  const step = Math.max(1, Math.ceil((n - 1) / (target - 1)))
  for (let i = 0; i < n - 1; i += step) out.add(i)
  const lastPicked = out.size ? Math.max(...out) : 0
  if (n - 1 - lastPicked < step / 2 && lastPicked !== 0) out.delete(lastPicked)
  out.add(n - 1)
  return out
}

function axisLabel(p: ChartPalette) {
  return { color: p.ink3, fontSize: 10.5, fontFamily: p.font }
}

function valueAxis(p: ChartPalette, range: { min: number; max: number; interval: number }) {
  return {
    type: 'value' as const,
    ...range,
    axisLine: { show: false },
    axisTick: { show: false },
    axisLabel: { ...axisLabel(p), margin: 6 },
    splitLine: { lineStyle: { color: p.grid, width: 1 } },
  }
}

function categoryAxis(p: ChartPalette, labels: string[], shown: Set<number>, boundaryGap: boolean) {
  return {
    type: 'category' as const,
    data: labels,
    boundaryGap,
    axisLine: { lineStyle: { color: p.axis, width: 1 } },
    axisTick: { show: false },
    axisLabel: { ...axisLabel(p), margin: 8, interval: (i: number) => shown.has(i) },
  }
}

const base = (p: ChartPalette) => ({
  animation: false,
  textStyle: { fontFamily: p.font },
  grid: { left: 26, right: 6, top: 22, bottom: 24, containLabel: false },
})

/** Max of the km axis: done km, planned km and today's planned km stacked on the last week. */
export function weeklyKmMax(weeks: WeeklyKm[], todayPlannedKm: number | null): number {
  const last = weeks[weeks.length - 1]
  const values = weeks.flatMap((w) => [w.km, w.plannedKm ?? 0])
  if (last && todayPlannedKm !== null) values.push(last.km + todayPlannedKm)
  return Math.max(0, ...values)
}

/**
 * Weekly km: bars (deload weeks lighter), plan as short ink ticks, today's planned km
 * as a dashed red box stacked on the last week with its value above.
 */
export function weeklyKmOption(weeks: WeeklyKm[], todayPlannedKm: number | null, p: ChartPalette): ChartOption {
  const n = weeks.length
  const series: ChartOption['series'] = [
    {
      type: 'bar',
      silent: true,
      barWidth: '54%',
      emphasis: { disabled: true },
      data: weeks.map((w) => ({ value: w.km, itemStyle: { color: w.deload ? p.chartSoft : p.chartMid } })),
    },
  ]

  const plan = weeks.flatMap((w, i) => (w.plannedKm === null ? [] : [[i, w.plannedKm]]))
  if (plan.length) {
    series.push({
      type: 'custom',
      silent: true,
      clip: false,
      z: 3,
      data: plan,
      renderItem: (_params, api) => {
        const [x = 0, y = 0] = api.coord([api.value(0), api.value(1)])
        const slot = slotWidth(api.size?.([1, 0]))
        const half = slot * 0.4
        return {
          type: 'line',
          shape: { x1: x - half, y1: y, x2: x + half, y2: y },
          style: { stroke: p.chart, lineWidth: 2 },
        }
      },
    })
  }

  const last = weeks[n - 1]
  if (last && todayPlannedKm !== null) {
    const label = num(todayPlannedKm)
    series.push({
      type: 'custom',
      silent: true,
      clip: false,
      z: 4,
      data: [[n - 1, last.km, last.km + todayPlannedKm]],
      renderItem: (_params, api) => {
        const [x = 0, bottom = 0] = api.coord([api.value(0), api.value(1)])
        const [, top = 0] = api.coord([api.value(0), api.value(2)])
        const w = slotWidth(api.size?.([1, 0])) * 0.45
        return {
          type: 'group',
          children: [
            {
              type: 'rect',
              shape: { x: x - w / 2, y: top, width: w, height: bottom - top },
              style: { fill: 'none', stroke: p.signal, lineWidth: 1.5, lineDash: [3, 2] },
            },
            {
              type: 'text',
              style: {
                text: label,
                x,
                y: top - 6,
                align: 'center',
                verticalAlign: 'bottom',
                fill: p.signal,
                font: `800 11.5px ${p.font}`,
              },
            },
          ],
        }
      },
    })
  }

  const range = niceRange(0, weeklyKmMax(weeks, todayPlannedKm))
  return {
    ...base(p),
    xAxis: categoryAxis(p, weeks.map((w) => dayMonth(w.weekStart)), labelIndices(n), true),
    yAxis: valueAxis(p, range),
    series,
  }
}

function slotWidth(size: number | number[] | undefined): number {
  return Array.isArray(size) ? (size[0] ?? 0) : (size ?? 0)
}

/** CTL (solid ink) and ATL (dashed mid) with square end markers and end labels. */
export function loadOption(points: LoadPoint[], p: ChartPalette): ChartOption {
  const n = points.length
  const lastPoint = points[n - 1]
  const ctlOnTop = !lastPoint || lastPoint.ctl >= lastPoint.atl

  const line = (
    key: 'ctl' | 'atl',
    color: string,
    width: number,
    dashed: boolean,
    labelColor: string,
    labelWeight: number,
    onTop: boolean,
  ): LineSeriesOption => ({
    type: 'line',
    silent: true,
    symbol: 'none',
    showAllSymbol: true,
    emphasis: { disabled: true },
    z: key === 'ctl' ? 3 : 2,
    lineStyle: { color, width, type: dashed ? [5, 3] : 'solid', join: 'round' },
    itemStyle: { color },
    data: points.map((pt, i) =>
      i === n - 1
        ? {
            value: pt[key],
            symbol: 'rect',
            symbolSize: 7,
            label: {
              show: true,
              formatter: `${key.toUpperCase()} ${num(pt[key])}`,
              position: onTop ? 'top' : 'bottom',
              distance: 8,
              align: 'right',
              offset: [-6, 0],
              color: labelColor,
              fontSize: 11.5,
              fontWeight: labelWeight,
              fontFamily: p.font,
            },
          }
        : pt[key],
    ),
  })

  const values = points.flatMap((pt) => [pt.ctl, pt.atl])
  const range = values.length ? niceRange(Math.min(...values), Math.max(...values)) : { min: 0, max: 10, interval: 10 }
  const ends = new Set(n ? [0, n - 1] : [])
  const xAxis = categoryAxis(p, points.map((pt) => dayMonth(pt.date)), ends, false)

  return {
    ...base(p),
    grid: { left: 26, right: 6, top: 22, bottom: 24, containLabel: false },
    xAxis: { ...xAxis, axisLabel: { ...xAxis.axisLabel, alignMinLabel: 'left', alignMaxLabel: 'right' } },
    yAxis: valueAxis(p, range),
    series: [
      line('atl', p.chartMid, 1.75, true, p.ink2, 700, !ctlOnTop),
      line('ctl', p.chart, 2.5, false, p.chart, 800, ctlOnTop),
    ],
  }
}
