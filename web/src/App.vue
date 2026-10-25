<script setup lang="ts">
import BottomNav from '@/components/ui/BottomNav.vue'
</script>

<template>
  <div class="app">
    <RouterView />
    <BottomNav />
  </div>
</template>

<style scoped>
.app { height: 100dvh; display: flex; flex-direction: column; background: var(--bg); }
</style>
