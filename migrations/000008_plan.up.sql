-- Plan (docs/design.md, "Stan trenera"): event-sourced. plan_event is the source of truth; plan_session and
-- plan_target are projections rewritten after every appended event (and safe to rebuild from the events).

create table plan_event (
  stream_id  uuid not null,       -- macrocycle; the latest PlanCreated starts the current one
  seq        int  not null,
  type       text not null check (type in ('PlanCreated', 'PlanChangeProposed', 'PlanChangeAccepted', 'PlanChangeRejected')),
  payload    jsonb not null,      -- versioned: {"v": 1, ...}
  actor      text not null check (actor in ('user', 'coach')),
  created_at timestamptz not null default now(),
  primary key (stream_id, seq)
);

-- One decision per proposal: a second accept/reject collides here (HTTP 409).
create unique index plan_event_decision_idx on plan_event (stream_id, (payload ->> 'proposalId'))
  where type in ('PlanChangeAccepted', 'PlanChangeRejected');

-- Proposals a validator rejected: never queued in plan_event (they change nothing), kept so the dashboard can
-- show them and the coach gets the reason.
create table plan_proposal_blocked (
  id          text primary key,
  stream_id   uuid not null,
  kind        text not null,
  origin      text not null,
  title       text not null,
  checks      jsonb not null,
  payload     jsonb not null,
  proposed_at timestamptz not null default now()
);

create table plan_session (
  date        date not null,
  slot        int  not null,
  kind        text not null check (kind in ('run', 'strength')),
  title       text not null,
  detail      text,
  distance_km numeric(5, 1),
  day_type    text not null,
  primary key (date, slot)
);

-- Nutrition target per day type, from a date on (a later row for the same day type replaces it from its date).
create table plan_target (
  from_date date not null,
  day_type  text not null,
  kcal      int not null,
  protein_g int not null,
  carbs_g   int not null,
  fat_g     int not null,
  primary key (from_date, day_type)
);

-- Athlete profile (markdown), edited by hand, versioned.
create table athlete_profile (
  version    serial primary key,
  content    text not null,
  created_at timestamptz not null default now()
);
