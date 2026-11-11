<script setup lang="ts">
// Line icons on a 24 px grid, drawn for the zero-radius style: 1.75 px stroke, square caps, miter joins.
// Decorative by default (aria-hidden); the control that holds the icon carries the label.
export type IconName =
  | 'today'
  | 'training'
  | 'nutrition'
  | 'body'
  | 'plan'
  | 'proposals'
  | 'sync'
  | 'chat'
  | 'history'
  | 'more'
  | 'theme'
  | 'arrow-right'
  | 'arrow-left'
  | 'close'

withDefaults(defineProps<{ name: IconName; size?: number }>(), { size: 22 })

const PATHS: Record<IconName, string> = {
  // calendar sheet; the filled "today" square is drawn separately
  today: 'M4 5h16v15H4zM4 10h16M8 3v4M16 3v4',
  // effort trace
  training: 'M2 12h5l3-7 4 14 3-7h5',
  // fork and knife
  nutrition: 'M7 3v18M4 3v5a3 3 0 0 0 6 0V3M17 21V3c-2 1.5-3 4-3 7v3h3',
  // bathroom scale with a dial
  body: 'M4 4h16v16H4zM8 11a4 4 0 0 1 8 0zM12 11l1.6-2.6',
  // flag on the finish line
  plan: 'M6 21V4h12l-3 4.5 3 4.5H6',
  // inbox tray: changes waiting for a decision
  proposals: 'M3 13l3-8h12l3 8v6H3zM3 13h5l1 3h6l1-3h5',
  sync: 'M3 12a9 9 0 0 1 15.7-6L21 8M21 3v5h-5M21 12a9 9 0 0 1-15.7 6L3 16M3 21v-5h5',
  chat: 'M4 4h16v12H10l-5 4v-4H4z',
  history: 'M12 4a8 8 0 1 1-8 8M4 5v4h4M12 8v4l3 2',
  more: 'M4 6h16M4 12h16M4 18h16',
  // half-filled disc; the filled half is drawn separately
  theme: 'M12 4a8 8 0 1 0 0 16a8 8 0 1 0 0-16z',
  'arrow-right': 'M5 12h14M13 6l6 6-6 6',
  'arrow-left': 'M19 12H5M11 6l-6 6 6 6',
  close: 'M5 5l14 14M19 5L5 19',
}
</script>

<template>
  <svg class="icon" viewBox="0 0 24 24" :width="size" :height="size" aria-hidden="true" focusable="false">
    <path :d="PATHS[name]" />
    <rect v-if="name === 'today'" x="7.5" y="13" width="4" height="4" class="fill" />
    <path v-else-if="name === 'theme'" d="M12 4a8 8 0 0 1 0 16z" class="fill" />
  </svg>
</template>

<style scoped>
.icon { flex: none; fill: none; stroke: currentColor; stroke-width: 1.75; stroke-linecap: square; stroke-linejoin: miter; }
.fill { fill: currentColor; stroke: none; }
</style>
