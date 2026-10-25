from datetime import datetime, timedelta

import pytest

from ingest import TZ, scheduler


def at(hour, minute):
    return datetime(2026, 10, 26, hour, minute, tzinfo=TZ)


@pytest.mark.parametrize(
    "t, expected",
    [
        (at(3, 0), ["garmin-backfill"]),
        (at(3, 5), ["hevy"]),
        (at(3, 10), ["garmin"]),
        (at(14, 5), ["hevy"]),
        (at(14, 0), []),
        (at(14, 30), []),
    ],
)
def test_due_jobs(t, expected):
    assert [j.name for j in scheduler.JOBS if j.due(t)] == expected


class FakeClock:
    def __init__(self, start):
        self.t = start

    def now(self):
        return self.t

    def sleep(self, seconds):
        self.t += timedelta(seconds=seconds)


def test_loop_runs_due_jobs_at_minute_boundaries_and_survives_failures():
    clock = FakeClock(at(3, 3) + timedelta(seconds=20))
    ran = []

    def run(job):
        ran.append((clock.now().strftime("%H:%M"), job.name))
        if job.name == "hevy":
            raise RuntimeError("boom")
        return 0

    scheduler.loop(run=run, now=clock.now, sleep=clock.sleep, ticks=8)
    assert ran == [("03:05", "hevy"), ("03:10", "garmin")]
