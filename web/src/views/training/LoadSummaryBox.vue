<script setup lang="ts">
import { computed } from 'vue'
import type { LoadSummary } from '@/api/types'
import BoxSection from '@/components/ui/BoxSection.vue'
import MetricCell from '@/components/ui/MetricCell.vue'
import { dateRange, num, signed } from '@/lib/format'
import { FORM_LABEL, FORM_TONE } from '@/lib/labels'
import { atlSub, ctlSub, kmSub } from './trainingData'

const props = defineProps<{ load: LoadSummary; weeks: number; wide?: boolean }>()

const meta = computed(() => `Garmin · tydzień ${dateRange(props.load.weekFrom, props.load.weekTo)}`)
</script>

<template>
  <BoxSection title="Obciążenie" :meta="meta" class="load-box">
    <div class="metric-grid" :class="{ four: wide }">
      <MetricCell label="Kilometraż" :value="num(load.weekKm)" unit="km" :sub="kmSub(load)" />
      <MetricCell v-if="load.ctl !== null" label="Fitness · CTL" :value="num(load.ctl)" :sub="ctlSub(load, weeks)" />
      <MetricCell v-if="load.atl !== null" label="Zmęczenie · ATL" :value="num(load.atl)" :sub="atlSub(load)" />
      <MetricCell
        v-if="load.tsb !== null"
        label="Forma · TSB"
        :value="signed(load.tsb)"
        :sub="load.form ? FORM_LABEL[load.form] : null"
        :sub-tone="load.form ? FORM_TONE[load.form] : undefined"
      />
    </div>
  </BoxSection>
</template>

<style scoped>
.load-box { gap: 0; padding-bottom: 4px; }
</style>
