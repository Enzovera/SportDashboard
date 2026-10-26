-- Nutrition from Fitatu (diet_day + settings). Projections, rebuilt from raw_payload.

-- One row per day with at least one logged product. Targets come from the Fitatu diet settings in force that
-- day; plan targets per day type (stage 4) override them in the API.
create table nutrition_day (
  date             date primary key,
  kcal             numeric(6, 0) not null,
  protein_g        numeric(5, 1) not null,
  carbs_g          numeric(5, 1) not null,
  fat_g            numeric(5, 1) not null,
  kcal_target      numeric(6, 0),
  protein_target_g numeric(5, 1),
  carbs_target_g   numeric(5, 1),
  fat_target_g     numeric(5, 1),
  updated_at       timestamptz not null default now()
);

-- Fitatu meal slots with at least one product.
create table nutrition_meal (
  date      date not null,
  slot      text not null,      -- breakfast | second_breakfast | lunch | dinner | snack | supper
  slot_no   int  not null,      -- Fitatu meal order
  at        timestamptz,        -- meal time set in Fitatu; null when not set
  title     text not null,      -- product names joined
  kcal      numeric(6, 0) not null,
  protein_g numeric(5, 1) not null,
  carbs_g   numeric(5, 1) not null,
  fat_g     numeric(5, 1) not null,
  primary key (date, slot)
);
