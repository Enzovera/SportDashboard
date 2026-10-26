<script setup lang="ts">
import type { ValidatorCheck } from '@/api/proposals'

// ✓ pass (good), ! warn, ✕ fail. `blocked`: the plain amber list of the "Zablokowane" box.
defineProps<{ checks: ValidatorCheck[]; blocked?: boolean }>()

const MARK: Record<ValidatorCheck['status'], string> = { pass: '✓', warn: '!', fail: '✕' }
const SR: Record<ValidatorCheck['status'], string> = { pass: 'spełnione', warn: 'ostrzeżenie', fail: 'złamane' }
</script>

<template>
  <ul class="checks" :class="{ blocked }" aria-label="Wynik walidatorów">
    <li v-for="c in checks" :key="c.rule + c.text" :class="`st-${c.status}`">
      <span class="mark" aria-hidden="true">{{ MARK[c.status] }}</span>
      <span><span class="sr-only">{{ SR[c.status] }}: </span>{{ c.text }}</span>
    </li>
  </ul>
</template>

<style scoped>
.checks {
  margin: 0; padding: 10px 12px; list-style: none; display: flex; flex-direction: column; gap: 5px;
  background: var(--surface2); border: 1px solid var(--rule); font-size: 13.5px;
}
.checks li { display: flex; gap: 8px; }
.mark { font-weight: 800; width: 10px; flex: none; text-align: center; }
.st-pass .mark { color: var(--good); }
.st-warn, .st-fail { color: var(--warn); }
.checks.blocked { padding: 0; background: none; border: 0; color: var(--warn); }
@media (min-width: 1024px) {
  .checks { gap: 6px; font-size: 14px; }
  .checks li { gap: 10px; }
}
</style>
