import { createServer, type IncomingMessage, type Server, type ServerResponse } from 'node:http'
import type { TurnRunner } from './agent.ts'
import { runChatTurn, SessionNotFoundError } from './chat.ts'
import type { ChatEvent } from './events.ts'
import type { ScreenContext } from './prompt.ts'
import type { ChatStore } from './store.ts'

// Reached only from the dashboard API / Vite proxy inside the docker network; auth is Cloudflare Access in front.
// Routes:
//   GET  /health
//   POST /api/chat/sessions                    -> 201 {id}
//   GET  /api/chat/sessions/:id/messages       -> {messages}
//   POST /api/chat/sessions/:id/messages       {text, screen?} -> text/event-stream of ChatEvent

const MAX_BODY = 16 * 1024
const MAX_TEXT = 4000
const HEARTBEAT_MS = 15_000

export interface ServerDeps {
  store: ChatStore
  runner: TurnRunner
  ready: { auth: boolean; mcp: boolean }
}

class HttpError extends Error {
  readonly status: number
  constructor(status: number, message: string) {
    super(message)
    this.status = status
  }
}

function json(res: ServerResponse, status: number, body: unknown) {
  res.writeHead(status, { 'content-type': 'application/json; charset=utf-8' })
  res.end(JSON.stringify(body))
}

async function readJson(req: IncomingMessage): Promise<unknown> {
  let size = 0
  const chunks: Buffer[] = []
  for await (const chunk of req) {
    size += chunk.length
    if (size > MAX_BODY) throw new HttpError(413, 'body too large')
    chunks.push(chunk)
  }
  try {
    return JSON.parse(Buffer.concat(chunks).toString('utf8') || '{}')
  } catch {
    throw new HttpError(400, 'invalid JSON')
  }
}

export function parseTurnBody(body: unknown): { text: string; screen: ScreenContext | null } {
  if (typeof body !== 'object' || body === null) throw new HttpError(400, 'expected an object')
  const { text, screen } = body as Record<string, unknown>
  if (typeof text !== 'string' || !text.trim()) throw new HttpError(400, 'text is required')
  if (text.length > MAX_TEXT) throw new HttpError(400, `text longer than ${MAX_TEXT} characters`)
  if (screen === undefined || screen === null) return { text: text.trim(), screen: null }
  const s = screen as Record<string, unknown>
  if (typeof s.view !== 'string' || (s.label !== undefined && typeof s.label !== 'string')) {
    throw new HttpError(400, 'screen must be {view: string, label?: string}')
  }
  return { text: text.trim(), screen: { view: s.view.slice(0, 40), ...(s.label ? { label: String(s.label).slice(0, 120) } : {}) } }
}

export function sseFrame(event: ChatEvent): string {
  const { type, ...data } = event
  return `event: ${type}\ndata: ${JSON.stringify(data)}\n\n`
}

export function createChatServer(deps: ServerDeps): Server {
  // One running turn per conversation: a second message waits for the first answer.
  const busy = new Set<string>()

  async function streamTurn(req: IncomingMessage, res: ServerResponse, sessionId: string) {
    if (!deps.ready.auth) throw new HttpError(503, 'Claude credentials are not configured')
    const input = parseTurnBody(await readJson(req))
    if (!(await deps.store.getSession(sessionId))) throw new HttpError(404, 'session not found')
    if (busy.has(sessionId)) throw new HttpError(409, 'previous message is still being answered')

    busy.add(sessionId)
    const abort = new AbortController()
    res.on('close', () => {
      if (!res.writableFinished) abort.abort()
    })
    res.writeHead(200, {
      'content-type': 'text/event-stream; charset=utf-8',
      'cache-control': 'no-cache, no-transform',
      'x-accel-buffering': 'no',
      connection: 'keep-alive',
    })
    const heartbeat = setInterval(() => res.write(': ping\n\n'), HEARTBEAT_MS)
    try {
      await runChatTurn(deps, { sessionId, ...input }, (e) => res.write(sseFrame(e)), abort.signal)
    } finally {
      clearInterval(heartbeat)
      busy.delete(sessionId)
      res.end()
    }
  }

  async function route(req: IncomingMessage, res: ServerResponse) {
    const url = new URL(req.url ?? '/', 'http://local')
    const parts = url.pathname.split('/').filter(Boolean)

    if (req.method === 'GET' && url.pathname === '/health') {
      return json(res, 200, { ok: true, ...deps.ready })
    }
    if (parts[0] !== 'api' || parts[1] !== 'chat' || parts[2] !== 'sessions') throw new HttpError(404, 'not found')

    if (parts.length === 3 && req.method === 'POST') {
      const session = await deps.store.createSession()
      return json(res, 201, { id: session.id })
    }
    const sessionId = parts[3]
    if (parts.length === 5 && sessionId && parts[4] === 'messages') {
      if (req.method === 'GET') {
        if (!(await deps.store.getSession(sessionId))) throw new HttpError(404, 'session not found')
        return json(res, 200, { messages: await deps.store.listMessages(sessionId) })
      }
      if (req.method === 'POST') return streamTurn(req, res, sessionId)
    }
    throw new HttpError(404, 'not found')
  }

  return createServer((req, res) => {
    route(req, res).catch((err) => {
      const status = err instanceof HttpError ? err.status : err instanceof SessionNotFoundError ? 404 : 500
      if (status === 500) console.error('coach-chat:', err)
      if (res.headersSent) return res.end()
      json(res, status, { error: status === 500 ? 'internal error' : err.message })
    })
  })
}
