import os
from datetime import date

import psycopg
import pytest

from ingest.store import PgStore, body_hash


@pytest.mark.parametrize(
    "a, b, same",
    [
        ({"x": 1, "y": [1, 2]}, {"y": [1, 2], "x": 1}, True),
        ({"x": 1}, {"x": 2}, False),
        ({"name": "Przysiad"}, {"name": "Przysiad"}, True),
        ({"l": [1, 2]}, {"l": [2, 1]}, False),
    ],
)
def test_body_hash_is_key_order_independent(a, b, same):
    assert (body_hash(a) == body_hash(b)) is same


@pytest.fixture
def pg():
    url = os.environ.get("DATABASE_URL")
    if not url:
        pytest.skip("DATABASE_URL not set")
    with psycopg.connect(url) as conn:
        yield PgStore(conn)
        conn.rollback()


def test_pg_put_raw_is_idempotent(pg):
    d = date(2026, 10, 25)
    assert pg.put_raw("test", "e", "1", d, {"a": 1}) is True
    assert pg.put_raw("test", "e", "1", d, {"a": 1}) is False
    assert pg.put_raw("test", "e", "1", d, {"a": 2}) is True  # changed content = new version
    assert pg.put_raw("test", "e", None, d, {"a": 1}) is True
    assert pg.put_raw("test", "e", None, d, {"a": 1}) is False  # nulls not distinct
    assert (pg.inserted, pg.unchanged) == (3, 2)


def test_pg_cursor_upsert(pg):
    assert pg.get_cursor("test", "e") is None
    pg.set_cursor("test", "e", "a")
    pg.set_cursor("test", "e", "b")
    assert pg.get_cursor("test", "e") == "b"
