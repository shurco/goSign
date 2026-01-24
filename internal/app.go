package app

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"

	"github.com/shurco/gosign/internal/assets"
	"github.com/shurco/gosign/internal/config"
	"github.com/shurco/gosign/internal/handlers/api"
	public "github.com/shurco/gosign/internal/handlers/public"
	"github.com/shurco/gosign/internal/models"
	"github.com/shurco/gosign/internal/queries"
	"github.com/shurco/gosign/internal/routes"
	"github.com/shurco/gosign/internal/services"
	"github.com/shurco/gosign/internal/services/submission"
	"github.com/shurco/gosign/internal/trust"
	"github.com/shurco/gosign/pkg/geolocation"
	"github.com/shurco/gosign/pkg/logging"
	"github.com/shurco/gosign/pkg/notification"
	"github.com/shurco/gosign/pkg/storage/postgres"
	"github.com/shurco/gosign/pkg/storage/redis"
)

// simpleTemplateRepository is a simple implementation of ResourceRepository for templates
type simpleTemplateRepository struct {
	templateQueries *queries.TemplateQueries
}

func (r *simpleTemplateRepository) List(page, pageSize int, filters map[string]string) ([]models.Template, int, error) {
	// TODO: Implement proper template listing with pagination and filters
	return []models.Template{}, 0, nil
}

func (r *simpleTemplateRepository) Get(id string) (*models.Template, error) {
	if r.templateQueries == nil {
		return nil, fmt.Errorf("template queries not initialized")
	}
	return r.templateQueries.Template(context.Background(), id)
}

func (r *simpleTemplateRepository) Create(item *models.Template) error {
	// TODO: Implement proper template creation
	return fmt.Errorf("not implemented")
}

// Update updates a template by ID using the template queries service.
func (r *simpleTemplateRepository) Update(id string, item *models.Template) error {
	if r.templateQueries == nil {
		return fmt.Errorf("template queries not initialized")
	}
	return r.templateQueries.UpdateTemplate(context.Background(), id, item)
}

func (r *simpleTemplateRepository) Delete(id string) error {
	// TODO: Implement proper template deletion
	return fmt.Errorf("not implemented")
}

// simpleSubmissionRepository is a simple implementation of ResourceRepository for submissions
type simpleSubmissionRepository struct {
	submissionRepo submission.Repository
}

func (r *simpleSubmissionRepository) List(page, pageSize int, filters map[string]string) ([]models.Submission, int, error) {
	// TODO: Implement proper submission listing with pagination and filters
	// For now, return empty list to avoid errors
	return []models.Submission{}, 0, nil
}

func (r *simpleSubmissionRepository) Get(id string) (*models.Submission, error) {
	if r.submissionRepo == nil {
		return nil, fmt.Errorf("submission repository not initialized")
	}
	return r.submissionRepo.GetSubmission(context.Background(), id)
}

func (r *simpleSubmissionRepository) Create(item *models.Submission) error {
	if r.submissionRepo == nil {
		return fmt.Errorf("submission repository not initialized")
	}
	return r.submissionRepo.CreateSubmission(context.Background(), item)
}

func (r *simpleSubmissionRepository) Update(id string, item *models.Submission) error {
	// TODO: Implement proper submission update
	return fmt.Errorf("not implemented")
}

func (r *simpleSubmissionRepository) Delete(id string) error {
	// TODO: Implement proper submission deletion
	return fmt.Errorf("not implemented")
}

// simpleWebhookRepository is a simple implementation of ResourceRepository for webhooks
type simpleWebhookRepository struct {
	// TODO: Add webhook queries when needed
}

func (r *simpleWebhookRepository) List(page, pageSize int, filters map[string]string) ([]models.Webhook, int, error) {
	// TODO: Implement proper webhook listing with pagination and filters
	// For now, return empty list to avoid errors
	return []models.Webhook{}, 0, nil
}

func (r *simpleWebhookRepository) Get(id string) (*models.Webhook, error) {
	// TODO: Implement proper webhook retrieval
	return nil, fmt.Errorf("not implemented")
}

