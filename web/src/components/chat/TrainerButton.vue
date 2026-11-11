<script setup lang="ts">
import Icon from '@/components/ui/Icon.vue'
import { useLayout } from '@/composables/useLayout'
import { useChat } from './useChat'

// Header "Trener" button, on every view. Cover: opens /czat. Fold: icon that toggles the chat panel (ChatHost).
// Desktop: the same toggle with the label.
const layout = useLayout()
const chat = useChat()
</script>

<template>
  <RouterLink v-if="layout === 'cover'" to="/czat" class="icon-btn" aria-label="Czat z trenerem" title="Trener">
    <Icon name="chat" />
  </RouterLink>
  <button
    v-else
    type="button"
    class="trainer"
    :class="[layout === 'desktop' ? 'btn' : 'icon-btn', { on: chat.state.open, 'btn--primary': layout === 'desktop' && chat.state.open }]"
    :aria-expanded="chat.state.open"
    aria-controls="chat-panel"
    :aria-label="chat.state.open ? 'Zamknij czat z trenerem' : 'Czat z trenerem'"
    :title="chat.state.open ? 'Zamknij czat' : 'Trener'"
    @click="chat.toggle"
  >
    <Icon :name="chat.state.open && layout !== 'desktop' ? 'close' : 'chat'" :size="layout === 'desktop' ? 18 : 22" />
    <span v-if="layout === 'desktop'" class="lbl">Trener</span>
  </button>
</template>

<style scoped>
.trainer { gap: 8px; white-space: nowrap; }
.icon-btn.on { background: var(--ink); color: var(--surface); }
@media (max-width: 1199px) {
  .btn { width: 44px; padding: 0; justify-content: center; }
  .lbl { display: none; }
}
</style>
