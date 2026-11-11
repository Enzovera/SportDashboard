"""Hevy -> Garmin: write the sets of each Hevy workout into the matching Garmin strength activity.

The athlete usually records the session on the watch too, so the watch activity (HR, load, calories) is the one
that gets the sets, exercises and the Hevy title. Only when the watch recorded nothing within GRACE after the
workout ended is a manual strength activity created, so Garmin still sees the session.

Hevy data comes from raw_payload (the hevy worker runs first), Garmin activities are read live. Side effects are
recorded in hevy_garmin_link and committed one by one, so a crash mid-run never creates an activity twice.
"""

import json
import logging
import re
from dataclasses import dataclass
from datetime import date, datetime, timedelta, timezone
from typing import Any, Protocol

import psycopg
from garminconnect import exercises as catalog

from . import TZ

log = logging.getLogger(__name__)

SOURCE = "hevy_garmin"
# Watch activity types a gym session may be recorded as.
STRENGTH_TYPES = {"strength_training", "hiit", "indoor_cardio", "fitness_equipment", "other"}
# Share of the shorter of the two sessions that must overlap to count as the same session.
MIN_OVERLAP = 0.5
# How long to wait for the watch activity to sync before creating a manual one.
GRACE = timedelta(hours=3)
# Assumed time under load of a set without a duration in Hevy.
DEFAULT_SET_S = 40


@dataclass(frozen=True)
class Workout:
    id: str
    hash: str
    body: dict

    @property
    def start(self) -> datetime:
        return datetime.fromisoformat(self.body["start_time"])

    @property
    def end(self) -> datetime:
        end = self.body.get("end_time")
        return datetime.fromisoformat(end) if end else self.start + timedelta(hours=1)

    @property
    def title(self) -> str:
        return self.body.get("title") or "Trening siłowy"


@dataclass(frozen=True)
class Link:
    hevy_workout_id: str
    workout_hash: str
    garmin_activity_id: int | None
    created_by_us: bool
    status: str


class Repo(Protocol):
    def workouts(self, since: date) -> list[Workout]:
        """Latest version of each Hevy workout starting on or after `since`, without deleted ones."""

    def deleted_links(self) -> list[Link]:
        """Synced links whose Hevy workout was deleted after its latest version."""

    def muscles(self) -> dict[str, str]:
        """Exercise template id -> primary muscle group."""

    def links(self) -> dict[str, Link]: ...

    def save(self, link: Link, error: str | None = None) -> None:
        """Upsert and commit at once: the Garmin side effect already happened."""


class GarminWriter(Protocol):
    def get_activities_by_date(self, startdate: str, enddate: str | None = None) -> list[dict]: ...
    def create_manual_activity(
        self, start_datetime: str, time_zone: str, type_key: str, distance_km: float, duration_min: int,
        activity_name: str,
    ) -> Any: ...
    def set_activity_name(self, activity_id: str, title: str) -> Any: ...
    def set_activity_description(self, activity_id: str, description: str) -> Any: ...
    def set_activity_exercise_sets(self, activity_id: int | str, payload: dict) -> Any: ...
    def delete_activity(self, activity_id: str) -> Any: ...


# ---------- exercise mapping ----------

_BY_LOWER = {name.lower(): e for name, e in catalog.BY_NAME.items()}
_CATEGORIES = set(catalog.CATEGORIES)

