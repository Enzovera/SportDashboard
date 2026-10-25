<script setup lang="ts">
import { computed } from 'vue'
import PageLayout from '@/components/ui/PageLayout.vue'
import SyncStatus from '@/components/ui/SyncStatus.vue'
import { getToday } from '@/api/client'
import { useApi } from '@/composables/useApi'
import { useLayout } from '@/composables/useLayout'
import { headerMeta, sectionColumns } from './today'
import CoachSection from './CoachSection.vue'
import SessionHero from './SessionHero.vue'
import WeekSection from './WeekSection.vue'
import TomorrowRow from './TomorrowRow.vue'
import ReadinessSection from './ReadinessSection.vue'
import NutritionSection from './NutritionSection.vue'
import SupplementsSection from './SupplementsSection.vue'
import ProposalsRow from './ProposalsRow.vue'
import WeightRow from './WeightRow.vue'

const layout = useLayout()
const { data, error } = useApi(getToday)

const cover = computed(() => layout.value === 'cover')
const meta = computed(() => (data.value ? headerMeta(data.value) : null))
const columns = computed(() => (data.value ? sectionColumns(data.value, layout.value) : []))
</script>

<template>
  <PageLayout title="Dziś" :meta="meta">
    <template #actions>
      <template v-if="data">
        <SyncStatus :sync="data.sync" :compact="cover" />
        <RouterLink v-if="data.coach" to="/plan" class="btn trainer">Trener</RouterLink>
      </template>
    </template>

    <div v-if="data" class="page" :class="{ cover }">
      <div v-for="(col, i) in columns" :key="i" class="col">
        <template v-for="key in col" :key="key">
          <CoachSection v-if="key === 'coach' && data.coach" :coach="data.coach" :compact="cover" />
          <SessionHero
            v-else-if="key === 'session' && data.session"
            :session="data.session"
            :week="data.week"
            :compact="cover"
          />
          <WeekSection
            v-else-if="key === 'week'"
            :week="data.week"
            :today="data.date"
            :compact="cover"
            :totals="!cover || !data.session"
          />
          <TomorrowRow v-else-if="key === 'tomorrow' && data.tomorrow" :tomorrow="data.tomorrow" />
          <ReadinessSection v-else-if="key === 'readiness'" :readiness="data.readiness" :compact="cover" />
          <NutritionSection
            v-else-if="key === 'nutrition' && data.nutrition"
            :nutrition="data.nutrition"
            :supplements="cover ? null : data.supplements"
            :compact="cover"
          />
          <SupplementsSection v-else-if="key === 'supplements' && data.supplements" :supplements="data.supplements" />
          <ProposalsRow
            v-else-if="key === 'proposals' && data.proposalsPending"
            :count="data.proposalsPending"
            :compact="cover"
          />
          <WeightRow v-else-if="key === 'weight' && data.weight" :weight="data.weight" />
        </template>
      </div>
    </div>
    <div v-else-if="error" class="page">
      <section class="box error" role="alert">{{ error }}</section>
    </div>
  </PageLayout>
</template>

<style scoped>
.trainer { min-height: 44px; }
.cover .col { gap: 10px; }
.error { grid-column: 1 / -1; font-size: 14px; }
@media (max-width: 599px) {
  .trainer { padding: 0 12px; }
}
</style>
