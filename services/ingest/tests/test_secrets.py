import base64

import pytest
from cryptography.exceptions import InvalidTag

from ingest import secrets

KEY = base64.b64encode(bytes(range(32))).decode()
# Vector shared with services/api/internal/secrets/secrets_test.go and services/coach-chat/src/secrets.test.ts.
NONCE = bytes.fromhex("a0a1a2a3a4a5a6a7a8a9aaab")
CT = bytes.fromhex("8e7d0a5468bf67cc1648ecb67e57f1ec4398dc023402d051cdcbaa7175fa95868d41")


class FakeConn:
    def __init__(self, rows):
        self.rows = rows

    def execute(self, _sql, params):
        row = self.rows.get(params[0])
        return type("Cur", (), {"fetchone": lambda _self: row})()


def test_go_vector(monkeypatch):
    monkeypatch.setenv("FORMA_SECRETS_KEY", KEY)
    assert secrets.open_value(secrets.cipher_from_env(), "HEVY_API_KEY", NONCE, CT) == "hevy-test-key-1234"
    with pytest.raises(InvalidTag):  # row moved to another name
        secrets.open_value(secrets.cipher_from_env(), "ANTHROPIC_API_KEY", NONCE, CT)


@pytest.mark.parametrize(
    "rows, key, env, want",
    [
        ({"HEVY_API_KEY": (NONCE, CT)}, KEY, "from-env", "hevy-test-key-1234"),  # app wins
        ({}, KEY, "from-env", "from-env"),
        ({}, None, None, None),
        ({"HEVY_API_KEY": (NONCE, CT)}, None, "from-env", "from-env"),  # no key: env, with a warning
    ],
)
def test_resolve(monkeypatch, rows, key, env, want):
    for name, value in (("FORMA_SECRETS_KEY", key), ("HEVY_API_KEY", env)):
        if value is None:
            monkeypatch.delenv(name, raising=False)
        else:
            monkeypatch.setenv(name, value)
    assert secrets.resolve(FakeConn(rows), "HEVY_API_KEY") == want


def test_bad_key(monkeypatch):
    monkeypatch.setenv("FORMA_SECRETS_KEY", base64.b64encode(b"short").decode())
    with pytest.raises(ValueError):
        secrets.cipher_from_env()