# Title keyword -> Garmin category, first match wins (more specific first).
_KEYWORDS = (
    ("romanian deadlift", "DEADLIFT"), ("deadlift", "DEADLIFT"), ("hip thrust", "HIP_RAISE"),
    ("glute bridge", "HIP_RAISE"), ("bench press", "BENCH_PRESS"), ("chest press", "BENCH_PRESS"),
    ("fly", "FLYE"), ("pec deck", "FLYE"), ("squat", "SQUAT"), ("leg press", "SQUAT"), ("lunge", "LUNGE"),
    ("split squat", "LUNGE"), ("step up", "LUNGE"), ("leg curl", "LEG_CURL"), ("leg extension", "SQUAT"),
    ("calf", "CALF_RAISE"), ("pull up", "PULL_UP"), ("pull-up", "PULL_UP"), ("chin up", "PULL_UP"),
    ("pulldown", "PULL_UP"), ("row", "ROW"), ("face pull", "SHOULDER_STABILITY"),
    ("lateral raise", "LATERAL_RAISE"), ("front raise", "LATERAL_RAISE"), ("overhead press", "SHOULDER_PRESS"),
    ("shoulder press", "SHOULDER_PRESS"), ("arnold", "SHOULDER_PRESS"), ("shrug", "SHRUG"),
    ("curl", "CURL"), ("triceps", "TRICEPS_EXTENSION"), ("tricep", "TRICEPS_EXTENSION"),
    ("skull", "TRICEPS_EXTENSION"), ("dip", "TRICEPS_EXTENSION"), ("push up", "PUSH_UP"),
    ("push-up", "PUSH_UP"), ("plank", "PLANK"), ("crunch", "CRUNCH"), ("sit up", "SIT_UP"),
    ("leg raise", "LEG_RAISE"), ("back extension", "HYPEREXTENSION"), ("hyperextension", "HYPEREXTENSION"),
    ("clean", "OLYMPIC_LIFT"), ("snatch", "OLYMPIC_LIFT"), ("swing", "HIP_SWING"), ("carry", "CARRY"),
    ("woodchop", "CHOP"), ("russian twist", "CORE"), ("ab wheel", "CORE"),
)

# Hevy primary muscle group -> Garmin category, when the title says nothing.
_MUSCLES = {
    "chest": "BENCH_PRESS", "shoulders": "SHOULDER_PRESS", "biceps": "CURL", "triceps": "TRICEPS_EXTENSION",
    "forearms": "CURL", "lats": "PULL_UP", "upper_back": "ROW", "lower_back": "HYPEREXTENSION", "traps": "SHRUG",
    "quadriceps": "SQUAT", "hamstrings": "LEG_CURL", "glutes": "HIP_RAISE", "calves": "CALF_RAISE",
    "abdominals": "CRUNCH", "abductors": "HIP_STABILITY", "adductors": "HIP_STABILITY", "cardio": "CARDIO",
    "full_body": "TOTAL_BODY",
}

_EQUIPMENT = re.compile(r"^(?P<base>.+?)\s*\((?P<equipment>[^)]+)\)\s*$")


def garmin_exercise(title: str, muscle: str | None) -> dict:
    """Garmin category and exercise (sub-category) for a Hevy exercise; name None when only the category is known.

    Hevy titles look like "Bench Press (Barbell)", Garmin's like "Barbell Bench Press".
    """
    m = _EQUIPMENT.match(title)
    base, equipment = (m["base"], m["equipment"]) if m else (title, None)
    candidates = [f"{equipment} {base}", f"{base} {equipment}"] if equipment else []
    candidates.append(base)
    for c in candidates:
        if e := _BY_LOWER.get(c.lower()):
            return {"category": e["category"], "name": e["exercise"] or None}
    low = base.lower()
    for keyword, category in _KEYWORDS:
        if keyword in low:
            return {"category": category, "name": None}
    category = _MUSCLES.get(muscle or "", "TOTAL_BODY")
    return {"category": category if category in _CATEGORIES else "TOTAL_BODY", "name": None}


# ---------- payloads ----------

def _garmin_ts(t: datetime) -> str:
    return t.astimezone(timezone.utc).strftime("%Y-%m-%dT%H:%M:%S.0")


