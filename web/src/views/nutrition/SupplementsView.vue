<script setup lang="ts">
import { computed, ref, shallowRef, watch } from 'vue'
import {
  decideReminderProposal,
  getSupplements,
  logSupplement,
  type SupplementItem,
} from '@/api/supplements'
import { useApi } from '@/composables/useApi'
import { useLayout } from '@/composables/useLayout'
import PageLayout from '@/components/ui/PageLayout.vue'
import ChecklistBox from './ChecklistBox.vue'
import NutritionSwitch from './NutritionSwitch.vue'
import RegularityBox from './RegularityBox.vue'
import ReminderProposalBox from './ReminderProposalBox.vue'
import StockBox from './StockBox.vue'
import { supplementsTabLabel } from './nutritionData'
import { counts, nextStatus, replaceItem, supplementsMeta, withStatus } from './supplementsData'

const layout = useLayout()
const { data, error, reload } = useApi(getSupplements)

// Local copy of the checklist for optimistic ticks; replaced on every (re)load.
const items = shallowRef<SupplementItem[]>([])
watch(data, (d) => (items.value = d?.checklist ?? []), { immediate: true })
const busy = ref(new Set<string>())
const actionError = ref<string | null>(null)
const deciding = ref(false)

const meta = computed(() => (data.value ? supplementsMeta(data.value.date) : null))
const tabLabel = computed(() => (data.value ? supplementsTabLabel({ ...counts(items.value), next: null }) : 'Suplementy'))

function setBusy(id: string, on: boolean) {
  const next = new Set(busy.value)
  if (on) next.add(id)
  else next.delete(id)
  busy.value = next
}

async function toggle(item: SupplementItem) {
  if (!data.value || busy.value.has(item.id)) return
  const status = nextStatus(item)
  const before = item.status
  actionError.value = null
  items.value = withStatus(items.value, item.id, status)
  setBusy(item.id, true)
  try {
    items.value = replaceItem(items.value, await logSupplement(item.id, status, data.value.date))
  } catch (e) {
    items.value = withStatus(items.value, item.id, before)
    actionError.value = e instanceof Error ? e.message : String(e)
  } finally {
    setBusy(item.id, false)
  }
}

async function decide(decision: 'accept' | 'reject') {
  const p = data.value?.proposal
  if (!p || deciding.value) return
  deciding.value = true
  actionError.value = null
  try {
    await decideReminderProposal(p.id, decision)
    await reload()
  } catch (e) {
    actionError.value = e instanceof Error ? e.message : String(e)
  } finally {
    deciding.value = false
  }
}
</script>

<template>
  <PageLayout title="Żywienie" :meta="meta">
    <template #actions>
      <NutritionSwitch :supplements-label="tabLabel" />
    </template>

    <div v-if="error && !data" :class="{ page: layout !== 'desktop' }">
      <section class="box" role="alert">{{ error }}</section>
    </div>

    <template v-else-if="data">
      <!-- design/screens/DesktopSuplementy.html: checklist hero (12), regularity (8) | proposal + stock (4). -->
      <div v-if="layout === 'desktop'" class="d-grid">
        <section v-if="actionError" class="box s-12 full" role="alert">{{ actionError }}</section>
        <ChecklistBox class="s-12 full" :items="items" :day-type="data.dayType" :busy="busy" hero @toggle="toggle" />
        <RegularityBox
          class="s-8"
          :items="data.regularity.items"
          :from="data.regularity.from"
          :to="data.regularity.to"
          strip
        />
        <div v-if="data.proposal || data.stock.length" class="stack s-4">
          <ReminderProposalBox v-if="data.proposal" :proposal="data.proposal" :busy="deciding" @decide="decide" />
          <StockBox :stock="data.stock" />
        </div>
      </div>

      <!-- design/screens/FoldSuplementy.html: checklist | regularity, stock, proposal. -->
      <div v-else-if="layout === 'fold'" class="page">
        <div class="col">
          <section v-if="actionError" class="box" role="alert">{{ actionError }}</section>
          <ChecklistBox :items="items" :day-type="data.dayType" :busy="busy" @toggle="toggle" />
        </div>
        <div class="col">
          <RegularityBox :items="data.regularity.items" :from="data.regularity.from" :to="data.regularity.to" />
          <StockBox :stock="data.stock" />
          <ReminderProposalBox v-if="data.proposal" :proposal="data.proposal" :busy="deciding" @decide="decide" />
        </div>
      </div>

      <!-- Cover: checklist first, then what needs a decision, then history and stock. -->
      <div v-else class="page">
        <div class="col">
          <section v-if="actionError" class="box" role="alert">{{ actionError }}</section>
          <ChecklistBox :items="items" :day-type="data.dayType" :busy="busy" @toggle="toggle" />
          <ReminderProposalBox v-if="data.proposal" :proposal="data.proposal" :busy="deciding" @decide="decide" />
          <RegularityBox :items="data.regularity.items" :from="data.regularity.from" :to="data.regularity.to" />
          <StockBox :stock="data.stock" />
        </div>
      </div>
    </template>
  </PageLayout>
</template>

<style scoped>
/* The hero checklist keeps the whole row also when the grid halves (chat open, narrow window). */
@container main (max-width: 1100px) {
  .d-grid > .full { grid-column: 1 / -1; }
}
</style>
