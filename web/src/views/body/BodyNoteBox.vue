<script setup lang="ts">
import { computed } from 'vue'
import type { BodyResponse } from '@/api/body'
import BoxSection from '@/components/ui/BoxSection.vue'
import { dayMonth } from '@/lib/format'

const props = defineProps<{ note: NonNullable<BodyResponse['coachNote']> }>()
const meta = computed(() => (props.note.kind === 'rule' ? 'zasada' : dayMonth(props.note.date)))
</script>

<template>
  <BoxSection title="Trener" :meta="meta" label="Uwaga trenera">
    <p class="text">{{ note.text }}</p>
  </BoxSection>
</template>

<style scoped>
.text { font-size: 14.5px; line-height: 1.5; }
@media (min-width: 1024px) { .text { font-size: 16px; line-height: 1.55; } }
</style>
