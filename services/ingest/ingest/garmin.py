"""Garmin Connect via unofficial python-garminconnect. Includes Xiaomi scale data synced to Garmin."""

import contextlib
import getpass
import logging
from collections.abc import Callable
from datetime import date, timedelta
from typing import Any, Protocol

from .store import Store

log = logging.getLogger(__name__)

SOURCE = "garmin"


class GarminApi(Protocol):
    def get_sleep_data(self, cdate: str) -> dict: ...
    def get_hrv_data(self, cdate: str) -> dict | None: ...
    def get_user_summary(self, cdate: str) -> dict: ...
    def get_body_composition(self, startdate: str, enddate: str | None = None) -> dict: ...
    def get_activities_by_date(self, startdate: str, enddate: str | None = None) -> list[dict]: ...
    def get_activity_hr_in_timezones(self, activity_id: str) -> Any: ...


# Per-day endpoints: name -> call. Sleep and HRV close later in the day, hence the nightly 7-day backfill.
DAILY: dict[str, Callable[[GarminApi, str], Any]] = {
    "sleep": lambda api, d: api.get_sleep_data(d),
    "hrv": lambda api, d: api.get_hrv_data(d),
    "user_summary": lambda api, d: api.get_user_summary(d),
    "body_composition": lambda api, d: api.get_body_composition(d, d),
}


def sync(api: GarminApi, store: Store, today: date, days: int) -> list[str]:
    """Fetch the last `days` days (today included). Returns per-call errors; the run continues past them."""
    errors: list[str] = []
    start = today - timedelta(days=days - 1)

    def attempt(what: str, fn: Callable[[], Any]) -> Any:
        try:
            return fn()
        except Exception as e:  # unofficial API: one failing endpoint must not drop the others
            log.warning("garmin: %s failed: %s", what, e)
            errors.append(f"{what}: {e}")
            return None

    d = start
    while d <= today:
        ds = d.isoformat()
        for endpoint, call in DAILY.items():
            body = attempt(f"{endpoint} {ds}", lambda: call(api, ds))
            if body:
                store.put_raw(SOURCE, endpoint, ds, d, body)
        d += timedelta(days=1)

    activities = attempt("activities", lambda: api.get_activities_by_date(start.isoformat(), today.isoformat()))
    for act in activities or []:
        act_id = str(act["activityId"])
        local = date.fromisoformat(act["startTimeLocal"][:10]) if act.get("startTimeLocal") else None
        # HR zones are a separate call; fetch them only for new or changed activities.
        if store.put_raw(SOURCE, "activities", act_id, local, act):
            zones = attempt(f"hr_zones {act_id}", lambda: api.get_activity_hr_in_timezones(act_id))
            if zones:
                store.put_raw(SOURCE, "activity_hr_zones", act_id, local, zones)
    return errors


def connect(tokenstore: str):
    """Resume a session from saved tokens (created by `ingest garmin-login`)."""
    from garminconnect import Garmin

    api = Garmin()
    api.login(tokenstore)
    return api


def save_tokens(api, tokenstore: str) -> None:
    # Tokens may have been refreshed during the run; persist them so the next run does not re-login.
    with contextlib.suppress(Exception):
        api.client.dump(tokenstore)


def interactive_login(email: str | None, tokenstore: str) -> None:
    """First login with password and MFA; only tokens are persisted, never the password."""
    from garminconnect import Garmin

    email = email or input("Garmin e-mail: ")
    password = getpass.getpass("Garmin hasło: ")
    api = Garmin(email=email, password=password, prompt_mfa=lambda: input("Kod MFA: "))
    api.login(tokenstore)
    api.client.dump(tokenstore)
    log.info("garmin: tokens saved to %s", tokenstore)
