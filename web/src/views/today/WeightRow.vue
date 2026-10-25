<script setup lang="ts">
import type { TodayResponse } from '@/api/types'
import { num, signed } from '@/lib/format'

defineProps<{ weight: NonNullable<TodayResponse['weight']> }>()
</script>

<template>
  <RouterLink to="/cialo" class="link-row weight">
    <span class="label">Waga <span class="muted">· śr. 7 dni</span></span>
    <span>
      <span class="value">{{ num(weight.avg7dKg, 1) }} kg</span>
      <span v-if="weight.deltaPerWeekKg !== null" class="delta"> {{ signed(weight.deltaPerWeekKg, 1) }} / tydz.</span>
    </span>
  </RouterLink>
</template>

<style scoped>
.weight { justify-content: space-between; }
.label { font-size: 13.5px; }
.muted { color: var(--ink3); }
.value { font-size: 20px; font-weight: 750; font-stretch: 78%; white-space: nowrap; }
.delta { font-size: 13px; color: var(--ink3); white-space: nowrap; margin-left: 4px; }
</style>
