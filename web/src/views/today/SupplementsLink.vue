<script setup lang="ts">
import Icon from '@/components/ui/Icon.vue'
import type { SupplementsToday } from '@/api/types'

// Supplements as one row at the bottom of the nutrition box (fold screen).
defineProps<{ supplements: SupplementsToday }>()
</script>

<template>
  <RouterLink to="/zywienie/suplementy" class="row">
    <span>
      <strong>Suplementy {{ supplements.taken }} / {{ supplements.total }}</strong>
      <span v-if="supplements.next" class="next"> · następna: {{ supplements.next }}</span>
    </span>
    <Icon name="arrow-right" :size="18" class="arrow" />
  </RouterLink>
</template>

<style scoped>
.row {
  display: flex; justify-content: space-between; align-items: center; gap: 8px; min-height: 44px;
  border-top: 1px solid var(--rule); margin-top: 4px; text-decoration: none; color: var(--ink); font-size: 13.5px;
}
.row:focus-visible { outline: 2px solid var(--ink); outline-offset: 2px; }
strong { font-weight: 700; }
.next { color: var(--ink3); }
.arrow { color: var(--ink2); }
</style>
