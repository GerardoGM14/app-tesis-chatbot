package config

import (
	"testing"
	"time"
)

func env(values map[string]string) func(string) string {
	return func(k string) string { return values[k] }
}

func TestLoadDefaults(t *testing.T) {
	cfg, err := Load(env(nil))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Env != "production" || cfg.HTTPAddr != ":8000" || cfg.AccessTokenTTL != 15*time.Minute ||
		cfg.RefreshTokenTTL != 168*time.Hour || !cfg.CookieSecure || cfg.TrustProxy {
		t.Errorf("valores por defecto inesperados: %+v", cfg)
	}
}

func TestLoadRejectsInvalidValues(t *testing.T) {
	_, err := Load(env(map[string]string{
		"ACCESS_TOKEN_TTL": "x", "REFRESH_TOKEN_TTL": "y", "COOKIE_SECURE": "talvez", "TRUST_PROXY": "z",
	}))
	if err == nil {
		t.Fatal("se esperaba error")
	}
}

func TestValidateServe(t *testing.T) {
	valid := map[string]string{
		"DATABASE_URL": "postgres://x", "JWT_SECRET": "0123456789abcdef0123456789abcdef",
	}
	cfg, _ := Load(env(valid))
	if err := cfg.ValidateServe(); err != nil {
		t.Errorf("configuración válida rechazada: %v", err)
	}

	cfg, _ = Load(env(map[string]string{"JWT_SECRET": "corto", "COOKIE_SECURE": "false", "ACCESS_TOKEN_TTL": "2h"}))
	if err := cfg.ValidateServe(); err == nil {
		t.Error("se esperaban errores de validación")
	}

	valid["APP_ENV"] = "development"
	valid["COOKIE_SECURE"] = "false"
	cfg, _ = Load(env(valid))
	if err := cfg.ValidateServe(); err != nil || !cfg.IsDevelopment() {
		t.Errorf("en development se permite cookie no segura: %v", err)
	}
}

func TestValidateMigrate(t *testing.T) {
	cfg, _ := Load(env(map[string]string{
		"MIGRATION_DATABASE_URL": "postgres://x", "APP_DB_PASSWORD": "una-clave-de-16-caracteres",
		"SUPERADMIN_EMAIL": "ADMIN@Sertech.pe", "SUPERADMIN_PASSWORD": "Clave-Inicial-2026",
	}))
	if err := cfg.ValidateMigrate(); err != nil {
		t.Errorf("configuración válida rechazada: %v", err)
	}
	if cfg.SuperadminEmail != "admin@sertech.pe" {
		t.Error("el email del superadmin debe normalizarse a minúsculas")
	}

	cfg, _ = Load(env(map[string]string{"APP_DB_PASSWORD": "corta", "SUPERADMIN_EMAIL": "a@b.pe"}))
	if err := cfg.ValidateMigrate(); err == nil {
		t.Error("se esperaban errores de validación")
	}
}
