<script setup lang="ts">
import { ref } from 'vue'

const props = defineProps<{ streaming: boolean; inputId: string }>()
const emit = defineEmits<{ send: [text: string]; abort: [] }>()

const CHIPS = ['Co zjeść przed biegiem?', 'Zdążę na 1:40?']
const MAX_TEXT = 4000

const draft = ref('')
const input = ref<HTMLInputElement | null>(null)

function submit() {
  const text = draft.value.trim()
  if (!text || props.streaming) return
  emit('send', text)
  draft.value = ''
}

function chip(text: string) {
  if (!props.streaming) emit('send', text)
}

defineExpose({ focus: () => input.value?.focus() })
</script>

<template>
  <form class="compose" @submit.prevent="submit">
    <div class="chips">
      <button v-for="c in CHIPS" :key="c" type="button" class="chip" :disabled="streaming" @click="chip(c)">{{ c }}</button>
    </div>
    <div class="row">
      <label :for="inputId" class="sr-only">Wiadomość do trenera</label>
      <input
        :id="inputId"
        ref="input"
        v-model="draft"
        class="field"
        type="text"
        enterkeyhint="send"
        autocomplete="off"
        :maxlength="MAX_TEXT"
        placeholder="Napisz do trenera…"
      />
      <button v-if="streaming" type="button" class="send" aria-label="Zatrzymaj odpowiedź" @click="emit('abort')">
        <svg width="14" height="14" viewBox="0 0 14 14" aria-hidden="true"><rect width="14" height="14" fill="currentColor" /></svg>
      </button>
      <button v-else type="submit" class="send" aria-label="Wyślij" :disabled="!draft.trim()">
        <svg width="18" height="18" viewBox="0 0 18 18" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="square" aria-hidden="true"><path d="M9 15V3M4 8l5-5 5 5" /></svg>
      </button>
    </div>
  </form>
</template>

<style scoped>
.compose { flex: none; display: flex; flex-direction: column; gap: 8px; padding: 10px 14px 12px; border-top: 1px solid var(--rule); background: var(--surface); }
.chips { display: flex; gap: 6px; overflow-x: auto; scrollbar-width: none; }
.chip {
  flex: none; min-height: 44px; padding: 0 12px; border: 1px solid var(--border); background: transparent;
  color: var(--ink2); font-size: 13px; cursor: pointer; white-space: nowrap;
}
.chip:disabled { color: var(--ink3); cursor: default; }
.row { display: flex; gap: 8px; align-items: center; }
.field {
  flex: 1; min-width: 0; min-height: 44px; padding: 0 12px; border: 1px solid var(--border-strong); border-radius: 0;
  font: inherit; font-size: 14.5px; color: var(--ink); background: var(--surface);
}
.field::placeholder { color: var(--ink3); }
.field:focus-visible { outline: 2px solid var(--ink); outline-offset: -1px; }
.send {
  flex: none; width: 44px; height: 44px; border: 0; background: var(--ink); color: var(--surface);
  display: flex; align-items: center; justify-content: center; cursor: pointer;
}
.send:disabled { background: var(--border-strong); cursor: default; }
.chip:focus-visible, .send:focus-visible { outline: 2px solid var(--ink); outline-offset: 2px; }

.cover { padding: 10px 12px 14px; border-top-color: var(--border); }
.desktop { gap: 10px; padding: 12px 18px 16px; }
.desktop .chips { gap: 8px; }
.desktop .chip { padding: 0 14px; font-size: 13.5px; }
.desktop .field { font-size: 15px; min-height: 48px; }
.desktop .send { width: 48px; height: 48px; }
</style>
