<script setup lang="ts">
import { computed } from 'vue'
import type { NutritionBalance, NutritionDay } from '@/api/nutrition'
import BoxSection from '@/components/ui/BoxSection.vue'
import EChart from '@/views/training/EChart.vue'
import ChartLegend, { type LegendItem } from '@/views/training/ChartLegend.vue'
import { proteinOption } from './nutritionCharts'
import { proteinAriaLabel, proteinHeadline } from './nutritionData'
import { useNutritionPalette } from './palette'

const props = defineProps<{ days: NutritionDay[]; balance: NutritionBalance; wide?: boolean; height?: number }>()
const palette = useNutritionPalette()

const option = computed(() => proteinOption(props.days, props.balance, palette.value, props.wide))
const complete = computed(() => props.days.filter((d) => d.complete && d.proteinGPerKg !== null))
const legend = computed(() => {
  const items: LegendItem[] = [{ kind: 'square', color: 'var(--chart)', label: 'g/kg dziennie' }]
  if (complete.value.some((d) => (d.proteinGPerKg ?? 0) < props.balance.proteinMinGPerKg))
    items.push({ kind: 'outline', color: 'var(--warn)', label: 'poniżej minimum' })
  return items
})
</script>

<template>
  <BoxSection
    v-if="complete.length"
    :title="proteinHeadline(days, balance)"
    :meta="wide ? `g/kg · ${complete.length} dni` : null"
    label="Białko na kilogram"
    class="chart-box"
  >
    <EChart :option="option" :label="proteinAriaLabel(days, balance)" :height="height ?? 196" />
    <ChartLegend :items="legend" class="legend" />
  </BoxSection>
</template>

<style scoped>
.chart-box { gap: 10px; }
/* "poniżej minimum" is a solid amber outline in the mockups, not the dashed "dziś" outline. */
.legend :deep(.outline) { border-style: solid; border-width: 2px; }
.legend :deep(.square) { width: 7px; height: 7px; }
</style>
