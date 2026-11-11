-- Hevy -> Garmin export (docs/design.md, "Integracje"): which Garmin activity carries the sets of a Hevy workout.
-- Not a projection: it records side effects in Garmin (sets written, activities created), so it is never rebuilt.

create table hevy_garmin_link (
  hevy_workout_id    text primary key,
  workout_hash       text not null,               -- body_hash of the Hevy version last written to Garmin
  garmin_activity_id bigint,
  created_by_us      boolean not null default false,  -- manual activity (no watch recording); deleted with the workout
  status             text not null check (status in ('synced', 'deleted', 'error')),
  error              text,
  attempts           int not null default 0,
  updated_at         timestamptz not null default now()
);
