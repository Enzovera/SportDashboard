<script setup lang="ts">
import type { ReminderProposal } from '@/api/supplements'
import BoxSection from '@/components/ui/BoxSection.vue'

// Coach proposal of a reminder time: needs a decision, so the red mark.
defineProps<{ proposal: ReminderProposal; busy: boolean }>()
const emit = defineEmits<{ decide: [decision: 'accept' | 'reject'] }>()
</script>

<template>
  <BoxSection title="Trener" label="Propozycja trenera" class="proposal">
    <template #meta>
      <span class="meta"><span class="mark" aria-hidden="true" />propozycja</span>
    </template>
    <p class="prose">{{ proposal.text }}</p>
    <div class="actions">
      <button type="button" class="btn btn--primary" :disabled="busy" @click="emit('decide', 'accept')">
        Przesuń na {{ proposal.proposedTime }}
      </button>
      <button type="button" class="btn" :disabled="busy" @click="emit('decide', 'reject')">
        Zostaw {{ proposal.currentTime }}
      </button>
    </div>
  </BoxSection>
</template>

<style scoped>
.mark { display: inline-block; width: 8px; height: 8px; background: var(--signal); margin-right: 6px; }
.prose { font-size: 14.5px; line-height: 1.5; }
.actions { display: flex; gap: 8px; margin-top: 4px; }
.actions .btn { flex: 1; justify-content: center; }
.btn:disabled { opacity: 0.6; cursor: progress; }
@media (min-width: 1024px) {
  .prose { font-size: 16px; line-height: 1.55; }
  .actions { gap: 12px; }
}
</style>
