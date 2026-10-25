<script setup lang="ts">
import type { RunRow } from '@/api/types'
import BoxSection from '@/components/ui/BoxSection.vue'
import { dayShort, num, pace } from '@/lib/format'

/** syncProblem: syncProblem() text for Garmin, or null when the sync is fine. */
defineProps<{ runs: RunRow[]; syncProblem: string | null }>()
</script>

<template>
  <BoxSection v-if="runs.length" title="Ostatnie biegi" class="table-box">
    <template #meta>
      <span v-if="syncProblem" class="status" :title="`Garmin: ${syncProblem}`">{{ syncProblem }}</span>
      <span v-else class="meta">Garmin</span>
    </template>
    <table class="table">
      <thead>
        <tr>
          <th scope="col"><span class="sr-only">Trening</span></th>
          <th scope="col" class="r w46">km</th>
          <th scope="col" class="r w46">/km</th>
          <th scope="col" class="r w40">bpm</th>
          <th scope="col" class="r w40">obc.</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="run in runs" :key="run.id">
          <th scope="row" class="name">
            <span class="title">{{ run.title }}</span>
            <span class="date">{{ dayShort(run.date) }}</span>
          </th>
          <td class="r strong">{{ num(run.distanceKm, 1) }}</td>
          <td class="r">{{ run.paceSPerKm !== null ? pace(run.paceSPerKm) : '' }}</td>
          <td class="r">{{ run.avgHr !== null ? num(run.avgHr) : '' }}</td>
          <td class="r dim">{{ run.load !== null ? num(run.load) : '' }}</td>
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
.w46 { width: 46px; }
.w40 { width: 40px; }
.name { text-align: left; padding: 7px 0; }
.title { display: block; font-weight: 700; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.date { font-size: 12px; color: var(--ink3); }
</style>
