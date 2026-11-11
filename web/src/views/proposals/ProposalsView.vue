<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, reactive, ref, shallowRef, watch } from 'vue'
import PageLayout from '@/components/ui/PageLayout.vue'
import BoxSection from '@/components/ui/BoxSection.vue'
import { type Decision, type ProposalDecision, decideProposal, getProposals } from '@/api/proposals'
import { useApi } from '@/composables/useApi'
import { useLayout } from '@/composables/useLayout'
import { dayShort } from '@/lib/format'
import { type QueuedDecision, createDecisionQueue } from './decisionQueue'
import { columns, mergeHistory, pendingText } from './proposals'
import BlockedBox from './BlockedBox.vue'
import HistoryBox from './HistoryBox.vue'
import ProposalCard from './ProposalCard.vue'
import WeekDiffBox from './WeekDiffBox.vue'

const layout = useLayout()
const { data, error, reload } = useApi(getProposals)

// ---- decisions with an undo window ----
const queued = shallowRef<QueuedDecision[]>([])
/** Sent and confirmed, until the reload no longer lists them as pending. */
const sent = ref(new Set<string>())
const localHistory = shallowRef<ProposalDecision[]>([])
const comments = reactive<Record<string, string>>({})

const queue = createDecisionQueue<ProposalDecision>({
  send: decideProposal,
  onChange: (items) => (queued.value = items),
  onSent: (item, result) => {
    sent.value = new Set(sent.value).add(item.id)
    localHistory.value = [result, ...localHistory.value]
    delete comments[item.id]
    void reload()
  },
})

function decide(id: string, decision: Decision) {
  queue.schedule(id, decision, comments[id] ?? null)
}

// Countdown tick only while something waits.
const now = ref(Date.now())
let tick: ReturnType<typeof setInterval> | null = null
watch(
  () => queued.value.some((q) => q.status === 'waiting'),
  (waiting) => {
    if (waiting && !tick) {
      now.value = Date.now()
      tick = setInterval(() => (now.value = Date.now()), 250)
    } else if (!waiting && tick) {
      clearInterval(tick)
      tick = null
    }
  },
)

// Leaving the view or closing the page sends what was decided (fetch keepalive).
const flush = () => void queue.flushAll()
onMounted(() => window.addEventListener('pagehide', flush))
onBeforeUnmount(() => {
  window.removeEventListener('pagehide', flush)
  if (tick) clearInterval(tick)
  flush()
})

// ---- derived ----
const pending = computed(() => (data.value?.pending ?? []).filter((p) => !sent.value.has(p.id)))
const pendingIds = computed(() => new Set(pending.value.map((p) => p.id)))
const waitingCount = computed(() => pending.value.filter((p) => !queued.value.some((q) => q.id === p.id && q.status !== 'failed')).length)
const history = computed(() => mergeHistory(data.value?.history ?? [], localHistory.value))
const blocked = computed(() => data.value?.blocked ?? [])
const week = computed(() => (data.value?.week && pending.value.length ? data.value.week : null))
const queuedFor = (id: string) => queued.value.find((q) => q.id === id)

const status = computed(() => pendingText(waitingCount.value))
// Cover: the pending count joins the meta line (the header has room for icons only).
const meta = computed(() => {
  if (!data.value) return null
  const day = dayShort(data.value.date)
  return layout.value === 'cover' && status.value ? `${day} · ${status.value}` : day
})
const cols = computed(() =>
  layout.value === 'desktop'
    ? []
    : columns(layout.value, pending.value.length, blocked.value.length > 0, history.value.length > 0),
)
const side = computed(() => week.value !== null || blocked.value.length > 0)
</script>

<template>
  <PageLayout title="Propozycje" :meta="meta">
    <template v-if="layout !== 'cover' && status" #actions>
      <span role="status" class="pending-status">{{ status }}</span>
    </template>

    <template v-if="data">
      <div v-if="layout === 'desktop'" class="d-grid">
        <div class="stack" :class="side ? 's-8' : 's-12'">
          <div class="pair">
            <ProposalCard
              v-for="p in pending"
              :key="p.id"
              v-model:comment="comments[p.id]"
              :proposal="p"
              :today="data.date"
              :wide="true"
              :queued="queuedFor(p.id)"
              :now="now"
              @decide="decide(p.id, $event)"
              @undo="queue.undo(p.id)"
            />
            <BoxSection v-if="!pending.length" title="Brak propozycji" />
          </div>
          <HistoryBox v-if="history.length" :history="history" :wide="true" />
        </div>
        <div v-if="side" class="stack s-4">
          <WeekDiffBox v-if="week" :week="week" :pending-ids="pendingIds" />
          <BlockedBox v-if="blocked.length" :blocked="blocked" />
        </div>
      </div>

      <div v-else class="page">
        <div v-for="(col, i) in cols" :key="i" class="col">
          <template v-for="item in col" :key="item.kind === 'card' ? pending[item.index]!.id : item.kind">
            <ProposalCard
              v-if="item.kind === 'card'"
              v-model:comment="comments[pending[item.index]!.id]"
              :proposal="pending[item.index]!"
              :today="data.date"
              :wide="false"
              :queued="queuedFor(pending[item.index]!.id)"
              :now="now"
              @decide="decide(pending[item.index]!.id, $event)"
              @undo="queue.undo(pending[item.index]!.id)"
            />
            <BoxSection v-else-if="item.kind === 'empty'" title="Brak propozycji" />
            <BlockedBox v-else-if="item.kind === 'blocked'" :blocked="blocked" />
            <HistoryBox v-else :history="history" :wide="false" />
          </template>
        </div>
      </div>
    </template>
    <div v-else-if="error" :class="{ page: layout !== 'desktop' }">
      <section class="box error" role="alert">{{ error }}</section>
    </div>
  </PageLayout>
</template>

<style scoped>
.pending-status { font-size: 14px; font-weight: 700; white-space: nowrap; }
.pair { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); column-gap: 24px; row-gap: 16px; align-items: start; }
.s-12 .pair { grid-template-columns: repeat(3, minmax(0, 1fr)); }
.error { grid-column: 1 / -1; font-size: 14px; }
@media (min-width: 1024px) {
  .pending-status { font-size: 15px; }
}
@container main (max-width: 1100px) {
  .pair { grid-template-columns: minmax(0, 1fr); }
}
</style>
