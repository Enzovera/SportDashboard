<script setup lang="ts">
import type { Tone } from '@/api/types'

// One cell of a .metric-grid: label, big narrow number with unit, one line below.
defineProps<{ label: string; value: string | null; unit?: string; sub?: string | null; subTone?: Tone }>()
</script>

<template>
  <div>
    <div class="metric-label">{{ label }}</div>
    <div class="metric-value">
      {{ value ?? '–' }}<span v-if="unit && value !== null" class="metric-unit">{{ unit }}</span>
    </div>
    <div v-if="sub" class="metric-sub" :class="subTone && subTone !== 'neutral' ? `tone-${subTone}` : null">
      {{ sub }}
    </div>
  </div>
</template>
