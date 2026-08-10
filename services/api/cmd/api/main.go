package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"flashx/services/api/internal/admin"
	"flashx/services/api/internal/auth"
	"flashx/services/api/internal/dispatch"
	"flashx/services/api/internal/drivers"
	"flashx/services/api/internal/payments"
	"flashx/services/api/internal/platform/config"
	"flashx/services/api/internal/platform/httpserver"
	"flashx/services/api/internal/platform/idempotency"
	"flashx/services/api/internal/platform/persistence"
	"flashx/services/api/internal/pricing"
	"flashx/services/api/internal/ratings"
	"flashx/services/api/internal/realtime"
	"flashx/services/api/internal/ride"
	"flashx/services/api/internal/routing"
	"flashx/services/api/internal/trips"
	"flashx/services/api/internal/users"
)

func main() {
	cfg := config.Load()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	var (
		tripStore        trips.Store
		driverStore      drivers.Store
		userStore        users.Store
		locationIndex    drivers.LocationIndex
		idempotencyStore idempotency.Store
		authStore        auth.Store
		adminStore       admin.Store
		paymentStore     payments.Store
		ratingStore      ratings.Store
		dispatchOffers   dispatch.OfferStore
		dispatchLocker   dispatch.Locker
		resources        *persistence.Resources
		readyCheck       func(context.Context) error
	)

	switch cfg.Persistence {
	case "memory":
		tripStore = trips.NewMemoryStore()
		driverStore = drivers.NewMemoryStore()
		userStore = users.NewMemoryStore()
		locationIndex = drivers.NewMemoryLocationIndex()
		idempotencyStore = idempotency.NewMemoryStore()
		authStore = auth.NewMemoryStore()
		adminStore = admin.NewMemoryStore()
		paymentStore = payments.NewMemoryStore()
		ratingStore = ratings.NewMemoryStore()
		dispatchOffers = dispatch.NewMemoryOfferStore()
		dispatchLocker = dispatch.NewMemoryLocker()
	case "postgres", "persistent":
		startupCtx, cancel := context.WithTimeout(ctx, 8*time.Second)
		var err error
		resources, err = persistence.Open(startupCtx, cfg)
		cancel()
		if err != nil {
			log.Fatalf("open persistence: %v", err)
		}
		defer resources.Close()
		tripStore = trips.NewPostgresStore(resources.Postgres)
		driverStore = drivers.NewPostgresStore(resources.Postgres)
		userStore = users.NewPostgresStore(resources.Postgres)
		locationIndex = drivers.NewRedisLocationIndex(resources.Redis, "flashx")
		idempotencyStore = idempotency.NewRedisStore(resources.Redis)
		authStore = auth.NewPostgresStore(resources.Postgres)
		adminStore = admin.NewPostgresStore(resources.Postgres)
		paymentStore = payments.NewPostgresStore(resources.Postgres)
		ratingStore = ratings.NewPostgresStore(resources.Postgres)
		dispatchOffers = dispatch.NewRedisOfferStore(resources.Redis, "flashx")
		dispatchLocker = dispatch.NewRedisLocker(resources.Redis, "flashx")
		readyCheck = resources.Ready
	default:
		log.Fatalf("unsupported PERSISTENCE mode %q", cfg.Persistence)
	}

	var otpSender auth.OTPSender
	switch cfg.SMSProvider {
	case "development", "":
		if cfg.AppEnv == "production" {
			log.Fatal("production requires a real SMS provider; SMS_PROVIDER=development is forbidden")
		}
		otpSender = auth.DevelopmentOTPSender{}
	case "webhook":
		sender, err := auth.NewWebhookOTPSender(cfg.SMSWebhookURL, cfg.SMSAPIKey)
		if err != nil {
			log.Fatalf("configure sms webhook: %v", err)
		}
		otpSender = sender
	default:
		log.Fatalf("unsupported SMS_PROVIDER %q", cfg.SMSProvider)
	}
	authService, err := auth.NewService(authStore, otpSender, cfg.JWTSecret, cfg.AppEnv, cfg.SMSProvider)
	if err != nil {
		log.Fatalf("configure auth: %v", err)
	}

	tripService := trips.NewService(tripStore)
	driverService := drivers.NewServiceWithLocationIndex(driverStore, locationIndex)
	userService := users.NewService(userStore)
	adminService := admin.NewService(adminStore)
	if cfg.AdminPhone != "" {
		if _, err := adminService.Bootstrap(cfg.AdminPhone, cfg.AdminEmail, cfg.AdminDisplayName); err != nil {
			log.Fatalf("bootstrap admin: %v", err)
		}
	}
	var routeProvider routing.Provider
	switch cfg.MapsProvider {
	case "mock", "development", "":
		if cfg.AppEnv == "production" {
			log.Fatal("production requires a real routing provider; MAPS_PROVIDER=mock is forbidden")
		}
		routeProvider = routing.NewFallbackProvider()
	case "google":
		provider, err := routing.NewGoogleProvider(cfg.GoogleMapsAPIKey)
		if err != nil {
			log.Fatalf("configure google routes: %v", err)
		}
		routeProvider = provider
	default:
		log.Fatalf("unsupported MAPS_PROVIDER %q", cfg.MapsProvider)
	}
	pricingService := pricing.NewServiceWithRouting(routeProvider)
	paymentService := payments.NewService(paymentStore)
	ratingService := ratings.NewService(ratingStore, tripService)
	dispatchEngine := dispatch.NewEngineWithStore(driverService, tripService, dispatchOffers, dispatchLocker)
	rideService := ride.NewService(tripService, driverService)
	realtimeHub := realtime.NewHub()

	server := httpserver.New(cfg.HTTPAddr, httpserver.Dependencies{
		AppEnv:           cfg.AppEnv,
		Persistence:      cfg.Persistence,
		Trips:            tripService,
		Drivers:          driverService,
		Users:            userService,
		Admin:            adminService,
		Dispatch:         dispatchEngine,
		Ride:             rideService,
		Pricing:          pricingService,
		Payments:         paymentService,
		Ratings:          ratingService,
		Idempotency:      idempotencyStore,
		Auth:             authService,
		Realtime:         realtimeHub,
		AllowDevIdentity: cfg.AllowDevIdentity,
		ReadyCheck:       readyCheck,
	})

	errCh := make(chan error, 1)
	go func() {
		log.Printf("flashx api listening on %s (%s, persistence=%s)", cfg.HTTPAddr, cfg.AppEnv, cfg.Persistence)
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
			log.Fatalf("http server: %v", err)
		}
	}
}
