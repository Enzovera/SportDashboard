import pg from 'pg'
import { sdkRunner } from './agent.ts'
import { loadConfig } from './config.ts'
import { createChatServer } from './server.ts'
import { pgSecretSource, resolveCredentials } from './secrets.ts'
import { PgChatStore } from './store.ts'

const config = loadConfig()
const pool = new pg.Pool({ connectionString: config.databaseUrl, max: 4 })

const stored = pgSecretSource(pool, config.secretsKey)
const credentials = async () => resolveCredentials(process.env, await stored())

const server = createChatServer({
  store: new PgChatStore(pool),
  runner: sdkRunner(config, credentials),
  ready: async () => {
    const c = await credentials()
    return { auth: c.auth, mcp: config.mcp !== null && c.mcpToken !== null }
  },
})

server.listen(config.port, () => {
  console.log(
    `coach-chat: listening on :${config.port}, model ${config.model}, effort ${config.effort}, ` +
      `mcp ${config.mcp ? config.mcp.url : 'off'}, secrets key ${config.secretsKey ? 'set' : 'missing'}`,
  )
})

for (const signal of ['SIGTERM', 'SIGINT'] as const) {
  process.on(signal, () => {
    server.close()
    void pool.end().finally(() => process.exit(0))
  })
}
