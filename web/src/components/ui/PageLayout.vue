<script setup lang="ts">
import type { RouteLocationRaw } from 'vue-router'
import { useLayout } from '@/composables/useLayout'
import TrainerButton from '@/components/chat/TrainerButton.vue'
import Icon from './Icon.vue'
import ThemeToggle from './ThemeToggle.vue'

// Frame of every view.
// cover:   56 px bar (title over meta, icon actions) and, when the view has controls, a tab strip below it
// fold:    one 56 px bar: title and meta, then the view controls and the icon actions
// desktop: the app header (DesktopHeader) holds the nav; the title becomes a page head above the grid
// Slots: `status`  = sync problems of this view's sources
//        `global`  = app-level actions next to Trener (sync) -> app header on desktop
//        `actions` = controls of this screen (range switch, sub-view switch) -> page head on desktop
defineProps<{
  title: string
  meta?: string | null
  /** Cover only: a back arrow before the title (full-screen chat). */
  back?: { to: RouteLocationRaw; label: string } | null
  /** No Trener button: the view is the chat itself. */
  plain?: boolean
}>()
const layout = useLayout()
</script>

<template>
  <div class="frame">
    <header v-if="layout !== 'desktop'" class="header" :class="layout">
      <div class="bar">
        <RouterLink v-if="back" :to="back.to" class="icon-btn back" :aria-label="back.label" :title="back.label">
          <Icon name="arrow-left" />
        </RouterLink>
        <div class="title">
          <h1>{{ title }}</h1>
          <div v-if="meta || (layout === 'cover' && $slots.status)" class="header-meta">
            <slot v-if="layout === 'cover'" name="status" />
            <span v-if="meta" class="meta-text">{{ meta }}</span>
          </div>
        </div>
        <div class="actions">
          <template v-if="layout === 'fold'"><slot name="status" /><slot name="actions" /></template>
          <div class="icons"><slot name="global" /><TrainerButton v-if="!plain" /><ThemeToggle /></div>
        </div>
      </div>
      <div v-if="layout === 'cover' && $slots.actions" class="toolbar"><slot name="actions" /></div>
    </header>
    <main class="main">
      <div v-if="layout === 'desktop'" class="d-content">
        <Teleport to="#d-actions" defer><slot name="status" /><slot name="global" /><TrainerButton v-if="!plain" /></Teleport>
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
  flex: none; padding-top: env(safe-area-inset-top);
  background: var(--surface); border-bottom: 1px solid var(--border);
}
.bar { min-height: var(--header-h); padding: 0 6px 0 16px; display: flex; align-items: center; gap: 12px; }
.title { flex: 1; min-width: 0; display: flex; align-items: baseline; gap: 10px; }
h1 { font-size: 24px; font-weight: 800; font-stretch: 78%; line-height: 1; white-space: nowrap; }
.header-meta { min-width: 0; display: flex; align-items: baseline; gap: 8px; font-size: 13px; color: var(--ink3); }
.meta-text { min-width: 0; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.actions { flex: none; display: flex; align-items: center; gap: 12px; min-width: 0; }
.icons { display: flex; align-items: center; }
.main { flex: 1; min-height: 0; overflow-y: auto; overscroll-behavior-y: contain; }

/* cover: title over meta on the left, icons on the right; view controls move to the strip below */
.cover .bar { padding: 0 2px 0 12px; gap: 4px; }
.cover .title { flex-direction: column; align-items: stretch; gap: 5px; }
.cover .header-meta { font-size: 12.5px; line-height: 1; }
.cover .back { margin-left: -10px; }
.toolbar { display: flex; align-items: stretch; min-height: 44px; padding: 0 12px; border-top: 1px solid var(--rule); }
.toolbar :deep(.segmented) { flex: 1; display: grid; grid-auto-flow: column; grid-auto-columns: minmax(0, 1fr); border: 0; }
.toolbar :deep(.segmented > *) {
  justify-content: center; padding: 0 4px; border: 0; background: transparent;
  font-size: 14px; color: var(--ink3);
}
.toolbar :deep(.segmented > [aria-pressed='true']),
.toolbar :deep(.segmented > [aria-current]) {
  background: transparent; color: var(--ink); font-weight: 800; box-shadow: inset 0 -3px 0 var(--ink);
}

.d-content { max-width: 1392px; margin: 0 auto; padding: 24px 32px 40px; }
.d-pagehead { display: flex; align-items: flex-end; justify-content: space-between; gap: 24px; margin-bottom: 20px; }
.d-title { display: flex; align-items: baseline; gap: 14px; min-width: 0; }
.d-title h1 { font-size: 32px; }
.d-meta { font-size: 14px; color: var(--ink3); }
</style>
