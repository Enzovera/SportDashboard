package httpapi

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"forma/api/internal/secrets"
)

// fakeSecrets validates like secrets.Store but keeps values in memory; noKey simulates a missing FORMA_SECRETS_KEY.
type fakeSecrets struct {
	noKey  bool
	stored map[string]string
}

func (f *fakeSecrets) List(context.Context) (secrets.Settings, error) {
	out := secrets.Settings{Writable: !f.noKey}
	for _, n := range secrets.Names {
		_, ok := f.stored[n]
		out.Secrets = append(out.Secrets, secrets.Status{Name: n, Set: ok})
	}
	return out, nil
}

func (f *fakeSecrets) Put(_ context.Context, name, value string) (secrets.Status, error) {
	if err := secrets.Validate(name, value); err != nil {
		return secrets.Status{}, err
	}
	if f.noKey {
		return secrets.Status{}, secrets.ErrNoKey
	}
	f.stored[name] = value
	return secrets.Status{Name: name, Set: true}, nil
}

func (f *fakeSecrets) Delete(_ context.Context, name string) (secrets.Status, error) {
	if !secrets.Known(name) {
		return secrets.Status{}, secrets.ErrUnknown
	}
	delete(f.stored, name)
	return secrets.Status{Name: name}, nil
}

func (f *fakeSecrets) Generate(_ context.Context, name string) (secrets.Generated, error) {
	if name != secrets.MCPToken {
		return secrets.Generated{}, fmt.Errorf("%w: only %s", secrets.ErrInvalid, secrets.MCPToken)
	}
	f.stored[name] = strings.Repeat("t", 43)
	return secrets.Generated{Status: secrets.Status{Name: name, Set: true}, Value: f.stored[name]}, nil
}

func TestSecretRoutes(t *testing.T) {
	const hevy = "0b7c2f1e-1234-4abc-9def-001122334455"
	tests := []struct {
		name       string
		method     string
		path       string
		body       string
		noKey      bool
		wantStatus int
		wantStored string // value of HEVY_API_KEY after the call
		wantNot    string // must not appear in the response
	}{
		{"list", "GET", "/api/settings/secrets", "", false, http.StatusOK, hevy, hevy},
		{"put", "PUT", "/api/settings/secrets/HEVY_API_KEY", `{"value":"new-hevy-key-0000"}`, false, http.StatusOK, "new-hevy-key-0000", "new-hevy-key-0000"},
		{"put unknown", "PUT", "/api/settings/secrets/GARMIN_PASSWORD", `{"value":"x"}`, false, http.StatusNotFound, hevy, ""},
		{"put newline", "PUT", "/api/settings/secrets/HEVY_API_KEY", `{"value":"abc\n"}`, false, http.StatusUnprocessableEntity, hevy, ""},
		{"put extra field", "PUT", "/api/settings/secrets/HEVY_API_KEY", `{"value":"abc","x":1}`, false, http.StatusBadRequest, hevy, ""},
		{"put without key", "PUT", "/api/settings/secrets/HEVY_API_KEY", `{"value":"new-hevy-key-0000"}`, true, http.StatusServiceUnavailable, hevy, ""},
		{"delete", "DELETE", "/api/settings/secrets/HEVY_API_KEY", "", false, http.StatusOK, "", ""},
		{"generate mcp", "POST", "/api/settings/secrets/FORMA_MCP_TOKEN/generate", "", false, http.StatusOK, hevy, hevy},
		{"generate hevy", "POST", "/api/settings/secrets/HEVY_API_KEY/generate", "", false, http.StatusUnprocessableEntity, hevy, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &fakeSecrets{noKey: tt.noKey, stored: map[string]string{secrets.HevyAPIKey: hevy}}
			rec := httptest.NewRecorder()
			New(nil, nil, f).ServeHTTP(rec, httptest.NewRequest(tt.method, tt.path, strings.NewReader(tt.body)))
			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d (%s)", rec.Code, tt.wantStatus, rec.Body)
			}
			if got := f.stored[secrets.HevyAPIKey]; got != tt.wantStored {
				t.Errorf("stored = %q, want %q", got, tt.wantStored)
			}
			if tt.wantNot != "" && strings.Contains(rec.Body.String(), tt.wantNot) {
				t.Errorf("response leaks the value: %s", rec.Body)
			}
		})
	}
}
