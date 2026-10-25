<script setup lang="ts">
import { computed } from 'vue'
import type { WeeklyKm } from '@/api/types'
import BoxSection from '@/components/ui/BoxSection.vue'
import EChart from './EChart.vue'
import ChartLegend, { type LegendItem } from './ChartLegend.vue'
import { weeklyKmOption } from './chartOptions'
import { useChartPalette } from './palette'
import { weeklyKmAriaLabel } from './trainingData'

const props = defineProps<{ weeks: WeeklyKm[]; todayPlannedKm: number | null }>()
const palette = useChartPalette()

const option = computed(() => weeklyKmOption(props.weeks, props.todayPlannedKm, palette.value))
const label = computed(() => weeklyKmAriaLabel(props.weeks, props.todayPlannedKm))
const legend = computed(() => {
  const items: LegendItem[] = [{ kind: 'square', color: 'var(--chart-mid)', label: 'wykonanie' }]
  if (props.weeks.some((w) => w.deload)) items.push({ kind: 'square', color: 'var(--chart-soft)', label: 'odciążenie' })
  if (props.weeks.some((w) => w.plannedKm !== null)) items.push({ kind: 'line', color: 'var(--chart)', label: 'plan' })
  if (props.todayPlannedKm !== null) items.push({ kind: 'outline', color: 'var(--signal)', label: 'dziś' })
  return items
})
</script>

<template>
  <BoxSection v-if="weeks.length" title="Kilometraż tygodniowy" class="chart-box">
    <EChart :option="option" :label="label" />
    <ChartLegend :items="legend" />
  </BoxSection>
</template>

<style scoped>
.chart-box { gap: 10px; }
</style>
