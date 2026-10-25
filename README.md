# Forma

Sportowy dashboard + AI trener. Projekt: [docs/design.md](docs/design.md), zasady pracy: [CLAUDE.md](CLAUDE.md).

## Środowisko dev (lokalnie, Docker Desktop)

Produkcja to Mikrus; lokalnie wszystko stoi w [deploy/dev/compose.yml](deploy/dev/compose.yml).

```bash
cp .env.example .env          # uzupełnij HEVY_API_KEY, GARMIN_EMAIL
docker compose -f deploy/dev/compose.yml --env-file .env up -d
```

`up` startuje Postgresa (127.0.0.1:5432), adminera (http://localhost:8081) i uruchamia migracje z `migrations/`.

### Ingest

```bash
# alias dla wygody
C="docker compose -f deploy/dev/compose.yml --env-file .env --profile jobs"

$C run --rm ingest garmin-login     # raz: hasło + MFA, tokeny lądują w wolumenie garmin_tokens
$C run --rm ingest garmin            # dziś i wczoraj (co godzinę)
$C run --rm ingest garmin --days 7   # nocny backfill
$C run --rm ingest hevy              # pierwszy raz pełny sync, potem zdarzenia od kursora
$C run --rm ingest hevy --full       # pełny resync
$C run --rm --entrypoint pytest ingest -q
```

Każdy run zapisuje się w `ingest_run` (status `ok | partial | error`, liczba nowych i niezmienionych payloadów, błędy).

### Toolchainy

Go i Node w kontenerach, bez instalacji na hoście:

```bash
docker compose -f deploy/dev/compose.yml --profile tools run --rm go go version
```

## Układ repo

| Ścieżka | Co |
| --- | --- |
| `services/ingest` | workery Python: Garmin, Hevy → `raw_payload` |
| `migrations` | migracje SQL (golang-migrate) |
| `deploy/dev` | compose dla dev |
| `docs`, `design` | projekt i makiety |
