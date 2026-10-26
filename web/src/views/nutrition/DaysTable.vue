<script setup lang="ts">
import { computed } from 'vue'
import type { NutritionDay } from '@/api/nutrition'
import BoxSection from '@/components/ui/BoxSection.vue'
import { dateRange, dayMonth, num } from '@/lib/format'
import { DAY_TYPE_SHORT, isKeyDay, pct } from './nutritionData'

// Desktop: days as columns, today last in red (design/screens/DesktopZywienie.html).
const props = defineProps<{ days: NutritionDay[] }>()
const meta = computed(() => {
  const first = props.days[0]
  const last = props.days[props.days.length - 1]
  return first && last ? dateRange(first.date, last.date) : null
})
</script>

<template>
  <BoxSection v-if="days.length" title="Dzień po dniu" :meta="meta">
    <div class="scroll">
      <table class="table days">
        <thead>
          <tr>
            <th scope="col"><span class="sr-only">Wskaźnik</span></th>
            <th v-for="d in days" :key="d.date" scope="col" :class="{ today: !d.complete }">{{ dayMonth(d.date) }}</th>
          </tr>
        </thead>
        <tbody>
          <tr>
            <th scope="row" class="dim">Typ dnia</th>
            <td v-for="d in days" :key="d.date" :class="!d.complete ? 'today' : isKeyDay(d) ? 'key' : 'dim'">
              {{ d.dayType ? DAY_TYPE_SHORT[d.dayType] : '' }}
            </td>
          </tr>
          <tr>
            <th scope="row">Zjedzone kcal</th>
            <td v-for="d in days" :key="d.date" :class="d.complete ? 'strong' : 'today'">
              {{ d.kcal !== null ? num(d.kcal) : '–' }}
            </td>
          </tr>
          <tr>
            <th scope="row">Cel kcal</th>
            <td v-for="d in days" :key="d.date" class="dim">{{ d.kcalTarget !== null ? num(d.kcalTarget) : '–' }}</td>
          </tr>
          <tr>
            <th scope="row">vs cel</th>
            <td v-for="d in days" :key="d.date" :class="{ warn: d.kcalTone === 'warn', muted: !d.complete }">
              {{ !d.complete ? 'w toku' : d.kcalVsTargetPct !== null ? pct(d.kcalVsTargetPct) : '–' }}
            </td>
          </tr>
          <tr>
            <th scope="row">Białko g/kg</th>
            <td v-for="d in days" :key="d.date" :class="{ warn: d.proteinTone === 'warn', muted: !d.complete }">
              {{ !d.complete ? 'w toku' : d.proteinGPerKg !== null ? num(d.proteinGPerKg, 2) : '–' }}
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </BoxSection>
</template>

<style scoped>
.scroll { overflow-x: auto; }
.days { table-layout: fixed; min-width: 760px; }
.days th:first-child { width: 120px; text-align: left; }
.days thead th { text-align: right; }
.days thead th.today { color: var(--signal); font-weight: 800; }
.days td { text-align: right; white-space: nowrap; }
.days td.today { color: var(--signal); font-weight: 700; }
.days td.key { font-weight: 700; }
.days td.warn { color: var(--warn); font-weight: 700; }
.days tbody tr:last-child { border-bottom: 0; }
</style>
