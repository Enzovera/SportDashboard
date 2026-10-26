<script setup lang="ts">
import BoxSection from '@/components/ui/BoxSection.vue'
import type { BlockedProposal } from '@/api/proposals'
import ValidatorChecks from './ValidatorChecks.vue'

defineProps<{ blocked: BlockedProposal[] }>()
</script>

<template>
  <BoxSection title="Zablokowane" meta="trener dostał powód" label="Zablokowane przez walidatory">
    <div v-for="b in blocked" :key="b.id" class="item">
      <div class="title">{{ b.title }}</div>
      <ValidatorChecks :checks="b.checks" blocked />
    </div>
  </BoxSection>
</template>

<style scoped>
.item { display: flex; flex-direction: column; gap: 8px; }
.item + .item { border-top: 1px solid var(--rule); padding-top: 10px; }
.title { font-size: 14.5px; font-weight: 700; }
@media (min-width: 1024px) {
  .title { font-size: 15px; }
}
</style>
