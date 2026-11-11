<script setup lang="ts">
import { syncNow, syncState } from '@/api/sync'
import { useLayout } from '@/composables/useLayout'
import Icon from './Icon.vue'

// "Odśwież": syncs all sources now; every view reloads when it is done (api/sync.ts).
// Fold/cover: icon button in the header; desktop: icon with the label.
const layout = useLayout()
</script>

<template>
  <span v-if="syncState.problem" class="status" :class="{ 'sr-only': layout === 'cover' }" role="status">
    {{ syncState.problem }}
  </span>
  <button
    type="button"
    class="sync"
    :class="[layout === 'desktop' ? 'btn' : 'icon-btn', { busy: syncState.busy, problem: syncState.problem }]"
    :disabled="syncState.busy"
    :aria-busy="syncState.busy"
    :aria-label="syncState.busy ? 'Odświeżam dane' : 'Odśwież dane'"
    :title="syncState.problem ?? (syncState.busy ? 'Odświeżam…' : 'Odśwież dane')"
    @click="syncNow"
  >
    <Icon name="sync" :size="layout === 'desktop' ? 18 : 22" />
    <span v-if="layout === 'desktop'" class="lbl">{{ syncState.busy ? 'Odświeżam…' : 'Odśwież' }}</span>
  </button>
</template>

<style scoped>
.sync { white-space: nowrap; gap: 8px; }
.sync:disabled { cursor: default; }
.busy :deep(.icon) { animation: spin 1s linear infinite; }
/* Cover has no room for the words: the failed refresh marks the icon (the text stays for screen readers). */
.icon-btn.problem { color: var(--warn); }
/* Narrow desktop: the header has no room for labels next to six tabs and a sync status. */
@media (max-width: 1199px) {
  .btn { width: 44px; padding: 0; justify-content: center; }
  .lbl { display: none; }
}
@keyframes spin { to { transform: rotate(360deg); } }
@media (prefers-reduced-motion: reduce) {
  .busy :deep(.icon) { animation: none; opacity: 0.45; }
}
</style>
