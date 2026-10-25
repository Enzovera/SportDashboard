"""Append-only raw_payload store, sync cursors and run log."""

import hashlib
import json
from datetime import date
from typing import Any, Protocol

import psycopg
from psycopg.types.json import Jsonb


def canonical_json(body: Any) -> str:
    return json.dumps(body, sort_keys=True, separators=(",", ":"), ensure_ascii=False)


def body_hash(body: Any) -> str:
    return hashlib.sha256(canonical_json(body).encode()).hexdigest()


class Store(Protocol):
    def put_raw(
        self, source: str, endpoint: str, external_id: str | None, for_date: date | None, body: Any
    ) -> bool:
        """Store a payload; returns True if it is new (unseen id or changed content)."""

    def get_cursor(self, source: str, endpoint: str) -> str | None: ...

    def set_cursor(self, source: str, endpoint: str, cursor: str) -> None: ...


class PgStore:
    """Store backed by Postgres. Caller owns the transaction (commit/rollback)."""

    def __init__(self, conn: psycopg.Connection):
        self.conn = conn
        self.inserted = 0
        self.unchanged = 0

    def put_raw(self, source, endpoint, external_id, for_date, body) -> bool:
        row = self.conn.execute(
            """
            insert into raw_payload (source, endpoint, external_id, for_date, body, body_hash)
            values (%s, %s, %s, %s, %s, %s)
            on conflict (source, endpoint, external_id, body_hash) do nothing
            returning id
            """,
            (source, endpoint, external_id, for_date, Jsonb(body), body_hash(body)),
        ).fetchone()
        if row is None:
            self.unchanged += 1
            return False
        self.inserted += 1
        return True

    def get_cursor(self, source, endpoint) -> str | None:
        row = self.conn.execute(
            "select cursor from sync_state where source = %s and endpoint = %s", (source, endpoint)
        ).fetchone()
        return row[0] if row else None

    def set_cursor(self, source, endpoint, cursor) -> None:
        self.conn.execute(
            """
            insert into sync_state (source, endpoint, cursor) values (%s, %s, %s)
            on conflict (source, endpoint) do update set cursor = excluded.cursor, updated_at = now()
            """,
            (source, endpoint, cursor),
        )

    def start_run(self, source: str, args: dict) -> int:
        run_id = self.conn.execute(
            "insert into ingest_run (source, args) values (%s, %s) returning id", (source, Jsonb(args))
        ).fetchone()[0]
        self.conn.commit()
        return run_id

    def finish_run(self, run_id: int, status: str, errors: list[str]) -> None:
        self.conn.execute(
            """
            update ingest_run
            set finished_at = now(), status = %s, inserted = %s, unchanged = %s, errors = %s
            where id = %s
            """,
            (status, self.inserted, self.unchanged, Jsonb(errors), run_id),
        )
        self.conn.commit()
