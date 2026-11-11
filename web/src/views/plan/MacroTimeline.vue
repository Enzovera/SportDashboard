<script setup lang="ts">
import { computed } from 'vue'
import type { PlanBlock } from '@/api/plan'
import BoxSection from '@/components/ui/BoxSection.vue'
import { fullDate, timeline } from './planData'

// Macrocycle to the race. horizontal (desktop): blocks side by side, one segment per week;
// vertical (fold/cover): blocks stacked, height grows with the number of weeks.
// Current block in ink, past weeks mid grey, today's week red.
const props = defineProps<{ blocks: PlanBlock[]; today: string; race: string; raceDate: string; horizontal?: boolean }>()

const tl = computed(() => timeline(props.blocks, props.today))
const title = computed(() => {
  const n = props.blocks.length
  const words = ['', 'Jeden blok', 'Dwa bloki', 'Trzy bloki', 'Cztery bloki', 'Pięć bloków', 'Sześć bloków']
  return `${words[n] ?? `${n} bloków`} do startu`
})
</script>

<template>
  <BoxSection :title="title" :meta="horizontal ? tl.meta : null" label="Oś makrocyklu" class="tl-box">
    <div class="tl" :class="horizontal ? 'h' : 'v'">
      <div class="now" :class="{ start: tl.beforeStart }"><span class="dot" aria-hidden="true" />{{ tl.now }}</div>
      <ol class="blocks">
        <li
          v-for="b in tl.blocks"
          :key="b.block.from"
          class="block"
          :class="{ current: b.current }"
          :style="horizontal ? { flex: `${b.block.weeks} 1 0` } : { minHeight: `${Math.max(44, b.block.weeks * 14)}px` }"
          :aria-current="b.current ? 'step' : undefined"
        >
          <div class="bar" aria-hidden="true">
            <span v-for="(s, i) in b.segments" :key="i" :class="s" />
          </div>
          <div class="text">
            <div class="name">{{ b.label }}<span v-if="horizontal" class="weeks">{{ b.block.weeks }} tyg.</span></div>
            <div class="dates">{{ horizontal ? b.range : b.dates }}</div>
            <div v-if="b.block.summary" class="desc">{{ b.block.summary }}</div>
          </div>
        </li>
        <li class="end">
          <div class="mark" aria-hidden="true" />
          <div class="text">
            <div class="name">{{ race }}</div>
            <div v-if="horizontal" class="dates">{{ fullDate(raceDate) }}</div>
          </div>
        </li>
      </ol>
    </div>
  </BoxSection>
</template>

<style scoped>
.tl-box { gap: 10px; }
.tl { display: flex; flex-direction: column; gap: 8px; position: relative; }
.now { display: flex; align-items: center; gap: 8px; font-size: 12.5px; font-weight: 700; color: var(--signal); }
.dot { width: 8px; height: 8px; background: var(--signal); flex: none; }
.blocks { list-style: none; margin: 0; padding: 0; display: flex; }
.name { font-size: 14.5px; font-weight: 800; }
.dates { font-size: 12.5px; color: var(--ink3); }
.desc { font-size: 13px; line-height: 1.45; color: var(--ink2); margin-top: 2px; }
.bar { display: flex; gap: 2px; }
.bar > span { flex: 1; background: var(--chart-soft); }
.bar > .past { background: var(--chart-mid); }
.bar > .block { background: var(--ink); }
.bar > .now { background: var(--signal); }
.mark::before { content: ''; display: block; width: 10px; height: 10px; background: var(--ink); transform: rotate(45deg); }

/* vertical: Fold / cover */
.v .blocks { flex-direction: column; gap: 4px; }
.v .block { display: flex; gap: 12px; }
.v .bar { flex: 0 0 8px; flex-direction: column; }
.v .text { display: flex; flex-direction: column; gap: 1px; padding: 2px 0; }
.v .end { display: flex; align-items: center; gap: 8px; padding-top: 4px; }
.v .end .name { font-size: 13.5px; }

/* horizontal: desktop */
.h { margin-top: auto; gap: 6px; }
.h .now { height: 22px; padding-left: 10px; font-size: 13.5px; }
.h .now .dot { display: none; }
.h .now:not(.start) { padding-left: 0; }
.h .now:not(.start) .dot { display: block; }
.h .now.start::before { content: ''; position: absolute; left: 0; top: 0; width: 2px; height: 64px; background: var(--signal); }
.h .blocks { gap: 8px; padding-left: 6px; }
.h .block { min-width: 0; display: flex; flex-direction: column; gap: 10px; }
.h .bar { height: 36px; }
.h .name { font-size: 15.5px; }
.h .name { white-space: nowrap; }
.h .weeks { margin-left: 8px; font-size: 13px; font-weight: 500; color: var(--ink3); }
.h .dates { font-size: 13px; white-space: nowrap; }
.h .desc { font-size: 13.5px; margin-top: 4px; }
.h .end { flex: none; width: 76px; display: flex; flex-direction: column; gap: 10px; }
.h .end .mark { height: 36px; display: flex; align-items: center; padding-left: 4px; }
.h .end .mark::before { width: 14px; height: 14px; }
</style>
