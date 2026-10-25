-- Projections built from raw_payload. Safe to truncate and rebuild.

create table day_summary (
  date              date primary key,
  sleep_s           int,
  sleep_score       int,
  hrv_last_night_ms int,
  hrv_weekly_avg_ms int,
  hrv_status        text,
  resting_hr        int,
  steps             int,
  active_kcal       int,
  total_kcal        int,
  weight_kg         numeric(5, 2),
  updated_at        timestamptz not null default now()
);

create table activity (
  source           text not null,
  external_id      text not null,
  started_at       timestamptz not null,
  local_date       date not null,
  sport            text not null,
  name             text,
  distance_m       numeric(10, 1),
  duration_s       numeric(10, 1),
  moving_s         numeric(10, 1),
  avg_hr           int,
  max_hr           int,
  elevation_gain_m numeric(7, 1),
  training_load    numeric(7, 1),
  hr_zones_s       jsonb,          -- seconds per HR zone, index 0 = zone 1
  updated_at       timestamptz not null default now(),
  primary key (source, external_id)
);

create index activity_local_date_idx on activity (local_date);

create table workout (
  id         text primary key,     -- Hevy workout id
  title      text,
  started_at timestamptz not null,
  ended_at   timestamptz,
  local_date date not null,
  volume_kg  numeric(10, 1),
  updated_at timestamptz not null default now()
);

create index workout_local_date_idx on workout (local_date);

create table strength_set (
  workout_id           text not null references workout (id) on delete cascade,
  exercise_index       int not null,
  set_no               int not null,
  exercise_template_id text,
  exercise             text not null,
  muscle_group         text,
  set_type             text,        -- normal | warmup | dropset | failure
  weight_kg            numeric(6, 2),
  reps                 int,
  rpe                  numeric(3, 1),
  distance_m           numeric(10, 1),
  duration_s           int,
  primary key (workout_id, exercise_index, set_no)
);
