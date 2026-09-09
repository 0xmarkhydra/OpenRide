package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	modulecatalog "github.com/0xmarkhydra/OpenRide/packages/modules-go/catalog"
	"github.com/0xmarkhydra/OpenRide/services/marketplace/internal/httpapi"
)

func main() {
	addr := strings.TrimSpace(os.Getenv("HTTP_ADDR"))
	if addr == "" {
		addr = ":8090"
	}

	readyCheck, err := dependencyReadiness(os.Getenv("DATABASE_URL"), os.Getenv("NATS_URL"))
	if err != nil {
		log.Fatalf("configure marketplace dependencies: %v", err)
	}

	server, err := httpapi.New(httpapi.Config{
		Addr:     addr,
		Services: modulecatalog.DefaultRegistry(),
		Ready:    readyCheck,
	})
	if err != nil {
		log.Fatalf("configure marketplace service: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, 1)
	go func() {
		log.Printf("marketplace-service listening on %s", addr)
		errCh <- server.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Printf("graceful shutdown failed: %v", err)
		}
	case err := <-errCh:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("marketplace service: %v", err)
		}
	}
}

// dependencyReadiness intentionally performs only connectivity checks here.
// Service-specific adapters will replace these probes with real DB/NATS health
// checks once persistence and event relays are wired into Marketplace Service.
func dependencyReadiness(databaseURL, natsURL string) (func(context.Context) error, error) {
	targets := make([]string, 0, 2)
	for _, raw := range []string{databaseURL, natsURL} {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			continue
		}
		target, err := networkTarget(raw)
		if err != nil {
			return nil, err
		}
		targets = append(targets, target)
	}

	if len(targets) == 0 {
		return nil, nil
	}

	return func(ctx context.Context) error {
		dialer := net.Dialer{Timeout: time.Second}
		for _, target := range targets {
			conn, err := dialer.DialContext(ctx, "tcp", target)
			if err != nil {
				return fmt.Errorf("dependency %s unavailable: %w", target, err)
			}
			_ = conn.Close()
		}
		return nil
	}, nil
}

func networkTarget(raw string) (string, error) {
	u, err := url.Parse(raw)
	if err != nil {
		return "", fmt.Errorf("parse dependency URL: %w", err)
	}
	if u.Hostname() == "" {
		return "", fmt.Errorf("dependency URL has no host: %q", raw)
	}
	port := u.Port()
	if port == "" {
		switch u.Scheme {
		case "postgres", "postgresql":
			port = "5432"
		case "nats":
			port = "4222"
		default:
			return "", fmt.Errorf("dependency URL has no port and unsupported scheme %q", u.Scheme)
		}
	}
	return net.JoinHostPort(u.Hostname(), port), nil
}
