<script setup lang="ts">
import type { SupplementRegularity } from '@/api/supplements'
import BoxSection from '@/components/ui/BoxSection.vue'
import ChartLegend from '@/views/training/ChartLegend.vue'
import { dateRange, dayMonth, num } from '@/lib/format'
import { regularityDays, stripCells } from './supplementsData'

/** strip = desktop: one square per day between the name and the counts. */
defineProps<{ items: SupplementRegularity[]; from: string; to: string; strip?: boolean }>()
</script>

<template>
  <BoxSection v-if="items.length" title="Regularność · 30 dni" :meta="strip ? dateRange(from, to) : null" label="Regularność z 30 dni" class="reg-box">
    <table class="table reg" :class="{ strip }">
      <thead>
        <tr>
          <th scope="col" class="c-name"><span class="sr-only">Suplement</span></th>
          <th v-if="strip" scope="col" class="c-strip">
            <span class="axis"><span>{{ dayMonth(from) }}</span><span>{{ dayMonth(to) }}</span></span>
          </th>
          <th scope="col" class="r c-days">dni</th>
          <th scope="col" class="r c-pct">%</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="r in items" :key="r.supplementId" :class="{ warn: r.tone === 'warn' }">
          <th scope="row">{{ r.name }}</th>
          <td v-if="strip" class="c-strip">
            <svg class="cells" :viewBox="`0 0 ${r.days.length * 13 - 3} 14`" aria-hidden="true">
              <template v-for="c in stripCells(r.days)" :key="c.x">
                <rect v-if="c.kind === 'taken'" class="ok" :x="c.x" y="2" width="10" height="10" />
                <rect v-else-if="c.kind === 'missed'" class="miss" :x="c.x + 0.75" y="2.75" width="8.5" height="8.5" />
                <rect v-else class="na" :x="c.x + 1" y="6.25" width="8" height="1.5" />
              </template>
            </svg>
          </td>
          <td class="r">{{ regularityDays(r) }}</td>
          <td class="r pct">{{ num(r.pct) }}</td>
        </tr>
      </tbody>
    </table>
    <ChartLegend
      v-if="strip"
      class="legend"
      :items="[
        { kind: 'square', color: 'var(--chart)', label: 'przyjęte' },
        { kind: 'outline', color: 'var(--chart-mid)', label: 'pominięte' },
        { kind: 'line', color: 'var(--axis)', label: 'nie dotyczy', weight: 1.5 },
      ]"
    />
  </BoxSection>
</template>

<style scoped>
.reg-box { gap: 4px; padding-bottom: 6px; }
.reg { table-layout: fixed; }
.reg thead th, .reg tbody th { text-align: left; }
.reg thead th.r { text-align: right; }
.reg tbody tr:last-child { border-bottom: 0; }
.c-days { width: 104px; }
.c-pct { width: 40px; }
.pct { font-weight: 700; }
.reg tr.warn th, .reg tr.warn td { color: var(--warn); }
.reg tr.warn th { font-weight: 700; }
.reg tr.warn .pct { font-weight: 800; }

.strip .c-name { width: 170px; }
.strip .c-strip { width: 50%; padding-right: 24px; }
.strip .c-days { width: 116px; }
.strip .c-pct { width: 52px; }
.axis { display: flex; justify-content: space-between; max-width: 387px; font-weight: 500; }
.cells { display: block; width: 100%; max-width: 387px; height: auto; }
.ok { fill: var(--chart); }
.miss { fill: none; stroke: var(--chart-mid); stroke-width: 1.5; }
.warn .miss { stroke: var(--warn); }
.na { fill: var(--axis); }
.legend { padding-top: 6px; }
.legend :deep(.outline) { border-style: solid; }
</style>
