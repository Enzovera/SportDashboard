<script setup lang="ts">
// Frame of every view: fixed 56 px header, scrolling main. The bottom nav lives in App.vue.
defineProps<{ title: string; meta?: string | null }>()
</script>

<template>
  <div class="frame">
    <header class="header">
      <div class="title">
        <h1>{{ title }}</h1>
        <span v-if="meta" class="header-meta">{{ meta }}</span>
      </div>
      <div class="actions"><slot name="actions" /></div>
    </header>
    <main class="main">
      <slot />
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

@media (max-width: 599px) {
  .header { padding: 8px 12px; grid-template-columns: minmax(0, 1fr) auto; column-gap: 12px; }
  .title { flex-direction: column; align-items: flex-start; gap: 4px; }
}
</style>
