package config

import (
	"errors"
	"testing"
	"time"
)

func TestBuildDatabaseURL(t *testing.T) {
	got := BuildDatabaseURL("h", "1", "u", "p", "n", "disable")
	want := "postgres://u:p@h:1/n?sslmode=disable"

	if got != want {
		t.Fatalf("BuildDatabaseURL() = %q, want %q", got, want)
	}
}

func TestDatabaseConfigDSNPrefersExplicitURL(t *testing.T) {
	cfg := DatabaseConfig{
		URL:      "postgres://u:p@h:1/n?sslmode=require",
		Host:     "h",
		Port:     "1",
		User:     "u",
		Password: "p",
		Name:     "n",
		SSLMode:  "disable",
	}

	if got := cfg.DSN(); got != cfg.URL {
		t.Fatalf("DSN() = %q, want explicit URL %q", got, cfg.URL)
	}
}

func TestLoadDatabaseConfigRequiresDBPartsWhenDatabaseURLMissing(t *testing.T) {
	t.Setenv("DATABASE_URL", "")
	t.Setenv("DB_HOST", "")

	_, err := LoadDatabaseConfig()
	if err == nil {
		t.Fatal("LoadDatabaseConfig() error = nil, want missing env error")
	}
}

func TestLoadJWTSecretRequiresMinimumLength(t *testing.T) {
	t.Setenv("JWT_SECRET", "short")

	_, err := loadJWTSecret()
	if !errors.Is(err, ErrInvalidEnv) {
		t.Fatalf("loadJWTSecret() error = %v, want ErrInvalidEnv", err)
	}
}

func TestLoadJWTAccessTTLUsesDefaultWhenMissing(t *testing.T) {
	t.Setenv("JWT_ACCESS_TTL", "")

	got, err := loadJWTAccessTTL()
	if err != nil {
		t.Fatalf("loadJWTAccessTTL() error = %v", err)
	}

	if got != 24*time.Hour {
		t.Fatalf("loadJWTAccessTTL() = %v, want 24h", got)
	}
}

func TestLoadJWTAccessTTLParsesDuration(t *testing.T) {
	t.Setenv("JWT_ACCESS_TTL", "2h")

	got, err := loadJWTAccessTTL()
	if err != nil {
		t.Fatalf("loadJWTAccessTTL() error = %v", err)
	}

	if got != 2*time.Hour {
		t.Fatalf("loadJWTAccessTTL() = %v, want 2h", got)
	}
}
