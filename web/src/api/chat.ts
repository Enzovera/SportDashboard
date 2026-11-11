// Client for services/coach-chat (docs/design.md "Czat z trenerem").
// Always talks to the real backend under /api/chat; `?chat=mock` in the URL switches to a scripted
// conversation (design/screens/FoldCzat, CoverCzat) for design checks without Claude.

/** Screen the message was written from, e.g. {view: 'plan', label: 'Plan'}. */
export interface ChatScreen {
  view: string
  label?: string
}

export type ChatRole = 'user' | 'assistant'
export type ChatMessageStatus = 'ok' | 'error' | 'aborted'

/** Stored message as returned by GET /api/chat/sessions/:id/messages. */
export interface ChatMessage {
  id: number
  role: ChatRole
  content: string
  screen: ChatScreen | null
  /** MCP tool names without the `mcp__forma__` prefix, e.g. `get_plan`. */
  tools: string[]
  status: ChatMessageStatus
  createdAt: string
}

/** One SSE event of a running turn (event name = `type`). */
export type ChatEvent =
  | { type: 'delta'; text: string }
  | { type: 'tool'; name: string }
  | { type: 'done'; messageId: number }
  | { type: 'error'; message: string }

export class ChatHttpError extends Error {
  readonly status: number
  constructor(status: number, message: string) {
    super(message)
    this.status = status
  }
}

const BASE = '/api/chat/sessions'

/** Polish message for a failed request; the backend's own error texts are English. */
export function httpErrorMessage(status: number): string {
  switch (status) {
    case 409:
      return 'Trener jeszcze odpowiada na poprzednią wiadomość.'
    case 503:
      return 'Czat nie ma dostępu do Claude.'
    case 404:
      return 'Rozmowa nie istnieje.'
    case 400:
    case 413:
      return 'Wiadomość jest nieprawidłowa albo za długa.'
    default:
      return 'Trener nie odpowiedział. Spróbuj ponownie.'
  }
}

async function ensureOk(resp: Response): Promise<Response> {
  if (!resp.ok) throw new ChatHttpError(resp.status, httpErrorMessage(resp.status))
  return resp
}

// ---------------------------------------------------------------------------------------------
// SSE parsing (EventSource cannot POST, so the stream is read from fetch).

export interface SseMessage {
  event: string
  data: string
}

/**
 * Incremental parser for text/event-stream. Feed raw bytes with `push`; chunks may split events,
 * lines and multi-byte UTF-8 characters anywhere. Comments (`: ping`) are ignored.
 */
export class SseParser {
  private readonly decoder = new TextDecoder('utf-8')
  private buffer = ''
  private event = ''
  private data: string[] = []

  push(chunk: Uint8Array | string): SseMessage[] {
    this.buffer += typeof chunk === 'string' ? chunk : this.decoder.decode(chunk, { stream: true })
    return this.drain()
  }

  /** Flushes the decoder; a trailing event without the closing blank line is dispatched too. */
  end(): SseMessage[] {
    this.buffer += this.decoder.decode()
    const out = this.drain()
    if (this.buffer) {
      this.line(this.buffer, out)
      this.buffer = ''
    }
    this.dispatch(out)
    return out
  }

  private drain(): SseMessage[] {
    const out: SseMessage[] = []
    let nl: number
    while ((nl = this.buffer.search(/\r\n|\n|\r/)) !== -1) {
      // A lone \r at the very end may be the first half of \r\n: wait for the next chunk.
      if (this.buffer[nl] === '\r' && nl === this.buffer.length - 1) break
      const len = this.buffer.startsWith('\r\n', nl) ? 2 : 1
      const line = this.buffer.slice(0, nl)
      this.buffer = this.buffer.slice(nl + len)
      this.line(line, out)
    }
    return out
  }

  private line(line: string, out: SseMessage[]) {
    if (line === '') return this.dispatch(out)
    if (line.startsWith(':')) return
    const colon = line.indexOf(':')
    const field = colon === -1 ? line : line.slice(0, colon)
    let value = colon === -1 ? '' : line.slice(colon + 1)
    if (value.startsWith(' ')) value = value.slice(1)
    if (field === 'event') this.event = value
    else if (field === 'data') this.data.push(value)
  }

