-- Coach chat (Agent SDK). The SDK keeps the model-side transcript itself (resumed via sdk_session_id);
-- these tables hold what the dashboard shows and lets the session survive a lost SDK transcript.

create table chat_session (
  id             uuid primary key default gen_random_uuid(),
  sdk_session_id text,             -- Agent SDK session id passed as `resume`; null until the first turn
  title          text,
  cost_usd_total numeric(10, 4),   -- SDK's running estimate for the whole session (cumulative across resumes)
  created_at     timestamptz not null default now(),
  updated_at     timestamptz not null default now()
);

create table chat_message (
  id         bigserial primary key,
  session_id uuid not null references chat_session (id) on delete cascade,
  role       text not null check (role in ('user', 'assistant')),
  content    text not null,
  screen     jsonb,                -- dashboard context the user wrote from, e.g. {"view": "plan", "label": "tydzień 26.10–1.11"}
  tools      jsonb not null default '[]',  -- MCP tools the assistant called during this turn
  status     text not null default 'ok' check (status in ('ok', 'error', 'aborted')),
  created_at timestamptz not null default now()
);

create index chat_message_session_idx on chat_message (session_id, id);
