<script setup lang="ts">
import { computed, type DeepReadonly } from 'vue'
import { time } from '@/lib/format'
import { dayLabel, toolTrace } from './chatText'
import CoachText from './CoachText.vue'
import type { UiMessage } from './useChat'

const props = defineProps<{ messages: readonly DeepReadonly<UiMessage>[]; busy: boolean; error: string | null }>()

type Item =
  | { kind: 'day'; key: string; label: string }
  | { kind: 'msg'; key: string; msg: DeepReadonly<UiMessage>; trace: string | null }

const items = computed<Item[]>(() => {
  const out: Item[] = []
  let day = ''
  for (const msg of props.messages) {
    const label = dayLabel(msg.createdAt)
    if (label !== day) {
      day = label
      out.push({ kind: 'day', key: `d-${msg.key}`, label })
    }
    out.push({ kind: 'msg', key: msg.key, msg, trace: msg.role === 'assistant' ? toolTrace(msg.tools) : null })
  }
  return out
})
</script>

<template>
  <!-- column-reverse with a single child keeps the log scrolled to the newest message without JS -->
  <div class="log" role="log" aria-label="Rozmowa z trenerem" :aria-busy="busy">
    <div class="thread">
      <template v-for="item in items" :key="item.key">
        <div v-if="item.kind === 'day'" class="day">{{ item.label }}</div>
        <div v-else-if="item.msg.role === 'user'" class="user">
          <div class="bubble">{{ item.msg.content }}</div>
          <span class="time">{{ time(item.msg.createdAt) }}</span>
        </div>
        <template v-else>
          <div v-if="item.trace" class="trace">{{ item.trace }}</div>
          <div v-if="item.msg.content || item.msg.status === 'streaming'" class="coach" :class="{ streaming: item.msg.status === 'streaming' }">
            <CoachText :text="item.msg.content" />
            <span v-if="item.msg.status === 'streaming'" class="cursor" aria-hidden="true" />
          </div>
          <div v-if="item.msg.status === 'error'" class="note warn">{{ item.msg.error }}</div>
          <div v-else-if="item.msg.status === 'aborted'" class="note">Przerwano</div>
          <!-- Extension point: item.msg.attachments (proposal cards) once the MCP server exists. -->
        </template>
      </template>
      <div v-if="error" class="note warn">{{ error }}</div>
    </div>
  </div>
</template>

<style scoped>
.log { flex: 1; min-height: 0; overflow-y: auto; display: flex; flex-direction: column-reverse; overscroll-behavior: contain; }
.thread { display: flex; flex-direction: column; gap: 10px; padding: 12px 14px; }
.day { align-self: center; font-size: 12px; color: var(--ink3); }
.user { align-self: flex-end; max-width: 86%; display: flex; flex-direction: column; align-items: flex-end; gap: 3px; }
.bubble { background: var(--ink); color: var(--surface); padding: 9px 12px; font-size: 14.5px; line-height: 1.45; white-space: pre-wrap; overflow-wrap: anywhere; }
.time { font-size: 11.5px; color: var(--ink3); }
.trace { align-self: flex-start; font-size: 12.5px; color: var(--ink3); }
.coach {
  align-self: flex-start; max-width: 94%; background: var(--surface2); border: 1px solid var(--rule);
  padding: 9px 12px; font-size: 14.5px; line-height: 1.45; overflow-wrap: anywhere;
}
/* Keep the cursor on the last line while the answer streams in. */
.coach.streaming :deep(p:last-of-type) { display: inline; }
.cursor { display: inline-block; width: 7px; height: 1em; margin-left: 3px; vertical-align: -0.15em; background: var(--ink3); animation: blink 1s steps(1) infinite; }
.note { align-self: flex-start; font-size: 12.5px; color: var(--ink3); }
.note.warn { color: var(--warn); font-weight: 600; }
@keyframes blink { 50% { opacity: 0; } }
@media (prefers-reduced-motion: reduce) { .cursor { animation: none; } }

/* cover (full screen): slightly larger text, 12 px margin */
.cover .thread { padding: 12px; }
.cover .bubble, .cover .coach { font-size: 15px; }

/* desktop scale (design/desktop.md) */
.desktop .thread { gap: 12px; padding: 16px 18px; }
.desktop .day { font-size: 12.5px; }
.desktop .user { gap: 4px; }
.desktop .bubble, .desktop .coach { padding: 10px 14px; font-size: 15px; line-height: 1.5; }
.desktop .time { font-size: 12px; }
.desktop .trace, .desktop .note { font-size: 13px; }
</style>
