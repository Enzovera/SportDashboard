from datetime import date

from ingest import garmin


class FakeApi:
    def __init__(self, activities=None, fail=()):
        self.activities = activities or []
        self.fail = set(fail)
        self.zone_calls: list[str] = []

    def _maybe_fail(self, name):
        if name in self.fail:
            raise RuntimeError(f"{name} down")

    def get_sleep_data(self, cdate):
        self._maybe_fail("sleep")
        return {"dailySleepDTO": {"calendarDate": cdate, "sleepTimeSeconds": 27000}}

    def get_hrv_data(self, cdate):
        return None  # no HRV for the day (e.g. watch not worn)

    def get_user_summary(self, cdate):
        return {"calendarDate": cdate, "totalSteps": 9000, "restingHeartRate": 48}

    def get_body_composition(self, startdate, enddate=None):
        return {"dateWeightList": []}

    def get_activities_by_date(self, startdate, enddate=None):
        return self.activities

    def get_activity_hr_in_timezones(self, activity_id):
        self.zone_calls.append(activity_id)
        return [{"zoneNumber": 1, "secsInZone": 600}]


TODAY = date(2026, 10, 25)
RUN = {"activityId": 123, "startTimeLocal": "2026-10-25 07:10:00", "activityType": {"typeKey": "running"}}


def test_sync_stores_each_day_and_skips_empty_responses(store):
    errors = garmin.sync(FakeApi(), store, TODAY, days=2)

    assert errors == []
    by_day = sorted((r[2], r[1]) for r in store.rows)
    assert by_day == [
        ("2026-10-24", "body_composition"), ("2026-10-24", "sleep"), ("2026-10-24", "user_summary"),
        ("2026-10-25", "body_composition"), ("2026-10-25", "sleep"), ("2026-10-25", "user_summary"),
    ]


def test_hr_zones_fetched_only_for_new_or_changed_activities(store):
    api = FakeApi(activities=[RUN])
    garmin.sync(api, store, TODAY, days=1)
    garmin.sync(api, store, TODAY, days=1)
    assert api.zone_calls == ["123"]

    api.activities = [{**RUN, "activityName": "Easy run"}]
    garmin.sync(api, store, TODAY, days=1)
    assert api.zone_calls == ["123", "123"]
    zones = [r for r in store.rows if r[1] == "activity_hr_zones"]
    assert zones[0][3] == TODAY


def test_failing_endpoint_is_reported_and_others_still_stored(store):
    errors = garmin.sync(FakeApi(activities=[RUN], fail={"sleep"}), store, TODAY, days=1)

    assert errors == ["sleep 2026-10-25: sleep down"]
    assert "sleep" not in store.endpoints()
    assert {"user_summary", "activities", "activity_hr_zones"} <= set(store.endpoints())
