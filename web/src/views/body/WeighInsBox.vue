<script setup lang="ts">
import { computed } from 'vue'
import type { WeighIn } from '@/api/body'
import BoxSection from '@/components/ui/BoxSection.vue'
import { dayShort, num } from '@/lib/format'
import { weighInRows } from './bodyData'

const props = defineProps<{ today: string; weighIns: WeighIn[]; source: string }>()

const rows = computed(() => weighInRows(props.today, props.weighIns))
const hasComposition = computed(() => props.weighIns.some((w) => w.fatPct !== null || w.muscleKg !== null))
const pct = (v: number | null) => (v === null ? '–' : `${num(v, 1)}%`)
const kg = (v: number | null) => (v === null ? '–' : num(v, 1))
</script>

<template>
  <BoxSection v-if="rows.length" title="Ostatnie ważenia" :meta="source" class="weighins">
    <table class="table">
      <thead>
        <tr>
          <th scope="col" class="day"><span class="sr-only">Dzień</span></th>
          <th scope="col" class="r w-kg">kg</th>
          <template v-if="hasComposition">
            <th scope="col" class="r w-fat">tłuszcz</th>
            <th scope="col" class="r w-muscle">mięśnie</th>
            <th scope="col" class="r w-water">woda</th>
          </template>
        </tr>
      </thead>
      <tbody>
        <tr v-for="row in rows" :key="row.date" :class="{ missing: !row.weighIn }">
          <th scope="row" class="day">{{ dayShort(row.date) }}</th>
          <template v-if="row.weighIn">
            <td class="r strong">{{ num(row.weighIn.weightKg, 1) }}</td>
            <template v-if="hasComposition">
              <td class="r">{{ pct(row.weighIn.fatPct) }}</td>
              <td class="r">{{ kg(row.weighIn.muscleKg) }}</td>
              <td class="r">{{ pct(row.weighIn.waterPct) }}</td>
            </template>
          </template>
          <template v-else>
            <td class="r">brak</td>
            <template v-if="hasComposition">
              <td class="r">–</td>
              <td class="r">–</td>
              <td class="r">–</td>
            </template>
          </template>
        </tr>
      </tbody>
    </table>
  </BoxSection>
</template>

<style scoped>
.weighins { gap: 4px; padding-bottom: 6px; }
.table { table-layout: fixed; }
.day { text-align: left; }
tbody .day { color: var(--ink2); }
.w-kg { width: 46px; }
.w-fat { width: 56px; }
.w-muscle { width: 58px; }
.w-water { width: 52px; }
tr.missing th, tr.missing td { color: var(--ink3); }
@media (min-width: 1024px) {
  .weighins { padding-bottom: 8px; }
  .w-kg { width: 52px; }
  .w-fat { width: 64px; }
  .w-muscle { width: 66px; }
  .w-water { width: 60px; }
}
</style>
