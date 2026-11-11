<script setup lang="ts">
// Theme toggle Auto -> Jasny -> Ciemny. Icon only, drawn like components/ui/Icon.vue;
// the mode is in aria-label and title.
import { useTheme } from '@/composables/useTheme'
import { THEME_LABEL } from '@/lib/theme'

const { mode, cycle } = useTheme()
</script>

<template>
  <button
    type="button"
    class="icon-btn theme"
    :aria-label="`Motyw: ${THEME_LABEL[mode]}`"
    :title="`Motyw: ${THEME_LABEL[mode]}`"
    @click="cycle"
  >
    <svg viewBox="0 0 24 24" width="22" height="22" aria-hidden="true">
      <!-- auto: half-filled circle -->
      <g v-if="mode === 'auto'">
        <circle cx="12" cy="12" r="8" />
        <path d="M12 4 a8 8 0 0 1 0 16 z" class="fill" />
      </g>
      <!-- light: sun -->
      <g v-else-if="mode === 'light'">
        <circle cx="12" cy="12" r="4" />
        <path d="M12 2v3M12 19v3M2 12h3M19 12h3M4.9 4.9l2.1 2.1M17 17l2.1 2.1M4.9 19.1L7 17M17 7l2.1-2.1" />
      </g>
      <!-- dark: moon -->
      <path v-else d="M20 14.5A8 8 0 0 1 9.5 4 8 8 0 1 0 20 14.5z" />
    </svg>
  </button>
</template>

<style scoped>
svg { fill: none; stroke: currentColor; stroke-width: 1.75; stroke-linecap: square; stroke-linejoin: miter; }
.fill { fill: currentColor; stroke: none; }
</style>
