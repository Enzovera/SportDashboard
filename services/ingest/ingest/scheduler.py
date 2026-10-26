"""Minimal in-container cron: runs ingest jobs at fixed local times, each in its own process.

Between minute boundaries it also serves on-demand sync requests queued by the dashboard (table sync_request).
"""

import logging
import subprocess
import sys
import time
from collections.abc import Callable
from dataclasses import dataclass
from datetime import datetime, timedelta
from typing import Protocol

import psycopg

from . import TZ

log = logging.getLogger(__name__)


@dataclass(frozen=True)
class Job:
    name: str
    argv: tuple[str, ...]
    minute: int
    hour: int | None = None  # None = every hour

    def due(self, t: datetime) -> bool:
        return t.minute == self.minute and (self.hour is None or t.hour == self.hour)


# Schedule from docs/design.md: Garmin and Hevy hourly, Fitatu 3 times a day, 7-day backfill of all sources at 3:00.
JOBS = (
    Job("hevy", ("hevy",), minute=5),
    Job("garmin", ("garmin", "--days", "2"), minute=10),
    Job("garmin-backfill", ("garmin", "--days", "7"), minute=0, hour=3),
    # After breakfast, lunch and dinner are usually logged; 2 days catch meals added the next morning.
    Job("fitatu-morning", ("fitatu", "--days", "2"), minute=15, hour=10),
    Job("fitatu-afternoon", ("fitatu", "--days", "2"), minute=15, hour=16),
    Job("fitatu-evening", ("fitatu", "--days", "2"), minute=15, hour=22),
    Job("fitatu-backfill", ("fitatu", "--days", "7"), minute=20, hour=3),
)

# On-demand sync ("Odśwież"): the same windows as the regular runs. minute=-1 is never due on the schedule.
SYNC_NOW = (
    Job("manual-hevy", ("hevy",), minute=-1),
    Job("manual-garmin", ("garmin", "--days", "2"), minute=-1),
    Job("manual-fitatu", ("fitatu", "--days", "2"), minute=-1),
)

# How often the queue is checked while waiting for the next minute.
POLL_SECONDS = 5.0
# Behind by more than this (host suspended, very long run) the missed minutes are skipped, not replayed.
MAX_CATCH_UP = timedelta(minutes=10)


class Requests(Protocol):
    def recover(self) -> None:
        """Close requests left running by a previous scheduler process."""

    def claim(self) -> int | None:
        """Mark the oldest pending request running and return its id, or None when the queue is empty."""

    def finish(self, request_id: int) -> None: ...


class PgRequests:
    """sync_request queue in Postgres. A short autocommit connection per call, so a database restart is harmless."""

    def __init__(self, dsn: str):
        self.dsn = dsn

    def recover(self) -> None:
        with psycopg.connect(self.dsn, autocommit=True) as conn:
            conn.execute("update sync_request set status = 'done', finished_at = now() where status = 'running'")

    def claim(self) -> int | None:
        with psycopg.connect(self.dsn, autocommit=True) as conn:
            row = conn.execute(
                """
                update sync_request set status = 'running', started_at = now()
                where id = (
                  select id from sync_request where status = 'pending' order by id limit 1 for update skip locked
                )
                returning id
                """
            ).fetchone()
        return row[0] if row else None

    def finish(self, request_id: int) -> None:
        with psycopg.connect(self.dsn, autocommit=True) as conn:
            conn.execute("update sync_request set status = 'done', finished_at = now() where id = %s", (request_id,))
            # The API listens and rebuilds projections right away instead of waiting for PROJECT_INTERVAL.
            conn.execute("select pg_notify('sync_done', %s)", (str(request_id),))


def serve_requests(requests: Requests, run: Callable[[Job], int]) -> None:
    try:
        request_id = requests.claim()
    except Exception:
        log.exception("scheduler: cannot read sync requests")
        return
    if request_id is None:
        return
    log.info("scheduler: sync request %d", request_id)
    for job in SYNC_NOW:
        try:
            run(job)
        except Exception:
            log.exception("scheduler: %s failed to run", job.name)
    try:
        requests.finish(request_id)
    except Exception:
        log.exception("scheduler: cannot finish sync request %d", request_id)


def next_minute(now: datetime) -> datetime:
    return now.replace(second=0, microsecond=0) + timedelta(minutes=1)


def run_job(job: Job) -> int:
    log.info("scheduler: start %s", job.name)
    # A separate process per job: a crash or hang in one worker cannot take the scheduler down.
    result = subprocess.run([sys.executable, "-m", "ingest", *job.argv], timeout=30 * 60)
    log.info("scheduler: %s exited with %d", job.name, result.returncode)
    return result.returncode


def loop(
    jobs=JOBS,
    run: Callable[[Job], int] = run_job,
    now: Callable[[], datetime] = lambda: datetime.now(TZ),
    sleep: Callable[[float], None] = time.sleep,
    ticks: int | None = None,
    requests: Requests | None = None,
) -> None:
    """Run due jobs sequentially at each minute boundary, serving sync requests in between.

    `ticks` limits iterations (tests). A minute that passed during a long sync request still runs its jobs.
    """
    log.info("scheduler: %s", ", ".join(f"{j.name}@{'*' if j.hour is None else j.hour}:{j.minute:02d}" for j in jobs))
    if requests is not None:
        try:
            requests.recover()
        except Exception:
            log.exception("scheduler: cannot recover sync requests")
    target = next_minute(now())
    while ticks is None or ticks > 0:
        while (left := (target - now()).total_seconds()) > 0:
            if requests is None:
                sleep(left)
                break
            serve_requests(requests, run)
            left = (target - now()).total_seconds()
            if left > 0:
                sleep(min(POLL_SECONDS, left))
        for job in jobs:
            if job.due(target):
                try:
                    run(job)
                except Exception:
                    log.exception("scheduler: %s failed to run", job.name)
        target += timedelta(minutes=1)
        if now() - target > MAX_CATCH_UP:
            target = next_minute(now())
        if ticks is not None:
            ticks -= 1
