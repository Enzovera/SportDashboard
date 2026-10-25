<script setup lang="ts">
import type { StrengthRow } from '@/api/types'
import BoxSection from '@/components/ui/BoxSection.vue'
import { dayShort, num } from '@/lib/format'

/** syncProblem: syncProblem() text for Hevy, or null when the sync is fine. */
defineProps<{ sessions: StrengthRow[]; syncProblem: string | null }>()
</script>

<template>
  <BoxSection v-if="sessions.length" title="Push / Pull / Legs" label="Treningi siłowe" class="table-box">
    <template #meta>
      <span v-if="syncProblem" class="status" :title="`Hevy: ${syncProblem}`">{{ syncProblem }}</span>
      <span v-else class="meta">Hevy</span>
    </template>
    <table class="table">
      <thead>
        <tr>
          <th scope="col"><span class="sr-only">Sesja</span></th>
          <th scope="col" class="r w44">min</th>
          <th scope="col" class="r w48">serie</th>
          <th scope="col" class="r w72">tonaż kg</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="s in sessions" :key="s.id">
          <th scope="row" class="name">
            <span class="title">{{ s.title }}</span>
            <span class="date">{{ dayShort(s.date) }}</span>
          </th>
          <td class="r">{{ num(s.durationMin) }}</td>
          <td class="r">{{ num(s.sets) }}</td>
          <td class="r strong">{{ num(s.tonnageKg) }}</td>
        </tr>
      </tbody>
    </table>
  </BoxSection>
</template>

<style scoped>
.table-box { gap: 4px; padding-bottom: 6px; }
.table { table-layout: fixed; }
.table tbody tr:last-child { border-bottom: 0; }
.table thead th { text-align: left; }
.table thead th.r { text-align: right; }
.w44 { width: 44px; }
.w48 { width: 48px; }
.w72 { width: 72px; }
.name { text-align: left; padding: 7px 0; }
.title { display: block; font-weight: 700; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.date { font-size: 12px; color: var(--ink3); }
</style>
