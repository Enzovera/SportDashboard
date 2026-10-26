<script setup lang="ts">
import { computed, nextTick, ref, watch } from 'vue'
import type { Decision, Proposal } from '@/api/proposals'
import { dayShort } from '@/lib/format'
import { type QueuedDecision, secondsLeft } from './decisionQueue'
import { KIND_LABEL, changedRows, diffHeading, diffValue, proposedMeta } from './proposals'
import ValidatorChecks from './ValidatorChecks.vue'

const props = defineProps<{
  proposal: Proposal
  today: string
  /** Desktop: before/after table with every row; otherwise only changed rows. */
  wide: boolean
  queued?: QueuedDecision
  /** Clock tick for the countdown. */
  now: number
}>()
const comment = defineModel<string>('comment', { default: '' })
const emit = defineEmits<{ decide: [decision: Decision]; undo: [] }>()

const p = computed(() => props.proposal)
const titleId = computed(() => `prop-${p.value.id}`)
const commentId = computed(() => `prop-${p.value.id}-comment`)
const rows = computed(() => changedRows(p.value))
/** Waiting or being sent: the card collapses to title + undo bar. */
const held = computed(() => props.queued && props.queued.status !== 'failed')
const seconds = computed(() => (props.queued ? secondsLeft(props.queued.deadline, props.now) : 0))
const verdict = computed(() => (props.queued?.decision === 'accept' ? 'Akceptacja' : 'Odrzucenie'))

// The clicked button disappears with the collapse: keep keyboard focus on "Cofnij".
const undoBtn = ref<HTMLButtonElement | null>(null)
watch(
  () => props.queued?.status === 'waiting',
  async (waiting) => {
    if (!waiting) return
    await nextTick()
    undoBtn.value?.focus()
  },
)
</script>

<template>
  <article class="box card" :class="{ held }" :aria-labelledby="titleId">
    <div class="card-meta">
      <span class="tag">{{ KIND_LABEL[p.kind] }}</span>
      <span class="meta">{{ proposedMeta(p, today) }}</span>
    </div>
    <h2 :id="titleId" class="title">{{ p.title }}</h2>

    <div v-if="held" class="hold" role="status">
      <span class="hold-text">
        <span :class="queued!.decision === 'accept' ? 'tone-good' : 'verdict-reject'">{{ verdict }}</span>
        <span v-if="queued!.status === 'waiting'" class="secs">{{ seconds }} s</span>
        <span v-else class="secs">wysyłanie</span>
      </span>
      <button v-if="queued!.status === 'waiting'" ref="undoBtn" type="button" class="btn" @click="emit('undo')">Cofnij</button>
    </div>

    <template v-else>
      <p class="prose">{{ p.rationale }}</p>

      <table v-if="wide" class="table diff">
        <colgroup><col /><col class="c-now" /><col class="c-after" /></colgroup>
        <thead>
          <tr>
            <th scope="col" class="diff-head">{{ diffHeading(p) }}</th>
            <th scope="col" class="r">teraz</th>
            <th scope="col" class="r">po</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="r in p.diff.rows" :key="r.label">
            <th scope="row" class="dim">{{ r.label }}</th>
            <template v-if="r.before !== r.after">
              <td class="r"><s v-if="r.before !== null">{{ diffValue(r.before, r.unit, r.unit !== 'kcal') }}</s><span v-else class="same">–</span></td>
              <td class="r">
                <span class="after">{{ diffValue(r.after, r.unit) }}</span>
                <template v-if="r.note"><br /><span class="note">{{ r.note }}</span></template>
              </td>
            </template>
            <template v-else>
              <td class="r same">{{ diffValue(r.before, r.unit) }}</td>
              <td class="r same">{{ diffValue(r.after, r.unit) }}</td>
            </template>
          </tr>
        </tbody>
        <tfoot v-if="p.diff.total">
          <tr>
            <th scope="row" class="strong">{{ p.diff.total.label }}</th>
            <td class="r"><s>{{ diffValue(p.diff.total.before, p.diff.total.unit) }}</s></td>
            <td class="r after">{{ diffValue(p.diff.total.after, p.diff.total.unit) }}</td>
          </tr>
        </tfoot>
      </table>

      <div v-else class="rows">
        <div v-for="r in rows" :key="r.label" class="row">
          <span class="dim">{{ r.label }}</span>
          <span class="val">
            <template v-if="r.before !== null"><s>{{ diffValue(r.before, r.unit, r.unit !== 'kcal') }}</s> → </template>
            <strong>{{ diffValue(r.after, r.unit) }}<template v-if="r.note && r.before === null">, {{ r.note }}</template></strong>
            <span v-if="r.note && r.before !== null" class="note"> ({{ r.note }})</span>
          </span>
        </div>
      </div>

      <ValidatorChecks :checks="p.checks" />

      <div v-if="!wide && p.effectiveFrom" class="from">Od: <strong>{{ dayShort(p.effectiveFrom) }}</strong></div>

      <div class="decide">
        <p v-if="queued?.status === 'failed'" class="tone-warn failed" role="alert">Nie wysłano: {{ queued.error }}</p>
        <label :for="commentId">Komentarz dla trenera (opcjonalnie)</label>
        <input :id="commentId" v-model="comment" class="field" type="text" maxlength="500" autocomplete="off" />
        <div class="decide-actions">
          <button type="button" class="btn" @click="emit('decide', 'reject')">Odrzuć</button>
          <button type="button" class="btn btn--primary" @click="emit('decide', 'accept')">Akceptuj</button>
        </div>
      </div>
    </template>
  </article>
