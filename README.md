# Forma

Sportowy dashboard + AI trener. Projekt: [docs/design.md](docs/design.md), zasady pracy: [CLAUDE.md](CLAUDE.md).

## Środowisko dev (lokalnie, Docker Desktop)

Produkcja to Mikrus; lokalnie wszystko stoi w [deploy/dev/compose.yml](deploy/dev/compose.yml).

```bash
cp .env.example .env          # uzupełnij FORMA_SECRETS_KEY (openssl rand -base64 32), GARMIN_EMAIL
docker compose -f deploy/dev/compose.yml --env-file .env up -d
```

`up` startuje Postgresa (127.0.0.1:5432), adminera (http://localhost:8081), uruchamia migracje z `migrations/` i startuje `scheduler`: Hevy co godzinę o :05, Garmin (2 dni) o :10, eksport serii Hevy → Garmin o :25, backfill Garmina 7 dni o 3:00. Harmonogram jest w [scheduler.py](services/ingest/ingest/scheduler.py), logi: `docker compose -f deploy/dev/compose.yml logs -f scheduler`. Przycisk „Odśwież” w dashboardzie kolejkuje sync wszystkich źródeł (`POST /api/sync`), który scheduler wykonuje w ciągu kilku sekund.

Na dev sync działa tylko, gdy komputer nie śpi i Docker Desktop jest uruchomiony (warto włączyć „Start Docker Desktop when you sign in”). Przerwy do 7 dni nadrabia nocny backfill.

### Bramka „2 tygodnie stabilnego syncu”

```sql
select source, date(started_at) as day, count(*) as runs,
       count(*) filter (where status = 'ok') as ok,
       count(*) filter (where status <> 'ok') as not_ok,
       sum(inserted) as inserted
from ingest_run
where started_at > now() - interval '14 days'
group by 1, 2 order by 2 desc, 1;
```

### Ingest

```bash
# alias dla wygody
C="docker compose -f deploy/dev/compose.yml --env-file .env --profile jobs"

$C run --rm ingest garmin-login     # raz: hasło + MFA, tokeny lądują w wolumenie garmin_tokens
$C run --rm ingest garmin            # dziś i wczoraj (co godzinę)
$C run --rm ingest garmin --days 7   # nocny backfill
$C run --rm ingest hevy              # pierwszy raz pełny sync, potem zdarzenia od kursora
$C run --rm ingest hevy --full       # pełny resync
$C run --rm ingest hevy-garmin --dry-run   # co zostałoby zapisane w Garminie (serie Hevy w aktywności z zegarka)
$C run --rm ingest hevy-garmin --days 14  # eksport serii Hevy -> Garmin, ręcznie dłuższe okno
$C run --rm ingest fitatu-login     # raz: e-mail + hasło, sesja ląduje w wolumenie fitatu_session
# konto bez hasła (logowanie Google/Apple/Facebook): w https://www.fitatu.com/app/ w konsoli przeglądarki
#   copy(JSON.stringify((u => ({token: u.token, refresh_token: u.refresh_token, id: u.id}))(JSON.parse(localStorage.user))))
# potem w PowerShellu (schowek -> sesja w wolumenie):
#   Get-Clipboard | docker compose -f deploy/dev/compose.yml --env-file .env --profile jobs run --rm -T --entrypoint fitatu-api ingest import-session
$C run --rm ingest fitatu           # dziś i wczoraj (3× dziennie: 10:15, 16:15, 22:15)
$C run --rm ingest fitatu --days 7  # nocny backfill (3:20)
$C run --rm --entrypoint pytest ingest -q
```

Każdy run zapisuje się w `ingest_run` (status `ok | partial | error`, liczba nowych i niezmienionych payloadów, błędy).

### Dashboard i czat

- Dashboard: http://localhost:5173 (Vite, HMR). Dane: domyślnie prawdziwe z API w Go (dziś Dziś i Trening); ciasteczko `mock=1` przełącza na wartości z makiet. Ustawia je `?mock=1` w adresie (powrót: `?mock=0`), parametr znika z URL po załadowaniu.
- API: `api` na :8080 (`/api/today`, `/api/training?range=4w|12w|block`, `/api/body`, `/api/nutrition`, `/api/supplements` + `POST …/{id}/log`, `/api/plan`, `/api/proposals` + `POST …/{id}/accept|reject`, `/health`), przebudowuje projekcje z `raw_payload` co 5 min. Dane do dev bez trenera i bez ekranu konfiguracji: `docker exec -i forma-dev-postgres-1 sh -c 'psql -U "$POSTGRES_USER" -d "$POSTGRES_DB"' < deploy/dev/seed/supplements.sql` (i tak samo `plan.sql`: plan do maratonu, propozycje z makiet). Ręcznie: `docker compose -f deploy/dev/compose.yml --env-file .env exec api go run ./cmd/api project`. Testy: `docker compose -f deploy/dev/compose.yml --profile tools run --rm -w /src/services/api go go test ./...` (w Git Bash poprzedź `MSYS_NO_PATHCONV=1`).
- Czat z trenerem: `coach-chat` na :8090, z dashboardu pod `/api/chat`. Wymaga `CLAUDE_CODE_OAUTH_TOKEN` (z `claude setup-token`) albo `ANTHROPIC_API_KEY`, w Ustawieniach albo w `.env`. Dane dostaje z serwera MCP (`FORMA_MCP_URL`, domyślnie `http://mcp:8091/mcp`).
- Serwer MCP: `mcp` na :8091 (`/mcp`, streamable HTTP, `/health`). Wymaga tokenu: „Wygeneruj” w Ustawieniach albo `FORMA_MCP_TOKEN` w `.env` (min. 32 znaki, np. `openssl rand -hex 32`); bez tokenu odpowiada 503. Ten sam token dostaje `coach-chat`. Po zmianach w Go: `docker compose -f deploy/dev/compose.yml restart mcp`.

```bash
docker compose -f deploy/dev/compose.yml --env-file .env run --rm --no-deps web npm test
docker compose -f deploy/dev/compose.yml --env-file .env run --rm --no-deps coach-chat npm test
```

### Ustawienia (klucze z poziomu aplikacji)

http://localhost:5173/ustawienia (na desktopie link w nagłówku, na Foldzie na dole Dziś): klucz Hevy, token subskrypcji Claude, klucz API Claude i token MCP („Wygeneruj” pokazuje nowy token raz). Wartość zapisana w aplikacji wygrywa ze zmienną z `.env`; „Usuń” wraca do zmiennej. Zmiana działa bez restartu: ingest czyta klucz przy każdym runie, czat przy każdej wiadomości, MCP odświeża token co minutę.

Wartości leżą w tabeli `app_secret` zaszyfrowane AES-256-GCM kluczem `FORMA_SECRETS_KEY` (dostają go `api`, `ingest`/`scheduler` i `coach-chat`, nie `mcp`). Bez tego klucza widok jest tylko do odczytu. Zgubiony klucz = sekrety z aplikacji do wpisania od nowa. Garmin i Fitatu nadal logują się przez `ingest garmin-login` / `fitatu-login`.

Po aktualizacji przebuduj obraz ingestu (nowa zależność `cryptography`): `docker compose -f deploy/dev/compose.yml build ingest scheduler`.

### Toolchainy

Go i Node w kontenerach, bez instalacji na hoście:

```bash
docker compose -f deploy/dev/compose.yml --profile tools run --rm go go version
```

## Układ repo

| Ścieżka | Co |
| --- | --- |
| `services/ingest` | workery Python: Garmin, Hevy → `raw_payload`, scheduler |
| `services/coach-chat` | czat z trenerem: Claude Agent SDK → SSE |
| `web` | dashboard: Vue 3 + ECharts |
| `migrations` | migracje SQL (golang-migrate) |
| `deploy/dev` | compose dla dev |
| `docs`, `design` | projekt i makiety |
