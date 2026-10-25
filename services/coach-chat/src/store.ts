import pg from 'pg'
import type { ScreenContext } from './prompt.ts'

export interface ChatSession {
  id: string
  sdkSessionId: string | null
  title: string | null
}

export interface ChatMessage {
  id: number
  role: 'user' | 'assistant'
  content: string
  screen: ScreenContext | null
  tools: string[]
  status: 'ok' | 'error' | 'aborted'
  createdAt: string
}

export type NewMessage = Omit<ChatMessage, 'id' | 'createdAt'> & { sessionId: string }

export interface ChatStore {
  createSession(): Promise<ChatSession>
  getSession(id: string): Promise<ChatSession | null>
  listMessages(sessionId: string): Promise<ChatMessage[]>
  addMessage(m: NewMessage): Promise<number>
  /** Called after each turn: SDK session id, title from the first message, running cost estimate. */
  updateSession(id: string, patch: { sdkSessionId: string | null; title: string | null; costUsdTotal: number | null }): Promise<void>
}

const UUID = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i

export class PgChatStore implements ChatStore {
  private readonly pool: pg.Pool
  constructor(pool: pg.Pool) {
    this.pool = pool
  }

  async createSession(): Promise<ChatSession> {
    const { rows } = await this.pool.query('insert into chat_session default values returning id')
    return { id: rows[0].id, sdkSessionId: null, title: null }
  }

  async getSession(id: string): Promise<ChatSession | null> {
    if (!UUID.test(id)) return null
    const { rows } = await this.pool.query('select id, sdk_session_id, title from chat_session where id = $1', [id])
    const r = rows[0]
    return r ? { id: r.id, sdkSessionId: r.sdk_session_id, title: r.title } : null
  }

  async listMessages(sessionId: string): Promise<ChatMessage[]> {
    const { rows } = await this.pool.query(
      `select id, role, content, screen, tools, status, created_at
       from chat_message where session_id = $1 order by id`,
      [sessionId],
    )
    return rows.map((r) => ({
      id: Number(r.id),
      role: r.role,
      content: r.content,
      screen: r.screen,
      tools: r.tools,
      status: r.status,
      createdAt: r.created_at.toISOString(),
    }))
  }

  async addMessage(m: NewMessage): Promise<number> {
    const { rows } = await this.pool.query(
      `insert into chat_message (session_id, role, content, screen, tools, status)
       values ($1, $2, $3, $4, $5, $6) returning id`,
      [m.sessionId, m.role, m.content, m.screen, JSON.stringify(m.tools), m.status],
    )
    return Number(rows[0].id)
  }

  async updateSession(id: string, patch: { sdkSessionId: string | null; title: string | null; costUsdTotal: number | null }) {
    await this.pool.query(
      `update chat_session
       set sdk_session_id = coalesce($2, sdk_session_id), title = coalesce(title, $3),
           cost_usd_total = coalesce($4, cost_usd_total), updated_at = now()
       where id = $1`,
      [id, patch.sdkSessionId, patch.title, patch.costUsdTotal],
    )
  }
}
