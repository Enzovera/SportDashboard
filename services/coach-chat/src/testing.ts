// Test helpers: fake SDK messages, in-memory store, scripted runner. Not used in production code.
import type { SDKMessage } from '@anthropic-ai/claude-agent-sdk'
import type { TurnInput, TurnRunner } from './agent.ts'
import type { ChatMessage, ChatSession, ChatStore, NewMessage } from './store.ts'

const msg = (m: Record<string, unknown>) => m as unknown as SDKMessage

export const fake = {
  init: (sessionId: string) => msg({ type: 'system', subtype: 'init', session_id: sessionId }),
  textStart: (sessionId = 's1') =>
    msg({ type: 'stream_event', parent_tool_use_id: null, session_id: sessionId, event: { type: 'content_block_start', index: 0, content_block: { type: 'text', text: '' } } }),
  delta: (text: string, parent: string | null = null) =>
    msg({ type: 'stream_event', parent_tool_use_id: parent, session_id: 's1', event: { type: 'content_block_delta', index: 0, delta: { type: 'text_delta', text } } }),
  toolUse: (name: string) =>
    msg({ type: 'assistant', parent_tool_use_id: null, session_id: 's1', message: { content: [{ type: 'tool_use', id: 't1', name, input: {} }] } }),
  result: (text: string, cost = 0.12) =>
    msg({ type: 'result', subtype: 'success', is_error: false, result: text, total_cost_usd: cost, session_id: 's1' }),
  resultError: (subtype: string) =>
    msg({ type: 'result', subtype, is_error: true, total_cost_usd: 0.3, session_id: 's1' }),
}

export class MemoryStore implements ChatStore {
  sessions = new Map<string, ChatSession & { costUsdTotal: number | null }>()
  messages: (ChatMessage & { sessionId: string })[] = []
  private seq = 0

  async createSession() {
    const id = `00000000-0000-4000-8000-${String(this.sessions.size + 1).padStart(12, '0')}`
    const s = { id, sdkSessionId: null, title: null, costUsdTotal: null }
    this.sessions.set(id, s)
    return { id, sdkSessionId: null, title: null }
  }
  async getSession(id: string) {
    const s = this.sessions.get(id)
    return s ? { id: s.id, sdkSessionId: s.sdkSessionId, title: s.title } : null
  }
  async listMessages(sessionId: string) {
    return this.messages.filter((m) => m.sessionId === sessionId)
  }
  async addMessage(m: NewMessage) {
    const id = ++this.seq
    this.messages.push({ ...m, id, createdAt: new Date(0).toISOString() })
    return id
  }
  async updateSession(id: string, patch: { sdkSessionId: string | null; title: string | null; costUsdTotal: number | null }) {
    const s = this.sessions.get(id)!
    s.sdkSessionId = patch.sdkSessionId ?? s.sdkSessionId
    s.title = s.title ?? patch.title
    s.costUsdTotal = patch.costUsdTotal ?? s.costUsdTotal
  }
}

/** Runner that replays a fixed script per turn and records its inputs. */
export function scriptedRunner(script: (input: TurnInput) => SDKMessage[] | Error): TurnRunner & { calls: TurnInput[] } {
  const calls: TurnInput[] = []
  const runner = async function* (input: TurnInput) {
    calls.push(input)
    const out = script(input)
    if (out instanceof Error) throw out
    for (const m of out) {
      if (input.signal.aborted) throw new Error('aborted')
      yield m
      await new Promise((r) => setImmediate(r))
    }
  }
  return Object.assign(runner, { calls })
}
