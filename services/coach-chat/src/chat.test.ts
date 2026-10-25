import assert from 'node:assert/strict'
import { describe, it } from 'node:test'
import { runChatTurn, SessionNotFoundError } from './chat.ts'
import type { ChatEvent } from './events.ts'
import { fake, MemoryStore, scriptedRunner } from './testing.ts'

describe('runChatTurn', () => {
  it('persists both messages and resumes the SDK session on the next turn', async () => {
    const store = new MemoryStore()
    const { id } = await store.createSession()
    const runner = scriptedRunner(() => [fake.init('s1'), fake.delta('Odpoczywaj.'), fake.result('Odpoczywaj.')])
    const events: ChatEvent[] = []
    const screen = { view: 'plan', label: 'tydzień 26.10–1.11' }

    await runChatTurn({ store, runner }, { sessionId: id, text: 'Czy jutro biegać?', screen }, (e) => events.push(e), new AbortController().signal)
    await runChatTurn({ store, runner }, { sessionId: id, text: 'A pojutrze?', screen: null }, () => {}, new AbortController().signal)

    assert.equal(runner.calls[0]!.resume, null)
    assert.equal(runner.calls[0]!.prompt, '[Ekran: plan, tydzień 26.10–1.11]\n\nCzy jutro biegać?')
    assert.equal(runner.calls[1]!.resume, 's1')
    assert.deepEqual(events.at(-1), { type: 'done', messageId: 2 })

    const msgs = await store.listMessages(id)
    assert.deepEqual(
      msgs.map((m) => [m.role, m.content, m.status]),
      [
        ['user', 'Czy jutro biegać?', 'ok'],
        ['assistant', 'Odpoczywaj.', 'ok'],
        ['user', 'A pojutrze?', 'ok'],
        ['assistant', 'Odpoczywaj.', 'ok'],
      ],
    )
    const session = store.sessions.get(id)!
    assert.equal(session.title, 'Czy jutro biegać?') // from the first message only
    assert.equal(session.costUsdTotal, 0.12)
  })

  it('stores a failed turn and emits one error, no done', async () => {
    const store = new MemoryStore()
    const { id } = await store.createSession()
    const runner = scriptedRunner(() => new Error('spawn failed'))
    const events: ChatEvent[] = []
    await runChatTurn({ store, runner }, { sessionId: id, text: 'Hej', screen: null }, (e) => events.push(e), new AbortController().signal)

    assert.deepEqual(events, [{ type: 'error', message: 'Trener nie odpowiedział. Spróbuj ponownie.' }])
    assert.equal((await store.listMessages(id)).at(-1)!.status, 'error')
  })

  it('marks an aborted turn without emitting an error', async () => {
    const store = new MemoryStore()
    const { id } = await store.createSession()
    const abort = new AbortController()
    const runner = scriptedRunner(() => [fake.delta('Zacz'), fake.delta('ynam'), fake.result('Zaczynam')])
    const events: ChatEvent[] = []
    await runChatTurn({ store, runner }, { sessionId: id, text: 'Hej', screen: null }, (e) => {
      events.push(e)
      abort.abort()
    }, abort.signal)

    assert.deepEqual(events, [{ type: 'delta', text: 'Zacz' }])
    const last = (await store.listMessages(id)).at(-1)!
    assert.deepEqual([last.status, last.content], ['aborted', 'Zacz'])
  })

  it('rejects unknown sessions', async () => {
    const store = new MemoryStore()
    await assert.rejects(
      runChatTurn({ store, runner: scriptedRunner(() => []) }, { sessionId: 'nope', text: 'x', screen: null }, () => {}, new AbortController().signal),
      SessionNotFoundError,
    )
  })
})
