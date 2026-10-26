/**
 * A nav tab is active on its own path and on its sub-pages (Żywienie on /zywienie/suplementy).
 * RouterLink's `router-link-active` only covers nested route records, not sibling routes with a common prefix.
 */
export function isTabActive(tabPath: string, routePath: string): boolean {
  if (tabPath === '/') return routePath === '/'
  return routePath === tabPath || routePath.startsWith(tabPath + '/')
}
