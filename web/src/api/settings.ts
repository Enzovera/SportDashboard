// Credentials set from the dashboard ("Ustawienia", docs/design.md, "Bezpieczeństwo").
// Values are write-only: the API stores them encrypted (AES-256-GCM, FORMA_SECRETS_KEY) and returns only
// status. A value stored here wins over the environment variable of the same name.
//
// Endpoints (services/api, Go):
//   GET    /api/settings/secrets                         -> SecretSettings
//   PUT    /api/settings/secrets/{name}  body {value}    -> SecretStatus   (422 invalid, 503 no FORMA_SECRETS_KEY)
//   DELETE /api/settings/secrets/{name}                  -> SecretStatus   (back to the env variable)
//   POST   /api/settings/secrets/FORMA_MCP_TOKEN/generate -> GeneratedSecret (the only response with a value)

import { mockMode } from './client'

export type SecretName = 'HEVY_API_KEY' | 'CLAUDE_CODE_OAUTH_TOKEN' | 'ANTHROPIC_API_KEY' | 'FORMA_MCP_TOKEN'

export interface SecretStatus {
  name: SecretName
  /** Stored in the app. False = services use the env variable, if there is one. */
  set: boolean
  /** Last 4 characters; null for short values or when not set. */
  hint: string | null
  updatedAt: string | null
}

export interface SecretSettings {
  /** False when the API has no FORMA_SECRETS_KEY: nothing can be stored. */
  writable: boolean
  secrets: SecretStatus[]
}

export interface GeneratedSecret extends SecretStatus {
  value: string
}

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const resp = await fetch(path, {
    ...init,
    headers: { accept: 'application/json', ...(init?.body ? { 'content-type': 'application/json' } : {}) },
  })
  if (!resp.ok) {
    const body = (await resp.json().catch(() => null)) as { error?: string } | null
    throw new Error(body?.error ?? `HTTP ${resp.status}`)
  }
  return (await resp.json()) as T
}

const path = (name: SecretName) => `/api/settings/secrets/${encodeURIComponent(name)}`

// ---------- mock (?mock=1): in memory, never leaves the browser ----------

const mockState: SecretSettings = {
  writable: true,
  secrets: [
    { name: 'HEVY_API_KEY', set: true, hint: '4f2a', updatedAt: '2026-11-02T19:14:00+01:00' },
    { name: 'CLAUDE_CODE_OAUTH_TOKEN', set: true, hint: 'Qw8x', updatedAt: '2026-10-20T08:01:00+02:00' },
    { name: 'ANTHROPIC_API_KEY', set: false, hint: null, updatedAt: null },
    { name: 'FORMA_MCP_TOKEN', set: false, hint: null, updatedAt: null },
  ],
}

const delay = <T>(value: T) => new Promise<T>((resolve) => setTimeout(() => resolve(structuredClone(value)), 150))

function mockPut(name: SecretName, value: string | null): SecretStatus {
  const s = mockState.secrets.find((x) => x.name === name)!
  Object.assign(s, value === null
    ? { set: false, hint: null, updatedAt: null }
    : { set: true, hint: value.length >= 12 ? value.slice(-4) : null, updatedAt: new Date().toISOString() })
  return s
}

// ---------- client ----------

export function getSecretSettings(): Promise<SecretSettings> {
  if (mockMode() !== 'off') return delay(mockState)
  return request('/api/settings/secrets')
}

export function putSecret(name: SecretName, value: string): Promise<SecretStatus> {
  if (mockMode() !== 'off') return delay(mockPut(name, value))
  return request(path(name), { method: 'PUT', body: JSON.stringify({ value }) })
}

export function deleteSecret(name: SecretName): Promise<SecretStatus> {
  if (mockMode() !== 'off') return delay(mockPut(name, null))
  return request(path(name), { method: 'DELETE' })
}

export function generateMcpToken(): Promise<GeneratedSecret> {
  if (mockMode() !== 'off') {
    const value = Array.from(crypto.getRandomValues(new Uint8Array(32)), (b) => b.toString(36).slice(-1)).join('') + 'mock000000a'
    return delay({ ...mockPut('FORMA_MCP_TOKEN', value), value })
  }
  return request(`${path('FORMA_MCP_TOKEN')}/generate`, { method: 'POST' })
}
