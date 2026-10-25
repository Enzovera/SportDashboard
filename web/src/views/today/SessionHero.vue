<script setup lang="ts">
import { computed } from 'vue'
import BoxSection from '@/components/ui/BoxSection.vue'
import type { TodaySession, WeekGrid } from '@/api/types'
import { dateRange, minutesWords, num, paceRange } from '@/lib/format'
import { sessionMeta, weekKm } from './today'

const props = defineProps<{ session: TodaySession; week: WeekGrid; compact?: boolean }>()

const distance = computed(() => {
  const km = props.session.distanceKm
  if (km === null) return null
  return num(km, Number.isInteger(km) ? 0 : 1)
})
const hasDetails = computed(
  () => props.session.zone !== null || props.session.paceSPerKm !== null || props.session.durationMin !== null,
)
</script>

<template>
  <BoxSection hero :title="session.title" label="Sesja na dziś" :meta="sessionMeta(session.action, !!compact)">
    <div v-if="distance !== null || hasDetails" class="main">
      <template v-if="distance !== null">
        <span class="distance">{{ distance }}</span>
        <span class="unit">km</span>
      </template>
      <span v-if="hasDetails" class="details">
        <span v-if="session.zone" class="line">{{ session.zone }}</span>
        <span v-if="session.paceSPerKm" class="line pace">{{ paceRange(session.paceSPerKm) }} /km</span>
        <span v-if="session.durationMin !== null" class="line">ok. {{ minutesWords(session.durationMin) }}</span>
      </span>
    </div>
    <div class="foot">
      <span>Tydzień {{ dateRange(week.from, week.to) }}</span>
      <span class="foot-km">{{ weekKm(week) }}</span>
    </div>
  </BoxSection>
</template>

<style scoped>
.main { display: flex; align-items: flex-end; gap: 10px; }
.distance {
  font-size: 80px; font-weight: 800; font-stretch: 70%; line-height: 0.82; letter-spacing: -0.02em;
  color: var(--hero-signal);
}
.unit { font-size: 22px; font-weight: 700; font-stretch: 80%; line-height: 1; }
@media (min-width: 1024px) {
  .distance { font-size: 96px; }
  .unit { font-size: 26px; }
}
.details {
  margin-left: auto; display: flex; flex-direction: column; align-items: flex-end;
  font-size: 13px; line-height: 1.45; color: var(--hero-ink2); text-align: right;
}
.pace { color: var(--hero-ink); font-weight: 600; }
.foot {
  display: flex; justify-content: space-between; gap: 8px;
  border-top: 1px solid var(--hero-rule); padding-top: 8px; font-size: 13px; color: var(--hero-ink2);
}
.foot-km { color: var(--hero-ink); font-weight: 700; }
</style>
