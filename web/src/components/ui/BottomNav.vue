<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import { useLayout } from '@/composables/useLayout'
import { isTabActive } from '@/lib/nav'
import Icon, { type IconName } from './Icon.vue'

const props = defineProps<{ proposalsPending?: number | null }>()
const layout = useLayout()
const route = useRoute()

interface Tab { to: string; label: string; icon: IconName; badge?: boolean }

// The same six tabs on both screens. Fold6: 3 | 3 around the hinge; cover: one row of six.
const left: Tab[] = [
  { to: '/', label: 'Dziś', icon: 'today' },
  { to: '/trening', label: 'Trening', icon: 'training' },
  { to: '/zywienie', label: 'Żywienie', icon: 'nutrition' },
]
const right: Tab[] = [
  { to: '/cialo', label: 'Ciało', icon: 'body' },
  { to: '/plan', label: 'Plan', icon: 'plan' },
  { to: '/propozycje', label: 'Propozycje', icon: 'proposals', badge: true },
]
const groups = computed(() => (layout.value === 'fold' ? [left, right] : [[...left, ...right]]))
</script>

<template>
  <nav aria-label="Główna nawigacja" class="nav" :class="layout">
    <div v-for="(group, i) in groups" :key="i" class="group" :style="{ '--n': group.length }">
      <RouterLink
        v-for="tab in group"
        :key="tab.to"
        :to="tab.to"
        class="tab"
        :class="{ current: isTabActive(tab.to, route.path) }"
        :aria-current="isTabActive(tab.to, route.path) ? 'page' : undefined"
      >
        <span class="glyph">
          <Icon :name="tab.icon" />
          <span v-if="tab.badge && props.proposalsPending" class="count">
            {{ props.proposalsPending }}<span class="sr-only"> do decyzji</span>
          </span>
        </span>
        <span class="label">{{ tab.label }}</span>
      </RouterLink>
    </div>
  </nav>
</template>

<style scoped>
.nav {
  flex: none; box-sizing: content-box; height: var(--nav-h);
  padding: 0 16px env(safe-area-inset-bottom);
  display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); column-gap: 24px;
  background: var(--surface); border-top: 1px solid var(--border);
}
.nav.cover { grid-template-columns: minmax(0, 1fr); padding-left: 0; padding-right: 0; }
.group { display: grid; grid-template-columns: repeat(var(--n), minmax(0, 1fr)); }
.tab {
  display: flex; flex-direction: column; align-items: center; justify-content: center; gap: 3px;
  min-width: 0; text-decoration: none; color: var(--ink3);
  -webkit-tap-highlight-color: transparent;
}
.glyph { position: relative; display: flex; }
.label { font-size: 12px; font-weight: 500; line-height: 1; white-space: nowrap; }
/* Six labels in 369 px: the narrow cut of Archivo keeps "Propozycje" inside its tab. */
.cover .label { font-size: 11.5px; font-stretch: 88%; }
.count { position: absolute; top: -6px; left: 13px; box-shadow: 0 0 0 2px var(--surface); }

.tab:active { background: var(--surface2); }
.tab.current { color: var(--ink); box-shadow: inset 0 3px 0 var(--ink); }
.tab.current .label { font-weight: 800; }
.tab.current :deep(.icon) { stroke-width: 2.25; }
.tab:focus-visible { outline: 2px solid var(--ink); outline-offset: -2px; }
</style>
