import { computed, reactive, readonly, watch } from 'vue'
import type { RouteLocationNormalizedLoaded } from 'vue-router'
import { ChatHttpError, chatMockEnabled, chatTransport, httpErrorMessage, type ChatMessage, type ChatScreen } from '@/api/chat'
import { router } from '@/router'

/**
 * Something attached to a coach message besides text. Proposal cards come with the MCP server
 * (`propose_plan_change` from chat); until then nothing produces attachments.
 */
export type ChatAttachment = { kind: 'proposal'; proposalId: string }

export interface UiMessage {
  /** Stable key for v-for: stored id, or a local one while the turn is running. */
  key: string
  id: number | null
  role: 'user' | 'assistant'
  content: string
  tools: string[]
  status: 'ok' | 'error' | 'aborted' | 'streaming'
  /** Shown in warn under the message when status is 'error'. */
  error: string | null
  createdAt: string
  attachments: ChatAttachment[]
}

const SESSION_KEY = 'forma.chat.session'

/** Screen context per route; the chat route itself reports the view it was opened from. */
const SCREENS: Record<string, ChatScreen> = {
  today: { view: 'today', label: 'Dziś' },
  training: { view: 'training', label: 'Trening' },
  nutrition: { view: 'nutrition', label: 'Żywienie' },
  supplements: { view: 'supplements', label: 'Suplementy' },
  body: { view: 'body', label: 'Ciało' },
  plan: { view: 'plan', label: 'Plan' },
  proposals: { view: 'proposals', label: 'Propozycje' },
}

export function screenFor(routeName: unknown): ChatScreen | null {
  return typeof routeName === 'string' ? (SCREENS[routeName] ?? null) : null
}

function loadSessionId(): string | null {
  try {
    return localStorage.getItem(SESSION_KEY)
  } catch {
    return null
  }
}

function saveSessionId(id: string | null) {
  try {
    if (id) localStorage.setItem(SESSION_KEY, id)
    else localStorage.removeItem(SESSION_KEY)
  } catch {
    /* private mode: the conversation lives until reload */
  }
}

const mock = chatMockEnabled()
const transport = chatTransport()

const state = reactive({
  open: false,
  sessionId: (null as string | null),
  messages: [] as UiMessage[],
  /** History request in flight. */
  loading: false,
  /** A turn is streaming. */
  streaming: false,
  /** Failure outside a message (history could not be loaded). */
  error: null as string | null,
  /** Last non-chat view: the screen context and the target of "back" on the cover screen. */
  from: { path: '/', screen: (SCREENS.today ?? null) as ChatScreen | null },
})

let historyLoaded = false
let controller: AbortController | null = null
let localSeq = 0

function trackRoute(route: RouteLocationNormalizedLoaded) {
  // Skip the chat route itself and the router's initial START_LOCATION (nothing matched yet).
  if (route.name === 'chat' || route.matched.length === 0) return
  state.from = { path: route.fullPath, screen: screenFor(route.name) }
}
watch(router.currentRoute, trackRoute, { immediate: true })

function fromStored(m: ChatMessage): UiMessage {
  const failed = m.role === 'assistant' && m.status === 'error'
  return {
    key: `m${m.id}`,
    id: m.id,
    role: m.role,
    content: m.content,
    tools: m.tools ?? [],
    status: m.status,
    error: failed ? httpErrorMessage(500) : null,
    createdAt: m.createdAt,
    attachments: [],
  }
}

/** Loads the stored conversation once (on reload the session id comes from localStorage). */
async function loadHistory() {
  if (historyLoaded || state.loading) return
  historyLoaded = true
  if (!state.sessionId) return
  state.loading = true
  state.error = null
  try {
    const stored = await transport.getMessages(state.sessionId)
    // Keep anything sent while the history was loading.
    state.messages = [...stored.map(fromStored), ...state.messages]
  } catch (err) {
    if (err instanceof ChatHttpError && err.status === 404) {
      // Session is gone (e.g. fresh database): start a new one with the next message.
      state.sessionId = null
      saveSessionId(null)
    } else {
      historyLoaded = false
      state.error = 'Nie udało się wczytać rozmowy.'
    }
  } finally {
    state.loading = false
  }
}

