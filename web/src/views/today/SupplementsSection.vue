<script setup lang="ts">
import type { SupplementsToday } from '@/api/types'

// Standalone supplements box (cover screen, or fold without nutrition data).
defineProps<{ supplements: SupplementsToday }>()
</script>

<template>
  <section class="box supplements" aria-label="Suplementy dziś">
    <div class="box-head">
      <h2>Suplementy</h2>
      <span class="taken">{{ supplements.taken }}<span class="of"> / {{ supplements.total }}</span></span>
    </div>
    <RouterLink to="/zywienie" class="row">
      <span>
        <template v-if="supplements.next">Następna: {{ supplements.next }} <span class="of">· wszystkie</span></template>
        <template v-else>Wszystkie</template>
      </span>
      <span class="arrow" aria-hidden="true">→</span>
    </RouterLink>
  </section>
</template>

<style scoped>
.supplements { padding-bottom: 0; gap: 4px; }
.taken { font-size: 14px; font-weight: 700; }
.of { color: var(--ink3); font-weight: 500; }
.row {
  display: flex; align-items: center; justify-content: space-between; gap: 8px; min-height: 44px;
  border-top: 1px solid var(--rule); font-size: 13.5px; text-decoration: none; color: var(--ink);
}
.row:focus-visible { outline: 2px solid var(--ink); outline-offset: 2px; }
.arrow { font-weight: 700; }
</style>