func (r *simpleWebhookRepository) Create(item *models.Webhook) error {
	// TODO: Implement proper webhook creation
	return fmt.Errorf("not implemented")
}

func (r *simpleWebhookRepository) Update(id string, item *models.Webhook) error {
	// TODO: Implement proper webhook update
	return fmt.Errorf("not implemented")
}

func (r *simpleWebhookRepository) Delete(id string) error {
	// TODO: Implement proper webhook deletion
	return fmt.Errorf("not implemented")
}

func New() error {
	fmt.Print("✍️ Sign documents without stress\n")

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

	// Create geolocation base directory (hardcoded path)
	const geolocationBaseDir = "./base"
	if err := os.MkdirAll(geolocationBaseDir, 0755); err != nil {
		log.Warn().Err(err).Str("path", geolocationBaseDir).Msg("Failed to create geolocation base directory")
	}

	// Ensure embedded assets are available on disk (fonts/images for certificate rendering).
	assetPaths, err := assets.EnsureOnDisk(assets.DefaultOutputDir())
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

	// db init
	if err := queries.Init(pool); err != nil {
		log.Err(err).Send()
		return err
	}

	// Initialize query services
	templateQueries := &queries.TemplateQueries{Pool: pool}
	organizationQueries := queries.NewOrganizationQueries(pool)
	userQueries := queries.NewUserQueries(pool)
	accountQueries := queries.NewAccountQueries(pool)

	// Create template repository (for now using a simple implementation)
	templateRepo := &simpleTemplateRepository{
		templateQueries: templateQueries,
	}

	// Initialize submission repository and service
	submissionRepo := queries.NewSubmissionRepository(pool)
	submissionRepoImpl := &simpleSubmissionRepository{
		submissionRepo: submissionRepo,
	}

	submissionService := submission.NewService(submissionRepo, nil, nil)

	// update trust certs
	if err = trust.Update(cfg.Trust); err != nil {
		log.Err(err).Send()
		return err
	}

	// Download GeoLite2 database if needed (after Adobe certificates update)
	if err = downloadGeoLite2IfNeeded(cfg, pool, log); err != nil {
		log.Warn().Err(err).Msg("Failed to download GeoLite2 database, continuing without it")
	}

	// Schedule trust certs update every 12 hours
	go func() {
		ticker := time.NewTicker(12 * time.Hour)
		defer ticker.Stop()

		// Calculate initial delay to align with 12-hour intervals
		now := time.Now()
		nextRun := now.Truncate(12 * time.Hour).Add(12 * time.Hour)
		if nextRun.Before(now) {
			nextRun = nextRun.Add(12 * time.Hour)
		}
		initialDelay := time.Until(nextRun)

		// Wait for initial delay
		time.Sleep(initialDelay)

		// Execute immediately on first run
		if err := trust.Update(cfg.Trust); err != nil {
			log.Err(err).Send()
		}

		// Periodic execution
		for range ticker.C {
			if err := trust.Update(cfg.Trust); err != nil {
				log.Err(err).Send()
			}
		}
	}()

	// web init
	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
		BodyLimit:             50 * 1024 * 1024,
	})

	// middleware.Fiber(app, log)
	// routes.SiteRoutes(app)
	app.Static("/drive/pages", "./lc_pages")
	app.Static("/drive/signed", "./lc_signed")
	app.Static("/drive/uploads", "./lc_uploads")

	// Initialize webhook repository
	webhookRepo := &simpleWebhookRepository{}

	// Initialize notification service (with nil repository - can work without it)
	notificationService := notification.NewService(nil)

	// Register email provider from config (best-effort).
	if cfg.Settings.Email != nil && cfg.Settings.Email["provider"] == "smtp" {
		port, _ := strconv.Atoi(cfg.Settings.Email["smtp_port"])
		if port == 0 {
			port = 1025
		}
		smtpCfg := notification.SMTPConfig{
			Host:      cfg.Settings.Email["smtp_host"],
			Port:      port,
			User:      cfg.Settings.Email["smtp_user"],
			Password:  cfg.Settings.Email["smtp_pass"],
			FromEmail: cfg.Settings.Email["from_email"],
			FromName:  cfg.Settings.Email["from_name"],
		}
		// Only register when it looks usable.
		if smtpCfg.Host != "" && smtpCfg.FromEmail != "" {
			notificationService.RegisterProvider(notification.NewEmailProvider(smtpCfg))
		}
	}

	// Register SMS provider if enabled via config (Twilio).
	// Config keys are optional; provider will return an error if incomplete.
	if cfg.Settings.Email["twilio_enabled"] == "true" {
		notificationService.RegisterProvider(notification.NewSMSProvider(notification.TwilioConfig{
			AccountSID: cfg.Settings.Email["twilio_account_sid"],
			AuthToken:  cfg.Settings.Email["twilio_auth_token"],
			FromNumber: cfg.Settings.Email["twilio_from_number"],
			Enabled:    true,
		}))
	}

	// Completed document builder (filesystem-backed cache).
	completedDoc := &services.CompletedDocumentBuilder{
		Pool:            pool,
		TemplateQueries: templateQueries,
		PagesDir:        "./lc_pages",
		SignedDir:       "./lc_signed",
		AssetsDir:       assetPaths.Dir,
	}

	// Initialize geolocation service (best-effort; works without database)
	geolocationDBPath := os.Getenv("GEOLITE2_DB_PATH")
	if geolocationDBPath == "" {
		// Hardcoded path (not configurable)
		geolocationDBPath = "./base/GeoLite2-City.mmdb"
	}
	geolocationSvc, geolocationErr := geolocation.NewService(geolocationDBPath)
	if geolocationErr != nil {
		log.Warn().Err(geolocationErr).Str("path", geolocationDBPath).Msg("Failed to initialize geolocation service, location will be empty. Download GeoLite2-City.mmdb from https://dev.maxmind.com/geoip/geoip2/geolite2/")
	}

	// Schedule GeoLite2 updates (Wed + Sat) and ensure DB exists.
	// Behavior:
	// - If DB file is missing: download immediately (best-effort) and Reload() the in-memory reader.
	// - If DB exists: force refresh on Wednesday and Saturday (once per day).
	scheduleGeoLite2Updates(pool, log, geolocationSvc)

	// Initialize API key repository and service
	// Convert pgxpool.Pool to sql.DB for APIKeyRepository
	sqlDB := stdlib.OpenDBFromPool(pool)
	apiKeyRepo := queries.NewAPIKeyRepository(sqlDB)
	apiKeyService := services.NewAPIKeyService(apiKeyRepo)

	// Initialize email template queries
	emailTemplateQueries := &queries.EmailTemplateQueries{Pool: pool}

	// Initialize API handlers
	apiHandlers := &routes.APIHandlers{
		Submissions:    api.NewSubmissionHandler(submissionRepoImpl, submissionService),
		Submitters:     nil, // TODO: initialize with repository and service
		SigningLinks:   api.NewSigningLinkHandler(pool, templateQueries, completedDoc),
		Templates:      api.NewTemplateHandler(templateRepo, templateQueries),
		Webhooks:       api.NewWebhookHandler(webhookRepo),
		Settings:       api.NewSettingsHandler(notificationService, accountQueries, userQueries, geolocationSvc),
		APIKeys:        api.NewAPIKeyHandler(apiKeyService),
		Stats:          api.NewStatsHandler(pool),
		Events:         api.NewEventHandler(pool),
		Organizations:  api.NewOrganizationHandler(organizationQueries, userQueries),
		Members:        api.NewMemberHandler(organizationQueries, userQueries),
		Invitations:    api.NewInvitationHandler(organizationQueries),
		Users:          api.NewUserHandler(userQueries),
		I18n:           api.NewI18nHandler(userQueries, accountQueries),
		Branding:       api.NewBrandingHandler(accountQueries, userQueries, nil), // TODO: initialize with storage
		EmailTemplates: api.NewEmailTemplateHandler(emailTemplateQueries, userQueries),
		PublicSigning:  public.NewPublicSigningHandler(pool, templateQueries, userQueries, notificationService, completedDoc, geolocationSvc),
	}

	routes.ApiRoutes(app, apiHandlers)
	routes.NotFoundRoute(app)

	fmt.Printf("├─[🚀] Admin UI: http://%s/_/\n", cfg.HTTPAddr)
	fmt.Printf("├─[🚀] Public UI: http://%s/\n", cfg.HTTPAddr)
	fmt.Printf("└─[🚀] Public API: http://%s/api/\n", cfg.HTTPAddr)

	// Listen on port
	go func() {
		if err := app.Listen(cfg.HTTPAddr); err != nil {
			log.Err(err).Send()
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit

	fmt.Print("\n✍️ Shutting down server...\n")
	return app.Shutdown()
}

func createDirs() error {
	dirs := []string{
		"./lc_pages",
		"./lc_signed",
		"./lc_uploads",
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	return nil
}

// scheduleGeoLite2Updates mirrors the Adobe trust-list updater loop:
// a frequent tick (12h) + a "staleness" check so downloads happen ~2x/week.
func scheduleGeoLite2Updates(pool *pgxpool.Pool, log *logging.Logger, geoSvc *geolocation.Service) {
	if pool == nil {
		return
	}

	const (
		dbPath            = "./base/GeoLite2-City.mmdb"
		checkEvery        = 12 * time.Hour
		initialAlignEvery = 12 * time.Hour
	)

	go func() {
		ticker := time.NewTicker(checkEvery)
		defer ticker.Stop()

		// Align initial run to a 12-hour boundary (same style as trust updater).
		now := time.Now()
		nextRun := now.Truncate(initialAlignEvery).Add(initialAlignEvery)
		if nextRun.Before(now) {
			nextRun = nextRun.Add(initialAlignEvery)
		}
		time.Sleep(time.Until(nextRun))

		// First run.
		updateGeoLite2OnSchedule(pool, log, geoSvc, dbPath)

		// Periodic runs.
		for range ticker.C {
			updateGeoLite2OnSchedule(pool, log, geoSvc, dbPath)
		}
	}()
}

func updateGeoLite2OnSchedule(pool *pgxpool.Pool, log *logging.Logger, geoSvc *geolocation.Service, dbPath string) {
	// If file is missing, we should try to download regardless of timestamps.
	_, statErr := os.Stat(dbPath)
	missing := statErr != nil

	// Find the configured method (MaxMind preferred, otherwise URL) and last_updated_at.
	accountID, method, licenseKey, downloadURL, lastUpdatedAt, err := pickGeoLite2Settings(context.Background(), pool)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to read GeoLite2 settings for scheduled update")
		return
	}
	if method == "" {
		// Not configured; nothing to do.
		return
	}

	now := time.Now().UTC()
	force := missing || shouldForceGeoLite2UpdateToday(now, lastUpdatedAt)
	if !force {
		return
	}

	if err := downloadGeoLite2(pool, log, licenseKey, downloadURL, method, accountID, true); err != nil {
		log.Warn().Err(err).Msg("Scheduled GeoLite2 update failed")
		return
	}

	// Reload in-memory DB so updates apply without restart.
	if geoSvc != nil {
		if err := geoSvc.Reload(); err != nil {
			log.Warn().Err(err).Msg("Failed to reload GeoLite2 database after scheduled update")
		}
	}
}

func shouldForceGeoLite2UpdateToday(now time.Time, lastUpdatedAt string) bool {
	// Only update on Wednesday and Saturday (unless file is missing, handled by caller).
	if now.Weekday() != time.Wednesday && now.Weekday() != time.Saturday {
		return false
	}

	// If we can't parse the last update timestamp, run the update.
	if lastUpdatedAt == "" {
		return true
	}
	ts, err := time.Parse(time.RFC3339, lastUpdatedAt)
	if err != nil {
		return true
	}
	ts = ts.UTC()

	// Only run once per scheduled day.
	return ts.Year() != now.Year() || ts.Month() != now.Month() || ts.Day() != now.Day()
}

// pickGeoLite2Settings selects one account's geolocation settings.
// Priority: MaxMind key first, then URL.
func pickGeoLite2Settings(ctx context.Context, pool *pgxpool.Pool) (accountID, method, licenseKey, downloadURL, lastUpdatedAt string, err error) {
	// MaxMind key first.
	{
		row := pool.QueryRow(ctx, `
			SELECT
				id,
				COALESCE(settings->'geolocation'->>'maxmind_license_key', ''),
				COALESCE(settings->'geolocation'->>'last_updated_at', '')
			FROM account
			WHERE COALESCE(settings->'geolocation'->>'maxmind_license_key', '') <> ''
			LIMIT 1
		`)
		var id, key, last string
		if scanErr := row.Scan(&id, &key, &last); scanErr == nil && key != "" {
			return id, "maxmind", key, "", last, nil
		}
	}

	// Fallback: URL.
	{
		row := pool.QueryRow(ctx, `
			SELECT
				id,
				COALESCE(settings->'geolocation'->>'download_url', ''),
				COALESCE(settings->'geolocation'->>'last_updated_at', '')
			FROM account
			WHERE COALESCE(settings->'geolocation'->>'download_url', '') <> ''
			LIMIT 1
		`)
		var id, url, last string
		if scanErr := row.Scan(&id, &url, &last); scanErr == nil && url != "" {
			return id, "url", "", url, last, nil
		}
	}

	return "", "", "", "", "", nil
}

// downloadGeoLite2IfNeeded downloads GeoLite2 database if it doesn't exist
// Tries to download from MaxMind if license_key is configured in database, otherwise skips
func downloadGeoLite2IfNeeded(cfg *config.Config, pool *pgxpool.Pool, log *logging.Logger) error {
	const dbPath = "./base/GeoLite2-City.mmdb"

	// Check if database already exists
	if _, err := os.Stat(dbPath); err == nil {
		return nil
	}

	accountID, method, licenseKey, downloadURL, _, err := pickGeoLite2Settings(context.Background(), pool)
	if err != nil {
		return err
	}
	if method == "" {
		log.Info().Msg("GeoLite2 database not found and no download method configured in database, skipping download")
		return nil
	}
	return downloadGeoLite2(pool, log, licenseKey, downloadURL, method, accountID, false)
}

func downloadGeoLite2(pool *pgxpool.Pool, log *logging.Logger, licenseKey, downloadURL, method, settingsAccountID string, force bool) error {
	// Hardcoded paths (not configurable)
	const baseDir = "./base"
	const dbPath = "./base/GeoLite2-City.mmdb"

	if !force {
		if _, err := os.Stat(dbPath); err == nil {
			return nil
		}
	}

	// Create base directory if it doesn't exist
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return fmt.Errorf("failed to create base directory: %w", err)
	}

	// Console notification (same style as Adobe trust-list updater).
	// Only printed when an actual download is about to happen.
	switch method {
	case "maxmind":
		fmt.Printf("├─[🌍] Updating GeoLite2 database (MaxMind)\n")
	case "url":
		fmt.Printf("├─[🌍] Updating GeoLite2 database (URL)\n")
	default:
		fmt.Printf("├─[🌍] Updating GeoLite2 database\n")
	}

	switch method {
	case "url":
		if downloadURL == "" {
			return fmt.Errorf("download method url selected but download_url is empty")
		}
		client := &http.Client{Timeout: 5 * time.Minute}
		resp, err := client.Get(downloadURL)
		if err != nil {
			return fmt.Errorf("failed to download from URL: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("failed to download from URL: HTTP status %d", resp.StatusCode)
		}

		tmpFile, err := os.CreateTemp("", "geolite2-*.tar.gz")
		if err != nil {
			return fmt.Errorf("failed to create temporary file: %w", err)
		}
		defer os.Remove(tmpFile.Name())
		defer tmpFile.Close()

		if _, err := io.Copy(tmpFile, resp.Body); err != nil {
			return fmt.Errorf("failed to save archive: %w", err)
		}

		if err := extractGeoLite2FromTarGz(tmpFile.Name(), dbPath); err != nil {
			if gzErr := extractGeoLite2FromGzipMMDB(tmpFile.Name(), dbPath); gzErr != nil {
				return fmt.Errorf("failed to extract database: tar.gz error: %w; gzip error: %v", err, gzErr)
			}
		}

		if pool != nil && settingsAccountID != "" {
			_ = queries.NewAccountQueries(pool).UpdateAccountGeolocationLastUpdate(context.Background(), settingsAccountID, time.Now(), "url")
		}
		return nil

	case "maxmind":
		if licenseKey == "" {
			return fmt.Errorf("download method maxmind selected but license key is empty")
		}

		maxmindDownloadURL := fmt.Sprintf("https://download.maxmind.com/app/geoip_download?edition_id=GeoLite2-City&license_key=%s&suffix=tar.gz", licenseKey)
		client := &http.Client{Timeout: 5 * time.Minute}
		resp, err := client.Get(maxmindDownloadURL)
		if err != nil {
			return fmt.Errorf("failed to download from MaxMind: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			log.Error().Int("status", resp.StatusCode).Bytes("body", body).Msg("MaxMind API error")
			return fmt.Errorf("MaxMind API returned status %d", resp.StatusCode)
		}

		tmpFile, err := os.CreateTemp("", "geolite2-*.tar.gz")
		if err != nil {
			return fmt.Errorf("failed to create temporary file: %w", err)
		}
		defer os.Remove(tmpFile.Name())
		defer tmpFile.Close()

		if _, err := io.Copy(tmpFile, resp.Body); err != nil {
			return fmt.Errorf("failed to save archive: %w", err)
		}

		if err := extractGeoLite2FromTarGz(tmpFile.Name(), dbPath); err != nil {
			if gzErr := extractGeoLite2FromGzipMMDB(tmpFile.Name(), dbPath); gzErr != nil {
				return fmt.Errorf("failed to extract database from tar.gz: %w; gzip error: %v", err, gzErr)
			}
		}

		if pool != nil && settingsAccountID != "" {
			_ = queries.NewAccountQueries(pool).UpdateAccountGeolocationLastUpdate(context.Background(), settingsAccountID, time.Now(), "maxmind")
		}
		return nil
	default:
		return fmt.Errorf("unknown geolocation download method: %s", method)
	}
}

// extractGeoLite2FromTarGz extracts GeoLite2-City.mmdb from tar.gz archive
func extractGeoLite2FromTarGz(tarGzPath, outputPath string) error {
	// Open tar.gz file
	file, err := os.Open(tarGzPath)
	if err != nil {
		return fmt.Errorf("failed to open tar.gz file: %w", err)
	}
	defer file.Close()

	// Create gzip reader
	gzReader, err := gzip.NewReader(file)
	if err != nil {
		return fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzReader.Close()

	// Create tar reader
	tarReader := tar.NewReader(gzReader)

	// Find and extract GeoLite2-City.mmdb
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read tar: %w", err)
		}

		// Look for GeoLite2-City.mmdb file
		if header.Typeflag == tar.TypeReg && strings.HasSuffix(header.Name, "GeoLite2-City.mmdb") {
			// Create output directory if needed
			if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
				return fmt.Errorf("failed to create output directory: %w", err)
			}

			// Create output file
			outFile, err := os.Create(outputPath)
			if err != nil {
				return fmt.Errorf("failed to create output file: %w", err)
			}
			defer outFile.Close()

			// Copy file content
			if _, err := io.Copy(outFile, tarReader); err != nil {
				return fmt.Errorf("failed to extract file: %w", err)
			}

			return nil
		}
	}

	return fmt.Errorf("GeoLite2-City.mmdb not found in archive")
}

// extractGeoLite2FromGzipMMDB extracts GeoLite2-City.mmdb from a gzip-compressed mmdb file (GeoLite2-City.mmdb.gz).
func extractGeoLite2FromGzipMMDB(gzPath, outputPath string) error {
	file, err := os.Open(gzPath)
	if err != nil {
		return fmt.Errorf("failed to open gzip file: %w", err)
	}
	defer file.Close()

	gzReader, err := gzip.NewReader(file)
	if err != nil {
		return fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzReader.Close()

	// Create output directory if needed
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	outFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer outFile.Close()

	if _, err := io.Copy(outFile, gzReader); err != nil {
		return fmt.Errorf("failed to extract gzip file: %w", err)
	}

	return nil
}
