package secrets

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Status is one row of GET /api/settings/secrets. Values never leave the server.
type Status struct {
	Name string `json:"name"`
	// Stored in the app (it wins over the env variable). False = the services use the env variable, if any.
	Set       bool       `json:"set"`
	Hint      *string    `json:"hint"` // last 4 characters; null for short values or when not set
	UpdatedAt *time.Time `json:"updatedAt"`
}

// Settings is the body of GET /api/settings/secrets.
type Settings struct {
	// False when FORMA_SECRETS_KEY is missing in the API: status is readable, nothing can be stored.
	Writable bool     `json:"writable"`
	Secrets  []Status `json:"secrets"`
}

// Generated is the body of POST /api/settings/secrets/FORMA_MCP_TOKEN/generate: the only response with a value.
type Generated struct {
	Status
	Value string `json:"value"`
}

// Store reads and writes app_secret. cipher is nil when FORMA_SECRETS_KEY is not set.
type Store struct {
	pool   *pgxpool.Pool
	cipher *Cipher
}

func NewStore(pool *pgxpool.Pool, c *Cipher) *Store { return &Store{pool: pool, cipher: c} }

func (s *Store) List(ctx context.Context) (Settings, error) {
	rows, err := s.pool.Query(ctx, `select name, hint, updated_at from app_secret`)
	if err != nil {
		return Settings{}, err
	}
	stored := map[string]Status{}
	for rows.Next() {
		var st Status
		var hint string
		var at time.Time
		if err := rows.Scan(&st.Name, &hint, &at); err != nil {
			return Settings{}, err
		}
		st.Set, st.UpdatedAt = true, &at
		if hint != "" {
			st.Hint = &hint
		}
		stored[st.Name] = st
	}
	if err := rows.Err(); err != nil {
		return Settings{}, err
	}
	out := Settings{Writable: s.cipher != nil, Secrets: make([]Status, 0, len(Names))}
	for _, n := range Names {
		st, ok := stored[n]
		if !ok {
			st = Status{Name: n}
		}
		out.Secrets = append(out.Secrets, st)
	}
	return out, nil
}

func (s *Store) Put(ctx context.Context, name, value string) (Status, error) {
	if err := Validate(name, value); err != nil {
		return Status{}, err
	}
	if s.cipher == nil {
		return Status{}, ErrNoKey
	}
	nonce, ct, err := s.cipher.Seal(name, value)
	if err != nil {
		return Status{}, err
	}
	hint := Hint(value)
	var at time.Time
	err = s.pool.QueryRow(ctx, `
		insert into app_secret (name, nonce, ciphertext, hint, digest) values ($1, $2, $3, $4, $5)
		on conflict (name) do update
		  set nonce = excluded.nonce, ciphertext = excluded.ciphertext, hint = excluded.hint,
		      digest = excluded.digest, updated_at = now()
		returning updated_at`, name, nonce, ct, hint, Digest(value)).Scan(&at)
	if err != nil {
		return Status{}, fmt.Errorf("store %s: %w", name, err)
	}
	st := Status{Name: name, Set: true, UpdatedAt: &at}
	if hint != "" {
		st.Hint = &hint
	}
	return st, nil
}

// Delete removes the stored value; the services fall back to the env variable. Deleting a missing one is fine.
func (s *Store) Delete(ctx context.Context, name string) (Status, error) {
	if !Known(name) {
		return Status{}, ErrUnknown
	}
	if _, err := s.pool.Exec(ctx, `delete from app_secret where name = $1`, name); err != nil {
		return Status{}, err
	}
	return Status{Name: name}, nil
}

// Generate stores a random value; only FORMA_MCP_TOKEN can be generated (the other keys come from their services).
func (s *Store) Generate(ctx context.Context, name string) (Generated, error) {
	if !Known(name) {
		return Generated{}, ErrUnknown
	}
	if name != MCPToken {
		return Generated{}, fmt.Errorf("%w: only %s can be generated", ErrInvalid, MCPToken)
	}
	value, err := GenerateToken()
	if err != nil {
		return Generated{}, err
	}
	st, err := s.Put(ctx, name, value)
	if err != nil {
		return Generated{}, err
	}
	return Generated{Status: st, Value: value}, nil
}

// LoadDigest returns the sha256 of a stored value, or nil when the app has none. Needs no key (MCP server).
func LoadDigest(ctx context.Context, pool *pgxpool.Pool, name string) ([]byte, error) {
	var d []byte
	err := pool.QueryRow(ctx, `select digest from app_secret where name = $1`, name).Scan(&d)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return d, err
}