def exercise_sets(workout: Workout, muscles: dict[str, str], start: datetime, end: datetime) -> list[dict]:
    """Garmin exerciseSets for the workout, spread evenly over [start, end].

    Hevy keeps no per-set timestamps, so each set gets an equal slot: the set itself, then rest until the next one.
    Warm-ups are included: they were done, Garmin only shows sets. Weight goes in grams.
    """
    flat = [
        (ex, s)
        for ex in sorted(workout.body.get("exercises") or [], key=lambda e: e.get("index", 0))
        for s in sorted(ex.get("sets") or [], key=lambda s: s.get("index", 0))
    ]
    if not flat:
        return []
    slot = max((end - start).total_seconds(), 0) / len(flat)
    out: list[dict] = []
    t = start
    for i, (ex, s) in enumerate(flat):
        active = float(min(s.get("duration_seconds") or DEFAULT_SET_S, slot)) if slot else float(DEFAULT_SET_S)
        weight = s.get("weight_kg")
        out.append({
            "exercises": [{"probability": 100.0, **garmin_exercise(ex.get("title") or "",
                                                                   muscles.get(ex.get("exercise_template_id") or ""))}],
            "duration": active,
            "repetitionCount": s.get("reps"),
            "weight": round(weight * 1000) if weight is not None else None,
            "setType": "ACTIVE",
            "startTime": _garmin_ts(t),
            "wktStepIndex": None,
            "messageIndex": len(out),
        })
        rest = slot - active
        if i < len(flat) - 1 and rest > 0:
            out.append({
                "exercises": [],
                "duration": rest,
                "repetitionCount": None,
                "weight": None,
                "setType": "REST",
                "startTime": _garmin_ts(t + timedelta(seconds=active)),
                "wktStepIndex": None,
                "messageIndex": len(out),
            })
        t += timedelta(seconds=slot)
    return out


def _num(x: float) -> str:
    return f"{x:g}".replace(".", ",")


def description(workout: Workout) -> str:
    """One line per exercise, e.g. "Bench Press (Barbell): 3 serie · 80 kg × 8, 80 kg × 8, 85 kg × 6"."""
    lines = []
    for ex in sorted(workout.body.get("exercises") or [], key=lambda e: e.get("index", 0)):
        sets = [s for s in sorted(ex.get("sets") or [], key=lambda s: s.get("index", 0)) if s.get("type") != "warmup"]
        parts = []
        for s in sets:
            reps, kg, secs = s.get("reps"), s.get("weight_kg"), s.get("duration_seconds")
            if reps is not None:
                parts.append(f"{_num(kg)} kg × {reps}" if kg else f"× {reps}")
            elif secs:
                parts.append(f"{secs} s")
        if parts:
            lines.append(f"{ex.get('title')}: {len(parts)} serie · {', '.join(parts)}")
    return "\n".join(lines + ["", "Z Hevy (Forma)"])


# ---------- matching ----------

def _activity_window(act: dict) -> tuple[datetime, datetime] | None:
    ts, dur = act.get("startTimeGMT"), act.get("duration")
    if not ts or dur is None:
        return None
    start = datetime.fromisoformat(ts).replace(tzinfo=timezone.utc)
    return start, start + timedelta(seconds=float(dur))


def find_match(workout: Workout, activities: list[dict], taken: set[int]) -> dict | None:
    """The watch strength activity overlapping the workout the most, if it covers enough of either session."""
    best, best_overlap = None, 0.0
    for act in activities:
        if (act.get("activityType") or {}).get("typeKey") not in STRENGTH_TYPES or act.get("activityId") in taken:
            continue
        window = _activity_window(act)
        if window is None:
            continue
        overlap = (min(workout.end, window[1]) - max(workout.start, window[0])).total_seconds()
        shorter = min((workout.end - workout.start).total_seconds(), (window[1] - window[0]).total_seconds())
        if shorter > 0 and overlap / shorter >= MIN_OVERLAP and overlap > best_overlap:
            best, best_overlap = act, overlap
    return best


# ---------- sync ----------

