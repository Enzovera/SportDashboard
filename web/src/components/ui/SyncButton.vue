<script setup lang="ts">
import { syncNow, syncState } from '@/api/sync'

// Header "Odśwież": syncs all sources now; every view reloads when it is done (api/sync.ts).
</script>

<template>
  <span v-if="syncState.problem" class="status" role="status">{{ syncState.problem }}</span>
  <button type="button" class="btn sync" :disabled="syncState.busy" :aria-busy="syncState.busy" @click="syncNow">
    {{ syncState.busy ? 'Odświeżam…' : 'Odśwież' }}
  </button>
</template>

<style scoped>
.sync { white-space: nowrap; }
</style>
