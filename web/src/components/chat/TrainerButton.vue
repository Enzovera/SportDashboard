<script setup lang="ts">
import { useLayout } from '@/composables/useLayout'
import { useChat } from './useChat'

// Header "Trener" button. Fold/desktop: toggles the chat panel (ChatHost). Cover: opens /czat.
const layout = useLayout()
const chat = useChat()
</script>

<template>
  <RouterLink v-if="layout === 'cover'" to="/czat" class="btn trainer">Trener</RouterLink>
  <button
    v-else
    type="button"
    class="btn trainer"
    :class="{ 'btn--primary': chat.state.open }"
    :aria-expanded="chat.state.open"
    aria-controls="chat-panel"
    :aria-label="chat.state.open ? 'Zamknij czat z trenerem' : undefined"
    @click="chat.toggle"
  >
    Trener<span v-if="chat.state.open" class="x" aria-hidden="true">✕</span>
  </button>
</template>

<style scoped>
.trainer { gap: 8px; white-space: nowrap; }
.x { font-weight: 400; }
</style>
