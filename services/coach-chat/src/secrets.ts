import { createDecipheriv } from 'node:crypto'
import type pg from 'pg'

// Credentials set from the dashboard ("Ustawienia", docs/design.md, "Bezpieczeństwo"). The API
// (services/api/internal/secrets) stores them in app_secret with AES-256-GCM under FORMA_SECRETS_KEY:
// 12-byte nonce, ciphertext with the 16-byte tag appended, the secret's name as additional data.
// They are read on every turn, so a new key works without restarting the container.

export const CLAUDE_SECRETS = ['CLAUDE_CODE_OAUTH_TOKEN', 'ANTHROPIC_API_KEY'] as const
export const MCP_SECRET = 'FORMA_MCP_TOKEN'
const NAMES = [...CLAUDE_SECRETS, MCP_SECRET]

/** Never handed to the Claude Code process: it has no tools, but it still does not need them. */
const HIDDEN_FROM_AGENT = ['FORMA_SECRETS_KEY', 'DATABASE_URL', MCP_SECRET]

export function parseKey(value: string | undefined): Buffer | null {
  if (!value?.trim()) return null
  const key = Buffer.from(value.trim(), 'base64')
  if (key.length !== 32) throw new Error(`FORMA_SECRETS_KEY must be 32 bytes, got ${key.length}`)
  return key
}

export function openValue(key: Buffer, name: string, nonce: Buffer, ciphertext: Buffer): string {
  if (nonce.length !== 12 || ciphertext.length < 16) throw new Error(`invalid stored ${name}`)
  const d = createDecipheriv('aes-256-gcm', key, nonce)
  d.setAAD(Buffer.from(name))
  d.setAuthTag(ciphertext.subarray(ciphertext.length - 16))
  return Buffer.concat([d.update(ciphertext.subarray(0, ciphertext.length - 16)), d.final()]).toString('utf8')
}

/** Values stored in the app, by name. */
export type SecretSource = () => Promise<Record<string, string>>

export function pgSecretSource(pool: pg.Pool, key: Buffer | null): SecretSource {
  let warned = false
  return async () => {
    const { rows } = await pool.query('select name, nonce, ciphertext from app_secret where name = any($1)', [NAMES])
    if (!key) {
      if (rows.length && !warned) console.warn('coach-chat: secrets stored in the app, but FORMA_SECRETS_KEY is not set')
      warned = true
      return {}
    }
    return Object.fromEntries(rows.map((r) => [r.name, openValue(key, r.name, r.nonce, r.ciphertext)]))
  }
}

export interface Credentials {
  /** Environment of the Claude Code process. */
  env: Record<string, string | undefined>
  /** Claude credentials present (subscription token or API key). */
  auth: boolean
  mcpToken: string | null
}

/**
 * App values win over the env. The Claude pair comes from one place: when either is stored in the app, both env
 * values are dropped, so switching between subscription and API key in Ustawienia is not overruled by the env
 * (where ANTHROPIC_API_KEY takes precedence over the token).
 */
export function resolveCredentials(base: NodeJS.ProcessEnv, stored: Record<string, string>): Credentials {
  const env: Record<string, string | undefined> = { ...base }
  for (const name of HIDDEN_FROM_AGENT) delete env[name]
  if (CLAUDE_SECRETS.some((n) => stored[n])) {
    for (const n of CLAUDE_SECRETS) {
      if (stored[n]) env[n] = stored[n]
      else delete env[n]
    }
  }
  return {
    env,
    auth: Boolean(env.CLAUDE_CODE_OAUTH_TOKEN || env.ANTHROPIC_API_KEY),
    mcpToken: stored[MCP_SECRET] || base[MCP_SECRET] || null,
  }
}
