<script setup lang="ts">
import BoxSection from '@/components/ui/BoxSection.vue'
import type { ProposalDecision } from '@/api/proposals'
import { dayMonth, localDate } from '@/lib/format'
import { DECISION_LABEL, ORIGIN_LABEL } from './proposals'

// Compact (fold/cover): date | change + source + comment | decision. Wide (desktop): separate columns.
defineProps<{ history: ProposalDecision[]; wide: boolean }>()

const date = (d: ProposalDecision) => dayMonth(localDate(d.decidedAt))
</script>

<template>
  <BoxSection title="Historia decyzji" class="history-box">
    <table v-if="wide" class="table history">
      <colgroup>
        <col class="c-date" /><col /><col class="c-src" /><col class="c-comment" /><col class="c-dec" />
      </colgroup>
      <thead>
        <tr>
          <th scope="col"><span class="sr-only">Data</span></th>
          <th scope="col">zmiana</th>
          <th scope="col">źródło</th>
          <th scope="col">Twój komentarz</th>
          <th scope="col" class="r">decyzja</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="d in history" :key="d.id">
          <th scope="row" class="strong">{{ date(d) }}</th>
          <td>{{ d.title }}</td>
          <td class="muted">{{ ORIGIN_LABEL[d.origin] }}</td>
          <td :class="d.comment ? 'dim' : 'muted'">{{ d.comment ? `„${d.comment}”` : '–' }}</td>
          <td class="r" :class="d.decision === 'accepted' ? 'tone-good' : 'rejected'">{{ DECISION_LABEL[d.decision] }}</td>
        </tr>
      </tbody>
    </table>

    <table v-else class="table history compact">
      <colgroup><col class="c-date" /><col /><col class="c-dec" /></colgroup>
      <thead class="sr-only">
        <tr><th scope="col">Data</th><th scope="col">Zmiana</th><th scope="col">Decyzja</th></tr>
      </thead>
      <tbody>
        <tr v-for="d in history" :key="d.id">
          <th scope="row" class="date">{{ date(d) }}</th>
          <td>
            <div class="change">{{ d.title }}</div>
            <div class="sub">{{ ORIGIN_LABEL[d.origin] }}</div>
            <div v-if="d.comment" class="sub">Twój komentarz: „{{ d.comment }}”</div>
          </td>
          <td class="r">
            <span class="dec" :class="d.decision === 'accepted' ? 'tone-good' : 'rejected'">{{ DECISION_LABEL[d.decision] }}</span>
          </td>
        </tr>
      </tbody>
    </table>
  </BoxSection>
</template>

<style scoped>
.history-box { padding-bottom: 6px; gap: 4px; }
.history { table-layout: fixed; }
.history th, .history td { vertical-align: top; text-align: left; }
.history .r { text-align: right; }
.history .rejected { color: var(--ink3); font-weight: 600; }

.compact .c-date { width: 44px; }
.compact .c-dec { width: 74px; }
.compact th, .compact td { padding: 8px 0; }
.compact tbody tr:first-child { border-top: 1px solid var(--rule); }
.compact tbody tr:last-child { border-bottom: 0; }
.compact .date { font-weight: 700; font-size: 13px; }
.compact .change { font-size: 14px; }
.compact .sub { font-size: 12.5px; color: var(--ink3); }
.compact .dec { font-size: 12.5px; font-weight: 700; }
.compact .dec.rejected { font-weight: 700; }

.history:not(.compact) .c-date { width: 64px; }
.history:not(.compact) .c-src { width: 140px; }
.history:not(.compact) .c-comment { width: 220px; }
.history:not(.compact) .c-dec { width: 90px; }
@media (min-width: 1024px) {
  .history-box { padding-bottom: 16px; gap: 10px; }
}
</style>
