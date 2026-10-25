from datetime import date, datetime, timezone

import pytest

from ingest import hevy


class FakeResponse:
    def __init__(self, status_code, data=None):
        self.status_code = status_code
        self._data = data

    def json(self):
        return self._data

    def raise_for_status(self):
        if self.status_code >= 400:
            raise RuntimeError(f"HTTP {self.status_code}")


class FakeSession:
    """Serves paginated responses: routes[path] = list of item lists, one per page."""

    def __init__(self, routes: dict[str, tuple[str, list[list[dict]]]]):
        self.routes = routes
        self.calls: list[tuple[str, dict]] = []

    def get(self, url, headers, params, timeout):
        path = url.removeprefix(hevy.BASE_URL)
        self.calls.append((path, params))
        assert headers["api-key"] == "key"
        key, pages = self.routes.get(path, ("", []))
        page = params["page"]
        if page > len(pages):
            return FakeResponse(404)
        return FakeResponse(200, {"page": page, "page_count": len(pages), key: pages[page - 1]})


NOW = datetime(2026, 10, 25, 18, 30, tzinfo=timezone.utc)
TEMPLATES = ("exercise_templates", [[{"id": "T1", "title": "Bench Press", "primary_muscle_group": "chest"}]])


def workout(wid, start="2026-10-24T22:30:00Z"):
    return {"id": wid, "title": "Push", "start_time": start, "exercises": []}


def test_first_run_does_full_sync_and_sets_cursor(store):
    session = FakeSession({
        "/v1/workouts": ("workouts", [[workout("w1"), workout("w2")], [workout("w3")]]),
        "/v1/exercise_templates": TEMPLATES,
    })
    hevy.sync(hevy.HevyClient("key", session), store, NOW)

    assert store.endpoints() == ["workouts"] * 3 + ["exercise_templates"]
    assert store.cursors[("hevy", "workouts")] == "2026-10-25T18:30:00Z"
    # 22:30 UTC on Oct 24 is already Oct 25 in Warsaw
    assert store.rows[0][3] == date(2026, 10, 25)


def test_incremental_run_uses_events_since_cursor(store):
    store.set_cursor("hevy", "workouts", "2026-10-25T10:00:00Z")
    store.set_cursor("hevy", "exercise_templates", "2026-10-25")
    session = FakeSession({
        "/v1/workouts/events": ("events", [[
            {"type": "updated", "workout": workout("w1")},
            {"type": "deleted", "id": "w0", "deleted_at": "2026-10-25T11:00:00Z"},
        ]]),
    })
    hevy.sync(hevy.HevyClient("key", session), store, NOW)

    assert [c[0] for c in session.calls] == ["/v1/workouts/events"]
    assert session.calls[0][1]["since"] == "2026-10-25T10:00:00Z"
    assert store.endpoints() == ["workouts", "workout_deleted"]


@pytest.mark.parametrize("full", [False, True])
def test_no_events_is_not_an_error(store, full):
    store.set_cursor("hevy", "workouts", "2026-10-25T10:00:00Z")
    session = FakeSession({"/v1/exercise_templates": TEMPLATES})
    assert hevy.sync(hevy.HevyClient("key", session), store, NOW, full=full) == []
    assert store.endpoints() == ["exercise_templates"]


def test_rerun_with_same_data_inserts_nothing(store):
    routes = {"/v1/workouts": ("workouts", [[workout("w1")]]), "/v1/exercise_templates": TEMPLATES}
    hevy.sync(hevy.HevyClient("key", FakeSession(routes)), store, NOW, full=True)
    hevy.sync(hevy.HevyClient("key", FakeSession(routes)), store, NOW, full=True)
    assert len(store.rows) == 2