async function ensureSession(): Promise<string> {
  if (state.sessionId) return state.sessionId
  const id = await transport.createSession()
  state.sessionId = id
  if (!mock) saveSessionId(id)
  return id
}

const BUSY_RETRIES = 4
const BUSY_WAIT_MS = 1500

function wait(ms: number, signal: AbortSignal): Promise<void> {
  return new Promise((resolve) => {
    const t = setTimeout(resolve, ms)
    signal.addEventListener('abort', () => (clearTimeout(t), resolve()), { once: true })
  })
}

async function runTurn(
  text: string,
  screen: ChatScreen | null,
  reply: UiMessage,
  retry: boolean,
  busyRetries = BUSY_RETRIES,
): Promise<void> {
  const signal = controller!.signal
  try {
    const sessionId = await ensureSession()
    let finished = false
    for await (const ev of transport.sendMessage(sessionId, { text, screen }, signal)) {
      if (ev.type === 'delta') reply.content += ev.text
      else if (ev.type === 'tool') reply.tools.push(ev.name)
      else if (ev.type === 'done') {
        reply.id = ev.messageId
        reply.status = 'ok'
        finished = true
      } else {
        reply.status = 'error'
        reply.error = ev.message
        finished = true
      }
    }
    if (!finished) {
      reply.status = 'error'
      reply.error = 'Połączenie z trenerem zostało przerwane.'
    }
  } catch (err) {
    if (signal.aborted) {
      reply.status = 'aborted'
      return
    }
    if (retry && err instanceof ChatHttpError && err.status === 404) {
      // Stored session id points at a conversation that no longer exists.
      state.sessionId = null
      saveSessionId(null)
      return runTurn(text, screen, reply, false, busyRetries)
    }
    if (busyRetries > 0 && err instanceof ChatHttpError && err.status === 409) {
      // After an abort the backend needs a moment to stop the previous turn; the message was not stored yet.
      await wait(BUSY_WAIT_MS, signal)
      if (signal.aborted) {
        reply.status = 'aborted'
        return
      }
      return runTurn(text, screen, reply, retry, busyRetries - 1)
    }
    reply.status = 'error'
    reply.error = err instanceof ChatHttpError ? err.message : 'Brak połączenia z trenerem.'
  }
}

async function send(raw: string): Promise<void> {
  const text = raw.trim()
  if (!text || state.streaming) return
  const screen = state.from.screen
  const now = new Date().toISOString()
  state.messages.push({
    key: `l${++localSeq}`, id: null, role: 'user', content: text, tools: [], status: 'ok', error: null, createdAt: now, attachments: [],
  })
  state.messages.push({
    key: `l${++localSeq}`, id: null, role: 'assistant', content: '', tools: [], status: 'streaming', error: null, createdAt: now, attachments: [],
  })
  // Mutate through the reactive array so the UI follows the stream.
  const reply = state.messages[state.messages.length - 1]!
  state.streaming = true
  controller = new AbortController()
  try {
    await runTurn(text, screen, reply, true)
  } finally {
    state.streaming = false
    controller = null
  }
}

function abort() {
  controller?.abort()
}

function setOpen(open: boolean) {
  state.open = open
  if (open) void loadHistory()
}

// The scripted conversation (?chat=mock) never touches the stored real session id.
state.sessionId = mock ? 'mock' : loadSessionId()

/** Shared chat state for the whole app (one conversation, one panel). */
export function useChat() {
  return {
    state: readonly(state),
    contextLabel: computed(() => state.from.screen?.label ?? null),
    send,
    abort,
    open: () => setOpen(true),
    close: () => setOpen(false),
    toggle: () => setOpen(!state.open),
    /** For the full-screen chat (cover), which shows the conversation without the panel. */
    loadHistory,
  }
}
