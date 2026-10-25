<script setup lang="ts">
import { computed } from 'vue'
import BoxSection from '@/components/ui/BoxSection.vue'
import type { NutritionToday, SupplementsToday } from '@/api/types'
import { num } from '@/lib/format'
import { nutritionLines } from './today'
import SupplementsLink from './SupplementsLink.vue'

const props = defineProps<{ nutrition: NutritionToday; supplements?: SupplementsToday | null; compact?: boolean }>()
const lines = computed(() => nutritionLines(props.nutrition))
const yesterdayTone = computed(() => {
  const tone = props.nutrition.yesterday?.proteinTone
  return tone && tone !== 'neutral' ? `tone-${tone}` : null
})
</script>

<template>
  <BoxSection
    class="nutrition"
    :class="{ 'has-link': !!supplements, compact }"
    :title="compact ? 'Żywienie' : 'Żywienie dziś'"
    label="Żywienie dziś"
    :meta="nutrition.targetLabel"
  >
    <table class="table">
      <thead>
        <tr>
          <th scope="col" class="name"><span class="sr-only">Składnik</span></th>
          <th scope="col" class="r">zjedzone</th>
          <th scope="col" class="r">cel</th>
          <th scope="col" class="r">zostało</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="l in lines" :key="l.label">
          <th scope="row" class="name">{{ l.label }} <span class="unit">{{ l.unit }}</span></th>
          <td class="r">{{ l.eaten }}</td>
          <td class="r dim">{{ l.target }}</td>
          <td class="r strong">{{ l.left }}</td>
        </tr>
      </tbody>
    </table>
    <p v-if="!compact && nutrition.yesterday" class="yesterday">
      Wczoraj {{ num(nutrition.yesterday.kcal.eaten) }} / {{ num(nutrition.yesterday.kcal.target) }} kcal, białko
      <span :class="yesterdayTone"
        >{{ num(nutrition.yesterday.proteinG.eaten) }} / {{ num(nutrition.yesterday.proteinG.target) }} g</span
      >
    </p>
    <SupplementsLink v-if="supplements" :supplements="supplements" />
  </BoxSection>
</template>

<style scoped>
.nutrition { gap: 4px; padding-bottom: 10px; }
.nutrition.has-link { padding-bottom: 0; }
.nutrition.compact { padding-bottom: 6px; }
.name { text-align: left; }
.unit { color: var(--ink3); font-weight: 400; }
.yesterday { margin-top: 2px; font-size: 13px; color: var(--ink2); }
.yesterday .tone-warn, .yesterday .tone-good { font-weight: 700; }
</style>
