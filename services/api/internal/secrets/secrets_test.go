package secrets

import (
	"bytes"
	"encoding/hex"
	"errors"
	"strings"
	"testing"
)

func testKey() []byte {
	k := make([]byte, 32)
	for i := range k {
		k[i] = byte(i)
	}
	return k
}

// Vector shared with services/ingest/tests/test_secrets.py and services/coach-chat/src/secrets.test.ts:
// key 00..1f, nonce a0..ab, name HEVY_API_KEY, value "hevy-test-key-1234".
const (
	vectorNonce = "a0a1a2a3a4a5a6a7a8a9aaab"
	vectorValue = "hevy-test-key-1234"
	vectorCT    = "8e7d0a5468bf67cc1648ecb67e57f1ec4398dc023402d051cdcbaa7175fa95868d41"
)

func TestCrossLanguageVector(t *testing.T) {
	c, err := NewCipher(testKey())
	if err != nil {
		t.Fatal(err)
	}
	nonce, _ := hex.DecodeString(vectorNonce)
	got := hex.EncodeToString(c.sealWith(nonce, HevyAPIKey, vectorValue))
	if got != vectorCT {
		t.Fatalf("ciphertext = %s, want %s", got, vectorCT)
	}
	ct, _ := hex.DecodeString(vectorCT)
	if v, err := c.Open(HevyAPIKey, nonce, ct); err != nil || v != vectorValue {
		t.Fatalf("Open = %q, %v", v, err)
	}
}

func TestSealOpen(t *testing.T) {
	c, _ := NewCipher(testKey())
	other := testKey()
	other[0] = 0xff
	c2, _ := NewCipher(other)
	nonce, ct, err := c.Seal(ClaudeOAuth, "sk-ant-oat01-abc")
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name    string
		cipher  *Cipher
		secret  string
		nonce   []byte
		wantErr bool
	}{
		{"same key and name", c, ClaudeOAuth, nonce, false},
		{"row moved to another name", c, AnthropicKey, nonce, true},
		{"different key", c2, ClaudeOAuth, nonce, true},
		{"short nonce", c, ClaudeOAuth, nonce[:4], true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v, err := tt.cipher.Open(tt.secret, tt.nonce, ct)
			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && v != "sk-ant-oat01-abc" {
				t.Fatalf("value = %q", v)
			}
		})
	}
	n2, _, _ := c.Seal(ClaudeOAuth, "sk-ant-oat01-abc")
	if bytes.Equal(nonce, n2) {
		t.Fatal("nonce reused")
	}
}

func TestParseKey(t *testing.T) {
	tests := []struct {
		in      string
		wantErr bool
	}{
		{"AAECAwQFBgcICQoLDA0ODxAREhMUFRYXGBkaGxwdHh8=", false},
		{" AAECAwQFBgcICQoLDA0ODxAREhMUFRYXGBkaGxwdHh8=\n", false},
		{"AAECAwQFBgcICQoLDA0ODw==", true}, // 16 bytes
		{"not base64!", true},
	}
	for _, tt := range tests {
		if _, err := ParseKey(tt.in); (err != nil) != tt.wantErr {
			t.Errorf("ParseKey(%q) err = %v, wantErr %v", tt.in, err, tt.wantErr)
		}
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name, secret, value string
		want                error
	}{
		{"hevy key", HevyAPIKey, "0b7c2f1e-1234-4abc-9def-001122334455", nil},
		{"unknown", "GARMIN_PASSWORD", "x", ErrUnknown},
		{"empty", HevyAPIKey, "", ErrInvalid},
		{"trailing newline", HevyAPIKey, "abc\n", ErrInvalid},
		{"inner space", AnthropicKey, "sk-ant api", ErrInvalid},
		{"too long", AnthropicKey, strings.Repeat("a", 4097), ErrInvalid},
		{"short mcp token", MCPToken, strings.Repeat("a", 31), ErrInvalid},
		{"mcp token", MCPToken, strings.Repeat("a", 32), nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Validate(tt.secret, tt.value); !errors.Is(err, tt.want) {
				t.Fatalf("err = %v, want %v", err, tt.want)
			}
		})
	}
}

func TestHint(t *testing.T) {
	tests := map[string]string{"sk-ant-api03-xyzw": "xyzw", "short": "", "": "", "ąęśćżźńółĄĘŚ": "łĄĘŚ"}
	for in, want := range tests {
		if got := Hint(in); got != want {
			t.Errorf("Hint(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestGenerateToken(t *testing.T) {
	a, _ := GenerateToken()
	b, _ := GenerateToken()
	if len(a) != 43 || a == b || Validate(MCPToken, a) != nil {
		t.Fatalf("tokens %q %q", a, b)
	}
}
