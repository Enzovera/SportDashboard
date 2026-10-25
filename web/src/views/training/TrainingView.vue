<script setup lang="ts">
import { computed, ref } from 'vue'
import type { Source, TrainingRange } from '@/api/types'
import { getTraining } from '@/api/client'
import { useApi } from '@/composables/useApi'
import { useLayout } from '@/composables/useLayout'
import PageLayout from '@/components/ui/PageLayout.vue'
import { dayMonth } from '@/lib/format'
import { syncProblem } from '@/lib/labels'
import CoachNoteBox from './CoachNoteBox.vue'
import LoadChartBox from './LoadChartBox.vue'
import LoadSummaryBox from './LoadSummaryBox.vue'
import MuscleSetsBox from './MuscleSetsBox.vue'
import RecentRunsBox from './RecentRunsBox.vue'
import StrengthBox from './StrengthBox.vue'
import WeeklyKmBox from './WeeklyKmBox.vue'
import { rangeWeeks } from './trainingData'

const ranges: { value: TrainingRange; label: string }[] = [
  { value: '4w', label: '4 tyg.' },
  { value: '12w', label: '12 tyg.' },
  { value: 'block', label: 'Blok' },
]
const range = ref<TrainingRange>('12w')
const layout = useLayout()
const { data, error } = useApi(() => getTraining(range.value), [range])

const meta = computed(() =>
  data.value ? `Garmin + Hevy · ${dayMonth(data.value.from)}–${dayMonth(data.value.to)}` : null,
)
const weeks = computed(() => (data.value ? rangeWeeks(data.value.from, data.value.to) : 0))

function problemOf(source: Source): string | null {
  const s = data.value?.sync.find((x) => x.source === source)
  return s ? syncProblem(s) : null
}
const garminProblem = computed(() => problemOf('garmin'))
const hevyProblem = computed(() => problemOf('hevy'))
</script>

<template>
  <PageLayout title="Trening" :meta="meta">
    <template #actions>
      <div class="segmented" role="group" aria-label="Zakres czasu">
        <button
          v-for="r in ranges"
          :key="r.value"
          type="button"
          :aria-pressed="range === r.value"
          @click="range = r.value"
        >
          {{ r.label }}
        </button>
      </div>
    </template>

    <div v-if="error && !data" class="page">
      <section class="box" role="alert">{{ error }}</section>
    </div>

    <div v-else-if="data" class="page">
      <template v-if="layout === 'fold'">
        <div class="col">
          <section v-if="error" class="box" role="alert">{{ error }}</section>
          <LoadSummaryBox :load="data.load" :weeks="weeks" />
          <WeeklyKmBox :weeks="data.weeklyKm" :today-planned-km="data.todayPlannedKm" />
          <LoadChartBox :points="data.loadSeries" />
          <CoachNoteBox v-if="data.coachNote" :note="data.coachNote" />
        </div>
        <div class="col">
          <RecentRunsBox :runs="data.recentRuns" :sync-problem="garminProblem" />
          <MuscleSetsBox :items="data.setsPerMuscle.items" />
          <StrengthBox :sessions="data.strengthSessions" :sync-problem="hevyProblem" />
        </div>
      </template>
      <div v-else class="col">
        <section v-if="error" class="box" role="alert">{{ error }}</section>
        <LoadSummaryBox :load="data.load" :weeks="weeks" />
        <WeeklyKmBox :weeks="data.weeklyKm" :today-planned-km="data.todayPlannedKm" />
        <RecentRunsBox :runs="data.recentRuns" :sync-problem="garminProblem" />
        <LoadChartBox :points="data.loadSeries" />
        <MuscleSetsBox :items="data.setsPerMuscle.items" />
        <StrengthBox :sessions="data.strengthSessions" :sync-problem="hevyProblem" />
        <CoachNoteBox v-if="data.coachNote" :note="data.coachNote" />
      </div>
    </div>
  </PageLayout>
</template>
