// Command api es el backend de la plataforma.
//
// Subcomandos:
//
//	api serve        levanta el servidor HTTP (por defecto)
//	api migrate      aplica migraciones, configura el rol de la app y el superadmin
//	api seed-demo    crea empresas y usuarios de demostración (solo development)
//	api healthcheck  consulta /healthz; lo usa el HEALTHCHECK de Docker
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5"

	"backend/internal/app"
	"backend/internal/bootstrap"
	"backend/internal/platform/config"
	"backend/internal/platform/database"
	"backend/internal/platform/security"
)

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	cfg, err := config.Load(os.Getenv)
	if err != nil {
		logger.Error("configuración inválida", "error", err)
		return 1
	}

	command := "serve"
	if len(args) > 0 {
		command = args[0]
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	switch command {
	case "serve":
		err = serve(ctx, cfg, logger)
	case "migrate":
		err = migrate(ctx, cfg, logger)
	case "seed-demo":
		err = seedDemo(ctx, cfg, logger)
	case "healthcheck":
		err = healthcheck(cfg)
	default:
		err = fmt.Errorf("subcomando desconocido: %s", command)
	}
	if err != nil {
		logger.Error("el comando falló", "comando", command, "error", err)
		return 1
	}
	return 0
}

func serve(ctx context.Context, cfg config.Config, logger *slog.Logger) error {
	if err := cfg.ValidateServe(); err != nil {
		return err
	}
	db, err := database.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer db.Close()

	handler, err := app.New(cfg, db, logger, app.Options{})
	if err != nil {
		return err
	}
	srv := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    1 << 16,
	}

	errCh := make(chan error, 1)
	go func() {
		logger.Info("servidor escuchando", "addr", cfg.HTTPAddr, "env", cfg.Env)
		errCh <- srv.ListenAndServe()
	}()

	select {
	case err := <-errCh:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	case <-ctx.Done():
		logger.Info("apagando el servidor")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		return srv.Shutdown(shutdownCtx)
	}
}

func migrate(ctx context.Context, cfg config.Config, logger *slog.Logger) error {
	if err := cfg.ValidateMigrate(); err != nil {
		return err
	}
	conn, err := pgx.Connect(ctx, cfg.MigrationDatabaseURL)
	if err != nil {
		return err
	}
	defer func() { _ = conn.Close(context.WithoutCancel(ctx)) }()

	applied, err := bootstrap.Run(ctx, conn, cfg.AppDBPassword, bootstrap.Superadmin{
		Nombre: cfg.SuperadminName, Email: cfg.SuperadminEmail, Password: cfg.SuperadminPassword,
	}, security.NewPasswordHasher(security.DefaultArgon2))
	if err != nil {
		return err
	}
	logger.Info("base de datos lista", "migraciones_aplicadas", applied)
	return nil
}

func seedDemo(ctx context.Context, cfg config.Config, logger *slog.Logger) error {
	if !cfg.IsDevelopment() {
		return errors.New("seed-demo solo se permite con APP_ENV=development")
	}
	if cfg.MigrationDatabaseURL == "" {
		return errors.New("MIGRATION_DATABASE_URL es obligatoria")
	}
	conn, err := pgx.Connect(ctx, cfg.MigrationDatabaseURL)
	if err != nil {
		return err
	}
	defer func() { _ = conn.Close(context.WithoutCancel(ctx)) }()

	if err := bootstrap.SeedDemo(ctx, conn, security.NewPasswordHasher(security.DefaultArgon2)); err != nil {
		return err
	}
	logger.Info("datos de demostración listos", "password", bootstrap.DemoPassword)
	return nil
}

func healthcheck(cfg config.Config) error {
	_, port, err := net.SplitHostPort(cfg.HTTPAddr)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://127.0.0.1:"+port+"/healthz", nil)
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("healthz respondió %d", resp.StatusCode)
	}
	return nil
}
