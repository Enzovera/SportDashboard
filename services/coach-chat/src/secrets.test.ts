import assert from 'node:assert/strict'
import { describe, it } from 'node:test'
import { openValue, parseKey, resolveCredentials } from './secrets.ts'

// Vector shared with services/api/internal/secrets/secrets_test.go and services/ingest/tests/test_secrets.py.
const key = parseKey(Buffer.from(Array.from({ length: 32 }, (_, i) => i)).toString('base64'))!
const nonce = Buffer.from('a0a1a2a3a4a5a6a7a8a9aaab', 'hex')
const ct = Buffer.from('8e7d0a5468bf67cc1648ecb67e57f1ec4398dc023402d051cdcbaa7175fa95868d41', 'hex')

describe('openValue', () => {
  it('opens the Go vector', () => assert.equal(openValue(key, 'HEVY_API_KEY', nonce, ct), 'hevy-test-key-1234'))
  it('rejects a row moved to another name', () => assert.throws(() => openValue(key, 'ANTHROPIC_API_KEY', nonce, ct)))
  it('rejects a short key', () => assert.throws(() => parseKey(Buffer.alloc(16).toString('base64'))))
  it('treats an empty key as not configured', () => assert.equal(parseKey(' '), null))
})

describe('resolveCredentials', () => {
  const base = { PATH: '/bin', DATABASE_URL: 'pg', FORMA_SECRETS_KEY: 'k', FORMA_MCP_TOKEN: 'env-mcp', ANTHROPIC_API_KEY: 'env-api' }
  const cases: [string, NodeJS.ProcessEnv, Record<string, string>, { oauth?: string; api?: string; auth: boolean; mcp: string | null }][] = [
    ['env only', base, {}, { api: 'env-api', auth: true, mcp: 'env-mcp' }],
    ['app token replaces the env pair', base, { CLAUDE_CODE_OAUTH_TOKEN: 'app-oat' }, { oauth: 'app-oat', auth: true, mcp: 'env-mcp' }],
    ['app mcp token wins', base, { FORMA_MCP_TOKEN: 'app-mcp' }, { api: 'env-api', auth: true, mcp: 'app-mcp' }],
    ['nothing', { PATH: '/bin' }, {}, { auth: false, mcp: null }],
  ]
  for (const [name, env, stored, want] of cases) {
    it(name, () => {
      const c = resolveCredentials(env, stored)
      assert.equal(c.env.CLAUDE_CODE_OAUTH_TOKEN, want.oauth)
      assert.equal(c.env.ANTHROPIC_API_KEY, want.api)
      assert.equal(c.auth, want.auth)
      assert.equal(c.mcpToken, want.mcp)
      assert.equal(c.env.PATH, '/bin')
      for (const hidden of ['DATABASE_URL', 'FORMA_SECRETS_KEY', 'FORMA_MCP_TOKEN']) assert.equal(hidden in c.env, false)
    })
  }
})
