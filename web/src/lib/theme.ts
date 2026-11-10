/**
 * Color theme chosen by the user. `auto` follows `prefers-color-scheme`; `light`/`dark` set `data-theme`
 * on <html>, which tokens.css and the chart palettes already react to.
 * The same key is read by the inline script in index.html, so the first paint has the right theme.
 */
export type ThemeMode = 'auto' | 'light' | 'dark'

export const THEME_STORAGE_KEY = 'forma.theme'

export const THEME_LABEL: Record<ThemeMode, string> = { auto: 'Auto', light: 'Jasny', dark: 'Ciemny' }

// Matches the `theme-color` metas in index.html (= --bg of each theme).
const THEME_COLOR = { light: '#F4F3EE', dark: '#0E0E0C' }

const ORDER: ThemeMode[] = ['auto', 'light', 'dark']

export function nextTheme(mode: ThemeMode): ThemeMode {
  return ORDER[(ORDER.indexOf(mode) + 1) % ORDER.length] ?? 'auto'
}

export function parseTheme(value: string | null): ThemeMode {
  return value === 'light' || value === 'dark' ? value : 'auto'
}

export function readTheme(): ThemeMode {
  try {
    return parseTheme(localStorage.getItem(THEME_STORAGE_KEY))
  } catch {
    return 'auto'
  }
}

export function applyTheme(mode: ThemeMode): void {
  const root = document.documentElement
  if (mode === 'auto') root.removeAttribute('data-theme')
  else root.setAttribute('data-theme', mode)

  // Browser chrome (PWA status bar): forced theme overrides both media-scoped metas.
  document.querySelectorAll<HTMLMetaElement>('meta[name="theme-color"]').forEach((meta) => {
    const scheme = meta.media.includes('dark') ? 'dark' : 'light'
    meta.content = THEME_COLOR[mode === 'auto' ? scheme : mode]
  })

  try {
    if (mode === 'auto') localStorage.removeItem(THEME_STORAGE_KEY)
    else localStorage.setItem(THEME_STORAGE_KEY, mode)
  } catch {
    // Storage blocked: the choice lasts until reload.
  }
}
