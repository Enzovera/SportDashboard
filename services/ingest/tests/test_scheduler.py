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
        (at(3, 20), ["fitatu-backfill"]),
        (at(10, 15), ["fitatu-morning"]),
        (at(16, 15), ["fitatu-afternoon"]),
        (at(22, 15), ["fitatu-evening"]),
        (at(22, 10), ["garmin"]),
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


class FakeRequests:
    def __init__(self, pending):
        self.pending = list(pending)
        self.recovered = False
        self.finished = []

    def recover(self):
        self.recovered = True

    def claim(self):
        return self.pending.pop(0) if self.pending else None

    def finish(self, request_id):
        self.finished.append(request_id)


def test_loop_serves_sync_request_between_minutes_and_still_runs_due_jobs():
    clock = FakeClock(at(3, 4) + timedelta(seconds=50))
    requests = FakeRequests([7])
    ran = []

    def run(job):
        ran.append((clock.now().strftime("%H:%M"), job.name))
        clock.t += timedelta(seconds=40)  # a sync request spills over the 3:05 boundary
        if job.name == "manual-garmin":
            raise RuntimeError("boom")
        return 0

    scheduler.loop(run=run, now=clock.now, sleep=clock.sleep, ticks=1, requests=requests)
    assert requests.recovered
    assert requests.finished == [7]
    assert [name for _, name in ran] == ["manual-hevy", "manual-garmin", "manual-fitatu", "hevy"]


def test_loop_skips_minutes_after_long_stall():
    clock = FakeClock(at(3, 3) + timedelta(seconds=50))
    ran = []

    def run(job):
        ran.append((clock.now().strftime("%H:%M"), job.name))
        return 0

    stalled = False

    def sleep(seconds):
        nonlocal stalled
        clock.sleep(seconds)
        if not stalled:  # host suspended for an hour right after start
            stalled = True
            clock.t += timedelta(hours=1)

    scheduler.loop(run=run, now=clock.now, sleep=sleep, ticks=8)
    # 3:05 (hevy) is not replayed; the loop resumes at 4:05
    assert ran == [("04:05", "hevy"), ("04:10", "garmin")]
