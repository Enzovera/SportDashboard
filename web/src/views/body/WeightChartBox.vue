<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import type { BodyWeek } from '@/api/body'
import BoxSection from '@/components/ui/BoxSection.vue'
import { DEFAULT_PAD, trendTitle, weightChart, weightChartLabel } from './bodyData'

// Weekly weight: average (squares + line), min–max of the week's weigh-ins (whiskers), current week in red.
// Plain SVG drawn at the box's real pixel width, so text stays the same size on every screen.
const props = withDefaults(
  defineProps<{ weeks: BodyWeek[]; today: string; trendKgPerWeek: number | null; height?: number; wide?: boolean }>(),
  { height: 196 },
)

const el = ref<HTMLDivElement | null>(null)
const width = ref(0)
let observer: ResizeObserver | null = null
onMounted(() => {
  if (!el.value) return
  width.value = el.value.clientWidth
  observer = new ResizeObserver(([entry]) => {
    if (entry) width.value = Math.round(entry.contentRect.width)
  })
  observer.observe(el.value)
})
onBeforeUnmount(() => observer?.disconnect())

const pad = computed(() => (props.wide ? { ...DEFAULT_PAD, left: 36, right: 52, bottom: 30 } : DEFAULT_PAD))
const chart = computed(() =>
  width.value > 0 ? weightChart(props.weeks, props.today, width.value, props.height, pad.value, props.wide ? 60 : 52) : null,
)
const title = computed(() => trendTitle(props.trendKgPerWeek))
const label = computed(() => weightChartLabel(props.weeks))
const hasRange = computed(() => props.weeks.some((w) => w.minKg !== null && w.maxKg !== null && w.maxKg > w.minKg))
const sq = computed(() => (props.wide ? 8 : 7))
const sqNow = computed(() => (props.wide ? 12 : 10))
</script>

<template>
  <BoxSection :title="title" :meta="wide ? 'kg · średnia tygodnia' : null" label="Waga tydzień po tygodniu" class="chart-box">
    <div ref="el" class="wrap" :style="{ height: `${height}px` }">
      <svg
        v-if="chart"
        class="chart"
        :class="{ wide }"
        :viewBox="`0 0 ${chart.width} ${chart.height}`"
        :width="chart.width"
        :height="chart.height"
        role="img"
        :aria-label="label"
      >
        <g class="grid">
          <line v-for="t in chart.yTicks" :key="t.label" :x1="chart.plotLeft" :x2="chart.width - 4" :y1="t.y" :y2="t.y" />
        </g>
        <line class="axis" :x1="chart.plotLeft" :x2="chart.width - 4" :y1="chart.baseline" :y2="chart.baseline" />
        <g text-anchor="end">
          <text v-for="t in chart.yTicks" :key="t.label" :x="chart.plotLeft - 8" :y="t.y + 4">{{ t.label }}</text>
        </g>
        <g class="range">
          <path
            v-for="w in chart.whiskers"
            :key="w.x"
            :d="`M${w.x} ${w.yMax}V${w.yMin}M${w.x - 4} ${w.yMax}h8M${w.x - 4} ${w.yMin}h8`"
          />
        </g>
        <polyline v-for="s in chart.segments" :key="s" class="line-main" :points="s" />
        <template v-for="p in chart.points" :key="p.x">
          <rect
            v-if="p.current"
            class="pt-now"
            :x="p.x - sqNow / 2"
            :y="p.y - sqNow / 2"
            :width="sqNow"
            :height="sqNow"
          />
          <rect v-else class="pt-main" :x="p.x - sq / 2" :y="p.y - sq / 2" :width="sq" :height="sq" />
        </template>
        <text
          v-if="chart.last"
          class="val"
          :class="{ now: chart.last.current }"
          :x="chart.last.x + 10"
          :y="chart.last.y + 5"
        >{{ chart.last.label }}</text>
        <g>
          <text
            v-for="l in chart.xLabels"
            :key="l.label"
            :x="l.x"
            :y="chart.height - 6"
            :text-anchor="l.anchor"
          >{{ l.label }}</text>
        </g>
      </svg>
    </div>
    <div class="legend" aria-hidden="true">
      <span><span class="sw sw-pt" />średnia tygodnia</span>
      <span v-if="hasRange"><span class="sw sw-range" />min–max ważeń w tygodniu</span>
    </div>
  </BoxSection>
</template>

<style scoped>
.chart-box { gap: 10px; }
.wrap { width: 100%; }
.chart { display: block; overflow: visible; }
.chart text { font-family: var(--font); font-size: 10.5px; fill: var(--ink3); font-variant-numeric: tabular-nums; }
.chart.wide text { font-size: 12px; }
.grid line { stroke: var(--grid); stroke-width: 1; }
.axis { stroke: var(--axis); stroke-width: 1; }
.range path { stroke: var(--chart-mid); stroke-width: 1.25; fill: none; }
.chart.wide .range path { stroke-width: 1.5; }
.line-main { fill: none; stroke: var(--chart); stroke-width: 2.25; stroke-linejoin: round; }
.chart.wide .line-main { stroke-width: 2.5; }
.pt-main { fill: var(--chart); }
.pt-now { fill: var(--signal); }
.chart .val { fill: var(--ink); font-size: 12px; font-weight: 800; }
.chart .val.now { fill: var(--signal); }
.chart.wide .val { font-size: 15px; }
.legend { display: flex; flex-wrap: wrap; gap: 6px 14px; font-size: 12.5px; color: var(--ink2); }
.legend > span { display: inline-flex; align-items: center; gap: 6px; }
.sw { display: inline-block; flex: none; }
.sw-pt { width: 8px; height: 8px; background: var(--chart); }
.sw-range { width: 2px; height: 12px; background: var(--chart-mid); }
@media (min-width: 1024px) { .legend { font-size: 13px; } }
</style>
