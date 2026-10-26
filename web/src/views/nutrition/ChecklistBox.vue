<script setup lang="ts">
import { computed } from 'vue'
import type { DayType } from '@/api/nutrition'
import type { SupplementItem } from '@/api/supplements'
import BoxSection from '@/components/ui/BoxSection.vue'
import { todayTitle } from './nutritionData'
import { counts, itemSub, slotGroups, whenLabel } from './supplementsData'

// Today's checklist, one tap per item. hero = desktop: dark block, time-of-day groups as four columns.
const props = defineProps<{ items: SupplementItem[]; dayType: DayType | null; busy: Set<string>; hero?: boolean }>()
const emit = defineEmits<{ toggle: [item: SupplementItem] }>()

const groups = computed(() => slotGroups(props.items, props.dayType))
const count = computed(() => counts(props.items))
</script>

<template>
  <BoxSection :title="todayTitle(dayType)" :hero="hero" label="Suplementy dziś" class="checklist" :class="{ hero }">
    <template #meta>
      <span v-if="items.length" class="count-total">
        {{ count.taken }}<span class="of"> / {{ count.total }}</span>
      </span>
    </template>
    <p v-if="!items.length" class="empty">Brak suplementów</p>
    <div v-else class="slots">
      <div v-for="g in groups" :key="g.key" class="slot">
        <div class="slot-label">{{ g.label }}</div>
        <label v-for="item in g.items" :key="item.id" class="item" :class="item.status">
          <input
            class="sq"
            type="checkbox"
            :checked="item.status === 'taken'"
            :disabled="busy.has(item.id)"
            @change="emit('toggle', item)"
          />
          <span class="text">
            <span class="name">{{ item.name }}</span>
            <span class="sub">{{ itemSub(item) }}</span>
          </span>
          <span class="when" :class="{ done: item.status === 'taken' }">{{ whenLabel(item) }}</span>
        </label>
      </div>
    </div>
  </BoxSection>
</template>

<style scoped>
.checklist { gap: 0; padding-bottom: 6px; }
.count-total { font-size: 20px; font-weight: 800; font-stretch: 75%; line-height: 1; }
.of { color: var(--ink3); font-weight: 600; }
.empty { padding: 8px 0; font-size: 14px; color: var(--ink2); }
.slot-label { font-size: 12.5px; font-weight: 700; color: var(--ink3); padding: 10px 0 2px; }
.item { display: flex; align-items: center; gap: 12px; min-height: 52px; border-top: 1px solid var(--rule); cursor: pointer; }
.text { flex: 1; min-width: 0; display: flex; flex-direction: column; gap: 1px; }
.name { font-size: 14.5px; font-weight: 600; }
.sub { font-size: 12.5px; color: var(--ink3); }
.when { flex: none; font-size: 12.5px; font-weight: 600; color: var(--ink3); }
.when.done { color: var(--good); }
.item.taken .name { text-decoration: line-through; color: var(--ink3); }

.sq {
  -webkit-appearance: none; appearance: none; flex: none; width: 22px; height: 22px; margin: 0;
  border: 1.5px solid var(--ink); border-radius: 0; background: transparent;
  display: grid; place-content: center; cursor: pointer;
}
.sq:checked { background: var(--ink); }
.sq:checked::after {
  content: ''; width: 10px; height: 5px; border: 2px solid var(--surface); border-top: 0; border-right: 0;
  transform: translateY(-1px) rotate(-45deg);
}
.sq:disabled { opacity: 0.5; cursor: progress; }
.sq:focus-visible { outline: 2px solid var(--ink); outline-offset: 2px; }

/* Desktop hero (design/screens/DesktopSuplementy.html). */
.hero .count-total { font-size: 40px; font-stretch: 72%; }
.hero .of { font-size: 22px; color: var(--hero-ink2); }
.hero .empty, .hero .slot-label, .hero .sub, .hero .when { color: var(--hero-ink2); }
.hero .when.done { color: var(--hero-good); font-weight: 700; }
.hero .item { border-top-color: var(--hero-rule); min-height: 60px; }
.hero .item.taken .name { color: var(--hero-ink2); }
.hero .name { font-size: 15.5px; }
.hero .sub, .hero .when { font-size: 13px; }
.hero .slot-label { font-size: 13px; padding: 4px 0 6px; }
.hero .slots { display: grid; grid-template-columns: repeat(4, minmax(0, 1fr)); column-gap: 24px; padding-top: 6px; }
.hero .sq { width: 24px; height: 24px; border-color: var(--hero-ink); }
.hero .sq:checked { background: var(--hero-ink); }
.hero .sq:checked::after { width: 11px; border-color: var(--hero); }
.hero .sq:focus-visible { outline-color: var(--hero-ink); }
/* Chat panel open or narrow window: two groups per row. */
@container main (max-width: 1100px) {
  .hero .slots { grid-template-columns: repeat(2, minmax(0, 1fr)); }
}
</style>
