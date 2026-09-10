package config

import "testing"

func TestVersionAndUserAgent(t *testing.T) {
	t.Setenv("TRAILFINDER_VERSION", "")
	t.Setenv("TRAILFINDER_USER_AGENT", "")
	if Version() != DefaultVersion {
		t.Fatalf("Version() = %q", Version())
	}
	if got := UserAgent(); got == "" || got[:16] != "trail-finder-mcp" {
		t.Fatalf("UserAgent() = %q", got)
	}

	t.Setenv("TRAILFINDER_VERSION", "9.9.9")
	t.Setenv("TRAILFINDER_USER_AGENT", "custom-ua")
	if Version() != "9.9.9" {
		t.Fatalf("Version override = %q", Version())
	}
	if UserAgent() != "custom-ua" {
		t.Fatalf("UserAgent override = %q", UserAgent())
	}
}

func TestEnv(t *testing.T) {
	t.Setenv("TRAILFINDER_TEST_ENV", "  value  ")
	if got := Env("TRAILFINDER_TEST_ENV", "fallback"); got != "value" {
		t.Fatalf("Env trimmed = %q", got)
	}
	if got := Env("TRAILFINDER_TEST_ENV_MISSING", "fallback"); got != "fallback" {
		t.Fatalf("Env fallback = %q", got)
	}
}
