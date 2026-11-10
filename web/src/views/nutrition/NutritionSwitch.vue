<script setup lang="ts">
// Jedzenie | Suplementy: two routes of the Żywienie tab, rendered as a segmented link pair.
// Keeps the query (e.g. ?chat=mock) when switching.
defineProps<{ supplementsLabel: string }>()
</script>

<template>
  <nav class="segmented switch" aria-label="Widok żywienia">
    <RouterLink :to="{ path: '/zywienie', query: $route.query }" exact-active-class="on">Jedzenie</RouterLink>
    <RouterLink :to="{ path: '/zywienie/suplementy', query: $route.query }" exact-active-class="on">
      {{ supplementsLabel }}
    </RouterLink>
  </nav>
</template>

<style scoped>
.switch a {
  display: inline-flex; align-items: center; min-height: 44px; padding: 0 12px;
  font-size: 13px; color: var(--ink2); text-decoration: none; white-space: nowrap;
}
.switch a + a { border-left: 1px solid var(--border); }
.switch a.on { background: var(--ink); color: var(--surface); font-weight: 700; }
.switch a:focus-visible { outline: 2px solid var(--ink); outline-offset: -2px; }
@media (max-width: 599px) {
  .switch a { padding: 0 10px; }
}
@media (min-width: 1024px) {
  .switch a { padding: 0 16px; font-size: 13.5px; }
}
</style>
