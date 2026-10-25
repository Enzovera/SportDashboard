<script setup lang="ts">
// Minimal ECharts wrapper: SVG renderer, tree-shaken modules, resizes with its box.
import { onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { init, use, type ECharts } from 'echarts/core'
import { BarChart, CustomChart, LineChart } from 'echarts/charts'
import { GridComponent } from 'echarts/components'
import { SVGRenderer } from 'echarts/renderers'
import type { ChartOption } from './chartOptions'

use([BarChart, LineChart, CustomChart, GridComponent, SVGRenderer])

const props = withDefaults(defineProps<{ option: ChartOption; label: string; height?: number }>(), { height: 192 })

const el = ref<HTMLDivElement | null>(null)
let chart: ECharts | null = null
let observer: ResizeObserver | null = null

onMounted(() => {
  if (!el.value) return
  chart = init(el.value, undefined, { renderer: 'svg' })
  chart.setOption(props.option)
  observer = new ResizeObserver(() => chart?.resize())
  observer.observe(el.value)
})

watch(
  () => props.option,
  (option) => chart?.setOption(option, { notMerge: true }),
)

onBeforeUnmount(() => {
  observer?.disconnect()
  chart?.dispose()
  chart = null
})
</script>

<template>
  <div ref="el" class="chart" role="img" :aria-label="label" :style="{ height: `${height}px` }" />
</template>

<style scoped>
.chart { width: 100%; }
</style>
