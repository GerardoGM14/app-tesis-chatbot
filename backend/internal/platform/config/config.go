// Package config carga la configuración desde variables de entorno.
package config

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Config es la configuración del backend.
type Config struct {
	Env                  string
	HTTPAddr             string
	DatabaseURL          string
	MigrationDatabaseURL string
	AppDBPassword        string
	JWTSecret            []byte
	AccessTokenTTL       time.Duration
	RefreshTokenTTL      time.Duration
	CookieSecure         bool
	TrustProxy           bool
	SuperadminEmail      string
	SuperadminPassword   string
	SuperadminName       string
}

// IsDevelopment indica si se ejecuta en el entorno local.
func (c Config) IsDevelopment() bool { return c.Env == "development" }

// Load lee las variables de entorno con getenv (os.Getenv en producción).
func Load(getenv func(string) string) (Config, error) {
	get := func(key, fallback string) string {
		if v := strings.TrimSpace(getenv(key)); v != "" {
			return v
		}
		return fallback
	}

	cfg := Config{
		Env:                  get("APP_ENV", "production"),
		HTTPAddr:             get("HTTP_ADDR", ":8000"),
		DatabaseURL:          get("DATABASE_URL", ""),
		MigrationDatabaseURL: get("MIGRATION_DATABASE_URL", ""),
		AppDBPassword:        get("APP_DB_PASSWORD", ""),
		JWTSecret:            []byte(get("JWT_SECRET", "")),
		SuperadminEmail:      strings.ToLower(get("SUPERADMIN_EMAIL", "")),
		SuperadminPassword:   get("SUPERADMIN_PASSWORD", ""),
		SuperadminName:       get("SUPERADMIN_NAME", "Superadministrador"),
	}

	var errs []error
	var err error
	if cfg.AccessTokenTTL, err = time.ParseDuration(get("ACCESS_TOKEN_TTL", "15m")); err != nil {
		errs = append(errs, fmt.Errorf("ACCESS_TOKEN_TTL: %w", err))
	}
	if cfg.RefreshTokenTTL, err = time.ParseDuration(get("REFRESH_TOKEN_TTL", "168h")); err != nil {
		errs = append(errs, fmt.Errorf("REFRESH_TOKEN_TTL: %w", err))
	}
	if cfg.CookieSecure, err = strconv.ParseBool(get("COOKIE_SECURE", "true")); err != nil {
		errs = append(errs, fmt.Errorf("COOKIE_SECURE: %w", err))
	}
	if cfg.TrustProxy, err = strconv.ParseBool(get("TRUST_PROXY", "false")); err != nil {
		errs = append(errs, fmt.Errorf("TRUST_PROXY: %w", err))
	}
	return cfg, errors.Join(errs...)
}

// ValidateServe verifica lo necesario para levantar el servidor HTTP.
func (c Config) ValidateServe() error {
	var errs []error
	if c.DatabaseURL == "" {
		errs = append(errs, errors.New("DATABASE_URL es obligatoria"))
	}
	if len(c.JWTSecret) < 32 {
		errs = append(errs, errors.New("JWT_SECRET debe tener al menos 32 caracteres"))
	}
	if c.AccessTokenTTL <= 0 || c.AccessTokenTTL > time.Hour {
		errs = append(errs, errors.New("ACCESS_TOKEN_TTL debe estar entre 1s y 1h"))
	}
	if c.RefreshTokenTTL < c.AccessTokenTTL {
		errs = append(errs, errors.New("REFRESH_TOKEN_TTL debe ser mayor que ACCESS_TOKEN_TTL"))
	}
	if !c.IsDevelopment() && !c.CookieSecure {
		errs = append(errs, errors.New("COOKIE_SECURE solo puede ser false en development"))
	}
	return errors.Join(errs...)
}

// ValidateMigrate verifica lo necesario para migrar y preparar la base.
func (c Config) ValidateMigrate() error {
	var errs []error
	if c.MigrationDatabaseURL == "" {
		errs = append(errs, errors.New("MIGRATION_DATABASE_URL es obligatoria"))
	}
	if len(c.AppDBPassword) < 16 {
		errs = append(errs, errors.New("APP_DB_PASSWORD debe tener al menos 16 caracteres"))
	}
	if (c.SuperadminEmail == "") != (c.SuperadminPassword == "") {
		errs = append(errs, errors.New("SUPERADMIN_EMAIL y SUPERADMIN_PASSWORD van juntas"))
	}
	return errors.Join(errs...)
}
