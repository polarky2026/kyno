package config

import (
	"strings"
	"testing"
)

func validEnv() map[string]string {
	return map[string]string{
		"DATABASE_URL":   "postgres://user:pass@localhost:5432/kyno",
		"APP_URL":        "http://localhost:5173/",
		"RESEND_API_KEY": "re_test",
		"EMAIL_FROM":     "Kyno <kyno@example.com>",
	}
}

func envFrom(values map[string]string) func(string) string {
	return func(key string) string { return values[key] }
}

func TestParseAppliesDefaults(t *testing.T) {
	cfg, err := parse(envFrom(validEnv()))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.AppEnv != "development" || cfg.Port != 3000 {
		t.Fatalf("got %+v, want development on port 3000", cfg)
	}
	if cfg.AppURL != "http://localhost:5173" {
		t.Fatalf("AppURL = %q, want the trailing slash removed", cfg.AppURL)
	}
}

func TestParseReadsPort(t *testing.T) {
	env := validEnv()
	env["PORT"] = "8080"

	cfg, err := parse(envFrom(env))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Port != 8080 {
		t.Fatalf("Port = %d, want 8080", cfg.Port)
	}
}

func TestParseRequiresEveryVariable(t *testing.T) {
	_, err := parse(envFrom(nil))
	if err == nil {
		t.Fatal("expected an error")
	}
	for _, name := range []string{"DATABASE_URL", "APP_URL", "RESEND_API_KEY", "EMAIL_FROM"} {
		if !strings.Contains(err.Error(), name) {
			t.Errorf("error does not mention %s", name)
		}
	}
}

func TestParseReportsInvalidValues(t *testing.T) {
	env := validEnv()
	env["PORT"] = "abc"
	env["APP_ENV"] = "staging"
	env["APP_URL"] = "localhost:5173"

	_, err := parse(envFrom(env))
	if err == nil {
		t.Fatal("expected an error")
	}
	for _, name := range []string{"PORT", "APP_ENV", "APP_URL"} {
		if !strings.Contains(err.Error(), name) {
			t.Errorf("error does not mention %s", name)
		}
	}
}
