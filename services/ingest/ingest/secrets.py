"""Credentials set from the dashboard ("Ustawienia", docs/design.md, "Bezpieczeństwo").

The API (services/api/internal/secrets) stores them in app_secret encrypted with AES-256-GCM under
FORMA_SECRETS_KEY (base64 of 32 bytes): 12-byte nonce, ciphertext with the tag appended, the secret's name as
additional data. A stored value wins over the environment variable of the same name. Ingest only reads
HEVY_API_KEY; Garmin and Fitatu keep their tokens in volumes.
"""

import base64
import logging
import os

from cryptography.hazmat.primitives.ciphers.aead import AESGCM

log = logging.getLogger("ingest.secrets")


def cipher_from_env() -> AESGCM | None:
    value = os.environ.get("FORMA_SECRETS_KEY", "").strip()
    if not value:
        return None
    key = base64.b64decode(value, validate=True)
    if len(key) != 32:
        raise ValueError(f"FORMA_SECRETS_KEY must be 32 bytes, got {len(key)}")
    return AESGCM(key)


def open_value(cipher: AESGCM, name: str, nonce: bytes, ciphertext: bytes) -> str:
    return cipher.decrypt(nonce, ciphertext, name.encode()).decode()


def resolve(conn, name: str) -> str | None:
    """Value stored in the app, else the environment variable, else None."""
    row = conn.execute("select nonce, ciphertext from app_secret where name = %s", (name,)).fetchone()
    if row:
        cipher = cipher_from_env()
        if cipher is not None:
            return open_value(cipher, name, bytes(row[0]), bytes(row[1]))
        log.warning("%s is stored in the app but FORMA_SECRETS_KEY is not set; using the environment", name)
    return os.environ.get(name) or None
