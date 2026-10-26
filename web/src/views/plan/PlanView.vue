<script setup lang="ts">
import { computed } from 'vue'
import { getPlan } from '@/api/plan'
import { useApi } from '@/composables/useApi'
import { useLayout } from '@/composables/useLayout'
import PageLayout from '@/components/ui/PageLayout.vue'
import GoalHero from './GoalHero.vue'
import MacroTimeline from './MacroTimeline.vue'
import WeekBox from './WeekBox.vue'
import { fullDate, orderWeeks } from './planData'

// Read-only view of the plan_session projection. Changes go through /propozycje only.
const layout = useLayout()
const { data, error } = useApi(getPlan)

const plan = computed(() => data.value?.plan ?? null)
const meta = computed(() => (plan.value ? `Makrocykl · ${plan.value.goal.race} ${fullDate(plan.value.goal.date)}` : null))
const weeks = computed(() =>
  plan.value && data.value ? orderWeeks(plan.value.currentWeek, plan.value.nextWeek, data.value.date) : [],
)
const pending = computed(() => data.value?.proposalsPending ?? 0)
const pendingText = computed(() => {
  const n = pending.value
  const noun = n === 1 ? 'propozycja zmiany czeka' : n % 10 >= 2 && n % 10 <= 4 && (n % 100 < 10 || n % 100 >= 20) ? 'propozycje zmian czekają' : 'propozycji zmian czeka'
  return noun
})
</script>

<template>
  <PageLayout title="Plan" :meta="meta">
    <template #actions>
      <RouterLink v-if="plan" :to="{ path: '/propozycje', hash: '#historia' }" class="btn">Historia zmian</RouterLink>
    </template>

    <div v-if="error && !data" class="page">
      <section class="box" role="alert">{{ error }}</section>
    </div>

    <div v-else-if="data && !plan" :class="layout === 'desktop' ? 'd-grid' : 'page'">
      <div :class="layout === 'desktop' ? 'stack s-6' : 'col'">
        <section class="box" aria-label="Plan">
          <div class="box-head"><h2>Brak planu</h2></div>
        </section>
        <RouterLink to="/czat" class="link-row">
          <span class="grow">Ułóż plan z trenerem</span><span aria-hidden="true">→</span>
        </RouterLink>
      </div>
    </div>

    <!-- design/screens/DesktopPlan.html: goal (4) + macrocycle (8); this week (5) + next week (7). -->
    <div v-else-if="data && plan && layout === 'desktop'" class="d-grid">
      <section v-if="error" class="box s-12" role="alert">{{ error }}</section>
      <div class="stack s-4"><GoalHero :goal="plan.goal" wide /></div>
      <div class="stack s-8 stretch">
        <MacroTimeline
          :blocks="plan.blocks"
          :today="data.date"
          :race="plan.goal.race"
          :race-date="plan.goal.date"
          horizontal
        />
      </div>
      <div class="stack s-5">
        <WeekBox v-if="plan.currentWeek" :week="plan.currentWeek" :today="data.date" wide />
        <RouterLink v-if="pending" to="/propozycje" class="link-row">
          <span class="count">{{ pending }}</span><span class="grow strong">{{ pendingText }}</span><span>Przejrzyj →</span>
        </RouterLink>
      </div>
      <div class="stack s-7">
        <WeekBox v-if="plan.nextWeek" :week="plan.nextWeek" :today="data.date" wide />
      </div>
    </div>

    <!-- design/screens/FoldPlan.html; cover = the same boxes in one column. -->
    <div v-else-if="data && plan" class="page">
      <div class="col">
        <section v-if="error" class="box" role="alert">{{ error }}</section>
        <GoalHero :goal="plan.goal" />
        <RouterLink v-if="pending" to="/propozycje" class="link-row">
          <span class="count">{{ pending }}</span><span class="grow strong">{{ pendingText }}</span><span>Przejrzyj →</span>
        </RouterLink>
        <template v-if="layout === 'cover'">
          <WeekBox v-for="w in weeks" :key="w.from" :week="w" :today="data.date" />
        </template>
        <MacroTimeline :blocks="plan.blocks" :today="data.date" :race="plan.goal.race" :race-date="plan.goal.date" />
      </div>
      <div v-if="layout === 'fold'" class="col">
        <WeekBox v-for="w in weeks" :key="w.from" :week="w" :today="data.date" />
      </div>
    </div>
  </PageLayout>
</template>

<style scoped>
.grow { flex: 1; }
.strong { font-weight: 700; }
</style>
