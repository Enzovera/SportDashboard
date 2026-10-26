// Pure helpers for the chat UI: tool trace, light markdown, day separators.
import { dayShort, localDate } from '@/lib/format'

/** Polish labels for MCP tool names (docs/design.md "Narzędzia MCP"). */
export const TOOL_LABELS: Record<string, string> = {
  get_day_summary: 'dzień',
  get_trends: 'trendy',
  get_activities: 'aktywności',
  get_strength_volume: 'objętość siłowa',
  get_nutrition: 'żywienie',
  get_supplements: 'suplementy',
  get_plan: 'plan',
  get_profile: 'profil',
  get_coach_log: 'dziennik trenera',
  save_coach_note: 'notatka trenera',
  propose_plan_change: 'propozycja zmiany',
}

export function toolLabel(name: string): string {
  const bare = name.replace(/^mcp__[^_]+__/, '')
  return TOOL_LABELS[bare] ?? bare.replace(/^get_/, '').replace(/_/g, ' ')
}

/** ['get_plan', 'get_trends', 'get_plan'] -> "Sprawdzono: plan, trendy"; no tools -> null. */
export function toolTrace(tools: readonly string[]): string | null {
  const labels = [...new Set(tools.map(toolLabel))]
  return labels.length ? `Sprawdzono: ${labels.join(', ')}` : null
}

// ---- light markdown: paragraphs, line breaks and **bold**; everything else stays literal text ----

export type Inline = { kind: 'text'; text: string; bold: boolean } | { kind: 'br' }
export type Paragraph = Inline[]

function inline(line: string): Inline[] {
  const out: Inline[] = []
  // Only closed pairs become bold; an unclosed "**" (e.g. mid-stream) stays literal.
  const re = /\*\*(?=\S)([\s\S]*?\S)\*\*/g
  let last = 0
  for (let m = re.exec(line); m; m = re.exec(line)) {
    if (m.index > last) out.push({ kind: 'text', text: line.slice(last, m.index), bold: false })
    out.push({ kind: 'text', text: m[1] ?? '', bold: true })
    last = m.index + m[0].length
  }
  if (last < line.length) out.push({ kind: 'text', text: line.slice(last), bold: false })
  return out
}

/** Model text -> paragraphs of text runs. Rendered with text interpolation only (no v-html). */
export function parseCoachText(text: string): Paragraph[] {
  return text
    .replace(/\r\n?/g, '\n')
    .split(/\n{2,}/)
    .map((p) => p.trim())
    .filter(Boolean)
    .map((p) =>
      p.split('\n').flatMap((line, i): Inline[] => (i === 0 ? inline(line) : [{ kind: 'br' }, ...inline(line)])),
    )
}

// ---- day separators ----

/** "dziś", "wczoraj" or "Nd 25.10" for the local (Warsaw) day of a timestamp. */
export function dayLabel(iso: string, now: Date = new Date()): string {
  const day = localDate(iso)
  if (day === localDate(now)) return 'dziś'
  const y = new Date(now.getTime() - 24 * 3600 * 1000)
  if (day === localDate(y)) return 'wczoraj'
  return dayShort(day)
}
