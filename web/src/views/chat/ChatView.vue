<script setup lang="ts">
import { computed, watchEffect } from 'vue'
import { useRouter } from 'vue-router'
import PageLayout from '@/components/ui/PageLayout.vue'
import ChatPanel from '@/components/chat/ChatPanel.vue'
import { useChat } from '@/components/chat/useChat'
import { useLayout } from '@/composables/useLayout'

// /czat: full-screen chat on the cover screen (CoverCzat). On fold/desktop the chat is a panel next to
// the view it was opened from, so the route opens the panel and goes back to that view.
const layout = useLayout()
const router = useRouter()
const chat = useChat()

const from = computed(() => chat.state.from)
const meta = computed(() => (chat.contextLabel.value ? `kontekst: ${chat.contextLabel.value}` : null))
const backLabel = computed(() => from.value.screen?.label ?? 'Dziś')

watchEffect(() => {
  if (layout.value === 'cover') return
  chat.open()
  void router.replace(from.value.path)
})
</script>

<template>
  <PageLayout
    v-if="layout === 'cover'"
    title="Trener"
    :meta="meta"
    :back="{ to: from.path, label: `Wróć do widoku ${backLabel}` }"
    plain
  >
    <ChatPanel layout="cover" variant="screen" />
  </PageLayout>
</template>
