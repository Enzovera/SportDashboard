<script setup lang="ts">
// One section of the Dziś screen by key; the same markup serves cover, fold and desktop.
import type { TodayResponse } from '@/api/types'
import type { Layout } from '@/composables/useLayout'
import type { SectionKey } from './today'
import CoachSection from './CoachSection.vue'
import SessionHero from './SessionHero.vue'
import WeekSection from './WeekSection.vue'
import TomorrowRow from './TomorrowRow.vue'
import ReadinessSection from './ReadinessSection.vue'
import NutritionSection from './NutritionSection.vue'
import SupplementsSection from './SupplementsSection.vue'
import ProposalsRow from './ProposalsRow.vue'
import WeightRow from './WeightRow.vue'

const props = defineProps<{ name: SectionKey; data: TodayResponse; layout: Layout }>()
const cover = props.layout === 'cover'
</script>

<template>
  <CoachSection v-if="name === 'coach' && data.coach" :coach="data.coach" :compact="cover" />
  <SessionHero v-else-if="name === 'session' && data.session" :session="data.session" :week="data.week" :compact="cover" />
  <WeekSection
    v-else-if="name === 'week'"
    :week="data.week"
    :today="data.date"
    :compact="cover"
    :totals="!cover || !data.session"
  />
  <TomorrowRow v-else-if="name === 'tomorrow' && data.tomorrow" :tomorrow="data.tomorrow" />
  <ReadinessSection v-else-if="name === 'readiness'" :readiness="data.readiness" :compact="cover" />
  <NutritionSection
    v-else-if="name === 'nutrition' && data.nutrition"
    :nutrition="data.nutrition"
    :supplements="layout === 'fold' ? data.supplements : null"
    :compact="cover"
  />
  <SupplementsSection v-else-if="name === 'supplements' && data.supplements" :supplements="data.supplements" />
  <ProposalsRow v-else-if="name === 'proposals' && data.proposalsPending" :count="data.proposalsPending" :compact="cover" />
  <WeightRow v-else-if="name === 'weight' && data.weight" :weight="data.weight" />
</template>
