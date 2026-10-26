-- On-demand sync from the dashboard ("Odśwież", docs/design.md, "Integracje"). The API only queues a request;
-- the ingest scheduler, which alone holds source credentials, claims it, runs the workers and marks it done.
-- Results per source are the ingest_run rows started after started_at.

create table sync_request (
  id           bigserial primary key,
  requested_at timestamptz not null default now(),
  started_at   timestamptz,
  finished_at  timestamptz,
  status       text not null default 'pending' check (status in ('pending', 'running', 'done'))
);

-- At most one active request: a second click joins the one already queued or running.
create unique index sync_request_active_idx on sync_request ((true)) where status in ('pending', 'running');
