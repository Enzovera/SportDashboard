import pg from 'pg'
import { sdkRunner } from './agent.ts'
import { loadConfig } from './config.ts'
import { createChatServer } from './server.ts'
import { PgChatStore } from './store.ts'

const config = loadConfig()
const pool = new pg.Pool({ connectionString: config.databaseUrl, max: 4 })

const server = createChatServer({
  store: new PgChatStore(pool),
  runner: sdkRunner(config),
  ready: { auth: config.authConfigured, mcp: config.mcp !== null },
})

server.listen(config.port, () => {
  console.log(
    `coach-chat: listening on :${config.port}, model ${config.model}, effort ${config.effort}, ` +
      `mcp ${config.mcp ? config.mcp.url : 'off'}, auth ${config.authConfigured ? 'ok' : 'missing'}`,
  )
})

for (const signal of ['SIGTERM', 'SIGINT'] as const) {
  process.on(signal, () => {
    server.close()
    void pool.end().finally(() => process.exit(0))
  })
}
