-- Dev seed: the supplement set from the mock-ups (example data, not recommendations).
-- Idempotent: re-running replaces the set and rules, keeps the log; stock is only added when empty.
begin;

insert into supplement (id, name, dose_amount, dose_unit, pack_unit, pack_per_dose, counted_in_fitatu, active_from, sort) values
  ('d3k2',         'Witamina D3 + K2',  2000, 'IU',      'capsule', 1, false, current_date, 10),
  ('omega3',       'Omega-3',              2, 'capsule', 'capsule', 2, false, current_date, 20),
  ('caffeine',     'Kofeina',            200, 'mg',      'tablet',  1, false, current_date, 30),
  ('electrolytes', 'Elektrolity',          1, 'sachet',  'sachet',  1, false, current_date, 40),
  ('creatine',     'Kreatyna',             5, 'g',       'g',       5, false, current_date, 50),
  ('whey',         'Białko serwatkowe',   30, 'g',       'g',      30, true,  '2026-09-01', 60),
  ('magnesium',    'Magnez',             300, 'mg',      'tablet',  1, false, current_date, 70)
on conflict (id) do update set name = excluded.name, dose_amount = excluded.dose_amount, dose_unit = excluded.dose_unit,
  pack_unit = excluded.pack_unit, pack_per_dose = excluded.pack_per_dose, counted_in_fitatu = excluded.counted_in_fitatu,
  sort = excluded.sort;

delete from supplement_rule;
insert into supplement_rule (supplement_id, "when", slot, remind_at, offset_min, note) values
  ('d3k2',         'daily',        'morning', '07:30', null, 'z posiłkiem'),
  ('omega3',       'daily',        'morning', '07:30', null, 'z posiłkiem'),
  ('caffeine',     'quality_run',  'before',  null,    -60,  '60 min przed'),
  ('electrolytes', 'long_run',     'during',  null,    null, 'bieg > 90 min'),
  ('creatine',     'daily',        'after',   null,    null, 'pora dowolna'),
  ('whey',         'training_day', 'after',   null,    null, null),
  ('magnesium',    'daily',        'evening', '20:00', null, 'przed snem');

insert into supplement_stock (supplement_id, amount, purchased_on)
select * from (values
  ('d3k2', 12::numeric, current_date), ('electrolytes', 6, current_date), ('magnesium', 20, current_date),
  ('omega3', 54, current_date), ('creatine', 205, current_date)) v
where not exists (select 1 from supplement_stock);

commit;
