<script setup lang="ts">
import type { Meal } from '@/api/nutrition'
import type { NutritionRow } from '@/api/types'
import BoxSection from '@/components/ui/BoxSection.vue'
import { num } from '@/lib/format'
import { MEAL_LABEL, mealMacros, mealTime } from './nutritionData'

/**
 * kcal: today's kcal vs target in the head (fold/cover, where there is no hero).
 * syncMeta: "Fitatu 08:15" in the head instead (desktop, the hero shows kcal).
 */
defineProps<{
  meals: Meal[]
  kcal: NutritionRow | null
  coachNote: { text: string } | null
  syncMeta?: string | null
  wide?: boolean
}>()
</script>

<template>
  <BoxSection title="Posiłki dziś" class="meals">
    <template #meta>
      <span v-if="wide" class="meta">{{ syncMeta }}</span>
      <span v-else-if="kcal" class="sum">
        <b>{{ num(kcal.eaten) }}</b><span class="of"> / {{ num(kcal.target) }} kcal</span>
      </span>
    </template>
    <div v-for="m in meals" :key="m.id" class="meal">
      <div class="top"><span class="slot">{{ MEAL_LABEL[m.slot] }}</span><span>{{ mealTime(m) }}</span></div>
      <div class="name">{{ m.title }}</div>
      <div class="macro">{{ mealMacros(m) }}</div>
    </div>
    <div v-if="!meals.length" class="meal empty">–</div>
    <p v-if="coachNote" class="coach"><b>Trener:</b> {{ coachNote.text }}</p>
  </BoxSection>
</template>

<style scoped>
.meals { gap: 2px; }
.sum { font-size: 13.5px; white-space: nowrap; }
.of { color: var(--ink3); }
.meal { display: flex; flex-direction: column; gap: 2px; padding: 8px 0; border-top: 1px solid var(--rule); }
.top { display: flex; justify-content: space-between; gap: 8px; font-size: 12.5px; color: var(--ink3); }
.slot { font-weight: 600; }
.name { font-size: 14.5px; font-weight: 600; }
.macro { font-size: 13px; color: var(--ink2); }
.empty { color: var(--ink3); }
.coach { padding: 8px 0 2px; border-top: 1px solid var(--rule); font-size: 13.5px; line-height: 1.5; }
@media (min-width: 1024px) {
  .meal { gap: 3px; padding: 10px 0; }
  .top { font-size: 13px; }
  .name { font-size: 15.5px; }
  .macro { font-size: 13.5px; }
  .coach { padding-top: 10px; font-size: 16px; line-height: 1.55; }
}
</style>
