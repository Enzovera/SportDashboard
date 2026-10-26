<script setup lang="ts">
// Desktop app header (design/screens/DesktopTemplate.html): brand, text nav, and a slot that views
// fill through <Teleport to="#d-actions"> (sync status, Trener).
import { useRoute } from 'vue-router'
import { isTabActive } from '@/lib/nav'

defineProps<{ proposalsPending?: number | null }>()
const route = useRoute()

const tabs = [
  { to: '/', label: 'Dziś' },
  { to: '/trening', label: 'Trening' },
  { to: '/zywienie', label: 'Żywienie' },
  { to: '/cialo', label: 'Ciało' },
  { to: '/plan', label: 'Plan' },
  { to: '/propozycje', label: 'Propozycje', badge: true },
]
</script>

<template>
  <header class="d-header">
    <RouterLink to="/" class="brand">Forma</RouterLink>
    <nav aria-label="Główna nawigacja" class="nav">
      <RouterLink
        v-for="tab in tabs"
        :key="tab.to"
        :to="tab.to"
        class="tab"
        :class="{ current: isTabActive(tab.to, route.path) }"
        :aria-current="isTabActive(tab.to, route.path) ? 'page' : undefined"
      >
        {{ tab.label }}
        <span v-if="tab.badge && proposalsPending" class="count">{{ proposalsPending }}</span>
      </RouterLink>
    </nav>
    <div id="d-actions" class="actions" />
  </header>
</template>

<style scoped>
.d-header {
  flex: none; height: 64px; padding: 0 32px; display: flex; align-items: stretch; gap: 40px;
  background: var(--surface); border-bottom: 1px solid var(--border);
}
.brand { display: flex; align-items: center; font-size: 22px; font-weight: 800; font-stretch: 78%; text-decoration: none; letter-spacing: -0.01em; }
.nav { display: flex; align-items: stretch; gap: 4px; }
.tab { display: flex; align-items: center; gap: 6px; padding: 0 14px; text-decoration: none; font-size: 15px; font-weight: 500; color: var(--ink2); }

.tab.current { font-weight: 800; color: var(--ink); box-shadow: inset 0 -3px 0 var(--ink); }
.tab:focus-visible, .brand:focus-visible { outline: 2px solid var(--ink); outline-offset: -2px; }
.actions { margin-left: auto; display: flex; align-items: center; gap: 16px; }
</style>
