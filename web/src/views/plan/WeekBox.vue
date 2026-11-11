<script setup lang="ts">
import { computed } from 'vue'
import type { PlanWeek } from '@/api/plan'
import BoxSection from '@/components/ui/BoxSection.vue'
import Icon from '@/components/ui/Icon.vue'
import { num, weekday } from '@/lib/format'
import { DAY_TYPE_LABEL, dayLines, dayNumber, decideCount, doneCell, isCurrentWeek, weekMeta, weekTitle } from './planData'

// One week of plan_session as a table. Current week: what was done; upcoming week: daily nutrition targets.
// Days touched by pending proposals are "do decyzji" (row-decide) and link to /propozycje.
const props = defineProps<{ week: PlanWeek; today: string; wide?: boolean }>()

const current = computed(() => isCurrentWeek(props.week, props.today))
const decide = computed(() => decideCount(props.week))
const rows = computed(() =>
  props.week.days.map((d) => ({
    day: d,
    lines: dayLines(d, !props.wide),
    done: doneCell(d, props.today),
    today: d.date === props.today,
    decide: d.proposals.length > 0,
  })),
)
const km = (v: number | null) => (v === null ? '–' : num(v))
const kcal = (t: PlanWeek['days'][number]['target']) => (t ? num(t.kcal) : '–')
</script>

<template>
  <BoxSection
    :title="weekTitle(week, today)"
    :label="current ? 'Ten tydzień' : 'Następny tydzień'"
    class="week"
    :class="{ wide, current }"
  >
    <template #meta>
      <span class="meta">
        {{ weekMeta(week) }}<template v-if="decide"> · <span class="decide-label">{{ decide }} do decyzji</span></template>
      </span>
    </template>
    <table class="table wk">
      <colgroup>
        <col class="c-day" />
        <col />
        <template v-if="wide">
          <col class="c-km" />
          <template v-if="current"><col class="c-done" /><col class="c-last" /></template>
          <template v-else>
            <col class="c-type" /><col class="c-kcal" /><col class="c-g" /><col class="c-g" /><col class="c-g-last" />
          </template>
        </template>
        <col v-else :class="current ? 'c-done-last' : 'c-last'" />
      </colgroup>
      <thead>
        <tr>
          <th scope="col"><span class="sr-only">Dzień</span></th>
          <th scope="col"><span :class="{ 'sr-only': !wide }">sesja</span></th>
          <template v-if="wide">
            <th scope="col" class="r">km</th>
            <template v-if="current">
              <th scope="col" class="r">wykonane</th>
              <th scope="col" class="r">cel kcal</th>
            </template>
            <template v-else>
              <th scope="col" class="type">typ dnia</th>
              <th scope="col" class="r">kcal</th>
              <th scope="col" class="r">B</th>
              <th scope="col" class="r">W</th>
              <th scope="col" class="r">T</th>
            </template>
          </template>
          <th v-else scope="col" class="r">{{ current ? 'wykonane' : 'cel kcal' }}</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="r in rows" :key="r.day.date" :class="{ today: r.today, 'row-decide': r.decide }">
          <th scope="row" class="day">
            <template v-if="wide">{{ weekday(r.day.date) }} {{ dayNumber(r.day.date) }}</template>
            <template v-else>{{ weekday(r.day.date) }}<br />{{ dayNumber(r.day.date) }}</template>
          </th>
          <td class="session">
            <div :class="{ inline: wide }">
              <span class="s-name" :class="{ rest: r.lines.rest }">{{ r.lines.title }}</span>
              <span v-if="r.lines.detail" class="s-sub">{{ r.lines.detail }}</span>
            </div>
            <RouterLink
              v-for="p in r.day.proposals"
              :key="p.id"
              class="prop-link"
              :to="{ path: '/propozycje', hash: `#${p.id}` }"
            ><span>propozycja: {{ p.summary }} <Icon name="arrow-right" :size="14" class="go" /></span></RouterLink>
          </td>
          <template v-if="wide">
            <td class="r km" :class="{ muted: r.day.runKm === null }">{{ km(r.day.runKm) }}</td>
            <template v-if="current">
              <td class="r" :class="`done-${r.done.tone}`">{{ r.done.text }}</td>
              <td class="r dim">{{ kcal(r.day.target) }}</td>
            </template>
            <template v-else>
              <td class="dim type">{{ DAY_TYPE_LABEL[r.day.dayType] }}</td>
              <td class="r">{{ kcal(r.day.target) }}</td>
              <td class="r dim">{{ r.day.target ? num(r.day.target.proteinG) : '–' }}</td>
              <td class="r dim">{{ r.day.target ? num(r.day.target.carbsG) : '–' }}</td>
              <td class="r dim">{{ r.day.target ? num(r.day.target.fatG) : '–' }}</td>
            </template>
          </template>
          <td v-else-if="current" class="r" :class="`done-${r.done.tone}`">{{ r.done.text }}</td>
          <td v-else class="r dim">{{ kcal(r.day.target) }}</td>
        </tr>
      </tbody>
      <tfoot v-if="wide && !current">
        <tr>
          <th scope="row" colspan="2">Tydzień</th>
          <td class="r">{{ num(week.plannedKm) }}</td>
          <td colspan="5" class="dim type unit">makro w g</td>
        </tr>
      </tfoot>
    </table>
  </BoxSection>
