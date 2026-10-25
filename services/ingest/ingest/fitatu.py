"""Fitatu via unofficial fitatu-api (reverse-engineered web API). Read-only: never calls the library's write methods."""

import getpass
import logging
from collections.abc import Callable
from datetime import date, timedelta
from typing import Any, Protocol

from .store import Store

log = logging.getLogger(__name__)

SOURCE = "fitatu"


class FitatuApi(Protocol):
    """The read calls we use, bound to the logged-in user (see FitatuReader)."""

    def diet_day(self, day: date) -> Any: ...
    def settings(self, day: date) -> Any: ...
    def water(self, day: date) -> Any: ...


# Per-day endpoints. Meals are often logged later in the day or the next morning, hence the 7-day backfill.
DAILY: dict[str, Callable[[FitatuApi, date], Any]] = {
    "diet_day": lambda api, d: api.diet_day(d),  # meals with items, eaten flags, kcal and macros per item
    "settings": lambda api, d: api.settings(d),  # targets for the day (kcal, macros), BMR/TMR
    "water": lambda api, d: api.water(d),
}


def sync(api: FitatuApi, store: Store, today: date, days: int) -> list[str]:
    """Fetch the last `days` days (today included). Returns per-call errors; the run continues past them."""
    errors: list[str] = []
    d = today - timedelta(days=days - 1)
    while d <= today:
        ds = d.isoformat()
        for endpoint, call in DAILY.items():
            try:
                body = call(api, d)
            except Exception as e:  # unofficial API: one failing endpoint must not drop the others
                log.warning("fitatu: %s %s failed: %s", endpoint, ds, e)
                errors.append(f"{endpoint} {ds}: {e}")
                continue
            if body:
                store.put_raw(SOURCE, endpoint, ds, d, body)
        d += timedelta(days=1)
    return errors


class FitatuReader:
    """Adapter over fitatu_api.FitatuApiClient exposing only the read calls above."""

    def __init__(self, client: Any):
        self.client = client
        self.user_id = client.auth.fitatu_user_id
        if not self.user_id:
            raise RuntimeError("Fitatu session has no user id; run `ingest fitatu-login` again")

    def diet_day(self, day: date) -> Any:
        return self.client.planner.get_day(self.user_id, day)

    def settings(self, day: date) -> Any:
        return self.client.user_settings.get_new(self.user_id, day)

    def water(self, day: date) -> Any:
        return self.client.water.get_day(self.user_id, day)


def connect(session_path: str) -> FitatuReader:
    """Resume from the session file written by `ingest fitatu-login`; refreshed tokens are merged back into it."""
    from fitatu_api import FitatuApiClient
    from fitatu_api.exceptions import FitatuAuthError

    try:
        client = FitatuApiClient.from_session(session_path)
    except FitatuAuthError as e:
        # The library suggests its own `fitatu-api login`, which writes elsewhere; point to ours.
        raise RuntimeError(f"no usable Fitatu session at {session_path}; run `ingest fitatu-login` ({e})") from None
    return FitatuReader(client)


def interactive_login(email: str | None, session_path: str) -> None:
    """First login with e-mail and password; only the session tokens are stored, never the password."""
    from fitatu_api import FitatuApiClient

    email = email or input("Fitatu e-mail: ")
    password = getpass.getpass("Fitatu hasło: ")
    client = FitatuApiClient.login(email, password, token_store_path=session_path)
    log.info("fitatu: session saved to %s (user %s)", session_path, client.auth.fitatu_user_id)
