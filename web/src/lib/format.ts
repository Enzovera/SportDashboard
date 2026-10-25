// Polish number/date formatting matching design/screens: "2 950", "8,0", "−5", "4:58", "Nd 25.10".

const NBSP = ' '
const MINUS = '−'
const TZ = 'Europe/Warsaw'

export const WEEKDAYS = ['Pn', 'Wt', 'Śr', 'Cz', 'Pt', 'Sb', 'Nd'] as const

/** 2950 -> "2 950", 8 -> "8,0" (decimals=1), -0.3 -> "−0,3". Groups thousands also for 4-digit numbers. */
export function num(value: number, decimals = 0): string {
  const fixed = Math.abs(value).toFixed(decimals)
  const [int = '0', frac] = fixed.split('.')
  const grouped = int.replace(/\B(?=(\d{3})+(?!\d))/g, NBSP)
  const isZero = Number(fixed) === 0
  return (value < 0 && !isZero ? MINUS : '') + grouped + (frac ? ',' + frac : '')
}

/** 4 -> "+4", -5 -> "−5", 0 -> "0". */
export function signed(value: number, decimals = 0): string {
  const s = num(value, decimals)
  return value > 0 && Number(Math.abs(value).toFixed(decimals)) !== 0 ? '+' + s : s
}

/** Seconds per km -> "4:58". */
export function pace(sPerKm: number): string {
  const total = Math.round(sPerKm)
  return `${Math.floor(total / 60)}:${String(total % 60).padStart(2, '0')}`
}

/** [350, 370] -> "5:50–6:10". */
export function paceRange([a, b]: [number, number]): string {
  return `${pace(a)}–${pace(b)}`
}

/** Seconds -> "7:40" (hours:minutes). */
export function hm(seconds: number): string {
  const m = Math.round(seconds / 60)
  return `${Math.floor(m / 60)}:${String(m % 60).padStart(2, '0')}`
}

/** Minutes -> "1 h 36 min", "45 min", "2 h". */
export function minutesWords(minutes: number): string {
  const m = Math.round(minutes)
  const h = Math.floor(m / 60)
  const rest = m % 60
  if (h === 0) return `${rest} min`
  return rest === 0 ? `${h} h` : `${h} h ${rest} min`
}

function parseDate(iso: string): Date {
  // Plain dates are local calendar days; parse at noon UTC to stay on the same day in any zone.
  return iso.length === 10 ? new Date(iso + 'T12:00:00Z') : new Date(iso)
}

function parts(iso: string) {
  const d = parseDate(iso)
  const day = new Intl.DateTimeFormat('en-GB', { timeZone: TZ, day: '2-digit' }).format(d)
  const month = new Intl.DateTimeFormat('en-GB', { timeZone: TZ, month: '2-digit' }).format(d)
  const wd = new Intl.DateTimeFormat('en-GB', { timeZone: TZ, weekday: 'short' }).format(d)
  const idx = ['Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat', 'Sun'].indexOf(wd)
  return { day, month, weekday: WEEKDAYS[idx] ?? '' }
}

/** "2026-10-25" -> "25.10". */
export function dayMonth(iso: string): string {
  const p = parts(iso)
  return `${p.day}.${p.month}`
}

/** "2026-10-25" -> "Nd 25.10". */
export function dayShort(iso: string): string {
  const p = parts(iso)
  return `${p.weekday} ${p.day}.${p.month}`
}

/** "2026-10-25" -> "Nd". */
export function weekday(iso: string): string {
  return parts(iso).weekday
}

/** ("2026-10-19", "2026-10-25") -> "19–25.10"; across months -> "28.09–04.10". */
export function dateRange(from: string, to: string): string {
  const a = parts(from)
  const b = parts(to)
  return a.month === b.month ? `${a.day}–${b.day}.${b.month}` : `${a.day}.${a.month}–${b.day}.${b.month}`
}

/** Timestamp -> "06:12" in Warsaw time. */
export function time(iso: string): string {
  return new Intl.DateTimeFormat('en-GB', { timeZone: TZ, hour: '2-digit', minute: '2-digit', hour12: false }).format(
    parseDate(iso),
  )
}

/** Local date of a timestamp, "YYYY-MM-DD". */
export function localDate(iso: string | Date): string {
  const d = typeof iso === 'string' ? parseDate(iso) : iso
  return new Intl.DateTimeFormat('en-CA', { timeZone: TZ }).format(d)
}
