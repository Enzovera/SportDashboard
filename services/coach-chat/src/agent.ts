import { mkdirSync } from 'node:fs'
import { query, type Options, type SDKMessage } from '@anthropic-ai/claude-agent-sdk'
import type { Config } from './config.ts'
import { COACH_SYSTEM_PROMPT } from './prompt.ts'

export interface TurnInput {
  prompt: string
  /** Agent SDK session id from the previous turn; null starts a new conversation. */
  resume: string | null
  signal: AbortSignal
}

/** One chat turn as a stream of SDK messages. Injected so the chat logic can be tested without Claude. */
export type TurnRunner = (input: TurnInput) => AsyncIterable<SDKMessage>

/** query() options from docs/design.md ("Czat z trenerem"): MCP-only, no built-in tools, no prompts. */
export function buildOptions(config: Config, resume: string | null, abortController: AbortController): Options {
  const headers: Record<string, string> = {}
  if (config.mcp?.token) headers.authorization = `Bearer ${config.mcp.token}`

  return {
    model: config.model,
    effort: config.effort,
    systemPrompt: COACH_SYSTEM_PROMPT,
    // No built-in tools at all (Bash, Read, Write, WebFetch, ...): the coach only reaches data through MCP.
    tools: [],
    mcpServers: config.mcp ? { forma: { type: 'http', url: config.mcp.url, headers } } : {},
    strictMcpConfig: true,
    allowedTools: ['mcp__forma'],
    // Loose chat questions must not end up in the coach log; the daily run owns save_coach_note.
    disallowedTools: ['mcp__forma__save_coach_note'],
    // Anything not pre-approved above is denied instead of waiting for an approval nobody can give.
    permissionMode: 'dontAsk',
    settingSources: [],
    includePartialMessages: true,
    maxTurns: config.maxTurns,
    cwd: config.workdir,
    ...(resume ? { resume } : {}),
    abortController,
  }
}

export function sdkRunner(config: Config): TurnRunner {
  mkdirSync(config.workdir, { recursive: true })
  return ({ prompt, resume, signal }) => {
    const abortController = new AbortController()
    signal.addEventListener('abort', () => abortController.abort(), { once: true })
    return query({ prompt, options: buildOptions(config, resume, abortController) })
  }
}
