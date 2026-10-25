"""Minimal in-container cron: runs ingest jobs at fixed local times, each in its own process."""

import logging
import subprocess
import sys
import time
from collections.abc import Callable
from dataclasses import dataclass
from datetime import datetime, timedelta

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


# Schedule from docs/design.md: Garmin and Hevy hourly, 7-day backfill at 3:00.
JOBS = (
    Job("hevy", ("hevy",), minute=5),
    Job("garmin", ("garmin", "--days", "2"), minute=10),
    Job("garmin-backfill", ("garmin", "--days", "7"), minute=0, hour=3),
)


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
) -> None:
    """Wake at each minute boundary and run due jobs sequentially. `ticks` limits iterations (tests)."""
    log.info("scheduler: %s", ", ".join(f"{j.name}@{'*' if j.hour is None else j.hour}:{j.minute:02d}" for j in jobs))
    while ticks is None or ticks > 0:
        target = next_minute(now())
        sleep(max(0.0, (target - now()).total_seconds()))
        for job in jobs:
            if job.due(target):
                try:
                    run(job)
                except Exception:
                    log.exception("scheduler: %s failed to run", job.name)
        if ticks is not None:
            ticks -= 1
