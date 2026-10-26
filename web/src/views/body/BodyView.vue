<script setup lang="ts">
import { computed } from 'vue'
import { getBody } from '@/api/body'
import { useApi } from '@/composables/useApi'
import { useLayout } from '@/composables/useLayout'
import PageLayout from '@/components/ui/PageLayout.vue'
import SyncStatus from '@/components/ui/SyncStatus.vue'
import { num } from '@/lib/format'
import BodyNoteBox from './BodyNoteBox.vue'
import CompositionBox from './CompositionBox.vue'
import WeighInsBox from './WeighInsBox.vue'
import WeightChartBox from './WeightChartBox.vue'

const layout = useLayout()
const { data, error } = useApi(getBody)

const meta = computed(() => {
  if (!data.value) return null
  const src = data.value.device ? `${data.value.device} → Garmin` : 'Garmin'
  return `${src} · ${data.value.rangeWeeks} tygodni`
})
const goal = computed(() => {
  const g = data.value?.goal
  return g ? `Cel wagi: ${num(g.targetKg, 1)} kg` : 'Cel wagi: brak'
})
const source = computed(() => data.value?.device ?? 'Garmin')
const hasWeighIns = computed(() => (data.value?.composition.rangeCount ?? 0) > 0)
</script>

<template>
  <PageLayout title="Ciało" :meta="meta">
    <template #global>
      <SyncStatus v-if="data" :sync="data.sync" :compact="layout === 'cover'" />
    </template>
    <template #actions>
      <span v-if="data && hasWeighIns" class="goal">{{ goal }}</span>
    </template>

    <div v-if="error && !data" class="page">
      <section class="box" role="alert">{{ error }}</section>
    </div>

    <div v-else-if="data && !hasWeighIns" :class="layout === 'desktop' ? 'd-grid' : 'page'">
      <div :class="layout === 'desktop' ? 'stack s-6' : 'col'">
        <section class="box" aria-label="Ważenia">
          <div class="box-head"><h2>Brak ważeń w {{ data.rangeWeeks }} tyg.</h2></div>
        </section>
      </div>
    </div>

    <!-- design/screens/DesktopCialo.html: composition and weekly trend left (8), weigh-ins and coach right (4). -->
    <div v-else-if="data && layout === 'desktop'" class="d-grid">
      <div class="stack s-8">
        <section v-if="error" class="box" role="alert">{{ error }}</section>
        <CompositionBox :composition="data.composition" :range-weeks="data.rangeWeeks" wide />
        <WeightChartBox
          :weeks="data.weeks"
          :today="data.date"
          :trend-kg-per-week="data.trendKgPerWeek"
          :height="278"
          wide
        />
      </div>
      <div class="stack s-4">
        <WeighInsBox :today="data.date" :weigh-ins="data.weighIns" :source="source" />
        <BodyNoteBox v-if="data.coachNote" :note="data.coachNote" />
      </div>
    </div>

    <!-- design/screens/FoldCialo.html; on the cover screen the same boxes in one column. -->
    <div v-else-if="data" class="page">
      <div class="col">
        <section v-if="error" class="box" role="alert">{{ error }}</section>
        <CompositionBox :composition="data.composition" :range-weeks="data.rangeWeeks" />
        <WeightChartBox :weeks="data.weeks" :today="data.date" :trend-kg-per-week="data.trendKgPerWeek" />
        <template v-if="layout === 'cover'">
          <WeighInsBox :today="data.date" :weigh-ins="data.weighIns" :source="source" />
          <BodyNoteBox v-if="data.coachNote" :note="data.coachNote" />
        </template>
      </div>
      <div v-if="layout === 'fold'" class="col">
        <WeighInsBox :today="data.date" :weigh-ins="data.weighIns" :source="source" />
        <BodyNoteBox v-if="data.coachNote" :note="data.coachNote" />
      </div>
    </div>
  </PageLayout>
</template>

<style scoped>
.goal { font-size: 13px; color: var(--ink3); white-space: nowrap; }
@media (min-width: 1024px) { .goal { font-size: 14px; } }
</style>
