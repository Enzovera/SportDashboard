<script setup lang="ts">
import { computed } from 'vue'
import BoxSection from '@/components/ui/BoxSection.vue'
import type { WeekDiff } from '@/api/proposals'
import { dateRange, num } from '@/lib/format'
import { dayNumbers, weekKm, weekRows } from './proposals'

// Desktop: the planned week now vs after accepting what still waits. Rows touched by a pending
// proposal are "needs a decision" (signal tint + 3 px bar), as in Plan.
const props = defineProps<{ week: WeekDiff; pendingIds: ReadonlySet<string> }>()

const rows = computed(() => weekRows(props.week, props.pendingIds))
</script>

<template>
  <BoxSection :title="`Tydzień ${dateRange(week.from, week.to)}`" :meta="week.block" label="Tydzień przed i po akceptacji">
    <table class="table wk">
      <colgroup><col class="c-day" /><col /><col /></colgroup>
      <thead>
        <tr>
          <th scope="col"><span class="sr-only">Dzień</span></th>
          <th scope="col">teraz</th>
          <th scope="col">po akceptacji</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="r in rows" :key="r.date" :class="{ 'row-decide': r.decide }">
          <th scope="row" class="day">{{ r.label }}</th>
          <td>
            <div class="s">{{ r.now.session }}</div>
            <div class="n">{{ dayNumbers(r.now) }}</div>
          </td>
          <td :class="{ 'after-same': !r.decide }">
            <div class="s" :class="{ chg: r.sessionChanged }">{{ r.after.session }}</div>
            <div class="n" :class="{ chg: r.numbersChanged }">{{ dayNumbers(r.after) }}</div>
          </td>
        </tr>
      </tbody>
      <tfoot>
        <tr>
          <th scope="row">Bieg</th>
          <td>{{ num(weekKm(rows, 'now')) }} km</td>
          <td>{{ num(weekKm(rows, 'after')) }} km</td>
        </tr>
      </tfoot>
    </table>
  </BoxSection>
</template>

<style scoped>
.wk { margin: 0 -18px; width: calc(100% + 36px); table-layout: fixed; }
.wk .c-day { width: 76px; }
.wk th, .wk td { text-align: left; vertical-align: top; }
.wk th:first-child, .wk td:first-child { padding-left: 18px; }
.wk th:last-child, .wk td:last-child { padding-right: 18px; }
.wk td + td { padding-left: 12px; }
.wk .day { white-space: nowrap; font-size: 13px; font-weight: 700; color: var(--ink3); line-height: 1.35; }
.wk .s { font-weight: 600; }
.wk .n { font-size: 12.5px; color: var(--ink3); }
.wk .after-same { color: var(--ink3); }
.wk .after-same .s { font-weight: 500; }
.wk tr.row-decide { background: var(--signal-tint); }
.wk tr.row-decide > :first-child { box-shadow: inset 3px 0 0 var(--signal); }
.wk tr.row-decide .day { color: var(--signal-tint-ink); }
.wk .chg { font-weight: 800; color: var(--ink); }
.wk tfoot tr { border-top: 1px solid var(--rule); }
.wk tfoot th, .wk tfoot td { font-weight: 700; padding-top: 10px; }
</style>
