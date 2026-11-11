-- Credentials set from the dashboard ("Ustawienia", docs/design.md, "Bezpieczeństwo"). Values are encrypted
-- with AES-256-GCM under FORMA_SECRETS_KEY (env only, never in the database), the secret's name is the
-- additional authenticated data, so a row cannot be moved to another name. pg_dump therefore carries only
-- ciphertext. A value stored here wins over the environment variable of the same name.
--   hint    last 4 characters, for the settings table without decrypting
--   digest  sha256 of the value: the MCP server verifies its bearer token with it and never gets the key

create table app_secret (
  name       text primary key check (name in (
               'HEVY_API_KEY', 'CLAUDE_CODE_OAUTH_TOKEN', 'ANTHROPIC_API_KEY', 'FORMA_MCP_TOKEN')),
  nonce      bytea not null check (length(nonce) = 12),
  ciphertext bytea not null,
  hint       text not null,
  digest     bytea not null check (length(digest) = 32),
  updated_at timestamptz not null default now()
);
