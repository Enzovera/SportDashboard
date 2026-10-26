-- Supplements (docs/design.md, "Suplementacja"). No source exports them: the set and rules are entered once,
-- the log is appended by ticking in the dashboard. Checklist, regularity and stock forecast are computed by the API.

create table supplement (
  id                text primary key,         -- slug, e.g. d3k2
  name              text not null,
  dose_amount       numeric(8, 2) not null,
  dose_unit         text not null check (dose_unit in ('IU', 'mg', 'g', 'capsule', 'tablet', 'sachet')),
  pack_unit         text not null check (pack_unit in ('IU', 'mg', 'g', 'capsule', 'tablet', 'sachet')),
  pack_per_dose     numeric(8, 2) not null,   -- pack units used by one dose, e.g. 2 capsules of omega-3
  counted_in_fitatu boolean not null default false,  -- whey: logged as a meal in Fitatu, here only ticked
  active_from       date not null,
  active_to         date,
  sort              int not null default 0
);

create table supplement_rule (
  id            bigserial primary key,
  supplement_id text not null references supplement (id),
  "when"        text not null check ("when" in ('daily', 'run_day', 'long_run', 'quality_run', 'strength_day', 'training_day')),
  slot          text not null check (slot in ('morning', 'before', 'during', 'after', 'evening')),
  remind_at     time,                         -- morning/evening reminders; null = tied to the session
  offset_min    int,                          -- before/after the session, e.g. -60 for caffeine
  note          text
);

create index supplement_rule_supplement_idx on supplement_rule (supplement_id);

-- Append-only: SupplementTaken / SupplementSkipped; the latest event of the day wins.
create table supplement_log (
  id            bigserial primary key,
  supplement_id text not null references supplement (id),
  date          date not null,
  type          text not null check (type in ('SupplementTaken', 'SupplementSkipped')),
  source        text not null default 'app' check (source in ('app', 'push')),
  payload       jsonb not null default '{"v": 1}',
  created_at    timestamptz not null default now()
);

create index supplement_log_date_idx on supplement_log (date, supplement_id, id);

-- Purchases / stock counts, in pack units. Stock left = latest entry minus doses taken after it.
create table supplement_stock (
  id            bigserial primary key,
  supplement_id text not null references supplement (id),
  amount        numeric(8, 2) not null,
  purchased_on  date not null,
  created_at    timestamptz not null default now()
);

create index supplement_stock_supplement_idx on supplement_stock (supplement_id, created_at);
