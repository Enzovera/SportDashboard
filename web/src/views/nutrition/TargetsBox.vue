<script setup lang="ts">
import Icon from '@/components/ui/Icon.vue'
import type { DayTarget, DayType, TargetProposal } from '@/api/nutrition'
import BoxSection from '@/components/ui/BoxSection.vue'
import { num } from '@/lib/format'
import { DAY_TYPE_LABEL, proposalValue, proposalWhere } from './nutritionData'

defineProps<{ targets: DayTarget[]; todayType: DayType | null; proposals: TargetProposal[] }>()
</script>

<template>
  <BoxSection v-if="targets.length" title="Cele per typ dnia" meta="makro w g" class="targets">
    <table class="table">
      <thead>
        <tr>
          <th scope="col"><span class="sr-only">Typ dnia</span></th>
          <th scope="col" class="r w-kcal">kcal</th>
          <th scope="col" class="r w-m">B</th>
          <th scope="col" class="r w-m">W</th>
          <th scope="col" class="r w-s">T</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="t in targets" :key="t.dayType">
          <th scope="row" :class="{ strong: t.dayType === todayType }">
            {{ DAY_TYPE_LABEL[t.dayType] }}<span v-if="t.dayType === todayType" class="sr-only"> (dziś)</span>
          </th>
          <td class="r strong">{{ num(t.kcal) }}</td>
          <td class="r">{{ num(t.proteinG) }}</td>
          <td class="r">{{ num(t.carbsG) }}</td>
          <td class="r">{{ num(t.fatG) }}</td>
        </tr>
      </tbody>
    </table>
    <RouterLink v-for="p in proposals" :key="p.id" to="/propozycje" class="proposal">
      <span class="text">
        <span class="mark" aria-hidden="true" />
        <span>Propozycja: <b>{{ proposalValue(p) }}</b> {{ proposalWhere(p) }}</span>
      </span>
      <Icon name="arrow-right" :size="18" class="arrow" />
    </RouterLink>
  </BoxSection>
</template>

<style scoped>
.targets { gap: 4px; padding-bottom: 0; }
.targets:not(:has(.proposal)) { padding-bottom: 6px; }
.table { table-layout: fixed; }
.table thead th, .table tbody th { text-align: left; }
.table thead th.r { text-align: right; }
.w-kcal { width: 52px; }
.w-m { width: 38px; }
.w-s { width: 34px; }
.proposal {
  display: flex; align-items: center; justify-content: space-between; gap: 12px; min-height: 44px;
  text-decoration: none; color: var(--ink); font-size: 13.5px;
}
.proposal + .proposal { border-top: 1px solid var(--rule); }
.proposal:focus-visible { outline: 2px solid var(--ink); outline-offset: -2px; }
.text { display: flex; align-items: center; gap: 10px; }
.mark { flex: none; width: 8px; height: 8px; background: var(--signal); }
.arrow { color: var(--ink2); }
@media (min-width: 1024px) {
  .w-kcal { width: 64px; }
  .w-m { width: 48px; }
  .w-s { width: 44px; }
  .proposal { font-size: 14px; }
}
</style>
