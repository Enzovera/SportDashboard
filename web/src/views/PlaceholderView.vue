<script setup lang="ts">
import PageLayout from '@/components/ui/PageLayout.vue'

defineProps<{ title: string }>()
</script>

<template>
  <PageLayout :title="title" />
</template>
