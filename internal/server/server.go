// Package server bootstraps the goSign HTTP API: configuration, storage,
// queries, services, routes, and graceful shutdown. Scheduled maintenance
// (trust lists, GeoLite2 downloads) lives in internal/worker.
package server

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/static"

	"github.com/shurco/gosign/internal/assets"
	"github.com/shurco/gosign/internal/config"
	"github.com/shurco/gosign/internal/handlers/api"
	public "github.com/shurco/gosign/internal/handlers/public"
	"github.com/shurco/gosign/internal/middleware"
	"github.com/shurco/gosign/internal/queries"
	"github.com/shurco/gosign/internal/routes"
	"github.com/shurco/gosign/internal/rpc/workerclient"
	"github.com/shurco/gosign/internal/services"
	"github.com/shurco/gosign/internal/services/submission"
	"github.com/shurco/gosign/pkg/appdir"
	"github.com/shurco/gosign/pkg/logging"
	"github.com/shurco/gosign/pkg/storage/postgres"
	"github.com/shurco/gosign/pkg/storage/redis"
	"github.com/shurco/gosign/pkg/webhook"
)

// Run starts the HTTP server and blocks until an interrupt signal.
func Run() error {
	fmt.Print("✍️ Sign documents without stress\n")

	appdir.Init()

	log := logging.Log

	if err := config.Load(); err != nil {
		log.Err(err).Send()
		return err
	}
	cfg := config.Data()

	// directories create
	if err := createDirs(); err != nil {
		log.Err(err).Send()
		return err
	}

	// Ensure embedded assets are available on disk (fonts/images for certificate rendering).
	assetsDir := filepath.Join(appdir.DataDir(), "assets")
	assetPaths, err := assets.EnsureOnDisk(assetsDir)
	if err != nil {
		log.Err(err).Msg("Failed to extract embedded assets")
		return err
	}

	// redis init
	redis.New(context.Background(), cfg.Redis.Address, cfg.Redis.Password)
	if err := redis.Conn.Ping(); err != nil {
		log.Err(err).Send()
		return err
	}
	defer redis.Conn.Close()

	// postgresql init
	pool, err := postgres.New(context.Background(), cfg.Postgres)
	if err != nil {
		log.Err(err).Send()
		return err
	}
	defer pool.Close()

	// db init; migrations are applied externally (scripts/migration or migrate container)
	queries.New(pool)
	if err := queries.CheckSchema(context.Background(), pool); err != nil {
		log.Err(err).Send()
		return err
	}

	// Initialize query services
	templateQueries := &queries.TemplateQueries{Pool: pool}
	organizationQueries := queries.NewOrganizationQueries(pool)
	userQueries := queries.NewUserQueries(pool)
	accountQueries := queries.NewAccountQueries(pool)
	settingQueries := queries.NewSettingQueries(pool)

	// Create template repository (for now using a simple implementation)
	templateRepo := &simpleTemplateRepository{
		templateQueries: templateQueries,
	}

	// Initialize submission repository and service
	submissionRepo := queries.NewSubmissionRepository(pool)
	submissionRepoImpl := &simpleSubmissionRepository{
		submissionRepo: submissionRepo,
	}

	notificationService := initNotificationService(settingQueries)

	// Webhook storage + async delivery (HMAC-signed, retries, auto-disable).
	webhookQueries := queries.NewWebhookQueries(pool)
	webhookNotifier := services.NewWebhookNotifier(webhookQueries, webhook.NewDispatcher(3, 30*time.Second))

	submissionService := submission.NewService(submissionRepo, notificationService, webhookNotifier)

	// Internal authenticated gRPC client to worker replicas (lazy connect;
	// the API works without workers, degrading geolocation and health data).
	workerConn, workerClient, err := workerclient.New(cfg.WorkerGRPCAddr, cfg.WorkerGRPCToken)
	if err != nil {
		log.Warn().Err(err).Str("addr", cfg.WorkerGRPCAddr).Msg("Worker gRPC client init failed, worker-backed features are disabled")
	} else {
		defer workerConn.Close()
	}

	// web init
	app := fiber.New(fiber.Config{
		BodyLimit: 50 * 1024 * 1024,
	})

	middleware.Fiber(app, log, cfg)
	routes.SiteRoutes(app, public.NewHealthHandler(workerClient))
	app.Use("/drive/pages", static.New(appdir.LcPages()))
	app.Use("/drive/signed", static.New(appdir.LcSigned()))
	app.Use("/drive/uploads", static.New(appdir.LcUploads()))

	// Completed document builder (filesystem-backed cache).
	completedDoc := &services.CompletedDocumentBuilder{
		Pool:            pool,
		TemplateQueries: templateQueries,
		PagesDir:        appdir.LcPages(),
		SignedDir:       appdir.LcSigned(),
		AssetsDir:       assetPaths.Dir,
	}

	// Geolocation lookups go through the worker over gRPC (best-effort):
	// the worker owns the GeoLite2 database, the server never touches it.
	geoLocator := workerclient.NewGeoLocator(workerClient)

	// Initialize API key repository and service
	apiKeyRepo := queries.NewAPIKeyRepository(pool)
	apiKeyService := services.NewAPIKeyService(apiKeyRepo)
	// Required for X-API-Key authentication in middleware.Protected().
	middleware.SetAPIKeyValidator(apiKeyService)

	// Initialize email template queries
	emailTemplateQueries := &queries.EmailTemplateQueries{Pool: pool}

	// Initialize API handlers
	apiHandlers := &routes.APIHandlers{
		Submissions:    api.NewSubmissionHandler(submissionRepoImpl, submissionService),
		Submitters:     api.NewSubmitterHandler(queries.NewSubmitterResourceRepository(pool), submissionService),
		SigningLinks:   api.NewSigningLinkHandler(pool, templateQueries, completedDoc),
		Templates:      api.NewTemplateHandler(templateRepo, templateQueries),
		Webhooks:       api.NewWebhookHandler(webhookQueries),
		Settings:       api.NewSettingsHandler(notificationService, accountQueries, userQueries, workerClient, settingQueries),
		APIKeys:        api.NewAPIKeyHandler(apiKeyService),
		Stats:          api.NewStatsHandler(pool),
		Events:         api.NewEventHandler(pool),
		Organizations:  api.NewOrganizationHandler(organizationQueries, userQueries),
		Members:        api.NewMemberHandler(organizationQueries, userQueries),
		Invitations:    api.NewInvitationHandler(organizationQueries),
		Users:          api.NewUserHandler(userQueries),
		I18n:           api.NewI18nHandler(userQueries, accountQueries),
		Branding:       api.NewBrandingHandler(accountQueries),
		EmailTemplates: api.NewEmailTemplateHandler(emailTemplateQueries, userQueries),
		PublicSigning:  public.NewPublicSigningHandler(pool, templateQueries, userQueries, notificationService, completedDoc, geoLocator, webhookNotifier),
		Embed:          public.NewEmbedHandler(submissionRepo),
	}

	routes.APIRoutes(app, apiHandlers)
	routes.NotFoundRoute(app)

	fmt.Printf("├─[🚀] API: http://%s/v1/\n", cfg.HTTPAddr)
	fmt.Printf("└─[🚀] Health: http://%s/health\n", cfg.HTTPAddr)

	// Listen on port; a bind failure must terminate the process (fail fast)
	listenErr := make(chan error, 1)
	go func() {
		listenErr <- app.Listen(cfg.HTTPAddr, fiber.ListenConfig{DisableStartupMessage: true})
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)

	select {
	case err := <-listenErr:
		if err != nil {
			log.Err(err).Send()
			return err
		}
		return nil
	case <-quit:
		fmt.Print("\n✍️ Shutting down server...\n")
		return app.Shutdown()
	}
}

func createDirs() error {
	dirs := []string{
		appdir.LcPages(),
		appdir.LcSigned(),
		appdir.LcUploads(),
		appdir.LcTmp(),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	return nil
}
