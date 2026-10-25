<script setup lang="ts">
import { computed } from 'vue'
import type { LoadPoint } from '@/api/types'
import BoxSection from '@/components/ui/BoxSection.vue'
import EChart from './EChart.vue'
import ChartLegend, { type LegendItem } from './ChartLegend.vue'
import { loadOption } from './chartOptions'
import { useChartPalette } from './palette'
import { loadAriaLabel } from './trainingData'

const props = defineProps<{ points: LoadPoint[] }>()
const palette = useChartPalette()

const option = computed(() => loadOption(props.points, palette.value))
const label = computed(() => loadAriaLabel(props.points))
const legend: LegendItem[] = [
  { kind: 'line', color: 'var(--chart)', label: 'CTL · 42 dni', weight: 3 },
  { kind: 'dash', color: 'var(--chart-mid)', label: 'ATL · 7 dni' },
]
</script>

<template>
  <BoxSection v-if="points.length" title="Fitness i zmęczenie" class="chart-box">
    <EChart :option="option" :label="label" />
    <ChartLegend :items="legend" />
  </BoxSection>
</template>

<style scoped>
.chart-box { gap: 10px; }
</style>
