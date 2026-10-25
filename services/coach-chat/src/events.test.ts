import assert from 'node:assert/strict'
import { describe, it } from 'node:test'
import { consume, TurnState, type ChatEvent } from './events.ts'
import { fake } from './testing.ts'

function run(...msgs: ReturnType<typeof fake.delta>[]): { st: TurnState; events: ChatEvent[] } {
  const st = new TurnState()
  const events = msgs.flatMap((m) => consume(m, st))
  return { st, events }
}

describe('consume', () => {
  it('streams text deltas and keeps the SDK session id', () => {
    const { st, events } = run(fake.init('abc'), fake.delta('Cześć'), fake.delta('!'), fake.result('Cześć!'))
    assert.deepEqual(events, [
      { type: 'delta', text: 'Cześć' },
      { type: 'delta', text: '!' },
    ])
    assert.equal(st.sdkSessionId, 's1') // the last message wins; all carry the same id in practice
    assert.equal(st.resultText, 'Cześć!')
    assert.equal(st.costUsdTotal, 0.12)
    assert.equal(st.failed, false)
  })

  it('reports MCP tool calls without the server prefix', () => {
    const { st, events } = run(fake.toolUse('mcp__forma__get_trends'))
    assert.deepEqual(events, [{ type: 'tool', name: 'get_trends' }])
    assert.deepEqual(st.tools, ['get_trends'])
  })

  it('separates text after a tool call with a blank line', () => {
    const { events } = run(fake.textStart(), fake.delta('Sprawdzam.'), fake.toolUse('mcp__forma__get_plan'), fake.textStart(), fake.delta('Plan'))
    assert.deepEqual(
      events.filter((e) => e.type === 'delta').map((e) => (e as { text: string }).text),
      ['Sprawdzam.', '\n\n', 'Plan'],
    )
  })

  it('ignores subagent stream events', () => {
    assert.deepEqual(run(fake.delta('inner', 'tool-1')).events, [])
  })

  it('turns error results into a Polish error event', () => {
    const { st, events } = run(fake.resultError('error_max_turns'))
    assert.deepEqual(events, [{ type: 'error', message: 'Trener przekroczył limit kroków na jedną wiadomość.' }])
    assert.equal(st.failed, true)
  })
})
