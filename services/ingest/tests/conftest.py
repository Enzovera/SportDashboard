import pytest

from ingest.store import body_hash


class MemoryStore:
    """In-memory Store with the same idempotency key as raw_payload."""

    def __init__(self):
        self.rows: list[tuple] = []
        self.keys: set[tuple] = set()
        self.cursors: dict[tuple, str] = {}

    def put_raw(self, source, endpoint, external_id, for_date, body) -> bool:
        key = (source, endpoint, external_id, body_hash(body))
        if key in self.keys:
            return False
        self.keys.add(key)
        self.rows.append((source, endpoint, external_id, for_date, body))
        return True

    def get_cursor(self, source, endpoint):
        return self.cursors.get((source, endpoint))

    def set_cursor(self, source, endpoint, cursor):
        self.cursors[(source, endpoint)] = cursor

    def endpoints(self) -> list[str]:
        return [r[1] for r in self.rows]


@pytest.fixture
def store():
    return MemoryStore()
