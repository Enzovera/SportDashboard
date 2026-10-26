<script setup lang="ts">
import type { SupplementStock } from '@/api/supplements'
import BoxSection from '@/components/ui/BoxSection.vue'
import { stockLasts, stockLeft } from './supplementsData'

defineProps<{ stock: SupplementStock[] }>()
</script>

<template>
  <BoxSection v-if="stock.length" title="Zapasy" class="stock">
    <table class="table">
      <thead>
        <tr>
          <th scope="col"><span class="sr-only">Suplement</span></th>
          <th scope="col" class="r c-left">zostało</th>
          <th scope="col" class="r c-lasts">na ile</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="s in stock" :key="s.supplementId">
          <th scope="row" :class="{ strong: s.reorder }">{{ s.name }}</th>
          <td class="r">{{ stockLeft(s) }}</td>
          <td class="r" :class="{ 'tone-warn': s.reorder, lasts: s.reorder }">{{ stockLasts(s) }}</td>
        </tr>
      </tbody>
    </table>
  </BoxSection>
</template>

<style scoped>
.stock { gap: 4px; padding-bottom: 6px; }
.table { table-layout: fixed; }
.table thead th, .table tbody th { text-align: left; }
.table thead th.r { text-align: right; }
.table tbody tr:last-child { border-bottom: 0; }
.c-left { width: 72px; }
.c-lasts { width: 106px; }
.lasts { font-weight: 700; }
@media (min-width: 1024px) {
  .c-left { width: 84px; }
  .c-lasts { width: 124px; }
}
</style>
