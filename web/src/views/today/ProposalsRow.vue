<script setup lang="ts">
import { proposalsText } from './today'

defineProps<{ count: number; compact?: boolean }>()
</script>

<template>
  <RouterLink to="/propozycje" class="link-row">
    <span class="badge">{{ count }}</span>
    <span class="text">{{ proposalsText(count, !!compact) }}</span>
    <span class="cta">Przejrzyj <span aria-hidden="true">→</span></span>
  </RouterLink>
</template>

<style scoped>
.badge {
  min-width: 28px; height: 28px; padding: 0 6px; flex: none;
  display: flex; align-items: center; justify-content: center;
  background: var(--signal); color: var(--signal-ink); font-size: 15px; font-weight: 800;
}
.text { flex: 1; min-width: 0; font-size: 14.5px; font-weight: 700; }
.cta { font-size: 13px; font-weight: 600; white-space: nowrap; }
</style>
