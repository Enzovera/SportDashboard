<script setup lang="ts">
import { computed } from 'vue'
import type { NutritionDay } from '@/api/nutrition'
import BoxSection from '@/components/ui/BoxSection.vue'
import EChart from '@/views/training/EChart.vue'
import ChartLegend from '@/views/training/ChartLegend.vue'
import { kcalOption } from './nutritionCharts'
import { DAY_TYPE_LEGEND, kcalAriaLabel, kcalHeadline } from './nutritionData'
import { useNutritionPalette } from './palette'

const props = defineProps<{ days: NutritionDay[]; wide?: boolean; height?: number }>()
const palette = useNutritionPalette()

const option = computed(() => kcalOption(props.days, palette.value, props.wide))
const legend = computed(() => {
  const items = [
    { kind: 'square' as const, color: 'var(--chart-mid)', label: 'zjedzone' },
    { kind: 'line' as const, color: 'var(--chart)', label: 'cel dnia' },
  ]
  return props.days.some((d) => !d.complete && d.kcal) ? [...items, { kind: 'outline' as const, color: 'var(--signal)', label: 'dziś' }] : items
})
</script>

<template>
  <BoxSection
    :title="kcalHeadline(days)"
    :meta="wide ? `kcal · ${days.length} dni` : null"
    label="Kalorie dzień po dniu"
    class="chart-box"
  >
    <EChart :option="option" :label="kcalAriaLabel(days)" :height="height ?? 212" />
    <div class="legend-row">
      <ChartLegend :items="legend" />
      <span class="codes">{{ DAY_TYPE_LEGEND }}</span>
    </div>
  </BoxSection>
</template>

<style scoped>
.chart-box { gap: 10px; }
.legend-row { display: flex; flex-wrap: wrap; gap: 6px 14px; }
.codes { font-size: 12.5px; color: var(--ink3); }
</style>