  private dispatch(out: SseMessage[]) {
    if (this.data.length > 0) out.push({ event: this.event || 'message', data: this.data.join('\n') })
    this.event = ''
    this.data = []
  }
}

/** Maps a raw SSE message to a typed chat event; unknown or malformed events are dropped. */
export function toChatEvent(msg: SseMessage): ChatEvent | null {
  let data: Record<string, unknown>
  try {
    data = JSON.parse(msg.data) as Record<string, unknown>
  } catch {
    return null
  }
  switch (msg.event) {
    case 'delta':
      return typeof data.text === 'string' ? { type: 'delta', text: data.text } : null
    case 'tool':
      return typeof data.name === 'string' ? { type: 'tool', name: data.name } : null
    case 'done':
      return typeof data.messageId === 'number' ? { type: 'done', messageId: data.messageId } : null
    case 'error':
      return { type: 'error', message: typeof data.message === 'string' ? data.message : httpErrorMessage(500) }
    default:
      return null
  }
}

/** Reads a fetch body as a stream of chat events. */
export async function* readChatEvents(body: ReadableStream<Uint8Array>): AsyncGenerator<ChatEvent> {
  const parser = new SseParser()
  const reader = body.getReader()
  try {
    for (;;) {
      const { done, value } = await reader.read()
      const messages = done ? parser.end() : parser.push(value)
      for (const m of messages) {
        const ev = toChatEvent(m)
        if (ev) yield ev
      }
      if (done) return
    }
  } finally {
    reader.releaseLock()
  }
}

// ---------------------------------------------------------------------------------------------
// Transport

export interface ChatTransport {
  createSession(): Promise<string>
  getMessages(sessionId: string): Promise<ChatMessage[]>
  /** Streams one turn. Aborting `signal` closes the connection; the backend then stops the turn. */
  sendMessage(sessionId: string, input: { text: string; screen?: ChatScreen | null }, signal: AbortSignal): AsyncIterable<ChatEvent>
}

export const httpTransport: ChatTransport = {
  async createSession() {
    const resp = await ensureOk(await fetch(BASE, { method: 'POST', headers: { accept: 'application/json' } }))
    return ((await resp.json()) as { id: string }).id
  },

  async getMessages(sessionId) {
    const resp = await ensureOk(
      await fetch(`${BASE}/${encodeURIComponent(sessionId)}/messages`, { headers: { accept: 'application/json' } }),
    )
    return ((await resp.json()) as { messages: ChatMessage[] }).messages
  },

  async *sendMessage(sessionId, input, signal) {
    const resp = await ensureOk(
      await fetch(`${BASE}/${encodeURIComponent(sessionId)}/messages`, {
        method: 'POST',
        headers: { 'content-type': 'application/json', accept: 'text/event-stream' },
        body: JSON.stringify(input.screen ? { text: input.text, screen: input.screen } : { text: input.text }),
        signal,
      }),
    )
    if (!resp.body) throw new ChatHttpError(500, httpErrorMessage(500))
    yield* readChatEvents(resp.body)
  },
}

// ---------------------------------------------------------------------------------------------
// Scripted conversation for `?chat=mock` (texts from design/screens/DesktopCzat, FoldCzat).

function todayAt(h: number, m: number): string {
  const d = new Date()
  d.setHours(h, m, 0, 0)
  return d.toISOString()
}

