<script setup lang="ts">
// Standard box: title (h2) + optional meta on the right. `hero` = the one dominant dark block per screen.
defineProps<{ title: string; meta?: string | null; hero?: boolean; label?: string }>()
</script>

<template>
  <section class="box" :class="{ 'box--hero': hero }" :aria-label="label ?? title">
    <div class="box-head">
      <h2>{{ title }}</h2>
      <slot name="meta">
        <span v-if="meta" class="meta">{{ meta }}</span>
      </slot>
    </div>
    <slot />
  </section>
</template>
