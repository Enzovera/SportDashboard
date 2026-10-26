<script setup lang="ts">
import { computed } from 'vue'
import type { NutritionBalance } from '@/api/nutrition'
import BoxSection from '@/components/ui/BoxSection.vue'
import MetricCell from '@/components/ui/MetricCell.vue'
import { dateRange } from '@/lib/format'
import { balanceCells } from './nutritionData'

const props = defineProps<{ balance: NutritionBalance; wide?: boolean }>()
const cells = computed(() => balanceCells(props.balance))
</script>

<template>
  <BoxSection title="Bilans" :meta="dateRange(balance.from, balance.to)" class="balance">
    <div class="metric-grid" :class="{ four: wide }">
      <MetricCell
        v-for="c in cells"
        :key="c.label"
        :label="c.label"
        :value="c.value"
        :unit="c.unit"
        :sub="c.sub"
        :sub-tone="c.subTone"
      />
    </div>
  </BoxSection>
</template>

<style scoped>
.balance { gap: 0; padding-bottom: 4px; }
</style>
