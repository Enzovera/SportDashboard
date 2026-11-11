import assert from 'node:assert/strict'
import type { AddressInfo } from 'node:net'
import { after, before, describe, it } from 'node:test'
import { buildOptions } from './agent.ts'
import { loadConfig } from './config.ts'
import { resolveCredentials } from './secrets.ts'
import { createChatServer, parseTurnBody } from './server.ts'
import { fake, MemoryStore, scriptedRunner } from './testing.ts'

describe('parseTurnBody', () => {
  const cases: [unknown, string | null][] = [
    [{ text: ' Hej ' }, null],
    [{ text: 'Hej', screen: { view: 'plan', label: 'tydzień' } }, null],
    [{ text: '' }, 'text is required'],
    [{ text: 'x'.repeat(4001) }, 'text longer than 4000 characters'],
    [{ text: 'Hej', screen: { label: 'x' } }, 'screen must be {view: string, label?: string}'],
    [null, 'expected an object'],
  ]
  for (const [body, error] of cases) {
    it(`${JSON.stringify(body)?.slice(0, 50)} -> ${error ?? 'ok'}`, () => {
      if (error) assert.throws(() => parseTurnBody(body), { message: error })
      else assert.equal(parseTurnBody(body).text, 'Hej')
    })
  }
})

describe('buildOptions', () => {
  it('locks the coach to MCP tools only', () => {
    const env = { DATABASE_URL: 'x', FORMA_MCP_URL: 'http://mcp:8091/mcp', FORMA_MCP_TOKEN: 't', CLAUDE_CODE_OAUTH_TOKEN: 'oat' }
    const o = buildOptions(loadConfig(env), resolveCredentials(env, {}), 'sess', new AbortController())
    assert.deepEqual(o.tools, [])
    assert.deepEqual(o.mcpServers, { forma: { type: 'http', url: 'http://mcp:8091/mcp', headers: { authorization: 'Bearer t' } } })
    assert.equal(o.strictMcpConfig, true)
    assert.equal(o.permissionMode, 'dontAsk')
    assert.deepEqual(o.settingSources, [])
    assert.deepEqual(o.disallowedTools, ['mcp__forma__save_coach_note'])
    assert.equal(o.resume, 'sess')
    assert.equal(o.model, 'claude-opus-5-5')
    assert.equal(o.env?.CLAUDE_CODE_OAUTH_TOKEN, 'oat')
    assert.equal(o.env?.DATABASE_URL, undefined)
  })

  it('starts a new conversation without resume and runs without MCP when not configured', () => {
    const o = buildOptions(loadConfig({ DATABASE_URL: 'x' }), resolveCredentials({}, {}), null, new AbortController())
    assert.equal('resume' in o, false)
    assert.deepEqual(o.mcpServers, {})
  })
})

describe('HTTP + SSE', () => {
  const store = new MemoryStore()
  const runner = scriptedRunner(() => [fake.init('s1'), fake.toolUse('mcp__forma__get_day_summary'), fake.delta('HRV w normie.'), fake.result('HRV w normie.')])
  const server = createChatServer({ store, runner, ready: async () => ({ auth: true, mcp: true }) })
  let base = ''

  before(async () => {
    await new Promise<void>((r) => server.listen(0, '127.0.0.1', r))
    base = `http://127.0.0.1:${(server.address() as AddressInfo).port}`
  })
  after(() => server.close())

  it('creates a session, streams a turn and lists history', async () => {
    const created = await fetch(`${base}/api/chat/sessions`, { method: 'POST' })
    assert.equal(created.status, 201)
    const { id } = (await created.json()) as { id: string }

    const resp = await fetch(`${base}/api/chat/sessions/${id}/messages`, {
      method: 'POST',
      headers: { 'content-type': 'application/json' },
      body: JSON.stringify({ text: 'Jak gotowość?', screen: { view: 'dzis' } }),
    })
    assert.equal(resp.headers.get('content-type'), 'text/event-stream; charset=utf-8')
    const body = await resp.text()
    assert.equal(
      body,
      'event: tool\ndata: {"name":"get_day_summary"}\n\n' +
        'event: delta\ndata: {"text":"HRV w normie."}\n\n' +
        'event: done\ndata: {"messageId":2}\n\n',
    )

    const history = (await (await fetch(`${base}/api/chat/sessions/${id}/messages`)).json()) as { messages: { role: string; tools: string[] }[] }
    assert.deepEqual(history.messages.map((m) => [m.role, m.tools]), [['user', []], ['assistant', ['get_day_summary']]])
  })

  it('returns 404 for unknown sessions and 400 for bad bodies', async () => {
    const missing = await fetch(`${base}/api/chat/sessions/00000000-0000-4000-8000-999999999999/messages`, {
      method: 'POST',
      body: JSON.stringify({ text: 'x' }),
    })
    assert.equal(missing.status, 404)
    const { id } = (await (await fetch(`${base}/api/chat/sessions`, { method: 'POST' })).json()) as { id: string }
    const bad = await fetch(`${base}/api/chat/sessions/${id}/messages`, { method: 'POST', body: '{"text": ""}' })
    assert.equal(bad.status, 400)
  })

  it('refuses to chat without Claude credentials', async () => {
    const noAuth = createChatServer({ store, runner, ready: async () => ({ auth: false, mcp: false }) })
    await new Promise<void>((r) => noAuth.listen(0, '127.0.0.1', r))
    const url = `http://127.0.0.1:${(noAuth.address() as AddressInfo).port}`
    const resp = await fetch(`${url}/api/chat/sessions/x/messages`, { method: 'POST', body: '{"text":"x"}' })
    assert.equal(resp.status, 503)
    noAuth.close()
  })
})
