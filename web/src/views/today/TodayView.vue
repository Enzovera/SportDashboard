<script setup lang="ts">
import { computed, watch } from 'vue'
import TrainerButton from '@/components/chat/TrainerButton.vue'
import { proposalsPendingCount } from '@/api/proposals'
import PageLayout from '@/components/ui/PageLayout.vue'
import SyncStatus from '@/components/ui/SyncStatus.vue'
import SyncButton from '@/components/ui/SyncButton.vue'
import { getToday } from '@/api/client'
import { useApi } from '@/composables/useApi'
import { useLayout } from '@/composables/useLayout'
import { desktopCells, headerMeta, sectionColumns } from './today'
import TodaySection from './TodaySection.vue'

const layout = useLayout()
const { data, error } = useApi(getToday)

// Dziś is the usual entry point: seed the nav badge before Propozycje is ever opened.
watch(data, (d) => {
  if (d) proposalsPendingCount.value = d.proposalsPending
})

const cover = computed(() => layout.value === 'cover')
const meta = computed(() => (data.value ? headerMeta(data.value) : null))
const columns = computed(() => (data.value && layout.value !== 'desktop' ? sectionColumns(data.value, layout.value) : []))
const cells = computed(() => (data.value && layout.value === 'desktop' ? desktopCells(data.value) : []))
</script>

<template>
  <PageLayout title="Dziś" :meta="meta">
    <template #global>
      <template v-if="data">
        <SyncStatus :sync="data.sync" :compact="cover" />
        <SyncButton />
        <TrainerButton />
      </template>
    </template>

    <template v-if="data">
      <div v-if="layout === 'desktop'" class="d-grid">
        <div
          v-for="cell in cells"
          :key="cell.keys.join()"
          class="stack"
          :class="[`s-${cell.span}`, { stretch: cell.stretch }]"
        >
          <TodaySection v-for="key in cell.keys" :key="key" :name="key" :data="data" :layout="layout" />
        </div>
      </div>
      <div v-else class="page" :class="{ cover }">
        <div v-for="(col, i) in columns" :key="i" class="col">
          <TodaySection v-for="key in col" :key="key" :name="key" :data="data" :layout="layout" />
          <!-- Fold/cover have no header link: Ustawienia close the last column of Dziś. -->
          <RouterLink v-if="i === columns.length - 1" to="/ustawienia" class="link-row">Ustawienia</RouterLink>
        </div>
      </div>
    </template>
    <div v-else-if="error" :class="{ page: layout !== 'desktop' }">
      <section class="box error" role="alert">{{ error }}</section>
    </div>
  </PageLayout>
</template>

<style scoped>
.trainer { min-height: 44px; }
.cover .col { gap: 10px; }
.error { grid-column: 1 / -1; font-size: 14px; }
@media (max-width: 599px) {
  .trainer { padding: 0 12px; }
}
</style>
