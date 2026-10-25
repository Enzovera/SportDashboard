"""Hevy official API v1 (requires Hevy Pro). Docs: https://api.hevyapp.com/docs/"""

import logging
from collections.abc import Iterator
from datetime import date, datetime, timezone
from typing import Any

import requests

from . import TZ
from .store import Store

log = logging.getLogger(__name__)

SOURCE = "hevy"
BASE_URL = "https://api.hevyapp.com"
WORKOUTS_PAGE_SIZE = 10  # API maximum for workouts and events
TEMPLATES_PAGE_SIZE = 100


class HevyClient:
    def __init__(self, api_key: str, session: requests.Session | None = None, base_url: str = BASE_URL):
        self.api_key = api_key
        self.session = session or requests.Session()
        self.base_url = base_url

    def pages(self, path: str, key: str, page_size: int, **params: Any) -> Iterator[dict]:
        """Yield items from all pages of a paginated endpoint."""
        page = 1
        while True:
            resp = self.session.get(
                self.base_url + path,
                headers={"api-key": self.api_key, "accept": "application/json"},
                params={"page": page, "pageSize": page_size, **params},
                timeout=30,
            )
            # Hevy answers 404 for a page past the end (also when there is nothing at all).
            if resp.status_code == 404:
                return
            resp.raise_for_status()
            data = resp.json()
            items = data.get(key) or []
            yield from items
            if not items or page >= data.get("page_count", 0):
                return
            page += 1


def local_date(iso_ts: str | None) -> date | None:
    if not iso_ts:
        return None
    return datetime.fromisoformat(iso_ts).astimezone(TZ).date()


def iso_utc(ts: datetime) -> str:
    return ts.astimezone(timezone.utc).strftime("%Y-%m-%dT%H:%M:%SZ")


def put_workout(store: Store, workout: dict) -> None:
    store.put_raw(SOURCE, "workouts", workout["id"], local_date(workout.get("start_time")), workout)


def sync(client: HevyClient, store: Store, now: datetime, full: bool = False) -> list[str]:
    """Full backfill on first run (or --full), then incremental via workout events."""
    cursor = store.get_cursor(SOURCE, "workouts")
    if full or cursor is None:
        log.info("hevy: full sync of workouts")
        for workout in client.pages("/v1/workouts", "workouts", WORKOUTS_PAGE_SIZE):
            put_workout(store, workout)
    else:
        log.info("hevy: workout events since %s", cursor)
        for event in client.pages("/v1/workouts/events", "events", WORKOUTS_PAGE_SIZE, since=cursor):
            match event.get("type"):
                case "updated":
                    put_workout(store, event["workout"])
                case "deleted":
                    store.put_raw(SOURCE, "workout_deleted", event["id"], None, event)
                case other:
                    log.warning("hevy: unknown event type %r", other)
    # The cursor is the time before fetching, so nothing created during the run is missed;
    # re-fetched overlap is absorbed by raw_payload idempotency.
    store.set_cursor(SOURCE, "workouts", iso_utc(now))

    # Exercise templates carry muscle groups for strength_set; refresh once a day.
    today = now.astimezone(TZ).date().isoformat()
    if full or store.get_cursor(SOURCE, "exercise_templates") != today:
        for template in client.pages("/v1/exercise_templates", "exercise_templates", TEMPLATES_PAGE_SIZE):
            store.put_raw(SOURCE, "exercise_templates", template["id"], None, template)
        store.set_cursor(SOURCE, "exercise_templates", today)
    return []
