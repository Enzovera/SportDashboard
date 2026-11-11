<script setup lang="ts">
import { useRoute, useRouter } from 'vue-router'
import BottomNav from '@/components/ui/BottomNav.vue'
import DesktopHeader from '@/components/ui/DesktopHeader.vue'
import ChatHost from '@/components/chat/ChatHost.vue'
import { getProposals, proposalsPendingCount } from '@/api/proposals'
import { useLayout } from '@/composables/useLayout'

const layout = useLayout()
const route = useRoute()

// Dziś, Plan and Propozycje report the pending count themselves; entering through any other view
// asks once, so the nav badge does not depend on where the app was opened.
void useRouter()
  .isReady()
  .then(() => {
    if (!['today', 'plan', 'proposals'].includes(String(route.name))) return getProposals()
  })
  .catch(() => {})
</script>

<template>
  <div class="app" :class="layout">
    <DesktopHeader v-if="layout === 'desktop'" :proposals-pending="proposalsPendingCount" />
    <!-- Wrapper keeps the nav at the bottom while a lazy view chunk is still loading. -->
    <div class="view"><RouterView /></div>
    <!-- Cover chat is full screen without the nav (CoverCzat). -->
    <BottomNav
      v-if="layout !== 'desktop' && !(layout === 'cover' && route.name === 'chat')"
      :proposals-pending="proposalsPendingCount"
    />
    <!-- Positions itself: desktop aside 420 px, fold right column, nothing on cover. -->
    <ChatHost />
  </div>
</template>

<style scoped>
.app { height: 100dvh; display: flex; flex-direction: column; background: var(--bg); }
.view { flex: 1; min-height: 0; display: flex; flex-direction: column; }
</style>
