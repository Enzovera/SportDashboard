import type { TurnRunner } from './agent.ts'
import { consume, TurnState, type ChatEvent } from './events.ts'
import { formatUserMessage, type ScreenContext } from './prompt.ts'
import type { ChatStore } from './store.ts'

export class SessionNotFoundError extends Error {}

const TITLE_MAX = 60

function titleFrom(text: string): string {
  const line = text.trim().split('\n')[0] ?? ''
  return line.length > TITLE_MAX ? line.slice(0, TITLE_MAX - 1) + '…' : line
}

/**
 * One user message -> one coach answer. Persists both messages, streams events through `emit`
 * and remembers the SDK session id so the next turn resumes the same conversation.
 */
export async function runChatTurn(
  deps: { store: ChatStore; runner: TurnRunner },
  input: { sessionId: string; text: string; screen: ScreenContext | null },
  emit: (event: ChatEvent) => void,
  signal: AbortSignal,
): Promise<void> {
  const { store, runner } = deps
  const session = await store.getSession(input.sessionId)
  if (!session) throw new SessionNotFoundError(input.sessionId)

  await store.addMessage({
    sessionId: session.id,
    role: 'user',
    content: input.text,
    screen: input.screen,
    tools: [],
    status: 'ok',
  })

  const st = new TurnState()
  try {
    const messages = runner({ prompt: formatUserMessage(input.text, input.screen), resume: session.sdkSessionId, signal })
    for await (const msg of messages) {
      for (const event of consume(msg, st)) emit(event)
    }
  } catch (err) {
    st.failed = true
    if (!signal.aborted) {
      console.error('coach-chat: turn failed', err)
      emit({ type: 'error', message: 'Trener nie odpowiedział. Spróbuj ponownie.' })
    }
  }

  const status = signal.aborted ? 'aborted' : st.failed ? 'error' : 'ok'
  await store.updateSession(session.id, {
    sdkSessionId: st.sdkSessionId,
    title: session.title ? null : titleFrom(input.text),
    costUsdTotal: st.costUsdTotal,
  })
  const messageId = await store.addMessage({
    sessionId: session.id,
    role: 'assistant',
    content: st.resultText ?? st.text,
    screen: null,
    tools: st.tools,
    status,
  })
  if (status === 'ok') emit({ type: 'done', messageId })
}
