import os
from datetime import date, datetime, timezone

import psycopg
import pytest

from ingest import hevy_garmin as hg
from ingest.hevy_garmin import Link, Workout
from ingest.store import PgStore

NOW = datetime(2026, 10, 25, 20, 0, tzinfo=timezone.utc)


def workout(wid="w1", start="2026-10-25T16:00:00+00:00", end="2026-10-25T17:00:00+00:00", h="h1", title="Push"):
    return Workout(wid, h, {
        "id": wid, "title": title, "start_time": start, "end_time": end,
        "exercises": [
            {"index": 0, "title": "Bench Press (Barbell)", "exercise_template_id": "T1", "sets": [
                {"index": 0, "type": "warmup", "weight_kg": 40, "reps": 10},
                {"index": 1, "type": "normal", "weight_kg": 80, "reps": 8},
            ]},
            {"index": 1, "title": "Plank", "exercise_template_id": "T2", "sets": [
                {"index": 0, "type": "normal", "weight_kg": None, "reps": None, "duration_seconds": 60},
            ]},
        ],
    })


def watch(aid=900, start="2026-10-25 15:58:00", duration=3700.0, type_key="strength_training"):
    return {"activityId": aid, "startTimeGMT": start, "duration": duration, "activityType": {"typeKey": type_key}}


class FakeRepo:
    def __init__(self, workouts=(), links=(), deleted=()):
        self._workouts = list(workouts)
        self.saved: dict[str, tuple[Link, str | None]] = {}
        self._links = {l.hevy_workout_id: l for l in links}
        self._deleted = list(deleted)

    def workouts(self, since):
        return [w for w in self._workouts if w.start.date() >= since]

    def deleted_links(self):
        return self._deleted

    def muscles(self):
        return {"T2": "abdominals"}

    def links(self):
        return dict(self._links)

    def save(self, link, error=None):
        self.saved[link.hevy_workout_id] = (link, error)


class FakeGarmin:
    def __init__(self, activities=(), fail_sets=False):
        self.activities = list(activities)
        self.fail_sets = fail_sets
        self.calls: list[tuple] = []

    def get_activities_by_date(self, startdate, enddate=None):
        self.calls.append(("list", startdate, enddate))
        return self.activities

    def create_manual_activity(self, start, tz, type_key, km, minutes, name):
        self.calls.append(("create", start, tz, type_key, minutes, name))
        return {"activityId": 555}

    def set_activity_name(self, activity_id, title):
        self.calls.append(("name", activity_id, title))

    def set_activity_description(self, activity_id, text):
        self.calls.append(("description", activity_id))

    def set_activity_exercise_sets(self, activity_id, payload):
        if self.fail_sets:
            raise RuntimeError("400 Invalid Sub-Category Passed")
        self.calls.append(("sets", activity_id, payload))

    def delete_activity(self, activity_id):
        self.calls.append(("delete", activity_id))

    def kinds(self):
        return [c[0] for c in self.calls]


@pytest.mark.parametrize(
    "title, muscle, want",
    [
        ("Bench Press (Barbell)", None, {"category": "BENCH_PRESS", "name": "BARBELL_BENCH_PRESS"}),
        ("Bench Press (Dumbbell)", None, {"category": "BENCH_PRESS", "name": "DUMBBELL_BENCH_PRESS"}),
        ("Lat Pulldown (Cable)", "lats", {"category": "PULL_UP", "name": None}),
        ("Some Gym Machine Thing", "quadriceps", {"category": "SQUAT", "name": None}),
        ("Mystery", None, {"category": "TOTAL_BODY", "name": None}),
    ],
)
def test_garmin_exercise(title, muscle, want):
    got = hg.garmin_exercise(title, muscle)
    assert got["category"] == want["category"]
    if want["name"] is not None:
        assert got["name"] == want["name"]
    assert got["category"] in hg._CATEGORIES


def test_exercise_sets_spread_over_window_with_rests_and_grams():
    w = workout()
    sets = hg.exercise_sets(w, {}, w.start, w.end)

    active = [s for s in sets if s["setType"] == "ACTIVE"]
    assert [(s["repetitionCount"], s["weight"]) for s in active] == [(10, 40000), (8, 80000), (None, None)]
    assert [s["duration"] for s in active] == [40.0, 40.0, 60.0]
    assert [s["setType"] for s in sets] == ["ACTIVE", "REST", "ACTIVE", "REST", "ACTIVE"]  # no rest after the last
    assert [s["messageIndex"] for s in sets] == [0, 1, 2, 3, 4]
    assert active[0]["startTime"] == "2026-10-25T16:00:00.0"
    assert active[1]["startTime"] == "2026-10-25T16:20:00.0"  # 60 min / 3 sets


def test_description_skips_warmups():
    assert hg.description(workout()).splitlines()[:2] == [
        "Bench Press (Barbell): 1 serie · 80 kg × 8",
        "Plank: 1 serie · 60 s",
    ]


def test_sets_go_into_the_overlapping_watch_activity():
    api = FakeGarmin([watch(aid=1, type_key="running"), watch(aid=900)])
    repo = FakeRepo([workout()])

    assert hg.sync(api, repo, NOW, days=3) == []

    assert api.kinds() == ["list", "name", "description", "sets"]
    assert api.calls[1] == ("name", "900", "Push")
    payload = api.calls[3][2]
    assert payload["activityId"] == 900
    assert payload["exerciseSets"][0]["startTime"] == "2026-10-25T15:58:00.0"  # watch window, not Hevy's
    assert repo.saved["w1"] == (Link("w1", "h1", 900, False, "synced"), None)


