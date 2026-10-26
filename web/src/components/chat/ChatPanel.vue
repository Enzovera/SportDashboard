<script setup lang="ts">
import { onMounted, ref } from 'vue'
import type { Layout } from '@/composables/useLayout'
import ChatComposer from './ChatComposer.vue'
import ChatLog from './ChatLog.vue'
import { useChat } from './useChat'

// Conversation with the coach: head (panel only), log, composer.
// panel  = next to a view (fold: right column, desktop: aside), with its own head
// screen = full-screen chat on the cover; the head is the page header (ChatView)
const props = withDefaults(defineProps<{ layout: Layout; variant?: 'panel' | 'screen' }>(), { variant: 'panel' })

const chat = useChat()
const composer = ref<InstanceType<typeof ChatComposer> | null>(null)

function onKeydown(e: KeyboardEvent) {
  if (e.key !== 'Escape') return
  if (chat.state.streaming) chat.abort()
  else if (props.variant === 'panel') chat.close()
  else return
  e.preventDefault()
}

onMounted(() => {
  void chat.loadHistory()
  // On touch screens focusing would pop the keyboard over the log.
  if (props.layout === 'desktop') composer.value?.focus()
})
</script>

<template>
  <section id="chat-panel" class="chat" :class="[layout, variant]" aria-label="Czat z trenerem" @keydown="onKeydown">
    <div v-if="variant === 'panel'" class="head">
      <h2>Trener</h2>
      <span v-if="chat.contextLabel.value" class="meta">kontekst: {{ chat.contextLabel.value }}</span>
    </div>
    <ChatLog :class="layout" :messages="chat.state.messages" :busy="chat.state.streaming || chat.state.loading" :error="chat.state.error" />
    <ChatComposer
      ref="composer"
      :class="layout"
      :input-id="`chat-input-${variant}`"
      :streaming="chat.state.streaming"
      @send="chat.send"
      @abort="chat.abort"
    />
  </section>
</template>

<style scoped>
.chat { min-width: 0; min-height: 0; height: 100%; display: flex; flex-direction: column; background: var(--surface); overflow: hidden; }
.panel.fold { border: 1px solid var(--border); }
.screen { background: var(--bg); }
.head { flex: none; display: flex; justify-content: space-between; align-items: baseline; gap: 8px; padding: 12px 14px; border-bottom: 1px solid var(--rule); }
.head h2 { font-size: 14px; font-weight: 700; }
.desktop .head { padding: 16px 18px 14px; gap: 12px; }
.desktop .head h2 { font-size: 15px; }
</style>
