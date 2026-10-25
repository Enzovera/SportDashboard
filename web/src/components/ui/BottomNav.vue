<script setup lang="ts">
import { computed } from 'vue'
import { useLayout } from '@/composables/useLayout'

const props = defineProps<{ proposalsPending?: number | null }>()
const layout = useLayout()

interface Tab { to: string; label: string; badge?: boolean }

// Fold6: 3 | 3 around the hinge. Cover: five tabs, Ciało is reachable from Dziś.
const left: Tab[] = [
  { to: '/', label: 'Dziś' },
  { to: '/trening', label: 'Trening' },
  { to: '/zywienie', label: 'Żywienie' },
]
const right: Tab[] = [
  { to: '/cialo', label: 'Ciało' },
  { to: '/plan', label: 'Plan' },
  { to: '/propozycje', label: 'Propozycje', badge: true },
]
const groups = computed(() =>
  layout.value === 'fold' ? [left, right] : [[...left, ...right.filter((t) => t.to !== '/cialo')]],
)
</script>

<template>
  <nav aria-label="Główna nawigacja" class="nav" :class="layout">
    <div v-for="(group, i) in groups" :key="i" class="group" :style="{ '--n': group.length }">
      <RouterLink v-for="tab in group" :key="tab.to" :to="tab.to" class="tab">
        {{ tab.label }}
        <span v-if="tab.badge && props.proposalsPending" class="count">{{ props.proposalsPending }}</span>
      </RouterLink>
    </div>
  </nav>
</template>

<style scoped>
.nav {
  flex: none; height: var(--nav-h); padding: 0 16px; padding-bottom: env(safe-area-inset-bottom);
  display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); column-gap: 24px;
  background: var(--surface); border-top: 1px solid var(--border);
}
.nav.cover { grid-template-columns: minmax(0, 1fr); padding: 0 4px; }
.group { display: grid; grid-template-columns: repeat(var(--n), minmax(0, 1fr)); }
.tab {
  display: flex; align-items: center; justify-content: center; gap: 6px; min-height: 44px;
  text-decoration: none; font-size: 14px; font-weight: 500; color: var(--ink2);
}
.cover .tab { font-size: 13px; }
.tab.router-link-exact-active { font-weight: 800; color: var(--ink); box-shadow: inset 0 3px 0 var(--ink); }
.tab:focus-visible { outline: 2px solid var(--ink); outline-offset: -2px; }
</style>
