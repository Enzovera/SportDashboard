// Package secrets keeps credentials set from the dashboard ("Ustawienia", docs/design.md, "Bezpieczeństwo").
//
// Values live in app_secret encrypted with AES-256-GCM under FORMA_SECRETS_KEY (32 bytes, base64, env only).
// The secret's name is the additional authenticated data. The same format is read by ingest (Python) and
// coach-chat (TypeScript): 12-byte nonce, ciphertext with the 16-byte tag appended. A stored value wins over
// the environment variable of the same name; without a stored value the services fall back to the env.
//
// The API only writes values and reports status (set, last 4 characters, date); it never returns a value,
// except a freshly generated MCP token, once. The MCP server does not get the key: it checks its bearer
// token against the sha256 digest.
package secrets

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"unicode"
)

// Names of the secrets the dashboard can set, in display order. The list is closed (also a CHECK in the table).
const (
	HevyAPIKey   = "HEVY_API_KEY"
	ClaudeOAuth  = "CLAUDE_CODE_OAUTH_TOKEN"
	AnthropicKey = "ANTHROPIC_API_KEY"
	MCPToken     = "FORMA_MCP_TOKEN"
)

var Names = []string{HevyAPIKey, ClaudeOAuth, AnthropicKey, MCPToken}

// MinMCPToken is the minimum length of the MCP bearer token, the same as for the env variable.
const MinMCPToken = 32

const maxValue = 4096

var (
	ErrUnknown = errors.New("unknown secret")
	ErrInvalid = errors.New("invalid value")
	// ErrNoKey means FORMA_SECRETS_KEY is not set, so nothing can be stored.
	ErrNoKey = errors.New("FORMA_SECRETS_KEY is not configured")
)

func Known(name string) bool {
	for _, n := range Names {
		if n == name {
			return true
		}
	}
	return false
}

// Validate checks a value before it is stored: no whitespace or control characters (pasted keys often
// carry a newline), at most 4 KiB, and the MCP token at least MinMCPToken characters.
func Validate(name, value string) error {
	if !Known(name) {
		return ErrUnknown
	}
	if value == "" {
		return fmt.Errorf("%w: empty", ErrInvalid)
	}
	if len(value) > maxValue {
		return fmt.Errorf("%w: longer than %d bytes", ErrInvalid, maxValue)
	}
	for _, r := range value {
		if unicode.IsSpace(r) || unicode.IsControl(r) {
			return fmt.Errorf("%w: contains whitespace or control characters", ErrInvalid)
		}
	}
	if name == MCPToken && len(value) < MinMCPToken {
		return fmt.Errorf("%w: %s must have at least %d characters", ErrInvalid, MCPToken, MinMCPToken)
	}
	return nil
}

// Hint is the last 4 characters shown in the settings table; short values get no hint at all.
func Hint(value string) string {
	r := []rune(value)
	if len(r) < 12 {
		return ""
	}
	return string(r[len(r)-4:])
}

func Digest(value string) []byte {
	sum := sha256.Sum256([]byte(value))
	return sum[:]
}

// Cipher seals and opens values with AES-256-GCM.
type Cipher struct{ aead cipher.AEAD }

// ParseKey decodes FORMA_SECRETS_KEY (standard base64 of 32 bytes, e.g. `openssl rand -base64 32`).
func ParseKey(s string) (*Cipher, error) {
	key, err := base64.StdEncoding.DecodeString(strings.TrimSpace(s))
	if err != nil {
		return nil, fmt.Errorf("FORMA_SECRETS_KEY: %w", err)
	}
	return NewCipher(key)
}

func NewCipher(key []byte) (*Cipher, error) {
	if len(key) != 32 {
		return nil, fmt.Errorf("FORMA_SECRETS_KEY must be 32 bytes, got %d", len(key))
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	return &Cipher{aead: aead}, nil
}

// Seal encrypts value under a random nonce with the name as additional data.
func (c *Cipher) Seal(name, value string) (nonce, ciphertext []byte, err error) {
	nonce = make([]byte, c.aead.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, nil, err
	}
	return nonce, c.sealWith(nonce, name, value), nil
}

func (c *Cipher) sealWith(nonce []byte, name, value string) []byte {
	return c.aead.Seal(nil, nonce, []byte(value), []byte(name))
}

// Open decrypts a stored value; it fails for a different key or a row moved to another name.
func (c *Cipher) Open(name string, nonce, ciphertext []byte) (string, error) {
	if len(nonce) != c.aead.NonceSize() {
		return "", errors.New("invalid nonce")
	}
	plain, err := c.aead.Open(nil, nonce, ciphertext, []byte(name))
	if err != nil {
		return "", fmt.Errorf("decrypt %s: %w", name, err)
	}
	return string(plain), nil
}

// GenerateToken returns a random URL-safe token (32 bytes, 43 characters) for FORMA_MCP_TOKEN.
func GenerateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
