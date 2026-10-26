-- Dev seed: a marathon plan (PlanCreated) and proposals as in the mock-ups. Example data, not a training plan.
-- Runs only on an empty plan_event. Proposals are inserted directly here, bypassing the validators that
-- plan.Store.Propose (MCP propose_plan_change) runs; the checks are copied from the mock-ups.
-- `api project` (or the next projection tick of `api serve`) rewrites plan_session / plan_target afterwards.
begin;

do $$
declare
  stream uuid := gen_random_uuid();
  first_day date := date_trunc('week', current_date)::date;  -- Monday of this week
  race date := '2027-03-14';
  days jsonb;
begin
  if exists (select 1 from plan_event) then
    raise notice 'plan_event not empty, seed skipped';
    return;
  end if;

  -- Weekly pattern: Mo Push + easy 6, Tu intervals 10, We Legs, Th rest, Fr Pull + easy 6, Sa tempo 8, Su long 16.
  select jsonb_agg(jsonb_build_object(
           'date', d::date,
           'dayType', case extract(isodow from d) when 1 then 'strength_run' when 2 then 'run' when 3 then 'strength'
                                                  when 5 then 'strength_run' when 6 then 'run' else 'long_run' end,
           'sessions', case extract(isodow from d)
             when 1 then '[{"kind":"strength","title":"Push"},{"kind":"run","title":"bieg łatwy","distanceKm":6,"short":"bieg 6 km"}]'::jsonb
             when 2 then '[{"kind":"run","title":"Interwały","detail":"6×800 m","distanceKm":10}]'::jsonb
             when 3 then '[{"kind":"strength","title":"Legs","detail":"16 serii roboczych","short":"Legs 16 serii"}]'::jsonb
             when 5 then '[{"kind":"strength","title":"Pull"},{"kind":"run","title":"bieg łatwy","distanceKm":6,"short":"bieg 6 km"}]'::jsonb
             when 6 then '[{"kind":"run","title":"Tempo","distanceKm":8}]'::jsonb
             else '[{"kind":"run","title":"Długie wybieganie","detail":"strefa 2","distanceKm":16,"short":"Długie 16 km"}]'::jsonb
           end) order by d)
    into days
    from generate_series(first_day, race, interval '1 day') d
   where extract(isodow from d) <> 4;

  insert into plan_event (stream_id, seq, type, actor, created_at, payload) values
  (stream, 1, 'PlanCreated', 'user', now() - interval '8 days', jsonb_build_object(
    'v', 1,
    'goal', jsonb_build_object('race', 'Maraton', 'date', race, 'targetTimeS', 14400, 'targetPaceSPerKm', 341,
      'secondary', '[{"label":"Półmaraton","value":"< 1:40","detail":"najlepszy 1:47"},{"label":"Siła","value":"Push / Pull / Legs","detail":"3× w Bazie"}]'::jsonb),
    'blocks', jsonb_build_array(
      jsonb_build_object('kind', 'base',  'from', first_day + 7,  'to', first_day + 48, 'summary', '46 → ok. 60 km, głównie strefa 2. PPL 3×.'),
      jsonb_build_object('kind', 'build', 'from', first_day + 49, 'to', first_day + 90, 'summary', 'Tempo i interwały 2×. PPL 2–3×.'),
      jsonb_build_object('kind', 'peak',  'from', first_day + 91, 'to', race - 21,      'summary', 'Najdłuższe wybiegania, tempo 5:41 /km.'),
      jsonb_build_object('kind', 'taper', 'from', race - 20,      'to', race,           'summary', null)),
    'targets', '[
      {"dayType":"rest","kcal":2300,"proteinG":145,"carbsG":255,"fatG":80},
      {"dayType":"strength","kcal":2600,"proteinG":145,"carbsG":330,"fatG":80},
      {"dayType":"run","kcal":2700,"proteinG":145,"carbsG":355,"fatG":80},
      {"dayType":"strength_run","kcal":2750,"proteinG":145,"carbsG":370,"fatG":80},
      {"dayType":"long_run","kcal":2950,"proteinG":145,"carbsG":412,"fatG":80}]'::jsonb,
    'days', days)),

  (stream, 2, 'PlanChangeProposed', 'coach', now() - interval '1 day', jsonb_build_object(
    'v', 1, 'proposalId', 'p-legs', 'kind', 'strength', 'origin', 'daily_run',
    'title', 'Środa, Legs: 16 → 12 serii', 'summary', '12 serii',
    'rationale', 'W bloku Baza kilometraż rośnie, a Legs obciąża te same mięśnie co bieg. 12 serii utrzyma bodziec przy mniejszym zmęczeniu przed sobotnim tempem.',
    'effectiveFrom', null,
    'diff', '{"heading":"Serie robocze","rows":[
      {"label":"Przysiad","before":4,"after":3,"unit":null,"note":null},
      {"label":"Martwy ciąg rumuński","before":4,"after":3,"unit":null,"note":null},
      {"label":"Wykroki","before":4,"after":3,"unit":null,"note":null},
      {"label":"Łydki","before":4,"after":3,"unit":null,"note":null}],
      "total":{"label":"Razem","before":16,"after":12,"unit":null,"note":null}}'::jsonb,
    'change', jsonb_build_object('days', jsonb_build_array(jsonb_build_object('date', first_day + 9, 'dayType', 'strength',
      'sessions', '[{"kind":"strength","title":"Legs","detail":"12 serii roboczych","short":"Legs 12 serii"}]'::jsonb))),
    'checks', '[{"rule":"rest_day","status":"pass","text":"dzień wolny: czwartek"},{"rule":"km_increase","status":"pass","text":"kilometraż bez zmian (46 km)"}]'::jsonb)),

  (stream, 3, 'PlanChangeProposed', 'coach', now() - interval '2 hours', jsonb_build_object(
    'v', 1, 'proposalId', 'p-carbs', 'kind', 'nutrition', 'origin', 'daily_run',
    'title', 'Dzień długiego wybiegania: węgle 412 → 460 g', 'summary', 'W 460 g, 3 140 kcal',
    'rationale', 'Na dwóch ostatnich długich wybieganiach tempo spadało po 13 km, a węgle wyniosły 4,8 i 5,0 g/kg. Wyższy cel i posiłek przed biegiem powinny utrzymać tempo.',
    'effectiveFrom', first_day + 13,
    'diff', '{"heading":"cel dnia","rows":[
      {"label":"Węglowodany","before":412,"after":460,"unit":"g","note":"6,0 g/kg"},
      {"label":"Kalorie","before":2950,"after":3140,"unit":"kcal","note":null},
      {"label":"Białko","before":145,"after":145,"unit":"g","note":null},
      {"label":"Tłuszcz","before":80,"after":80,"unit":"g","note":null},
      {"label":"Przed biegiem","before":null,"after":"80 g węgli","unit":null,"note":"2–3 h przed"}],"total":null}'::jsonb,
    'change', '{"targets":[{"dayType":"long_run","kcal":3140,"carbsG":460}]}'::jsonb,
    'checks', '[{"rule":"kcal_floor","status":"pass","text":"kcal powyżej dolnego progu"},{"rule":"long_run_deficit","status":"pass","text":"brak deficytu w dniu długiego"},{"rule":"protein_min","status":"pass","text":"białko ≥ 1,6 g/kg"}]'::jsonb)),

  (stream, 4, 'PlanChangeProposed', 'coach', now() - interval '1 hour', jsonb_build_object(
    'v', 1, 'proposalId', 'p-magnesium-time', 'kind', 'nutrition', 'origin', 'weekly_run',
    'title', 'Magnez: przypomnienie 20:00 → 21:30', 'summary', 'Magnez 21:30',
    'rationale', 'Magnez pomijasz głównie po wieczornych treningach (7 z 9 pominięć). Przesunąć przypomnienie na 21:30?',
    'effectiveFrom', null,
    'diff', '{"heading":"przypomnienie","rows":[{"label":"Magnez","before":"20:00","after":"21:30","unit":null,"note":null}],"total":null}'::jsonb,
    'change', '{"reminder":{"supplementId":"magnesium","supplementName":"Magnez","from":"20:00","to":"21:30"}}'::jsonb,
    'checks', '[{"rule":"reminder_time","status":"pass","text":"przypomnienie 20:00 → 21:30"}]'::jsonb));

  insert into plan_proposal_blocked (id, stream_id, kind, origin, title, checks, payload, proposed_at) values
  ('p-long-22', stream, 'run', 'daily_run', 'Niedziela: długie 16 → 22 km',
   '[{"rule":"km_increase","status":"fail","text":"przyrost 46 → 52 km (+13%, limit 10%)"},{"rule":"long_run_share","status":"warn","text":"długie = 42% tygodnia (limit 35%)"}]',
   '{"v":1}', now() - interval '2 hours');
end $$;

commit;
