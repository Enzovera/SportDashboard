<script setup lang="ts">
import BoxSection from '@/components/ui/BoxSection.vue'
import { dayMonth } from '@/lib/format'

defineProps<{ note: { date: string; text: string } }>()
</script>

<template>
  <BoxSection title="Trener" :meta="dayMonth(note.date)" label="Uwaga trenera">
    <p class="text">{{ note.text }}</p>
  </BoxSection>
</template>

<style scoped>
.text { font-size: 14.5px; line-height: 1.5; }
</style>
