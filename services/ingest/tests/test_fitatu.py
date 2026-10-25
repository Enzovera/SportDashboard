from datetime import date

import pytest

from ingest import fitatu


class FakeApi:
    def __init__(self, fail=(), empty=()):
        self.fail = set(fail)
        self.empty = set(empty)

    def _reply(self, name, day, body):
        if name in self.fail:
            raise RuntimeError(f"{name} down")
        return None if name in self.empty else body

    def diet_day(self, day):
        return self._reply("diet_day", day, {"date": day.isoformat(), "dietPlan": {"breakfast": {"items": []}}})

    def settings(self, day):
        return self._reply("settings", day, {"userDietSettings": {"energy": 2950}})

    def water(self, day):
        return self._reply("water", day, {"water": {"waterConsumption": 1500}})


TODAY = date(2026, 10, 25)


def test_sync_stores_every_endpoint_per_day(store):
    assert fitatu.sync(FakeApi(), store, TODAY, days=2) == []
    assert sorted((r[2], r[1]) for r in store.rows) == [
        ("2026-10-24", "diet_day"), ("2026-10-24", "settings"), ("2026-10-24", "water"),
        ("2026-10-25", "diet_day"), ("2026-10-25", "settings"), ("2026-10-25", "water"),
    ]
    assert {r[0] for r in store.rows} == {"fitatu"}
    assert store.rows[0][3] == date(2026, 10, 24)


def test_unchanged_days_are_not_stored_twice(store):
    fitatu.sync(FakeApi(), store, TODAY, days=7)
    fitatu.sync(FakeApi(), store, TODAY, days=7)
    assert len(store.rows) == 21


@pytest.mark.parametrize("broken", ["diet_day", "water"])
def test_failing_endpoint_is_reported_and_others_still_stored(store, broken):
    errors = fitatu.sync(FakeApi(fail={broken}), store, TODAY, days=1)
    assert errors == [f"{broken} 2026-10-25: {broken} down"]
    assert broken not in store.endpoints()
    assert len(store.rows) == 2


def test_empty_reply_is_skipped(store):
    assert fitatu.sync(FakeApi(empty={"water"}), store, TODAY, days=1) == []
    assert "water" not in store.endpoints()


class FakeClient:
    """Stands in for fitatu_api.FitatuApiClient; records calls to check the reader stays read-only."""

    class _Mod:
        def __init__(self, calls, name):
            self.calls, self.name = calls, name

        def __getattr__(self, method):
            return lambda *a: self.calls.append((self.name, method, a)) or {"ok": True}

    def __init__(self, user_id="123"):
        self.calls = []
        self.auth = type("Auth", (), {"fitatu_user_id": user_id})()
        self.planner = self._Mod(self.calls, "planner")
        self.user_settings = self._Mod(self.calls, "user_settings")
        self.water = self._Mod(self.calls, "water")


def test_reader_uses_only_get_calls():
    client = FakeClient()
    reader = fitatu.FitatuReader(client)
    fitatu.sync(reader, type("S", (), {"put_raw": lambda *a: True})(), TODAY, days=1)
    assert [(m, f) for m, f, _ in client.calls] == [("planner", "get_day"), ("user_settings", "get_new"), ("water", "get_day")]
    assert all(args[0] == "123" for _, _, args in client.calls)


def test_reader_requires_a_user_id():
    with pytest.raises(RuntimeError, match="fitatu-login"):
        fitatu.FitatuReader(FakeClient(user_id=None))
