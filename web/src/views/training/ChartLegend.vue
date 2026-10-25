<script lang="ts">
// Legend under a chart: small swatch + Polish label. Colors are CSS variables, e.g. "var(--chart-mid)".
export interface LegendItem {
  /** square = bar, line = solid line/tick, dash = dashed line, outline = dashed outline ("dziś") */
  kind: 'square' | 'line' | 'dash' | 'outline'
  color: string
  label: string
  /** Line thickness in px for `line`. */
  weight?: number
}
</script>

<script setup lang="ts">
defineProps<{ items: LegendItem[] }>()
</script>

<template>
  <div class="legend" aria-hidden="true">
    <span v-for="item in items" :key="item.label" class="item">
      <span class="swatch" :class="item.kind" :style="{ '--c': item.color, '--w': `${item.weight ?? 2}px` }" />
      {{ item.label }}
    </span>
  </div>
</template>

<style scoped>
.legend { display: flex; flex-wrap: wrap; gap: 6px 14px; font-size: 12.5px; color: var(--ink2); }
.item { display: inline-flex; align-items: center; gap: 6px; }
.swatch { flex: none; }
.square { width: 10px; height: 10px; background: var(--c); }
.line { width: 14px; height: var(--w); background: var(--c); }
.dash { width: 14px; height: 0; border-top: 2px dashed var(--c); }
.outline { width: 10px; height: 10px; border: 1.5px dashed var(--c); }
</style>
