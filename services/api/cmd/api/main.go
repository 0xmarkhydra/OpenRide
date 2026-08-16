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
	"flashx/services/api/internal/custodyevidence"
	"flashx/services/api/internal/customervehicles"
	"flashx/services/api/internal/dispatch"
	"flashx/services/api/internal/driverdocs"
	"flashx/services/api/internal/drivers"
	"flashx/services/api/internal/inspectionchecklist"
	"flashx/services/api/internal/notifications"
	"flashx/services/api/internal/objectstorage"
	"flashx/services/api/internal/operationalsettings"
	"flashx/services/api/internal/payments"
	"flashx/services/api/internal/places"
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
		tripStore                trips.Store
		driverStore              drivers.Store
		vehicleStore             customervehicles.Store
		userStore                users.Store
		locationIndex            drivers.LocationIndex
		idempotencyStore         idempotency.Store
		authStore                auth.Store
		adminStore               admin.Store
		paymentStore             payments.Store
		ratingStore              ratings.Store
		documentStore            driverdocs.Store
		custodyEvidenceStore     custodyevidence.Store
		inspectionChecklistStore inspectionchecklist.Store
		notificationStore        notifications.Store
		pricingStore             pricing.Store
		operationalSettingsStore operationalsettings.Store
		dispatchOffers           dispatch.OfferStore
		dispatchLocker           dispatch.Locker
		resources                *persistence.Resources
		readyCheck               func(context.Context) error
	)

	switch cfg.Persistence {
	case "memory":
		tripStore = trips.NewMemoryStore()
		driverStore = drivers.NewMemoryStore()
		vehicleStore = customervehicles.NewMemoryStore()
		userStore = users.NewMemoryStore()
		locationIndex = drivers.NewMemoryLocationIndex()
		idempotencyStore = idempotency.NewMemoryStore()
		authStore = auth.NewMemoryStore()
		adminStore = admin.NewMemoryStore()
		paymentStore = payments.NewMemoryStore()
		ratingStore = ratings.NewMemoryStore()
		documentStore = driverdocs.NewMemoryStore()
		custodyEvidenceStore = custodyevidence.NewMemoryStore()
		inspectionChecklistStore = inspectionchecklist.NewMemoryStore()
		notificationStore = notifications.NewMemoryStore()
		pricingStore = pricing.NewMemoryStore()
		operationalSettingsStore = operationalsettings.NewMemoryStore()
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
		vehicleStore = customervehicles.NewPostgresStore(resources.Postgres)
		userStore = users.NewPostgresStore(resources.Postgres)
		locationIndex = drivers.NewRedisLocationIndex(resources.Redis, "flashx")
		idempotencyStore = idempotency.NewRedisStore(resources.Redis)
		authStore = auth.NewPostgresStore(resources.Postgres)
		adminStore = admin.NewPostgresStore(resources.Postgres)
		paymentStore = payments.NewPostgresStore(resources.Postgres)
		ratingStore = ratings.NewPostgresStore(resources.Postgres)
		documentStore = driverdocs.NewPostgresStore(resources.Postgres)
		custodyEvidenceStore = custodyevidence.NewPostgresStore(resources.Postgres)
		inspectionChecklistStore = inspectionchecklist.NewPostgresStore(resources.Postgres)
		notificationStore = notifications.NewPostgresStore(resources.Postgres)
		pricingStore = pricing.NewPostgresStore(resources.Postgres)
		operationalSettingsStore = operationalsettings.NewPostgresStore(resources.Postgres)
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

	var storageSigner objectstorage.Signer
	switch cfg.ObjectStorageProvider {
	case "disabled", "":
		if cfg.AppEnv == "production" {
			log.Fatal("production requires object storage; OBJECT_STORAGE_PROVIDER=disabled is forbidden")
		}
	case "s3":
		storageCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
		signer, err := objectstorage.NewS3Signer(storageCtx, objectstorage.S3Config{
			Endpoint:        cfg.S3Endpoint,
			Region:          cfg.S3Region,
			Bucket:          cfg.S3Bucket,
			AccessKeyID:     cfg.S3AccessKeyID,
			SecretAccessKey: cfg.S3SecretAccessKey,
			ForcePathStyle:  cfg.S3ForcePathStyle,
			TTL:             time.Duration(cfg.S3PresignTTLSeconds) * time.Second,
		})
		cancel()
		if err != nil {
			log.Fatalf("configure object storage signer: %v", err)
		}
		storageSigner = signer
	default:
		log.Fatalf("unsupported OBJECT_STORAGE_PROVIDER %q", cfg.ObjectStorageProvider)
	}

	var pushProvider notifications.Provider
	switch cfg.PushProvider {
	case "disabled", "":
		pushProvider = notifications.DisabledProvider{ProviderName: "disabled"}
	case "development":
		if cfg.AppEnv == "production" {
			log.Fatal("PUSH_PROVIDER=development is forbidden in production")
		}
		pushProvider = notifications.DisabledProvider{ProviderName: "development"}
	default:
		log.Fatalf("unsupported PUSH_PROVIDER %q; real FCM/APNs provider is not configured yet", cfg.PushProvider)
	}

	tripService := trips.NewService(tripStore)
	inspectionChecklistService, err := inspectionchecklist.NewService(inspectionChecklistStore, tripService)
	if err != nil {
		log.Fatalf("configure inspection checklist: %v", err)
	}
	driverService := drivers.NewServiceWithLocationIndex(driverStore, locationIndex)
	vehicleService := customervehicles.NewService(vehicleStore)
	userService := users.NewService(userStore)
	adminService := admin.NewService(adminStore)
	if cfg.AdminPhone != "" {
		if _, err := adminService.Bootstrap(cfg.AdminPhone, cfg.AdminEmail, cfg.AdminDisplayName); err != nil {
			log.Fatalf("bootstrap admin: %v", err)
		}
	}
	var routeProvider routing.Provider
	var placeProvider places.Provider
	switch cfg.MapsProvider {
	case "mock", "development", "":
		if cfg.AppEnv == "production" {
			log.Fatal("production requires a real routing provider; MAPS_PROVIDER=mock is forbidden")
		}
		routeProvider = routing.NewFallbackProvider()
		placeProvider = places.DisabledProvider{}
	case "google":
		routeGoogle, err := routing.NewGoogleProvider(cfg.GoogleMapsAPIKey)
		if err != nil {
			log.Fatalf("configure google routes: %v", err)
		}
		placesGoogle, err := places.NewGoogleProvider(cfg.GoogleMapsAPIKey)
		if err != nil {
			log.Fatalf("configure google places: %v", err)
		}
		routeProvider = routeGoogle
		placeProvider = placesGoogle
	default:
		log.Fatalf("unsupported MAPS_PROVIDER %q", cfg.MapsProvider)
	}
	pricingService, err := pricing.NewServiceWithStore(routeProvider, pricingStore)
	if err != nil {
		log.Fatalf("configure pricing: %v", err)
	}
	paymentService := payments.NewService(paymentStore)
	notificationService := notifications.NewService(notificationStore, pushProvider)
	ratingService := ratings.NewService(ratingStore, tripService)
	driverDocumentService := driverdocs.NewService(documentStore, storageSigner)
	custodyEvidenceService := custodyevidence.NewService(custodyEvidenceStore, storageSigner, tripService)
	dispatchEngine := dispatch.NewEngineWithStore(driverService, tripService, dispatchOffers, dispatchLocker)
	dispatchEngine.SetCustomerVehicles(vehicleService)
	operationalSettingsService, err := operationalsettings.NewService(operationalSettingsStore)
	if err != nil {
		log.Fatalf("configure operational settings: %v", err)
	}
	operationalSettingsService.SetApplyHook(func(settings operationalsettings.Config) {
		dispatchEngine.UpdatePolicy(settings.DispatchMaxDistanceM, settings.DriverLocationMaxAgeSeconds)
	})
	rideService := ride.NewService(tripService, driverService)
	realtimeHub := realtime.NewHub()

	server := httpserver.New(cfg.HTTPAddr, httpserver.Dependencies{
		AppEnv:              cfg.AppEnv,
		Persistence:         cfg.Persistence,
		Trips:               tripService,
		Drivers:             driverService,
		CustomerVehicles:    vehicleService,
		DriverDocuments:     driverDocumentService,
		CustodyEvidence:     custodyEvidenceService,
		InspectionChecklist: inspectionChecklistService,
		Users:               userService,
		Admin:               adminService,
		Dispatch:            dispatchEngine,
		Ride:                rideService,
		Pricing:             pricingService,
		OperationalSettings: operationalSettingsService,
		Payments:            paymentService,
		Notifications:       notificationService,
		Places:              placeProvider,
		Ratings:             ratingService,
		Idempotency:         idempotencyStore,
		Auth:                authService,
		Realtime:            realtimeHub,
		AllowDevIdentity:    cfg.AllowDevIdentity,
		ReadyCheck:          readyCheck,
	})

	errCh := make(chan error, 1)
	go server.RunScheduledDispatch(ctx, 5*time.Second)
	go server.RunNotificationDispatch(ctx, 5*time.Second)
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
