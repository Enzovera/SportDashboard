<script setup lang="ts">
import { computed } from 'vue'
import { parseCoachText } from './chatText'

// Model output is never passed to v-html: only paragraphs, line breaks and **bold** are interpreted.
const props = defineProps<{ text: string }>()
const paragraphs = computed(() => parseCoachText(props.text))
</script>

<template>
  <p v-for="(para, i) in paragraphs" :key="i">
    <template v-for="(run, j) in para" :key="j">
      <br v-if="run.kind === 'br'" />
      <strong v-else-if="run.bold">{{ run.text }}</strong>
      <template v-else>{{ run.text }}</template>
    </template>
  </p>
</template>

<style scoped>
p + p { margin-top: 0.6em; }
strong { font-weight: 700; }
</style>
