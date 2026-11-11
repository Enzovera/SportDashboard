<script setup lang="ts">
import { computed, nextTick, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import { useLayout } from '@/composables/useLayout'
import { useTheme } from '@/composables/useTheme'
import { TABS, coverTabs, isTabActive, type NavTab } from '@/lib/nav'
import { THEME_LABEL } from '@/lib/theme'
import Icon from './Icon.vue'

const props = defineProps<{ proposalsPending?: number | null }>()
const layout = useLayout()
const route = useRoute()
const { mode, cycle } = useTheme()

// Fold6: every tab, 3 | 3 around the hinge. Cover: the primary tabs and "Więcej" with the rest (lib/nav.ts).
const cover = coverTabs()
const HALF = Math.ceil(TABS.length / 2)
const groups = computed<NavTab[][]>(() =>
  layout.value === 'fold' ? [TABS.slice(0, HALF), TABS.slice(HALF)] : [cover.bar],
)
const active = (tab: NavTab) => isTabActive(tab.to, route.path)
const moreActive = computed(() => cover.more.some(active))
const pending = computed(() => props.proposalsPending ?? 0)

// ---- "Więcej" sheet (cover only) ----
const open = ref(false)
const sheet = ref<HTMLElement | null>(null)
// The button renders inside the v-for over groups, so the template ref is an array.
const moreButton = ref<HTMLButtonElement[]>([])

async function toggle() {
  open.value = !open.value
  if (!open.value) return
  await nextTick()
  sheet.value?.querySelector<HTMLElement>('a, button')?.focus()
}
function close(restoreFocus = false) {
  if (!open.value) return
  open.value = false
  if (restoreFocus) moreButton.value[0]?.focus()
}
watch(() => route.fullPath, () => close())
watch(layout, () => close())
</script>

<template>
  <nav aria-label="Główna nawigacja" class="nav" :class="layout" @keydown.esc="close(true)">
    <template v-if="layout === 'cover' && open">
      <div class="backdrop" @click="close()" />
      <div id="nav-more" ref="sheet" class="sheet" role="dialog" aria-label="Więcej">
        <RouterLink
          v-for="tab in cover.more"
          :key="tab.to"
          :to="tab.to"
          class="row"
          :class="{ current: active(tab) }"
          :aria-current="active(tab) ? 'page' : undefined"
          @click="close()"
        >
          <Icon :name="tab.icon" />
          <span class="grow">{{ tab.label }}</span>
          <span v-if="tab.badge && pending" class="count">{{ pending }}<span class="sr-only"> do decyzji</span></span>
          <Icon name="arrow-right" :size="18" class="go" />
        </RouterLink>
        <button type="button" class="row" @click="cycle">
          <Icon name="theme" />
          <span class="grow">Motyw</span>
          <span class="value">{{ THEME_LABEL[mode] }}</span>
        </button>
      </div>
    </template>

    <div v-for="(group, i) in groups" :key="i" class="group">
      <RouterLink
        v-for="tab in group"
        :key="tab.to"
        :to="tab.to"
        class="tab"
        :class="{ current: active(tab) && !open }"
        :aria-current="active(tab) ? 'page' : undefined"
      >
        <span class="glyph">
          <Icon :name="tab.icon" />
          <span v-if="tab.badge && pending" class="count">{{ pending }}<span class="sr-only"> do decyzji</span></span>
        </span>
        <span class="label">{{ tab.label }}</span>
      </RouterLink>
      <button
        v-if="layout === 'cover'"
        ref="moreButton"
        type="button"
        class="tab"
        :class="{ current: moreActive || open }"
        aria-haspopup="dialog"
        aria-controls="nav-more"
        :aria-expanded="open"
        @click="toggle"
      >
        <span class="glyph">
          <Icon :name="open ? 'close' : 'more'" />
          <span v-if="pending && !open" class="count">{{ pending }}<span class="sr-only"> propozycje do decyzji</span></span>
        </span>
        <span class="label">Więcej</span>
      </button>
    </div>
  </nav>
</template>

<style scoped>
.nav {
  --nav-full: calc(var(--nav-h) + env(safe-area-inset-bottom) + 1px);
  position: relative; z-index: 30; flex: none; box-sizing: content-box; height: var(--nav-h);
  padding: 0 16px env(safe-area-inset-bottom);
  display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); column-gap: 24px;
  background: var(--surface); border-top: 1px solid var(--border);
}
.nav.cover { grid-template-columns: minmax(0, 1fr); padding-left: 0; padding-right: 0; }
.group { display: grid; grid-auto-flow: column; grid-auto-columns: minmax(0, 1fr); }
.tab {
  display: flex; flex-direction: column; align-items: center; justify-content: center; gap: 3px;
  min-width: 0; padding: 0; border: 0; background: transparent; cursor: pointer;
  text-decoration: none; color: var(--ink3); -webkit-tap-highlight-color: transparent;
}
.glyph { position: relative; display: flex; }
.label { font-size: 12px; font-weight: 500; line-height: 1; white-space: nowrap; }
.glyph .count { position: absolute; top: -6px; left: 13px; box-shadow: 0 0 0 2px var(--surface); }

.tab:active { background: var(--surface2); }
.tab.current { color: var(--ink); box-shadow: inset 0 3px 0 var(--ink); }
.tab.current .label { font-weight: 800; }
.tab.current :deep(.icon) { stroke-width: 2.25; }
.tab:focus-visible, .row:focus-visible { outline: 2px solid var(--ink); outline-offset: -2px; }

/* "Więcej": a sheet that grows out of the nav; the page behind is dimmed and closes it on tap. */
.backdrop { position: fixed; inset: 0 0 var(--nav-full); background: rgb(20 20 16 / 0.45); }
.sheet {
  position: fixed; left: 0; right: 0; bottom: var(--nav-full);
  display: flex; flex-direction: column; background: var(--surface); border-top: 1px solid var(--border-strong);
  animation: rise 0.14s ease-out;
}
.row {
  display: flex; align-items: center; gap: 14px; min-height: 56px; padding: 0 16px; width: 100%;
  border: 0; background: transparent; cursor: pointer; text-align: left; text-decoration: none;
  font: inherit; font-size: 15px; font-weight: 600; color: var(--ink);
  -webkit-tap-highlight-color: transparent;
}
.row + .row { border-top: 1px solid var(--rule); }
.row:active { background: var(--surface2); }
.row.current { box-shadow: inset 3px 0 0 var(--ink); font-weight: 800; }
.grow { flex: 1; min-width: 0; }
.go { color: var(--ink2); }
.value { font-size: 13.5px; font-weight: 500; color: var(--ink3); }
@keyframes rise { from { transform: translateY(12px); opacity: 0; } }
@media (prefers-reduced-motion: reduce) { .sheet { animation: none; } }
</style>
