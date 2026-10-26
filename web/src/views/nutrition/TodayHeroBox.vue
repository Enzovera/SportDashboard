<script setup lang="ts">
import { computed } from 'vue'
import type { NutritionTodayDetail } from '@/api/nutrition'
import BoxSection from '@/components/ui/BoxSection.vue'
import { num } from '@/lib/format'
import { remaining, targetMeta, todayTitle } from './nutritionData'

// Desktop hero: today's intake vs the target of today's day type (design/screens/DesktopZywienie.html).
const props = defineProps<{ today: NutritionTodayDetail }>()
const rows = computed(() => [
  { label: 'Kalorie', unit: 'kcal', row: props.today.kcal },
  { label: 'Białko', unit: 'g', row: props.today.proteinG },
  { label: 'Węglowodany', unit: 'g', row: props.today.carbsG },
  { label: 'Tłuszcz', unit: 'g', row: props.today.fatG },
])
</script>

<template>
  <BoxSection hero :title="todayTitle(today.dayType)" :meta="targetMeta(today.dayType)" label="Dziś">
    <div class="hero-num">
      <span class="big">{{ num(today.kcal.eaten) }}</span><span class="of">/ {{ num(today.kcal.target) }} kcal</span>
    </div>
    <table class="table">
      <thead>
        <tr>
          <th scope="col"><span class="sr-only">Składnik</span></th>
          <th scope="col" class="r">zjedzone</th>
          <th scope="col" class="r">cel</th>
          <th scope="col" class="r">zostało</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="r in rows" :key="r.label">
          <th scope="row">{{ r.label }} <span class="unit">{{ r.unit }}</span></th>
          <td class="r strong">{{ num(r.row.eaten) }}</td>
          <td class="r dim">{{ num(r.row.target) }}</td>
          <td class="r strong">{{ num(remaining(r.row)) }}</td>
        </tr>
      </tbody>
    </table>
  </BoxSection>
</template>

<style scoped>
.hero-num { display: flex; align-items: baseline; gap: 10px; padding: 6px 0 8px; }
.big { font-size: 96px; font-weight: 800; font-stretch: 70%; line-height: 0.82; }
.of { font-size: 22px; font-weight: 700; font-stretch: 78%; color: var(--hero-ink2); }
.table { color: var(--hero-ink); }
.table thead tr, .table .dim, .unit { color: var(--hero-ink2); }
.table thead th, .table tbody th { text-align: left; }
.table thead th.r { text-align: right; }
.table tbody tr { border-color: var(--hero-rule); }
.table tbody tr:last-child { border-bottom: 0; }
.unit { font-size: 12.5px; }
</style>
