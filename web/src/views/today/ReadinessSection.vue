<script setup lang="ts">
import { computed } from 'vue'
import BoxSection from '@/components/ui/BoxSection.vue'
import MetricCell from '@/components/ui/MetricCell.vue'
import type { Readiness } from '@/api/types'
import { hm, minutesWords, num, signed, time } from '@/lib/format'
import { FORM_LABEL, FORM_TONE } from '@/lib/labels'

const props = defineProps<{ readiness: Readiness; compact?: boolean }>()

const hrvSub = computed(() => {
  const r = props.readiness
  const parts = [r.hrvBaselineMs !== null ? `baza ${num(r.hrvBaselineMs)}` : null, r.hrvLabel].filter(Boolean)
  return parts.length ? parts.join(' · ') : null
})
const sleepSub = computed(() => {
  const avg = props.readiness.sleepAvg7dS
  return avg === null ? null : `śr. ${minutesWords(avg / 60)}`
})
const hrSub = computed(() => {
  const d = props.readiness.restingHrDelta7d
  if (d === null) return null
  return `${signed(d)} vs ${props.compact ? '7 dni' : 'śr. 7 dni'}`
})
</script>

<template>
  <BoxSection
    class="readiness"
    title="Gotowość"
    :meta="readiness.measuredAt ? `Garmin ${time(readiness.measuredAt)}` : null"
  >
    <div class="metric-grid">
      <MetricCell
        label="HRV · 7 dni"
        :value="readiness.hrv7dMs !== null ? num(readiness.hrv7dMs) : null"
        unit="ms"
        :sub="hrvSub"
        :sub-tone="readiness.hrvTone"
      />
      <MetricCell
        label="Sen"
        :value="readiness.sleepS !== null ? hm(readiness.sleepS) : null"
        unit="h"
        :sub="sleepSub"
      />
      <MetricCell
        label="Tętno spocz."
        :value="readiness.restingHr !== null ? num(readiness.restingHr) : null"
        unit="bpm"
        :sub="hrSub"
      />
      <MetricCell
        label="Forma · TSB"
        :value="readiness.tsb !== null ? signed(readiness.tsb) : null"
        :sub="readiness.form ? FORM_LABEL[readiness.form] : null"
        :sub-tone="readiness.form ? FORM_TONE[readiness.form] : undefined"
      />
    </div>
  </BoxSection>
</template>

<style scoped>
.readiness { padding-bottom: 4px; gap: 0; }
</style>
