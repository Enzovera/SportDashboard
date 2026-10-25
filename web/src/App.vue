<script setup lang="ts">
import BottomNav from '@/components/ui/BottomNav.vue'
</script>

<template>
  <div class="app">
    <!-- Wrapper keeps the nav at the bottom while a lazy view chunk is still loading. -->
    <div class="view"><RouterView /></div>
    <BottomNav />
  </div>
</template>

<style scoped>
.app { height: 100dvh; display: flex; flex-direction: column; background: var(--bg); }
.view { flex: 1; min-height: 0; display: flex; flex-direction: column; }
</style>
