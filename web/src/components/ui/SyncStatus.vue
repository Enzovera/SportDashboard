<script setup lang="ts">
import { computed } from 'vue'
import type { SyncStatus } from '@/api/types'
import { SOURCE_LABEL, syncProblem } from '@/lib/labels'

// Shows only sources with a problem, e.g. "Hevy: błąd sync". Nothing when all is fine.
const props = defineProps<{ sync: SyncStatus[]; compact?: boolean }>()

const problems = computed(() =>
  props.sync
    .map((s) => ({ source: SOURCE_LABEL[s.source], text: syncProblem(s) }))
    .filter((p): p is { source: string; text: string } => p.text !== null),
)
</script>

<template>
  <span v-for="p in problems" :key="p.source" class="status" :title="`${p.source}: ${p.text}`">
    {{ compact ? p.source : `${p.source}: ${p.text}` }}
  </span>
</template>
