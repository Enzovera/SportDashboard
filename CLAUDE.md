# Forma — sportowy dashboard + AI trener

Osobista aplikacja jednego użytkownika (Patryk). Zbiera dane z Garmina, Hevy, Fitatu i wagi Xiaomi do jednej bazy, pokazuje je na dashboardzie zoptymalizowanym pod Galaxy Z Fold6 i daje AI trenera (Claude), który codziennie podsumowuje dzień, pilnuje planu do maratonu i układa żywienie pod trening.

Stan: **etap 1 w toku.** Jest środowisko dev (`deploy/dev/compose.yml`), migracje (`raw_payload`, `sync_state`, `ingest_run`, projekcje) i ingest Garmina i Hevy (`services/ingest`). Dalej: projektory i API w Go, widoki Dziś i Trening. Komendy dev w `README.md`.

## Gdzie co jest

- `docs/design.md` — dokument projektowy, źródło prawdy: integracje, model danych, stan trenera, metryki, runy AI, narzędzia MCP, czat (Agent SDK), dashboard, suplementacja, walidatory, hosting (Mikrus), roadmapa. Czytaj odpowiednią sekcję przed pracą nad danym obszarem.
- `design/screens/*.html` — statyczne makiety wszystkich ekranów (jasne i `-dark`), z prawdziwymi wartościami przykładowymi. Traktuj je jako referencję struktury i stylu, nie jako kod produkcyjny.
- `design/png/*.png` — zrzuty tych samych ekranów (2×), do szybkiego podglądu.
- `design/tokens.css`, `design/tokens.json` — kolory (jasny/ciemny), typografia, layout.
- `design/README.md` — zasady designu v2 i lista ekranów.
- Makiety źródłowe (interaktywne): https://claude.ai/artifact/KtXo6nwS3hwJma1m7nRzpi

## Architektura (skrót z docs/design.md)

| Część | Język | Rola |
|---|---|---|
| `ingest` | Python (cron) | workery Garmin (python-garminconnect), Hevy (oficjalne API v1, Pro), Fitatu (fitatu-api); zapis surowych odpowiedzi |
| `api` | Go | API dashboardu, projekcje, akceptacja zmian planu, SSE dla czatu |
| `mcp` | Go | serwer MCP — **jedyna droga AI do danych** |
| `coach-chat` | TypeScript | czat z trenerem przez Claude Agent SDK (`query()` z `resume`) |
| `web` | do ustalenia (propozycja: Vue + ECharts) | dashboard, PWA pod Fold6 |
| baza | PostgreSQL (własny kontener) | `raw_payload` + projekcje + zdarzenia planu |
| deploy | docker compose na Mikrus 3.0 | Cloudflare Tunnel + Cloudflare Access, nocny `pg_dump` poza serwer |

Runy AI: dzienny 6:30 i tygodniowy niedziela 20:00 przez `claude -p` z podpiętym serwerem MCP. Auth: `CLAUDE_CODE_OAUTH_TOKEN` (subskrypcja); ustawiony `ANTHROPIC_API_KEY` ma pierwszeństwo.

## Twarde zasady (nie łam bez rozmowy)

- `raw_payload` jest append-only; idempotencja po `(source, endpoint, external_id, body_hash)`. Wszystko inne to projekcje, które da się skasować i odbudować.
- Plan jest event-sourced: `plan_event` (`PlanCreated`, `PlanChangeProposed`, `PlanChangeAccepted`, `PlanChangeRejected`), `plan_session` to projekcja. Zmianę planu akceptuje tylko człowiek w dashboardzie.
- Walidatory (kilometraż +10%/tydz., długie ≤ 35%, dzień wolny, taper, dolny próg kcal, brak deficytu w dniu długiego, białko ≥ 1,6 g/kg) żyją w kodzie, nie w prompcie. Odrzucona propozycja nie trafia do kolejki, trener dostaje powód.
- Metryki (ATL/CTL/TSB, kilometraż, objętość siłowa, gotowość, bilans żywienia, średnia wagi 7 dni) liczy kod. Model dostaje gotowe agregaty i nigdy sam nie liczy.
- AI nie pisze do Garmina, Hevy ani Fitatu. Serwer MCP czyta projekcje i zapisuje tylko notatki trenera i propozycje.
- Suplementy: trener może tylko proponować pory przypomnień; nie dodaje suplementów i nie zmienia dawek. Białko serwatkowe liczy się wyłącznie w Fitatu (bez podwójnego liczenia).
- Sekrety (token Garmina, sesja Fitatu, klucz Hevy, tokeny Claude) tylko w zmiennych środowiskowych / menedżerze sekretów, nigdy w repo.

## Konwencje

- Identyfikatory, nazwy tabel i commity po angielsku; teksty UI po polsku.
- Go: standardowy layout, `context` w każdym I/O, testy tabelaryczne; migracje SQL wersjonowane.
- Zdarzenia nazywaj w czasie przeszłym, payload w `jsonb`, wersjonowany.
- Proponowany układ repo: `services/{ingest,api,mcp,coach-chat}`, `web/`, `migrations/`, `deploy/`, `docs/`, `design/`.
- Nie dodawaj zależności, które dublują to, co już jest w stacku, bez uzasadnienia.

## Design (v2) — w skrócie

Pełne zasady w `design/README.md`. Najważniejsze: zero zaokrągleń, boksy z ramką 1px, krój Archivo z wąskimi liczbami i `tabular-nums`, jeden kolor sygnałowy (czerwień) tylko dla „dziś / teraz / wymaga decyzji”, ostrzeżenia na bursztynowo, tabele zamiast pasków postępu, nawigacja tekstowa bez ikon, brak mikrotekstów tłumaczących działanie aplikacji. Teksty trenera mogą być prozą LLM.

Fold6: ekran wewnętrzny 707 × 823 CSS px, dwie kolumny po obu stronach zgięcia (odstęp 24, margines 16), nawigacja 3 | 3. Ekran zewnętrzny 369 px szerokości, jedna kolumna. Cele dotykowe ≥ 44 px.

## Roadmapa

1. **Dane** — ingest Garmina i Hevy do `raw_payload`, projekcje dnia i aktywności, widoki Dziś (bez trenera) i Trening. Dalej dopiero po 2 tygodniach stabilnego syncu.
2. **Żywienie i ciało** — Fitatu, waga przez Garmina, metryki, widoki Żywienie (z suplementami) i Ciało.
3. **Trener v1** — MCP (odczyt + `save_coach_note`), run dzienny z pushem, czat (Agent SDK).
4. **Plan i propozycje** — profil, `plan_event`, walidatory, widoki Plan i Propozycje, run tygodniowy.
5. **Dopracowanie** — alerty, raport tygodniowy, Health Connect jako zapasowe źródło.

## Otwarte decyzje

- Framework frontendu (Vue + ECharts to tylko propozycja).
- Telefon i model wagi — czy Mi Scale Exporter zadziała.
- Plan Claude: Pro czy Max (limity vs kredyty API dla runów).
- Mikrus 3.0 czy 3.5 — zależy od realnego zużycia RAM przez czat.
- Maraton 14.03.2027 jest jeszcze w rozważaniu.

## Współpraca

Patryk to senior backend (Go, PHP/Symfony, event sourcing, CQRS, serverless AWS). Rozmawiamy po polsku, zwięźle. Zanim zaczniesz większą zmianę, pokaż plan; przy decyzjach architektonicznych odwołuj się do `docs/design.md` i aktualizuj go, gdy decyzja się zmienia.
