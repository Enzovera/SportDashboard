<script setup lang="ts">
import BoxSection from '@/components/ui/BoxSection.vue'
import type { CoachSummary } from '@/api/types'
import { time } from '@/lib/format'

defineProps<{ coach: CoachSummary; compact?: boolean }>()
</script>

<template>
  <BoxSection
    title="Trener"
    label="Podsumowanie trenera"
    :meta="compact ? time(coach.at) : `run dzienny · ${time(coach.at)}`"
  >
    <p class="text">{{ coach.text }}</p>
    <div v-if="coach.flags.length" class="flags">
      <span
        v-for="(f, i) in coach.flags"
        :key="i"
        :class="f.tone === 'neutral' ? 'flag-neutral' : `tone-${f.tone}`"
      >{{ f.text }}</span>
    </div>
  </BoxSection>
</template>

<style scoped>
.text { font-size: 15.5px; line-height: 1.5; }
.flags { display: flex; flex-wrap: wrap; gap: 4px 14px; font-size: 13px; font-weight: 600; }
.flag-neutral { color: var(--ink2); }
</style>
