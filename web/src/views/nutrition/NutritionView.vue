<script setup lang="ts">
import { computed } from 'vue'
import { getNutrition } from '@/api/nutrition'
import { useApi } from '@/composables/useApi'
import { useLayout } from '@/composables/useLayout'
import PageLayout from '@/components/ui/PageLayout.vue'
import SyncStatus from '@/components/ui/SyncStatus.vue'
import SyncButton from '@/components/ui/SyncButton.vue'
import BalanceBox from './BalanceBox.vue'
import DaysTable from './DaysTable.vue'
import KcalChartBox from './KcalChartBox.vue'
import MealsBox from './MealsBox.vue'
import NutritionSwitch from './NutritionSwitch.vue'
import ProteinChartBox from './ProteinChartBox.vue'
import TargetsBox from './TargetsBox.vue'
import TodayHeroBox from './TodayHeroBox.vue'
import { fitatuMeta, headerMeta, supplementsTabLabel } from './nutritionData'

const layout = useLayout()
const { data, error } = useApi(() => getNutrition('14d'))

const meta = computed(() => (data.value ? headerMeta(data.value) : null))
const tabLabel = computed(() => supplementsTabLabel(data.value?.supplements ?? null))
/** Without Fitatu data (before stage 2) there is nothing to show but the targets, if a plan exists. */
const empty = computed(() => !!data.value && !data.value.today && !data.value.days.length && !data.value.meals.length)
const todayType = computed(() => data.value?.today?.dayType ?? null)
</script>

<template>
  <PageLayout title="Żywienie" :meta="meta">
    <template v-if="data" #status><SyncStatus :sync="data.sync" :compact="layout === 'cover'" /></template>
    <template #global><SyncButton /></template>
    <template #actions>
      <NutritionSwitch :supplements-label="tabLabel" />
    </template>

    <div v-if="error && !data" :class="{ page: layout !== 'desktop' }">
      <section class="box" role="alert">{{ error }}</section>
    </div>

    <template v-else-if="data">
      <div v-if="empty" :class="layout === 'desktop' ? 'd-grid' : 'page'">
        <section class="box empty-box" :class="{ 's-4': layout === 'desktop' }" aria-label="Posiłki dziś">
          <div class="box-head"><h2>Posiłki dziś</h2></div>
          <p class="empty">Brak danych z Fitatu</p>
        </section>
        <TargetsBox
          v-if="data.targets.length"
          :targets="data.targets"
          :today-type="todayType"
          :proposals="data.targetProposals"
          :class="{ 's-4': layout === 'desktop' }"
        />
      </div>

      <!-- design/screens/DesktopZywienie.html: today (hero), targets, meals left (4); balance and charts right (8). -->
      <div v-else-if="layout === 'desktop'" class="d-grid">
        <div class="stack s-4">
          <TodayHeroBox v-if="data.today" :today="data.today" />
          <TargetsBox :targets="data.targets" :today-type="todayType" :proposals="data.targetProposals" />
          <MealsBox :meals="data.meals" :kcal="null" :coach-note="data.coachNote" :sync-meta="fitatuMeta(data)" wide />
        </div>
        <div class="stack s-8">
          <BalanceBox v-if="data.balance" :balance="data.balance" wide />
          <KcalChartBox v-if="data.days.length" :days="data.days" wide :height="250" />
          <ProteinChartBox v-if="data.balance" :days="data.days" :balance="data.balance" wide :height="212" />
        </div>
        <div class="s-12 full"><DaysTable :days="data.days" /></div>
      </div>

      <!-- design/screens/FoldZywienie.html: balance, calories, meals | targets, protein. -->
      <div v-else-if="layout === 'fold'" class="page">
        <div class="col">
          <BalanceBox v-if="data.balance" :balance="data.balance" />
          <KcalChartBox v-if="data.days.length" :days="data.days" />
          <MealsBox :meals="data.meals" :kcal="data.today?.kcal ?? null" :coach-note="data.coachNote" />
        </div>
        <div class="col">
          <TargetsBox :targets="data.targets" :today-type="todayType" :proposals="data.targetProposals" />
          <ProteinChartBox v-if="data.balance" :days="data.days" :balance="data.balance" />
        </div>
      </div>

      <!-- Cover: today first, then the 14-day picture, targets last. -->
      <div v-else class="page">
        <div class="col">
          <MealsBox :meals="data.meals" :kcal="data.today?.kcal ?? null" :coach-note="data.coachNote" />
          <BalanceBox v-if="data.balance" :balance="data.balance" />
          <KcalChartBox v-if="data.days.length" :days="data.days" />
          <ProteinChartBox v-if="data.balance" :days="data.days" :balance="data.balance" />
          <TargetsBox :targets="data.targets" :today-type="todayType" :proposals="data.targetProposals" />
        </div>
      </div>
    </template>
  </PageLayout>
</template>

<style scoped>
.empty-box { gap: 6px; }
.empty { font-size: 14px; color: var(--ink2); }
/* The 14-column day table needs the whole row also when the grid halves (chat open, narrow window). */
@container main (max-width: 1100px) {
  .d-grid > .full { grid-column: 1 / -1; }
}
</style>
