import type { EffortLevel } from '@anthropic-ai/claude-agent-sdk'

export interface Config {
  port: number
  databaseUrl: string
  model: string
  effort: EffortLevel
  maxTurns: number
  /** Forma MCP server (services/mcp). Without it the coach has no data tools and must say so. */
  mcp: { url: string; token: string | null } | null
  /** Empty working directory for the Claude Code process; the chat never touches files. */
  workdir: string
  /** CLAUDE_CODE_OAUTH_TOKEN (subscription) or ANTHROPIC_API_KEY (API key wins when both are set). */
  authConfigured: boolean
}

const EFFORTS = ['low', 'medium', 'high', 'xhigh', 'max'] as const

export function loadConfig(env: NodeJS.ProcessEnv = process.env): Config {
  const required = (name: string) => {
    const v = env[name]
    if (!v) throw new Error(`missing environment variable ${name}`)
    return v
  }
  const effort = env.COACH_EFFORT ?? 'medium'
  if (!(EFFORTS as readonly string[]).includes(effort)) throw new Error(`invalid COACH_EFFORT ${effort}`)

  return {
    port: Number(env.PORT ?? 8090),
    databaseUrl: required('DATABASE_URL'),
    model: env.COACH_MODEL ?? 'claude-opus-5-5',
    effort: effort as EffortLevel,
    maxTurns: Number(env.COACH_MAX_TURNS ?? 8),
    mcp: env.FORMA_MCP_URL ? { url: env.FORMA_MCP_URL, token: env.FORMA_MCP_TOKEN || null } : null,
    workdir: env.COACH_WORKDIR ?? '/tmp/coach',
    authConfigured: Boolean(env.CLAUDE_CODE_OAUTH_TOKEN || env.ANTHROPIC_API_KEY),
  }
}
