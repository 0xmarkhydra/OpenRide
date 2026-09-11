package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/nats-io/nats.go"

	modulecatalog "github.com/0xmarkhydra/OpenRide/packages/modules-go/catalog"
	"github.com/0xmarkhydra/OpenRide/services/marketplace/internal/app"
	"github.com/0xmarkhydra/OpenRide/services/marketplace/internal/httpapi"
	"github.com/0xmarkhydra/OpenRide/services/marketplace/internal/outbox"
	"github.com/0xmarkhydra/OpenRide/services/marketplace/internal/store"
)

func main() {
	addr := strings.TrimSpace(os.Getenv("HTTP_ADDR"))
	if addr == "" {
		addr = ":8090"
	}
	databaseURL := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	natsURL := strings.TrimSpace(os.Getenv("NATS_URL"))
	gatewayToken := strings.TrimSpace(os.Getenv("MARKETPLACE_GATEWAY_TOKEN"))
	if databaseURL == "" || natsURL == "" {
		log.Fatal("DATABASE_URL and NATS_URL are required for durable marketplace mode")
	}
	if len(gatewayToken) < 24 {
		log.Fatal("MARKETPLACE_GATEWAY_TOKEN with at least 24 characters is required")
	}

	bootstrapCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	pg, err := store.New(bootstrapCtx, databaseURL)
	if err != nil {
		log.Fatalf("connect marketplace database: %v", err)
	}
	defer pg.Close()
	nc, err := nats.Connect(natsURL, nats.Name("openride-marketplace"), nats.Timeout(5*time.Second), nats.MaxReconnects(-1))
	if err != nil {
		log.Fatalf("connect nats: %v", err)
	}
	defer nc.Close()
	js, err := nc.JetStream()
	if err != nil {
		log.Fatalf("open jetstream: %v", err)
	}
	if err := ensureMarketplaceStream(js); err != nil {
		log.Fatalf("ensure marketplace stream: %v", err)
	}

	ready := func(ctx context.Context) error {
		if err := pg.Ping(ctx); err != nil {
			return err
		}
		if nc.Status() != nats.CONNECTED {
			return errors.New("nats not connected")
		}
		return nil
	}
	acceptance := app.AcceptanceService{Store: pg}
	server, err := httpapi.New(httpapi.Config{
		Addr: addr,
		Services: modulecatalog.DefaultRegistry(),
		Ready: ready,
		V2Store: pg,
		Acceptance: acceptance,
		GatewayToken: gatewayToken,
	})
	if err != nil {
		log.Fatalf("configure marketplace service: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	go outbox.Relay{Store: pg, JS: js, Interval: 500 * time.Millisecond, Batch: 50}.Run(ctx)

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

func ensureMarketplaceStream(js nats.JetStreamContext) error {
	config := &nats.StreamConfig{
		Name:       "MARKETPLACE",
		Subjects:   []string{"openride.marketplace.>"},
		Storage:    nats.FileStorage,
		Retention:  nats.LimitsPolicy,
		Duplicates: 10 * time.Minute,
	}
	info, err := js.StreamInfo(config.Name)
	if err != nil {
		_, addErr := js.AddStream(config)
		return addErr
	}
	if len(info.Config.Subjects) == 1 && info.Config.Subjects[0] == config.Subjects[0] && info.Config.Duplicates == config.Duplicates {
		return nil
	}
	info.Config.Subjects = config.Subjects
	info.Config.Duplicates = config.Duplicates
	_, err = js.UpdateStream(&info.Config)
	return err
}
