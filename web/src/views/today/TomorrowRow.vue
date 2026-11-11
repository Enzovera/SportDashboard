<script setup lang="ts">
import Icon from '@/components/ui/Icon.vue'
defineProps<{ tomorrow: { title: string; detail: string | null } }>()
</script>

<template>
  <RouterLink to="/plan" class="link-row">
    <span class="label">Jutro</span>
    <span class="text"><strong>{{ tomorrow.title }}</strong><template v-if="tomorrow.detail"> + {{ tomorrow.detail }}</template></span>
    <Icon name="arrow-right" :size="18" class="arrow" />
  </RouterLink>
</template>

<style scoped>
.label { font-size: 12.5px; color: var(--ink3); width: 40px; flex: none; }
.text { flex: 1; min-width: 0; font-size: 14px; }
.text strong { font-weight: 700; }
.arrow { color: var(--ink2); }
</style>
