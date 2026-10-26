-- Coach log (docs/design.md, "Stan trenera"): summaries and decisions written by AI runs through MCP save_coach_note.
-- Append-only: a re-run adds a new note, readers take the latest per (date, kind) when they need one.
create table coach_note (
  id         bigserial primary key,
  date       date not null,
  kind       text not null check (kind in ('daily', 'weekly', 'decision')),
  content    text not null check (length(content) between 1 and 8000),
  created_at timestamptz not null default now()
);

create index coach_note_date_idx on coach_note (date, id);
