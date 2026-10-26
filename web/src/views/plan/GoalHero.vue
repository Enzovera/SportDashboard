<script setup lang="ts">
import type { RaceGoal } from '@/api/plan'
import BoxSection from '@/components/ui/BoxSection.vue'
import { pace } from '@/lib/format'
import { fullDate, raceTime } from './planData'

defineProps<{ goal: RaceGoal; wide?: boolean }>()
</script>

<template>
  <BoxSection
    hero
    :title="`${goal.race} · ${fullDate(goal.date)}`"
    :meta="`za ${goal.weeksLeft} tyg.`"
    label="Cele"
    class="goal"
    :class="{ wide }"
  >
    <div v-if="goal.targetTimeS !== null" class="big">
      <span class="time">{{ raceTime(goal.targetTimeS) }}</span>
      <span v-if="goal.targetPaceSPerKm !== null" class="pace">cel<br /><strong>{{ pace(goal.targetPaceSPerKm) }} /km</strong></span>
    </div>
    <div v-for="s in goal.secondary" :key="s.label" class="row">
      <span class="k">{{ s.label }}</span>
      <span><span class="v">{{ s.value }}</span><span v-if="s.detail" class="k"> · {{ s.detail }}</span></span>
    </div>
  </BoxSection>
</template>

<style scoped>
.big { display: flex; align-items: flex-end; justify-content: space-between; gap: 10px; }
.time { font-size: 64px; font-weight: 800; font-stretch: 70%; line-height: 0.85; letter-spacing: -0.01em; }
.pace { text-align: right; font-size: 13px; line-height: 1.45; color: var(--hero-ink2); }
.pace strong { color: var(--hero-ink); font-weight: 700; }
.row { display: flex; justify-content: space-between; gap: 8px; border-top: 1px solid var(--hero-rule); padding-top: 8px; font-size: 13.5px; }
.k { color: var(--hero-ink2); }
.v { font-weight: 700; }
.wide .big { gap: 16px; padding: 4px 0 6px; }
.wide .time { font-size: 96px; line-height: 0.82; }
.wide .pace { font-size: 14px; line-height: 1.5; }
.wide .row { gap: 12px; padding-top: 10px; font-size: 14.5px; }
</style>
