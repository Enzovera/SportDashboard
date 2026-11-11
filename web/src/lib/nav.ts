import type { IconName } from '@/components/ui/Icon.vue'

export interface NavTab {
  to: string
  label: string
  icon: IconName
  /** Shows the count of proposals waiting for a decision. */
  badge?: boolean
  /** Has its own place in the phone's bottom nav; the rest sit under "Więcej". */
  primary: boolean
}

/**
 * All destinations, in nav order. Fold6 and desktop show every tab; the cover screen (369 px) shows the
 * primary ones and "Więcej", so a new view is one more entry here and never a seventh tab on the phone.
 */
export const TABS: NavTab[] = [
  { to: '/', label: 'Dziś', icon: 'today', primary: true },
  { to: '/trening', label: 'Trening', icon: 'training', primary: true },
  { to: '/zywienie', label: 'Żywienie', icon: 'nutrition', primary: true },
  { to: '/cialo', label: 'Ciało', icon: 'body', primary: false },
  { to: '/plan', label: 'Plan', icon: 'plan', primary: true },
  { to: '/propozycje', label: 'Propozycje', icon: 'proposals', badge: true, primary: false },
]

/** Cover nav: primary tabs in the bar, the rest in the "Więcej" sheet with what needs a decision first. */
export function coverTabs(tabs: NavTab[] = TABS): { bar: NavTab[]; more: NavTab[] } {
  const more = tabs.filter((t) => !t.primary)
  return {
    bar: tabs.filter((t) => t.primary),
    more: [...more.filter((t) => t.badge), ...more.filter((t) => !t.badge)],
  }
}

/**
 * A nav tab is active on its own path and on its sub-pages (Żywienie on /zywienie/suplementy).
 * RouterLink's `router-link-active` only covers nested route records, not sibling routes with a common prefix.
 */
export function isTabActive(tabPath: string, routePath: string): boolean {
  if (tabPath === '/') return routePath === '/'
  return routePath === tabPath || routePath.startsWith(tabPath + '/')
}
