<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { getPlan } from '@/api/plan'
import { proposalsPendingCount } from '@/api/proposals'
import { useApi } from '@/composables/useApi'
import { useElementWidth } from '@/composables/useElementWidth'
import { useLayout } from '@/composables/useLayout'
import Icon from '@/components/ui/Icon.vue'
import PageLayout from '@/components/ui/PageLayout.vue'
import ProposalsRow from '@/views/today/ProposalsRow.vue'
import GoalHero from './GoalHero.vue'
import MacroTimeline from './MacroTimeline.vue'
import WeekBox from './WeekBox.vue'
import { fullDate, orderWeeks } from './planData'

// Read-only view of the plan_session projection. Changes go through /propozycje only.
const layout = useLayout()
const { data, error } = useApi(getPlan)

// Desktop with the chat docked or a small window: the wide week tables and the horizontal timeline do not
// fit a halved grid, so the view falls back to the two-column arrangement of the Fold (compact boxes).
const WIDE_MIN_WIDTH = 1100
const root = ref<HTMLElement | null>(null)
const width = useElementWidth(root)
const wide = computed(() => layout.value === 'desktop' && (width.value === 0 || width.value > WIDE_MIN_WIDTH))
const twoCol = computed(() => layout.value !== 'cover')

const plan = computed(() => data.value?.plan ?? null)
const meta = computed(() => (plan.value ? `Makrocykl · ${plan.value.goal.race} ${fullDate(plan.value.goal.date)}` : null))
const weeks = computed(() =>
  plan.value && data.value ? orderWeeks(plan.value.currentWeek, plan.value.nextWeek, data.value.date) : [],
)
const pending = computed(() => data.value?.proposalsPending ?? 0)
// Keeps the nav badge in step when Plan is the entry point.
watch(data, (d) => {
  if (d) proposalsPendingCount.value = d.proposalsPending
})
</script>

<template>
  <PageLayout title="Plan" :meta="meta">
    <template v-if="plan && layout !== 'cover'" #actions>
      <RouterLink :to="{ path: '/propozycje', hash: '#historia' }" class="btn hist">
        <Icon name="history" :size="18" />Historia zmian
      </RouterLink>
    </template>

    <div ref="root">
    <div v-if="error && !data" class="page">
      <section class="box" role="alert">{{ error }}</section>
    </div>

    <div v-else-if="data && !plan" :class="layout === 'desktop' ? 'd-grid' : 'page'">
      <div :class="layout === 'desktop' ? 'stack s-6' : 'col'">
        <section class="box" aria-label="Plan">
          <div class="box-head"><h2>Brak planu</h2></div>
        </section>
        <RouterLink to="/czat" class="link-row">
          <span class="grow">Ułóż plan z trenerem</span><Icon name="arrow-right" :size="18" />
        </RouterLink>
      </div>
    </div>

    <!-- design/screens/DesktopPlan.html: goal (4) + macrocycle (8); this week (5) + next week (7). -->
    <div v-else-if="data && plan && wide" class="d-grid">
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
        <ProposalsRow v-if="pending" :count="pending" />
      </div>
      <div class="stack s-7">
        <WeekBox v-if="plan.nextWeek" :week="plan.nextWeek" :today="data.date" wide />
      </div>
    </div>

    <!-- design/screens/FoldPlan.html (also a narrow desktop column); cover = the same boxes in one column. -->
    <div v-else-if="data && plan" :class="layout === 'desktop' ? 'd-grid' : 'page'">
      <div :class="layout === 'desktop' ? 'stack s-6' : 'col'">
        <section v-if="error" class="box" role="alert">{{ error }}</section>
        <GoalHero :goal="plan.goal" />
        <ProposalsRow v-if="pending" :count="pending" :compact="layout === 'cover'" />
        <template v-if="!twoCol">
          <WeekBox v-for="w in weeks" :key="w.from" :week="w" :today="data.date" />
        </template>
        <MacroTimeline :blocks="plan.blocks" :today="data.date" :race="plan.goal.race" :race-date="plan.goal.date" />
        <!-- Cover: the header holds icons only, so the shortcut closes the column. -->
        <RouterLink v-if="layout === 'cover'" :to="{ path: '/propozycje', hash: '#historia' }" class="link-row">
          <Icon name="history" :size="20" /><span class="grow">Historia zmian</span><Icon name="arrow-right" :size="18" />
        </RouterLink>
      </div>
      <div v-if="twoCol" :class="layout === 'desktop' ? 'stack s-6' : 'col'">
        <WeekBox v-for="w in weeks" :key="w.from" :week="w" :today="data.date" />
      </div>
    </div>
    </div>
  </PageLayout>
</template>

<style scoped>
.grow { flex: 1; font-size: 14px; }
.hist { gap: 8px; white-space: nowrap; }
</style>