def test_unchanged_synced_workout_is_skipped_and_changed_one_rewritten():
    synced = Link("w1", "h1", 900, False, "synced")
    api = FakeGarmin([watch()])
    hg.sync(api, FakeRepo([workout()], links=[synced]), NOW, days=3)
    assert api.calls == []

    repo = FakeRepo([workout(h="h2")], links=[synced])
    hg.sync(api, repo, NOW, days=3)
    assert api.kinds() == ["name", "description", "sets"]  # known activity: no listing, no matching
    assert repo.saved["w1"][0] == Link("w1", "h2", 900, False, "synced")


def test_activity_already_linked_to_another_workout_is_not_matched_again():
    api = FakeGarmin([watch(aid=900)])
    repo = FakeRepo([workout()], links=[Link("other", "x", 900, False, "synced")])

    hg.sync(api, repo, datetime(2026, 10, 25, 17, 30, tzinfo=timezone.utc), days=3)

    assert api.kinds() == ["list"]  # within grace: wait for the watch
    assert repo.saved == {}


def test_manual_activity_after_grace_when_watch_recorded_nothing():
    api = FakeGarmin([watch(start="2026-10-25 08:00:00")])  # morning session, no overlap
    repo = FakeRepo([workout()])

    hg.sync(api, repo, NOW, days=3)

    assert api.calls[1] == ("create", "2026-10-25T17:00:00.000", "Europe/Warsaw", "strength_training", 60, "Push")
    assert api.kinds()[2:] == ["name", "description", "sets"]
    assert repo.saved["w1"] == (Link("w1", "h1", 555, True, "synced"), None)


def test_failed_sets_keep_the_created_activity_for_the_retry():
    api = FakeGarmin([], fail_sets=True)
    repo = FakeRepo([workout()])

    errors = hg.sync(api, repo, NOW, days=3)

    assert errors == ["w1: 400 Invalid Sub-Category Passed"]
    link, error = repo.saved["w1"]
    assert (link.garmin_activity_id, link.created_by_us, link.status, link.workout_hash) == (555, True, "error", "")
    assert error == "400 Invalid Sub-Category Passed"

    # Next run reuses the activity instead of creating another one.
    api2 = FakeGarmin([])
    hg.sync(api2, FakeRepo([workout()], links=[link]), NOW, days=3)
    assert "create" not in api2.kinds()


def test_deleted_workout_removes_only_activities_we_created():
    api = FakeGarmin()
    repo = FakeRepo(deleted=[Link("mine", "h", 555, True, "synced"), Link("watch", "h", 900, False, "synced")])

    hg.sync(api, repo, NOW, days=3)

    assert api.calls == [("delete", "555")]
    assert repo.saved["mine"][0].status == "deleted"
    assert repo.saved["watch"][0] == Link("watch", "h", 900, False, "deleted")


def test_old_workouts_outside_window_are_ignored():
    api = FakeGarmin()
    repo = FakeRepo([workout(start="2026-10-01T16:00:00+00:00", end="2026-10-01T17:00:00+00:00")])
    assert hg.sync(api, repo, NOW, days=3) == []
    assert api.calls == []



@pytest.fixture
def pg_repo():
    url = os.environ.get("DATABASE_URL")
    if not url:
        pytest.skip("DATABASE_URL not set")
    with psycopg.connect(url) as conn:
        yield hg.PgRepo(conn, commit=False)
        conn.rollback()


def test_pg_repo_reads_latest_versions_and_deletions(pg_repo):
    store = PgStore(pg_repo.conn)
    d = date(2026, 10, 25)
    w = workout(wid="pg-w1").body
    store.put_raw("hevy", "workouts", "pg-w1", d, w)
    store.put_raw("hevy", "workouts", "pg-w1", d, {**w, "title": "Push v2"})
    store.put_raw("hevy", "workouts", "pg-w2", d, workout(wid="pg-w2").body)
    store.put_raw("hevy", "workout_deleted", "pg-w2", None, {"id": "pg-w2", "type": "deleted"})
    store.put_raw("hevy", "workouts", "pg-old", date(2026, 9, 1), workout(wid="pg-old").body)
    store.put_raw("hevy", "exercise_templates", "pg-T", None, {"id": "pg-T", "primary_muscle_group": "chest"})

    got = [x for x in pg_repo.workouts(d) if x.id.startswith("pg-")]
    assert [(x.id, x.title) for x in got] == [("pg-w1", "Push v2")]
    assert pg_repo.muscles()["pg-T"] == "chest"

    pg_repo.save(Link("pg-w2", "h", 555, True, "synced"))
    pg_repo.save(Link("pg-w1", "h", 900, False, "error"), "boom")
    pg_repo.save(Link("pg-w1", "h", 900, False, "error"), "boom")
    assert [l.hevy_workout_id for l in pg_repo.deleted_links() if l.hevy_workout_id.startswith("pg-")] == ["pg-w2"]
    assert pg_repo.links()["pg-w1"] == Link("pg-w1", "h", 900, False, "error")
    attempts = pg_repo.conn.execute(
        "select attempts from hevy_garmin_link where hevy_workout_id = 'pg-w1'").fetchone()[0]
    assert attempts == 2
