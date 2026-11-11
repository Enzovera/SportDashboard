<script setup lang="ts">
import { computed } from 'vue'
import type { WeekGrid } from '@/api/types'
import { weekTable, weekTotals } from './today'

const props = defineProps<{ week: WeekGrid; today: string; totals?: boolean; compact?: boolean }>()
const table = computed(() => weekTable(props.week, props.today))
</script>

<template>
  <section class="box week" :class="{ compact }" aria-label="Ten tydzień">
    <div class="box-head head">
      <h2>Ten tydzień</h2>
      <RouterLink v-if="week.runKmPlanned !== null" to="/plan" class="plan-link">Plan</RouterLink>
    </div>
    <table class="table grid">
      <thead>
        <tr>
          <th scope="col" class="kind"><span class="sr-only">Rodzaj</span></th>
          <th v-for="h in table.heads" :key="h.label" scope="col" :class="{ today: h.today }">
            {{ h.label }}<span v-if="h.today" class="sr-only"> (dziś)</span>
          </th>
        </tr>
      </thead>
      <tbody>
        <tr>
          <th scope="row">Bieg</th>
          <td
            v-for="(c, i) in table.run"
            :key="i"
            :class="{ muted: c.empty, dim: c.planned, signal: c.signal }"
          >{{ c.text }}</td>
        </tr>
        <tr>
          <th scope="row">Siła</th>
          <td v-for="(c, i) in table.strength" :key="i" :class="{ muted: c.empty }">{{ c.text }}</td>
        </tr>
      </tbody>
    </table>
    <div v-if="totals" class="totals">{{ weekTotals(week) }}</div>
  </section>
</template>

<style scoped>
.week { padding: 6px 14px 10px; gap: 2px; }
.head { align-items: center; min-height: 36px; }
.plan-link {
  display: inline-flex; align-items: center; min-height: 44px; margin: -4px 0;
  font-size: 13px; font-weight: 600; color: var(--ink); text-decoration: underline; text-underline-offset: 3px;
}
.plan-link:focus-visible { outline: 2px solid var(--ink); outline-offset: 2px; }
.grid { table-layout: fixed; font-size: 13px; font-stretch: 88%; }
.compact .grid { font-stretch: 85%; }
.grid thead th { padding: 0 0 5px; font-weight: 500; }
.grid thead th.kind { width: 40px; text-align: left; }
.compact .grid thead th.kind { width: 38px; }
.grid thead th.today { font-weight: 800; color: var(--ink); }
.grid tbody { text-align: center; }
.grid tbody th { padding: 7px 0; text-align: left; font-weight: 600; color: var(--ink2); }
.grid td { padding: 0; }
.grid td.signal { background: var(--signal); color: var(--signal-ink); font-weight: 800; }
.totals { font-size: 12.5px; color: var(--ink3); padding-top: 6px; }
@media (min-width: 1024px) {
  .week { padding: 10px 18px 14px; gap: 4px; }
  .grid { font-size: 14px; font-stretch: 100%; }
  .grid thead tr { font-size: 12.5px; }
  .grid thead th { padding-bottom: 7px; }
  .grid thead th.kind { width: 56px; }
  .grid tbody th, .grid td { padding-top: 9px; padding-bottom: 9px; }
  .grid tbody th { padding-left: 0; padding-right: 0; }
  .totals { font-size: 13px; padding-top: 8px; }
}
</style>