</template>

<style scoped>
.card { padding: 14px; gap: 12px; }
.card-meta { display: flex; flex-wrap: wrap; align-items: center; gap: 6px 10px; }
.tag {
  display: inline-flex; align-items: center; min-height: 22px; padding: 0 7px;
  border: 1px solid var(--ink); color: var(--ink); font-size: 12px; font-weight: 700;
}
.title { font-size: 18px; font-weight: 800; line-height: 1.3; }
.prose { font-size: 14.5px; line-height: 1.5; color: var(--ink2); }

.rows { display: flex; flex-direction: column; }
.row {
  display: flex; justify-content: space-between; align-items: baseline; gap: 8px;
  padding: 8px 0; border-top: 1px solid var(--rule); font-size: 14px;
}
.row:last-child { border-bottom: 1px solid var(--rule); }
.row .dim { color: var(--ink2); }
.row .val { text-align: right; }
.row s { color: var(--ink3); }
.row strong { font-weight: 800; }
.row .note { font-size: 12.5px; color: var(--ink2); }

.diff { table-layout: fixed; }
.diff .c-now { width: 72px; }
.diff .c-after { width: 112px; }
.diff th { text-align: left; }
.diff th.r { text-align: right; }
.diff thead th { font-size: 12.5px; color: var(--ink3); }
.diff thead th.diff-head { color: var(--ink); font-weight: 700; font-size: 13.5px; }
.diff tfoot tr { border-top: 1px solid var(--rule); border-bottom: 1px solid var(--rule); }
.diff s { color: var(--ink3); }
.diff .after { font-weight: 800; }
.diff .same { color: var(--ink3); }
.diff .note { font-size: 12.5px; font-weight: 500; color: var(--ink2); }

.from { font-size: 13px; color: var(--ink2); }
.from strong { font-weight: 800; color: var(--ink); }

.decide { display: flex; flex-direction: column; gap: 8px; }
.decide label { font-size: 12.5px; font-weight: 600; color: var(--ink2); }
.field {
  min-height: 44px; padding: 0 12px; border: 1px solid var(--border-strong); border-radius: 0;
  font: inherit; font-size: 14px; color: var(--ink); background: var(--surface);
}
.field:focus-visible { outline: 2px solid var(--ink); outline-offset: 0; }
.decide-actions { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 8px; }
.decide-actions .btn { min-height: 48px; justify-content: center; font-size: 14.5px; }
.decide-actions .btn:not(.btn--primary) { font-weight: 600; }
.decide-actions .btn--primary { font-weight: 800; }
.failed { font-size: 13.5px; }

.hold {
  display: flex; align-items: center; justify-content: space-between; gap: 12px;
  padding-top: 10px; border-top: 1px solid var(--rule);
}
.hold-text { display: flex; align-items: baseline; gap: 10px; font-size: 14.5px; font-weight: 700; }
.verdict-reject { color: var(--ink2); }
.secs { font-weight: 800; font-stretch: 75%; font-size: 18px; min-width: 3ch; }
.hold .btn { min-height: 48px; padding: 0 20px; justify-content: center; }

@media (min-width: 1024px) {
  .card { padding: 16px 18px; gap: 14px; }
  .tag { min-height: 24px; padding: 0 8px; font-size: 12.5px; }
  .title { font-size: 20px; }
  .prose { font-size: 15px; }
  .decide label { font-size: 13px; }
  .field { font-size: 14.5px; }
  .decide-actions .btn { font-size: 15px; }
}
</style>
