<script setup lang="ts">
import { useLayout } from '@/composables/useLayout'
import ThemeToggle from './ThemeToggle.vue'

// Frame of every view.
// fold/cover: fixed 56 px header with title and all actions, scrolling main; the bottom nav lives in App.vue.
// desktop: the app header (DesktopHeader) holds the nav; the title becomes a page head above the grid.
// Slots: `global` = app-level actions (sync status, Trener) -> app header on desktop;
//        `actions` = actions of this screen (e.g. range switch) -> page head on desktop.
defineProps<{ title: string; meta?: string | null }>()
const layout = useLayout()
</script>

<template>
  <div class="frame">
    <header v-if="layout !== 'desktop'" class="header">
      <div class="title">
        <h1>{{ title }}</h1>
        <span v-if="meta" class="header-meta">{{ meta }}</span>
      </div>
      <div class="actions"><slot name="global" /><slot name="actions" /><ThemeToggle /></div>
    </header>
    <main class="main">
      <div v-if="layout === 'desktop'" class="d-content">
        <Teleport to="#d-actions" defer><slot name="global" /></Teleport>
        <div class="d-pagehead">
          <div class="d-title">
            <h1>{{ title }}</h1>
            <span v-if="meta" class="d-meta">{{ meta }}</span>
          </div>
          <div class="actions"><slot name="actions" /></div>
        </div>
        <slot />
      </div>
      <slot v-else />
    </main>
  </div>
</template>

<style scoped>
.frame { flex: 1; min-height: 0; display: flex; flex-direction: column; }
.header {
  flex: none; min-height: var(--header-h); padding: 0 16px;
  display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); column-gap: 24px; align-items: center;
  background: var(--surface); border-bottom: 1px solid var(--border);
}
.title { display: flex; align-items: baseline; gap: 10px; min-width: 0; }
h1 { font-size: 24px; font-weight: 800; font-stretch: 78%; line-height: 1; }
.header-meta { font-size: 13px; color: var(--ink3); white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.actions { display: flex; justify-content: flex-end; align-items: center; gap: 16px; min-width: 0; }
.main { flex: 1; min-height: 0; overflow-y: auto; }

.d-content { max-width: 1392px; margin: 0 auto; padding: 24px 32px 40px; }
.d-pagehead { display: flex; align-items: flex-end; justify-content: space-between; gap: 24px; margin-bottom: 20px; }
.d-title { display: flex; align-items: baseline; gap: 14px; min-width: 0; }
.d-title h1 { font-size: 32px; }
.d-meta { font-size: 14px; color: var(--ink3); }

@media (max-width: 599px) {
  .header { padding: 8px 12px; grid-template-columns: minmax(0, 1fr) auto; column-gap: 12px; }
  .title { flex-direction: column; align-items: flex-start; gap: 4px; }
}
</style>
