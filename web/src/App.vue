<script setup lang="ts">
import BottomNav from '@/components/ui/BottomNav.vue'
import DesktopHeader from '@/components/ui/DesktopHeader.vue'
import { useLayout } from '@/composables/useLayout'

const layout = useLayout()
</script>

<template>
  <div class="app" :class="layout">
    <DesktopHeader v-if="layout === 'desktop'" />
    <!-- Wrapper keeps the nav at the bottom while a lazy view chunk is still loading. -->
    <div class="view"><RouterView /></div>
    <BottomNav v-if="layout !== 'desktop'" />
  </div>
</template>

<style scoped>
.app { height: 100dvh; display: flex; flex-direction: column; background: var(--bg); }
.view { flex: 1; min-height: 0; display: flex; flex-direction: column; }
</style>
