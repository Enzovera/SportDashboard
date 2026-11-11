<script setup lang="ts">
import { computed, onScopeDispose, watchEffect } from 'vue'
import { useRoute } from 'vue-router'
import { useLayout } from '@/composables/useLayout'
import ChatPanel from './ChatPanel.vue'
import { useChat } from './useChat'

// Mount once in App.vue. Positions the chat panel itself:
//   desktop: aside 420 px on the right below the 64 px header; the view narrows (rule below)
//   fold:    right column from the hinge, between the view header and the bottom nav (FoldCzat)
//   cover:   nothing, the chat is the /czat route
const layout = useLayout()
const route = useRoute()
const chat = useChat()

const visible = computed(() => chat.state.open && layout.value !== 'cover' && route.name !== 'chat')
const docked = computed(() => visible.value && layout.value === 'desktop')

// Desktop: the content narrows instead of being covered (design/desktop.md).
watchEffect(() => document.documentElement.classList.toggle('chat-docked', docked.value))
onScopeDispose(() => document.documentElement.classList.remove('chat-docked'))
</script>

<template>
  <aside v-if="visible" class="chat-host" :class="layout">
    <ChatPanel :layout="layout" />
  </aside>
</template>

<style scoped>
.chat-host { position: fixed; right: 0; z-index: 20; display: flex; flex-direction: column; }
.chat-host.desktop {
  top: 64px; bottom: 0; width: 420px;
  border-left: 1px solid var(--border); background: var(--surface);
}
.chat-host.fold {
  top: calc(var(--header-h) + env(safe-area-inset-top) + 1px);
  bottom: calc(var(--nav-h) + env(safe-area-inset-bottom) + 1px); width: calc(50% - 12px);
  padding: 12px 16px 12px 0; background: var(--bg);
}
</style>

<style>
/* The view column of App.vue gives way to the docked panel; the .d-content container query then halves the grid. */
:root.chat-docked .app.desktop > .view { margin-right: 420px; }
</style>
