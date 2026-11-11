"""CLI: python -m ingest {hevy [--full] | garmin [--days N] | garmin-login | hevy-garmin [--days N] [--dry-run]
| fitatu [--days N] | fitatu-login | schedule}"""

import argparse
import logging
import os
import sys
from datetime import datetime, timezone

import psycopg

from . import TZ, fitatu, garmin, hevy, hevy_garmin, scheduler
from .store import PgStore

log = logging.getLogger("ingest")


def env(name: str) -> str:
    value = os.environ.get(name)
    if not value:
        sys.exit(f"missing environment variable {name}")
    return value


def run(source: str, args: dict, work) -> int:
    """Run one worker in a single transaction and record it in ingest_run."""
    with psycopg.connect(env("DATABASE_URL")) as conn:
        store = PgStore(conn)
        run_id = store.start_run(source, args)
        try:
            errors = work(store)
            conn.commit()
        except Exception as e:
            conn.rollback()
            log.exception("%s: run failed", source)
            store.finish_run(run_id, "error", [str(e)])
            return 1
        store.finish_run(run_id, "partial" if errors else "ok", errors)
        log.info("%s: %d new, %d unchanged, %d errors", source, store.inserted, store.unchanged, len(errors))
        return 0


def main() -> int:
    logging.basicConfig(level=logging.INFO, format="%(asctime)s %(levelname)s %(name)s %(message)s")
    parser = argparse.ArgumentParser(prog="ingest")
    sub = parser.add_subparsers(dest="cmd", required=True)
    p_hevy = sub.add_parser("hevy", help="sync Hevy workouts")
    p_hevy.add_argument("--full", action="store_true", help="refetch all workouts instead of events since cursor")
    p_garmin = sub.add_parser("garmin", help="sync Garmin days and activities")
    p_garmin.add_argument("--days", type=int, default=2, help="days back including today (nightly backfill: 7)")
    sub.add_parser("garmin-login", help="first Garmin login with password and MFA")
    p_hg = sub.add_parser("hevy-garmin", help="write Hevy sets into the matching Garmin strength activities")
    p_hg.add_argument("--days", type=int, default=3, help="Hevy workouts of the last N days (today included)")
    p_hg.add_argument("--dry-run", action="store_true", help="log what would be written to Garmin, change nothing")
    p_fitatu = sub.add_parser("fitatu", help="sync Fitatu meals, targets and water")
    p_fitatu.add_argument("--days", type=int, default=2, help="days back including today (nightly backfill: 7)")
    sub.add_parser("fitatu-login", help="first Fitatu login with e-mail and password")
    sub.add_parser("schedule", help="run hourly sync and nightly backfill forever")
    args = parser.parse_args()

    now = datetime.now(timezone.utc)
    match args.cmd:
        case "hevy":
            client = hevy.HevyClient(env("HEVY_API_KEY"))
            return run("hevy", {"full": args.full}, lambda store: hevy.sync(client, store, now, args.full))
        case "garmin":
            tokenstore = env("GARMINTOKENS")

            def work(store):
                # Login inside the run, so auth failures are recorded in ingest_run too.
                api = garmin.connect(tokenstore)
                try:
                    return garmin.sync(api, store, now.astimezone(TZ).date(), args.days)
                finally:
                    garmin.save_tokens(api, tokenstore)

            return run("garmin", {"days": args.days}, work)
        case "hevy-garmin":
            tokenstore = env("GARMINTOKENS")

            def work_hg(store):
                api = garmin.connect(tokenstore)
                try:
                    if args.dry_run:
                        return hevy_garmin.sync(hevy_garmin.DryRun(api), hevy_garmin.NoSave(store.conn), now, args.days)
                    return hevy_garmin.sync(api, hevy_garmin.PgRepo(store.conn), now, args.days)
                finally:
                    garmin.save_tokens(api, tokenstore)

            return run(hevy_garmin.SOURCE, {"days": args.days, "dry_run": args.dry_run}, work_hg)
        case "garmin-login":
            garmin.interactive_login(os.environ.get("GARMIN_EMAIL"), env("GARMINTOKENS"))
            return 0
        case "fitatu":
            session = env("FITATU_SESSION")

            def work_fitatu(store):
                # Login inside the run, so an expired session is recorded in ingest_run too.
                return fitatu.sync(fitatu.connect(session), store, now.astimezone(TZ).date(), args.days)

            return run("fitatu", {"days": args.days}, work_fitatu)
        case "fitatu-login":
            fitatu.interactive_login(os.environ.get("FITATU_EMAIL"), env("FITATU_SESSION"))
            return 0
        case "schedule":
            scheduler.loop(requests=scheduler.PgRequests(env("DATABASE_URL")))
            return 0
    return 2


if __name__ == "__main__":
    sys.exit(main())
