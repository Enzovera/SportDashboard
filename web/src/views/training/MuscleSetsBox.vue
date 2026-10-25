<script setup lang="ts">
import { computed } from 'vue'
import type { MuscleSets } from '@/api/types'
import BoxSection from '@/components/ui/BoxSection.vue'
import { num } from '@/lib/format'
import { muscleLabel } from '@/lib/labels'
import { barShares } from './trainingData'

const props = defineProps<{ items: MuscleSets[] }>()
const rows = computed(() => {
  const shares = barShares(props.items.map((i) => i.sets))
  return props.items.map((item, i) => ({ ...item, label: muscleLabel(item.muscle), share: shares[i] ?? 0 }))
})
</script>

<template>
  <BoxSection v-if="items.length" title="Serie robocze · tydzień" meta="Hevy" label="Serie robocze na partię" class="sets-box">
    <ul class="rows">
      <li v-for="row in rows" :key="row.muscle" class="row">
        <span class="name">{{ row.label }}</span>
        <span class="track" aria-hidden="true"><span class="fill" :style="{ width: `${row.share * 100}%` }" /></span>
        <span class="sets-n">{{ num(row.sets) }}</span>
      </li>
    </ul>
  </BoxSection>
</template>

<style scoped>
.sets-box { gap: 6px; }
.rows { list-style: none; margin: 0; padding: 0; display: flex; flex-direction: column; gap: 2px; }
.row { display: flex; align-items: center; gap: 10px; min-height: 26px; font-size: 13.5px; }
.name { flex: 0 0 100px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.track { flex: 1; height: 8px; background: var(--rule); }
.fill { display: block; height: 100%; background: var(--chart-mid); }
.sets-n { flex: 0 0 22px; text-align: right; font-weight: 700; }
</style>