</template>

<style scoped>
.week { --pad: 14px; gap: 4px; padding-bottom: 6px; }
.week.wide { --pad: 18px; padding-bottom: 8px; }
.decide-label { font-weight: 700; color: var(--signal); white-space: nowrap; }
.week :deep(.box-head) { flex-wrap: wrap; row-gap: 2px; }
.week :deep(.box-head h2) { white-space: nowrap; }

/* Rows run edge to edge so the "do decyzji" tint spans the box. */
.wk { margin: 0 calc(-1 * var(--pad)); width: calc(100% + 2 * var(--pad)); table-layout: fixed; }
.wk th:first-child, .wk td:first-child { padding-left: var(--pad); }
.wk th:last-child, .wk td:last-child { padding-right: var(--pad); }
.wk td, .wk th { vertical-align: top; }
.wk thead th { text-align: left; }
.wk thead th.r { text-align: right; }
.c-day { width: calc(40px + var(--pad)); }
.c-done-last { width: calc(112px + var(--pad)); }
.c-last { width: calc(64px + var(--pad)); }
.wide .c-day { width: calc(58px + var(--pad)); }
.wide .c-last { width: calc(74px + var(--pad)); }
.c-done { width: 128px; }
.c-km { width: 44px; }
.c-type { width: 158px; }
.c-kcal { width: 58px; }
.c-g { width: 40px; }
.c-g-last { width: calc(50px + var(--pad)); }

.day { text-align: left; white-space: nowrap; font-size: 12.5px; font-weight: 700; color: var(--ink3); line-height: 1.3; }
.wide .day { font-size: 13px; line-height: 1.35; }
tbody td, tbody th { padding-top: 7px; padding-bottom: 7px; }
.s-name { display: block; font-size: 14.5px; font-weight: 700; }
.s-name.rest { font-weight: 500; color: var(--ink2); }
.s-sub { display: block; font-size: 12.5px; color: var(--ink2); }
.inline .s-name, .inline .s-sub { display: inline; font-size: inherit; }
.inline .s-sub { margin-left: 4px; }
.type { padding-left: 16px; text-align: left; }
.done-signal { color: var(--signal); font-weight: 700; }
.done-warn { color: var(--warn); font-weight: 600; }
.done-dim { color: var(--ink2); }
.done-muted { color: var(--ink3); }

.wk tr.today .day, .wk tr.today .km { color: var(--signal); font-weight: 800; }
.wk tr.row-decide { background: var(--signal-tint); }
.wk tr.row-decide > :first-child { box-shadow: inset 3px 0 0 var(--signal); }
.wk tr.row-decide .day { color: var(--signal-tint-ink); }

.prop-link {
  display: flex; align-items: center; gap: 8px; min-height: 44px; margin: -6px 0 -8px;
  font-size: 12.5px; font-weight: 700; text-decoration: none; color: var(--ink);
}
.prop-link::before { content: ''; width: 7px; height: 7px; background: var(--signal); flex: none; }
.prop-link .go { vertical-align: -2px; }
.prop-link:focus-visible { outline: 2px solid var(--ink); outline-offset: -2px; }
/* One line: the cells next to it hold a single line at the top, so the link may run under them. */
.wide .prop-link { display: inline-flex; font-size: 13.5px; margin: -8px 0 -12px; white-space: nowrap; }
.wide .prop-link::before { width: 8px; height: 8px; }

/* Narrow desktop box (chat docked, small window): the detail drops under the session name instead of
   breaking the name in two. Thresholds (content width of the box) = fixed columns of each table + room
   for the longest session. */
.week { container: week / inline-size; }
@container week (max-width: 480px) {
  .current .inline .s-sub { display: block; margin-left: 0; font-size: 12.5px; }
}
@container week (max-width: 640px) {
  .week:not(.current) .inline .s-sub { display: block; margin-left: 0; font-size: 12.5px; }
}

.wk tfoot tr { border-top: 1px solid var(--rule); }
.wk tfoot th, .wk tfoot td { font-weight: 700; padding-top: 10px; text-align: left; }
.wk tfoot td.r { text-align: right; }
.wk tfoot td.unit { font-weight: 500; }
</style>
