<script setup lang="ts">
import { computed } from 'vue'
import { getSecretSettings, type SecretName, type SecretStatus } from '@/api/settings'
import { useApi } from '@/composables/useApi'
import { useLayout } from '@/composables/useLayout'
import PageLayout from '@/components/ui/PageLayout.vue'
import SecretBox from './SecretBox.vue'
import { COACH_SECRETS, SOURCE_SECRETS } from './settingsData'

// Ustawienia: credentials stored in the app (docs/design.md, "Bezpieczeństwo"). Fold: sources | coach.
const layout = useLayout()
const { data, error } = useApi(getSecretSettings)

const byName = computed(() => new Map((data.value?.secrets ?? []).map((s) => [s.name, s])))
const status = (name: SecretName) => byName.value.get(name)

function changed(s: SecretStatus) {
  if (!data.value) return
  data.value = { ...data.value, secrets: data.value.secrets.map((x) => (x.name === s.name ? s : x)) }
}

const columns = computed(() =>
  layout.value === 'cover' ? [[...SOURCE_SECRETS, ...COACH_SECRETS]] : [SOURCE_SECRETS, COACH_SECRETS],
)
</script>

<template>
  <PageLayout title="Ustawienia" meta="Klucze i tokeny">
    <div v-if="error && !data" :class="{ page: layout !== 'desktop' }">
      <section class="box" role="alert">{{ error }}</section>
    </div>
    <div v-else-if="data" :class="layout === 'desktop' ? 'd-grid' : 'page'">
      <section v-if="!data.writable" class="box notice s-12" role="alert">
        <span class="tone-warn">Zapis wyłączony: API nie ma FORMA_SECRETS_KEY.</span>
      </section>
      <div v-for="(col, i) in columns" :key="i" :class="layout === 'desktop' ? 'stack s-6' : 'col'">
        <SecretBox
          v-for="spec in col"
          :key="spec.name"
          :spec="spec"
          :status="status(spec.name)"
          :writable="data.writable"
          @changed="changed"
        />
      </div>
    </div>
  </PageLayout>
</template>

<style scoped>
.page > .notice { grid-column: 1 / -1; margin-bottom: 12px; }
.notice { font-size: 14px; }
</style>