function mockHistory(): ChatMessage[] {
  const screen = { view: 'today', label: 'Dziś' }
  return [
    { id: 1, role: 'user', content: 'Mogę dziś pobiec długie trochę szybciej niż 5:50?', screen, tools: [], status: 'ok', createdAt: todayAt(7, 12) },
    {
      id: 2,
      role: 'assistant',
      content:
        'Lepiej nie. Jesteś świeży (TSB +4, HRV 52 przy bazie 54), ale długie ma budować bazę — trzymaj **strefę 2, 5:50–6:10 /km**. Mocniejsze tempo zostaw na wtorkowe interwały.',
      screen: null,
      tools: ['get_day_summary', 'get_plan'],
      status: 'ok',
      createdAt: todayAt(7, 12),
    },
    { id: 3, role: 'user', content: 'We wtorek nie zdążę na interwały. Da się je przesunąć?', screen, tools: [], status: 'ok', createdAt: todayAt(13, 40) },
    {
      id: 4,
      role: 'assistant',
      content:
        'Da się: wtorek wolny, interwały w środę, Legs w czwartek. Kilometraż bez zmian.\n\nLegs bliżej weekendu daje 3 dni do długiego — przyjmij też czekającą propozycję 12 serii.',
      screen: null,
      tools: ['get_plan', 'get_trends', 'get_coach_log'],
      status: 'ok',
      createdAt: todayAt(13, 41),
    },
  ]
}

const MOCK_REPLIES: Record<string, { tools: string[]; text: string }> = {
  'Co zjeść przed biegiem?': {
    tools: ['get_plan', 'get_nutrition'],
    text: 'Dziś długie 16 km, więc **2–3 h przed**: 80–100 g węgli, mało tłuszczu i błonnika — owsianka z bananem i miodem.\n\nJeśli wychodzisz zaraz po wstaniu, wystarczy banan i kawa 30–45 min przed.',
  },
  'Zdążę na 1:40?': {
    tools: ['get_profile', 'get_activities', 'get_trends'],
    text: 'Na dziś realnie **1:43–1:45**. Tempo progowe 4:58 /km i CTL 41 to jeszcze za mało na 4:44 /km przez 21 km.\n\nDo startu jest 9 tygodni: przy +8% kilometrażu tygodniowo i dwóch akcentach 1:40 jest w zasięgu.',
  },
}

let mockId = 100
const sleep = (ms: number, signal: AbortSignal) =>
  new Promise<void>((resolve, reject) => {
    if (signal.aborted) return reject(new DOMException('Aborted', 'AbortError'))
    const t = setTimeout(resolve, ms)
    signal.addEventListener('abort', () => (clearTimeout(t), reject(new DOMException('Aborted', 'AbortError'))), { once: true })
  })

export const mockTransport: ChatTransport = {
  async createSession() {
    return 'mock'
  },
  async getMessages() {
    return mockHistory()
  },
  async *sendMessage(_sessionId, input, signal) {
    const reply = MOCK_REPLIES[input.text] ?? {
      tools: ['get_day_summary', 'get_plan'],
      text: 'To odpowiedź z makiety. Dziś trzymaj plan: **długie 16 km w strefie 2**, a po biegu 30 g białka w ciągu godziny.',
    }
    await sleep(400, signal)
    if (/błąd|error/i.test(input.text)) {
      yield { type: 'error', message: 'Trener nie odpowiedział. Spróbuj ponownie.' }
      return
    }
    for (const name of reply.tools) {
      yield { type: 'tool', name }
      await sleep(350, signal)
    }
    for (const word of reply.text.split(/(?<= )/)) {
      yield { type: 'delta', text: word }
      await sleep(35, signal)
    }
    yield { type: 'done', messageId: ++mockId }
  },
}

const MOCK_KEY = 'forma.chat.mock'

/** `?chat=mock` turns the scripted conversation on for this tab, `?chat=off` turns it off. */
export function chatMockEnabled(): boolean {
  if (import.meta.env.VITE_MOCK === '1') return true
  const param = new URLSearchParams(window.location.search).get('chat')
  try {
    if (param === 'mock') sessionStorage.setItem(MOCK_KEY, '1')
    else if (param === 'off') sessionStorage.removeItem(MOCK_KEY)
    return sessionStorage.getItem(MOCK_KEY) === '1'
  } catch {
    return param === 'mock'
  }
}

export function chatTransport(): ChatTransport {
  return chatMockEnabled() ? mockTransport : httpTransport
}
