import { onScopeDispose, shallowRef, type Ref } from 'vue'
import { readPalette } from '@/views/training/palette'
import type { NutritionPalette } from './nutritionCharts'

/** Training's chart palette plus the status colors the nutrition charts need. */
function read(): NutritionPalette {
  const s = getComputedStyle(document.documentElement)
  const v = (name: string) => s.getPropertyValue(name).trim()
  return { ...readPalette(), ink: v('--ink'), surface: v('--surface'), good: v('--good'), warn: v('--warn') }
}

/** Refreshed on color scheme / `data-theme` change and once web fonts are loaded (as training/palette.ts). */
export function useNutritionPalette(): Ref<NutritionPalette> {
  const palette = shallowRef(read())
  const update = () => (palette.value = read())

  const scheme = window.matchMedia('(prefers-color-scheme: dark)')
  scheme.addEventListener('change', update)
  const themeObserver = new MutationObserver(update)
  themeObserver.observe(document.documentElement, { attributes: true, attributeFilter: ['data-theme'] })
  let disposed = false
  document.fonts?.ready.then(() => !disposed && update())

  onScopeDispose(() => {
    disposed = true
    scheme.removeEventListener('change', update)
    themeObserver.disconnect()
  })
  return palette
}
