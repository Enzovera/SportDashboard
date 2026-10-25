import type { SDKMessage } from '@anthropic-ai/claude-agent-sdk'

/** Events streamed to the dashboard over SSE (event name = `type`). */
export type ChatEvent =
  | { type: 'delta'; text: string }
  | { type: 'tool'; name: string }
  | { type: 'done'; messageId: number }
  | { type: 'error'; message: string }

const MCP_PREFIX = 'mcp__forma__'

/** Accumulates one turn while SDK messages arrive. */
export class TurnState {
  sdkSessionId: string | null = null
  text = ''
  tools: string[] = []
  resultText: string | null = null
  costUsdTotal: number | null = null
  failed = false
}

const ERRORS: Record<string, string> = {
  error_max_turns: 'Trener przekroczył limit kroków na jedną wiadomość.',
  error_max_budget_usd: 'Przekroczony budżet tej rozmowy.',
}

/** Maps one SDK message to zero or more chat events, updating `st`. Only the main agent is shown (no subagents). */
export function consume(msg: SDKMessage, st: TurnState): ChatEvent[] {
  if ('session_id' in msg && typeof msg.session_id === 'string' && msg.session_id) st.sdkSessionId = msg.session_id

  switch (msg.type) {
    case 'stream_event': {
      if (msg.parent_tool_use_id !== null) return []
      const ev = msg.event
      // Separate text written after a tool call from the text before it.
      if (ev.type === 'content_block_start' && ev.content_block.type === 'text' && st.text && !st.text.endsWith('\n')) {
        st.text += '\n\n'
        return [{ type: 'delta', text: '\n\n' }]
      }
      if (ev.type === 'content_block_delta' && ev.delta.type === 'text_delta' && ev.delta.text) {
        st.text += ev.delta.text
        return [{ type: 'delta', text: ev.delta.text }]
      }
      return []
    }
    case 'assistant': {
      if (msg.parent_tool_use_id !== null) return []
      const events: ChatEvent[] = []
      for (const block of msg.message.content) {
        if (block.type === 'tool_use') {
          const name = block.name.startsWith(MCP_PREFIX) ? block.name.slice(MCP_PREFIX.length) : block.name
          st.tools.push(name)
          events.push({ type: 'tool', name })
        }
      }
      return events
    }
    case 'result': {
      st.costUsdTotal = msg.total_cost_usd
      if (msg.subtype === 'success' && !msg.is_error) {
        st.resultText = msg.result
        return []
      }
      st.failed = true
      const message =
        ERRORS[msg.subtype] ?? (msg.subtype === 'success' && msg.result ? msg.result : 'Trener nie odpowiedział. Spróbuj ponownie.')
      return [{ type: 'error', message }]
    }
    default:
      return []
  }
}
