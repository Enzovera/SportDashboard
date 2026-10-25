-- Raw API responses: append-only, the only source of truth.
create table raw_payload (
  id          bigserial primary key,
  source      text not null,      -- garmin | hevy | fitatu
  endpoint    text not null,      -- e.g. sleep, activities, workouts
  external_id text,
  for_date    date,
  body        jsonb not null,
  body_hash   text not null,      -- sha256 of canonical JSON
  fetched_at  timestamptz not null default now(),
  unique nulls not distinct (source, endpoint, external_id, body_hash)
);

create index raw_payload_source_endpoint_date_idx on raw_payload (source, endpoint, for_date);
create index raw_payload_fetched_at_idx on raw_payload (fetched_at);

-- Incremental sync cursors per source/endpoint.
create table sync_state (
  source     text not null,
  endpoint   text not null,
  cursor     text not null,
  updated_at timestamptz not null default now(),
  primary key (source, endpoint)
);

-- One row per ingest run; basis for the "2 weeks of stable sync" gate.
create table ingest_run (
  id          bigserial primary key,
  source      text not null,
  args        jsonb not null default '{}',
  started_at  timestamptz not null default now(),
  finished_at timestamptz,
  status      text not null default 'running',  -- running | ok | partial | error
  inserted    int not null default 0,
  unchanged   int not null default 0,
  errors      jsonb not null default '[]'
);

create index ingest_run_source_started_idx on ingest_run (source, started_at desc);
