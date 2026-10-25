import { onScopeDispose, ref } from 'vue'

/**
 * cover   = outer screen (~369 px): one column, bottom nav with 5 tabs
 * fold    = Fold6 inner screen (707 px) up to tablets: two columns around the hinge, bottom nav 3 | 3
 * desktop = browser from 1024 px: nav in the header, 12-column grid (design/desktop.md)
 */
export type Layout = 'cover' | 'fold' | 'desktop'

export const COVER_MAX_WIDTH = 599
export const DESKTOP_MIN_WIDTH = 1024

export function layoutFor(width: number): Layout {
  if (width <= COVER_MAX_WIDTH) return 'cover'
  return width >= DESKTOP_MIN_WIDTH ? 'desktop' : 'fold'
}

export function useLayout() {
  const cover = window.matchMedia(`(max-width: ${COVER_MAX_WIDTH}px)`)
  const desktop = window.matchMedia(`(min-width: ${DESKTOP_MIN_WIDTH}px)`)
  const current = (): Layout => (cover.matches ? 'cover' : desktop.matches ? 'desktop' : 'fold')
  const layout = ref<Layout>(current())
  const update = () => {
    const next = current()
    if (next !== layout.value) layout.value = next
  }
  // Media-query `change` is the precise signal; `resize` backs it up where it does not fire
  // (e.g. emulated viewports), so unfolding the Fold or resizing a window always switches the layout.
  cover.addEventListener('change', update)
  desktop.addEventListener('change', update)
  window.addEventListener('resize', update, { passive: true })
  onScopeDispose(() => {
    cover.removeEventListener('change', update)
    desktop.removeEventListener('change', update)
    window.removeEventListener('resize', update)
  })
  return layout
}
