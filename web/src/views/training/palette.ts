import { onScopeDispose, shallowRef, type Ref } from 'vue'
import type { ChartPalette } from './chartOptions'

/** Reads chart colors and font from the design tokens on :root. */
export function readPalette(): ChartPalette {
  const s = getComputedStyle(document.documentElement)
  const v = (name: string) => s.getPropertyValue(name).trim()
  return {
    chart: v('--chart'),
    chartMid: v('--chart-mid'),
    chartSoft: v('--chart-soft'),
    grid: v('--grid'),
    axis: v('--axis'),
    ink2: v('--ink2'),
    ink3: v('--ink3'),
    signal: v('--signal'),
    font: v('--font') || 'Archivo, system-ui, sans-serif',
  }
}

/**
 * Current chart palette; refreshed when the color scheme or `data-theme` changes,
 * and once web fonts are loaded (so label widths are measured with Archivo).
 */
export function useChartPalette(): Ref<ChartPalette> {
  const palette = shallowRef(readPalette())
  const update = () => (palette.value = readPalette())

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
