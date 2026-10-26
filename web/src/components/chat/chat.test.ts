import { describe, expect, it } from 'vitest'
import { readChatEvents, SseParser, toChatEvent, type ChatEvent } from '@/api/chat'
import { dayLabel, parseCoachText, toolTrace } from './chatText'

const enc = new TextEncoder()

/** Frames exactly as services/coach-chat writes them (sseFrame). */
function frame(event: ChatEvent): string {
  const { type, ...data } = event
  return `event: ${type}\ndata: ${JSON.stringify(data)}\n\n`
}

const EVENTS: ChatEvent[] = [
  { type: 'tool', name: 'get_plan' },
  { type: 'delta', text: 'Zjedz owsiankę — 80 g węgli' },
  { type: 'delta', text: ' i łyżkę miodu. Żółć, źdźbło 🏃' },
  { type: 'done', messageId: 42 },
]
const STREAM = enc.encode(': ping\n\n' + EVENTS.map(frame).join(''))

function parseChunks(chunks: Uint8Array[]): ChatEvent[] {
  const p = new SseParser()
  const msgs = [...chunks.flatMap((c) => p.push(c)), ...p.end()]
  return msgs.map(toChatEvent).filter((e): e is ChatEvent => e !== null)
}

function splitAt(bytes: Uint8Array, ...cuts: number[]): Uint8Array[] {
  const out: Uint8Array[] = []
  let last = 0
  for (const c of [...cuts, bytes.length]) {
    out.push(bytes.slice(last, c))
    last = c
  }
  return out
}

describe('SseParser', () => {
  it('parses a whole stream and ignores heartbeat comments', () => {
    expect(parseChunks([STREAM])).toEqual(EVENTS)
  })

  it('handles every single split point (mid-event, mid-line, mid-UTF-8 character)', () => {
    for (let i = 1; i < STREAM.length; i++) {
      expect(parseChunks(splitAt(STREAM, i)), `split at ${i}`).toEqual(EVENTS)
    }
  })

  it('handles byte-by-byte delivery', () => {
    expect(parseChunks(Array.from(STREAM, (b) => Uint8Array.of(b)))).toEqual(EVENTS)
  })

  it('splits inside a multi-byte character', () => {
    const bytes = enc.encode(frame({ type: 'delta', text: 'ż' }))
    const at = bytes.indexOf(0xc5) + 1 // between the two bytes of "ż"
    expect(parseChunks(splitAt(bytes, at))).toEqual([{ type: 'delta', text: 'ż' }])
    const emoji = enc.encode(frame({ type: 'delta', text: '🏃' }))
    const start = emoji.indexOf(0xf0)
    expect(parseChunks(splitAt(emoji, start + 1, start + 2, start + 3))).toEqual([{ type: 'delta', text: '🏃' }])
  })

  it('accepts CRLF line endings, also split between CR and LF', () => {
    const crlf = enc.encode(frame({ type: 'tool', name: 'get_trends' }).replace(/\n/g, '\r\n'))
    const cr = crlf.indexOf(13)
    expect(parseChunks(splitAt(crlf, cr + 1))).toEqual([{ type: 'tool', name: 'get_trends' }])
  })

  it('joins multi-line data and dispatches a trailing event without blank line', () => {
    const p = new SseParser()
    expect(p.push('event: error\ndata: {"message":\ndata: "Błąd"}')).toEqual([])
    expect(p.end()).toEqual([{ event: 'error', data: '{"message":\n"Błąd"}' }])
  })

  it('drops unknown and malformed events', () => {
    expect(toChatEvent({ event: 'delta', data: 'not json' })).toBeNull()
    expect(toChatEvent({ event: 'other', data: '{}' })).toBeNull()
    expect(toChatEvent({ event: 'done', data: '{"messageId":"x"}' })).toBeNull()
  })
})

describe('readChatEvents', () => {
  it('reads events from a ReadableStream', async () => {
    const chunks = splitAt(STREAM, 3, 40, 41, 90)
    const body = new ReadableStream<Uint8Array>({
      start(c) {
        chunks.forEach((ch) => c.enqueue(ch))
        c.close()
      },
    })
    const got: ChatEvent[] = []
    for await (const ev of readChatEvents(body)) got.push(ev)
    expect(got).toEqual(EVENTS)
  })
})

describe('toolTrace', () => {
  it('labels MCP tools in Polish, deduplicated in order', () => {
    expect(toolTrace(['get_plan', 'get_trends', 'get_plan', 'get_coach_log'])).toBe('Sprawdzono: plan, trendy, dziennik trenera')
    expect(toolTrace(['mcp__forma__get_nutrition', 'get_new_thing'])).toBe('Sprawdzono: żywienie, new thing')
    expect(toolTrace([])).toBeNull()
  })
})

describe('parseCoachText', () => {
  it('splits paragraphs and line breaks, bolds closed pairs only', () => {
    expect(parseCoachText('Trzymaj **strefę 2**.\nPotem lekko.\n\nDrugi akapit **niedomknięty')).toEqual([
      [
        { kind: 'text', text: 'Trzymaj ', bold: false },
        { kind: 'text', text: 'strefę 2', bold: true },
        { kind: 'text', text: '.', bold: false },
        { kind: 'br' },
        { kind: 'text', text: 'Potem lekko.', bold: false },
      ],
      [{ kind: 'text', text: 'Drugi akapit **niedomknięty', bold: false }],
    ])
  })

  it('keeps markup-like text literal (rendered by interpolation, never as HTML)', () => {
    expect(parseCoachText('<b>x</b> & **<i>y</i>**')).toEqual([
      [
        { kind: 'text', text: '<b>x</b> & ', bold: false },
        { kind: 'text', text: '<i>y</i>', bold: true },
      ],
    ])
  })

  it('returns nothing for empty text', () => {
    expect(parseCoachText('')).toEqual([])
    expect(parseCoachText('\n\n')).toEqual([])
  })
})

describe('dayLabel', () => {
  const now = new Date('2026-10-25T12:00:00Z')
  it('names today and yesterday, otherwise weekday + date', () => {
    expect(dayLabel('2026-10-25T05:12:00Z', now)).toBe('dziś')
    expect(dayLabel('2026-10-24T18:00:00Z', now)).toBe('wczoraj')
    expect(dayLabel('2026-10-21T08:00:00Z', now)).toBe('Śr 21.10')
  })
})
