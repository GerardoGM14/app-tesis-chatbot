// Command wa-service mantiene las sesiones de WhatsApp de cada empresa.
//
// Subcomandos:
//
//	wa-service serve        levanta el API interno (por defecto)
//	wa-service healthcheck  consulta /healthz; lo usa el HEALTHCHECK de Docker
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

	"wa-service/internal/server"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	addr := getenv("HTTP_ADDR", ":8001")

	command := "serve"
	if len(os.Args) > 1 {
		command = os.Args[1]
	}

	var err error
	switch command {
	case "serve":
		err = serve(addr, logger)
	case "healthcheck":
		err = healthcheck(addr)
	default:
		err = fmt.Errorf("subcomando desconocido: %s", command)
	}
	if err != nil {
		logger.Error("el comando falló", "comando", command, "error", err)
		os.Exit(1)
	}
}

func serve(addr string, logger *slog.Logger) error {
	secret := []byte(os.Getenv("INTERNAL_HMAC_SECRET"))
	if len(secret) < 32 {
		return errors.New("INTERNAL_HMAC_SECRET debe tener al menos 32 caracteres")
	}
	srv := &http.Server{
		Addr:              addr,
		Handler:           server.New(secret),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, 1)
	go func() {
		logger.Info("wa-service escuchando", "addr", addr)
		errCh <- srv.ListenAndServe()
	}()
	select {
	case err := <-errCh:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return srv.Shutdown(shutdownCtx)
	}
}

func healthcheck(addr string) error {
	_, port, err := net.SplitHostPort(addr)
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

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
