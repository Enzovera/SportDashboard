<script setup lang="ts">
import { computed } from 'vue'
import type { BodyComposition } from '@/api/body'
import BoxSection from '@/components/ui/BoxSection.vue'
import MetricCell from '@/components/ui/MetricCell.vue'
import { dayShort, localDate, time } from '@/lib/format'
import { compositionCells } from './bodyData'

const props = defineProps<{ composition: BodyComposition; rangeWeeks: number; wide?: boolean }>()

const cells = computed(() => compositionCells(props.composition, props.rangeWeeks))
const meta = computed(() => {
  const at = props.composition.measuredAt
  return at ? `${dayShort(localDate(at))} · ${time(at)}` : null
})
</script>

<template>
  <BoxSection title="Skład ciała" :meta="meta" class="composition">
    <div
      class="metric-grid"
      :class="{ four: wide }"
      :style="wide ? { gridTemplateColumns: `repeat(${Math.max(cells.length, 2)}, minmax(0, 1fr))` } : undefined"
    >
      <MetricCell
        v-for="c in cells"
        :key="c.label"
        :label="c.label"
        :value="c.value"
        :unit="c.unit || undefined"
        :sub="c.sub"
      />
    </div>
  </BoxSection>
</template>

<style scoped>
.composition { gap: 0; padding-bottom: 4px; }
@media (min-width: 1024px) { .composition { gap: 2px; padding-bottom: 8px; } }
</style>
