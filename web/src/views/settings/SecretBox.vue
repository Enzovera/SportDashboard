<script setup lang="ts">
import { computed, ref, useId } from 'vue'
import { deleteSecret, generateMcpToken, putSecret, type SecretStatus } from '@/api/settings'
import BoxSection from '@/components/ui/BoxSection.vue'
import { secretMeta, secretProblem, type SecretSpec } from './settingsData'

// One credential: status in the meta, a write-only field, Zapisz / Usuń (back to the env) and, for the
// MCP token, Wygeneruj. A generated token is shown once, until the page is left.
const props = defineProps<{ spec: SecretSpec; status: SecretStatus | undefined; writable: boolean }>()
const emit = defineEmits<{ changed: [SecretStatus] }>()

const id = useId()
const value = ref('')
const busy = ref(false)
const problem = ref<string | null>(null)
const generated = ref<string | null>(null)
const meta = computed(() => secretMeta(props.status))

async function act(work: () => Promise<SecretStatus>) {
  busy.value = true
  problem.value = null
  try {
    emit('changed', await work())
    value.value = ''
  } catch (e) {
    problem.value = e instanceof Error ? e.message : String(e)
  } finally {
    busy.value = false
  }
}

function save() {
  const v = value.value.trim()
  const p = secretProblem(props.spec, v)
  if (p) {
    problem.value = p
    return
  }
  generated.value = null
  void act(() => putSecret(props.spec.name, v))
}

function remove() {
  generated.value = null
  void act(() => deleteSecret(props.spec.name))
}

function generate() {
  void act(async () => {
    const g = await generateMcpToken()
    generated.value = g.value
    return g
  })
}
</script>

<template>
  <BoxSection :title="spec.title" :label="spec.title">
    <template #meta>
      <span class="meta" :class="{ set: status?.set }">{{ meta }}</span>
    </template>
    <form class="form" @submit.prevent="save">
      <label :for="id" class="sr-only">{{ spec.title }}</label>
      <input
        :id="id"
        v-model="value"
        class="field"
        type="password"
        autocomplete="off"
        autocapitalize="off"
        spellcheck="false"
        :placeholder="status?.set ? 'Nowa wartość' : spec.name"
        :disabled="!writable || busy"
      />
      <div class="actions">
        <button v-if="spec.generate" type="button" class="btn" :disabled="!writable || busy" @click="generate">
          Wygeneruj
        </button>
        <button v-if="status?.set" type="button" class="btn" :disabled="busy" @click="remove">Usuń</button>
        <button type="submit" class="btn btn--primary" :disabled="!writable || busy || !value">Zapisz</button>
      </div>
    </form>
    <p v-if="problem" class="problem tone-warn" role="alert">{{ problem }}</p>
    <div v-if="generated" class="generated">
      <span class="dim">Nowy token, widoczny tylko teraz:</span>
      <code>{{ generated }}</code>
    </div>
  </BoxSection>
</template>

<style scoped>
.meta { white-space: nowrap; }
.meta.set { color: var(--ink2); font-variant-numeric: tabular-nums; }
.form { display: flex; flex-direction: column; gap: 8px; }
.field {
  min-height: 44px; padding: 0 12px; border: 1px solid var(--border-strong); border-radius: 0;
  font: inherit; font-size: 14px; color: var(--ink); background: var(--surface); width: 100%;
}
.field:focus-visible { outline: 2px solid var(--ink); outline-offset: 0; }
.field:disabled { color: var(--ink3); }
.actions { display: flex; justify-content: flex-end; gap: 8px; }
.btn:disabled { opacity: 0.45; cursor: default; }
.problem { font-size: 13px; }
.generated { display: flex; flex-direction: column; gap: 4px; font-size: 13px; }
.generated .dim { color: var(--ink2); }
.generated code { font-size: 13px; word-break: break-all; user-select: all; padding: 8px; border: 1px solid var(--rule); }
@media (min-width: 1024px) {
  .field { font-size: 14.5px; }
}
</style>