def sync(api: GarminWriter, repo: Repo, now: datetime, days: int) -> list[str]:
    """Write new or changed Hevy workouts of the last `days` days into Garmin; drop activities we created for
    workouts deleted in Hevy. Returns per-workout errors; the run continues past them."""
    errors: list[str] = []
    links = repo.links()

    for link in repo.deleted_links():
        if link.created_by_us and link.garmin_activity_id:
            try:
                api.delete_activity(str(link.garmin_activity_id))
            except Exception as e:
                errors.append(f"delete {link.garmin_activity_id}: {e}")
                continue
        # A watch activity stays: it is the athlete's own recording, only the sets came from Hevy.
        repo.save(Link(link.hevy_workout_id, link.workout_hash, link.garmin_activity_id, link.created_by_us,
                       "deleted"))
        log.info("hevy_garmin: workout %s deleted in Hevy", link.hevy_workout_id)

    since = now.astimezone(TZ).date() - timedelta(days=days - 1)
    todo = [
        w for w in repo.workouts(since)
        if (l := links.get(w.id)) is None or l.status != "synced" or l.workout_hash != w.hash
    ]
    if not todo:
        return errors
    muscles = repo.muscles()
    taken = {l.garmin_activity_id for l in links.values() if l.garmin_activity_id and l.status != "deleted"}
    activities: list[dict] | None = None

    for w in todo:
        link = links.get(w.id)
        activity_id = link.garmin_activity_id if link else None
        created = link.created_by_us if link else False
        try:
            window = (w.start, w.end)
            if activity_id is None:
                if activities is None:
                    activities = api.get_activities_by_date(
                        (since - timedelta(days=1)).isoformat(), now.astimezone(TZ).date().isoformat()) or []
                match = find_match(w, activities, taken)
                if match is not None:
                    activity_id = int(match["activityId"])
                    window = _activity_window(match) or window
                elif now < w.end + GRACE:
                    log.info("hevy_garmin: %s waits for the watch activity", w.id)
                    continue
                else:
                    activity_id = create_manual(api, w)
                    created = True
                    # Record before writing sets: a failure below must not create a second activity.
                    repo.save(Link(w.id, "", activity_id, True, "error"), "sets not written yet")
                taken.add(activity_id)
            elif not created and activities is not None:
                match = next((a for a in activities if a.get("activityId") == activity_id), None)
                window = (_activity_window(match) if match else None) or window

            api.set_activity_name(str(activity_id), w.title)
            api.set_activity_description(str(activity_id), description(w))
            api.set_activity_exercise_sets(activity_id, {
                "activityId": activity_id,
                "exerciseSets": exercise_sets(w, muscles, *window),
            })
            repo.save(Link(w.id, w.hash, activity_id, created, "synced"))
            log.info("hevy_garmin: %s -> activity %s%s", w.id, activity_id, " (created)" if created else "")
        except Exception as e:
            log.warning("hevy_garmin: %s failed: %s", w.id, e)
            errors.append(f"{w.id}: {e}")
            repo.save(Link(w.id, link.workout_hash if link else "", activity_id, created, "error"), str(e))
    return errors


def create_manual(api: GarminWriter, w: Workout) -> int:
    local = w.start.astimezone(TZ)
    minutes = max(1, round((w.end - w.start).total_seconds() / 60))
    resp = api.create_manual_activity(
        local.strftime("%Y-%m-%dT%H:%M:%S.000"), str(TZ), "strength_training", 0, minutes, w.title)
    activity_id = (resp or {}).get("activityId")
    if not activity_id:
        raise RuntimeError(f"manual activity: no activityId in {resp!r}")
    return int(activity_id)


# ---------- Postgres ----------

