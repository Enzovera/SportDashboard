-- Single weigh-ins from Garmin body_composition (dateWeightList). Projection, rebuilt from raw_payload.
-- day_summary.weight_kg keeps the day's average for the Today view and the 7-day mean.
create table body_composition (
  measured_at timestamptz primary key,
  local_date  date not null,
  weight_kg   numeric(5, 2) not null,
  fat_pct     numeric(4, 1),
  muscle_kg   numeric(5, 2),
  water_pct   numeric(4, 1),
  source_type text,             -- Garmin sourceType: INDEX_SCALE | MANUAL | USER_SETTING ...
  updated_at  timestamptz not null default now()
);

create index body_composition_local_date_idx on body_composition (local_date);
