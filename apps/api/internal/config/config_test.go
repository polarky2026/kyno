package config

import (
	"strings"
	"testing"
)

func envFrom(values map[string]string) func(string) string {
	return func(key string) string { return values[key] }
}

func TestParseAppliesDefaults(t *testing.T) {
	cfg, err := parse(envFrom(nil))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.AppEnv != "development" || cfg.Port != 3000 {
		t.Fatalf("got %+v, want development on port 3000", cfg)
	}
}

func TestParseReadsPort(t *testing.T) {
	cfg, err := parse(envFrom(map[string]string{"PORT": "8080"}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Port != 8080 {
		t.Fatalf("Port = %d, want 8080", cfg.Port)
	}
}

func TestParseReportsEveryInvalidVariable(t *testing.T) {
	_, err := parse(envFrom(map[string]string{"PORT": "abc", "APP_ENV": "staging"}))
	if err == nil {
		t.Fatal("expected an error")
	}
	for _, name := range []string{"PORT", "APP_ENV"} {
		if !strings.Contains(err.Error(), name) {
			t.Errorf("error %q does not mention %s", err, name)
		}
	}
}