class PgRepo:
    """Reads Hevy from raw_payload, keeps hevy_garmin_link. `save` commits right away."""

    def __init__(self, conn: psycopg.Connection, commit: bool = True):
        self.conn = conn
        self.commit = commit  # False only in tests, which roll back

    def workouts(self, since: date) -> list[Workout]:
        rows = self.conn.execute(
            """
            select w.external_id, w.body_hash, w.body
            from (
              select distinct on (external_id) id, external_id, body_hash, body
              from raw_payload
              where source = 'hevy' and endpoint = 'workouts' and for_date >= %s
              order by external_id, id desc
            ) w
            where not exists (
              select 1 from raw_payload d
              where d.source = 'hevy' and d.endpoint = 'workout_deleted' and d.external_id = w.external_id
                and d.id > w.id
            )
            order by w.body->>'start_time'
            """,
            (since,),
        ).fetchall()
        return [Workout(r[0], r[1], r[2]) for r in rows]

    def deleted_links(self) -> list[Link]:
        rows = self.conn.execute(
            """
            select l.hevy_workout_id, l.workout_hash, l.garmin_activity_id, l.created_by_us, l.status
            from hevy_garmin_link l
            where l.status <> 'deleted' and exists (
              select 1 from raw_payload d
              where d.source = 'hevy' and d.endpoint = 'workout_deleted' and d.external_id = l.hevy_workout_id
                and d.id > (select coalesce(max(w.id), 0) from raw_payload w
                            where w.source = 'hevy' and w.endpoint = 'workouts'
                              and w.external_id = l.hevy_workout_id)
            )
            """
        ).fetchall()
        return [Link(*r) for r in rows]

    def muscles(self) -> dict[str, str]:
        rows = self.conn.execute(
            """
            select distinct on (external_id) external_id, body->>'primary_muscle_group'
            from raw_payload where source = 'hevy' and endpoint = 'exercise_templates'
            order by external_id, id desc
            """
        ).fetchall()
        return {r[0]: r[1] for r in rows if r[1]}

    def links(self) -> dict[str, Link]:
        rows = self.conn.execute(
            "select hevy_workout_id, workout_hash, garmin_activity_id, created_by_us, status from hevy_garmin_link"
        ).fetchall()
        return {r[0]: Link(*r) for r in rows}

    def save(self, link: Link, error: str | None = None) -> None:
        self.conn.execute(
            """
            insert into hevy_garmin_link
              (hevy_workout_id, workout_hash, garmin_activity_id, created_by_us, status, error, attempts)
            values (%s, %s, %s, %s, %s, %s, 1)
            on conflict (hevy_workout_id) do update set
              workout_hash = excluded.workout_hash, garmin_activity_id = excluded.garmin_activity_id,
              created_by_us = excluded.created_by_us, status = excluded.status, error = excluded.error,
              attempts = case when excluded.status = 'error' then hevy_garmin_link.attempts + 1 else 0 end,
              updated_at = now()
            """,
            (link.hevy_workout_id, link.workout_hash, link.garmin_activity_id, link.created_by_us, link.status,
             error),
        )
        if self.commit:
            self.conn.commit()


class DryRun:
    """Reads Garmin for real, logs writes instead of sending them; with NoSave nothing is recorded either."""

    def __init__(self, api: GarminWriter):
        self.api = api

    def get_activities_by_date(self, startdate, enddate=None):
        return self.api.get_activities_by_date(startdate, enddate)

    def create_manual_activity(self, *args):
        log.info("dry-run: create_manual_activity%r", args)
        return {"activityId": 1}

    def set_activity_name(self, activity_id, title):
        log.info("dry-run: activity %s name %r", activity_id, title)

    def set_activity_description(self, activity_id, text):
        log.info("dry-run: activity %s description:\n%s", activity_id, text)

    def set_activity_exercise_sets(self, activity_id, payload):
        log.info("dry-run: activity %s exerciseSets %s", activity_id, json.dumps(payload, ensure_ascii=False))

    def delete_activity(self, activity_id):
        log.info("dry-run: delete activity %s", activity_id)


class NoSave(PgRepo):
    def save(self, link: Link, error: str | None = None) -> None:
        log.info("dry-run: link %s", link)
