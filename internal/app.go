package app

import (
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

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/static"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/shurco/gosign/internal/assets"
	"github.com/shurco/gosign/internal/config"
	"github.com/shurco/gosign/internal/handlers/api"
	public "github.com/shurco/gosign/internal/handlers/public"
	"github.com/shurco/gosign/internal/middleware"
	"github.com/shurco/gosign/internal/queries"
	"github.com/shurco/gosign/internal/routes"
	"github.com/shurco/gosign/internal/services"
	"github.com/shurco/gosign/internal/services/submission"
	"github.com/shurco/gosign/internal/trust"
	"github.com/shurco/gosign/pkg/appdir"
	"github.com/shurco/gosign/pkg/geolocation"
	"github.com/shurco/gosign/pkg/logging"
	"github.com/shurco/gosign/pkg/notification"
	"github.com/shurco/gosign/pkg/storage/postgres"
	"github.com/shurco/gosign/pkg/storage/redis"
	"github.com/shurco/gosign/pkg/utils"
)

func New() error {
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

	// Create geolocation base directory next to executable
	if err := os.MkdirAll(appdir.Base(), 0755); err != nil {
		log.Warn().Err(err).Str("path", appdir.Base()).Msg("Failed to create geolocation base directory")
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

	submissionService := submission.NewService(submissionRepo, nil, nil)

	// update trust certs
	if err = trust.Update(); err != nil {
		log.Err(err).Send()
		return err
	}

	// Download GeoLite2 database if needed (after Adobe certificates update)
	if err = downloadGeoLite2IfNeeded(cfg, pool, log); err != nil {
		log.Warn().Err(err).Msg("Failed to download GeoLite2 database, continuing without it")
	}

	startPeriodicTask(12*time.Hour, func() {
		if err := trust.Update(); err != nil {
			log.Err(err).Send()
		}
	})

	// web init
	app := fiber.New(fiber.Config{
		BodyLimit: 50 * 1024 * 1024,
	})

	middleware.Fiber(app, log, cfg)
	routes.SiteRoutes(app)
	app.Use("/drive/pages", static.New(appdir.LcPages()))
	app.Use("/drive/signed", static.New(appdir.LcSigned()))
	app.Use("/drive/uploads", static.New(appdir.LcUploads()))

	// Initialize webhook repository
	webhookRepo := &simpleWebhookRepository{}

	notificationService := initNotificationService(settingQueries)

	// Completed document builder (filesystem-backed cache).
	completedDoc := &services.CompletedDocumentBuilder{
		Pool:            pool,
		TemplateQueries: templateQueries,
		PagesDir:        appdir.LcPages(),
		SignedDir:       appdir.LcSigned(),
		AssetsDir:       assetPaths.Dir,
	}

	// Initialize geolocation service (best-effort; works without database)
	geolocationDBPath := os.Getenv("GEOLITE2_DB_PATH")
	if geolocationDBPath == "" {
		geolocationDBPath = filepath.Join(appdir.Base(), "GeoLite2-City.mmdb")
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
	apiKeyRepo := queries.NewAPIKeyRepository(pool)
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
		Settings:       api.NewSettingsHandler(notificationService, accountQueries, userQueries, geolocationSvc, settingQueries),
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
		if err := app.Listen(cfg.HTTPAddr, fiber.ListenConfig{DisableStartupMessage: true}); err != nil {
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

// startPeriodicTask launches a goroutine that aligns to the given interval,
// executes fn once after the alignment, then repeats every interval.
func startPeriodicTask(interval time.Duration, fn func()) {
	go func() {
		now := time.Now()
		nextRun := now.Truncate(interval).Add(interval)
		if nextRun.Before(now) {
			nextRun = nextRun.Add(interval)
		}
		time.Sleep(time.Until(nextRun))

		fn()

		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for range ticker.C {
			fn()
		}
	}()
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

func initNotificationService(settingQueries *queries.SettingQueries) *notification.Service {
	svc := notification.NewService(nil)
	ctx := context.Background()

	if smtpMap, err := settingQueries.GetGlobalSetting(ctx, "smtp"); err == nil && utils.GetStringFromMap(smtpMap, "provider", "") == "smtp" {
		port, _ := strconv.Atoi(utils.GetStringFromMap(smtpMap, "smtp_port", "1025"))
		if port == 0 {
			port = 1025
		}
		smtpCfg := notification.SMTPConfig{
			Host:      utils.GetStringFromMap(smtpMap, "smtp_host", ""),
			Port:      port,
			User:      utils.GetStringFromMap(smtpMap, "smtp_user", ""),
			Password:  utils.GetStringFromMap(smtpMap, "smtp_pass", ""),
			FromEmail: utils.GetStringFromMap(smtpMap, "from_email", ""),
			FromName:  utils.GetStringFromMap(smtpMap, "from_name", ""),
		}
		if smtpCfg.Host != "" && smtpCfg.FromEmail != "" {
			svc.RegisterProvider(notification.NewEmailProvider(smtpCfg))
		}
	}

	if smsMap, err := settingQueries.GetGlobalSetting(ctx, "sms"); err == nil && utils.GetStringFromMap(smsMap, "twilio_enabled", "") == "true" {
		svc.RegisterProvider(notification.NewSMSProvider(notification.TwilioConfig{
			AccountSID: utils.GetStringFromMap(smsMap, "twilio_account_sid", ""),
			AuthToken:  utils.GetStringFromMap(smsMap, "twilio_auth_token", ""),
			FromNumber: utils.GetStringFromMap(smsMap, "twilio_from_number", ""),
			Enabled:    true,
		}))
	}

	return svc
}

// scheduleGeoLite2Updates mirrors the Adobe trust-list updater loop:
// a frequent tick (12h) + a "staleness" check so downloads happen ~2x/week.
func scheduleGeoLite2Updates(pool *pgxpool.Pool, log *logging.Logger, geoSvc *geolocation.Service) {
	if pool == nil {
		return
	}

	dbPath := filepath.Join(appdir.Base(), "GeoLite2-City.mmdb")

	startPeriodicTask(12*time.Hour, func() {
		updateGeoLite2OnSchedule(pool, log, geoSvc, dbPath)
	})
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

// updateGlobalGeolocationLastUpdate stores last download time and source in global setting table (key geolocation).
func updateGlobalGeolocationLastUpdate(ctx context.Context, pool *pgxpool.Pool, updatedAt time.Time, source string) {
	_, err := pool.Exec(ctx, `
		UPDATE setting
		SET value = value || jsonb_build_object('last_updated_at', $1::text, 'last_updated_source', $2::text)
		WHERE key = 'geolocation'
	`, updatedAt.UTC().Format(time.RFC3339), source)
	if err != nil {
		// Best-effort; do not fail the download
		return
	}
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

	// Fallback: URL from account.
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
			return id, "url", "", strings.TrimSpace(url), last, nil
		}
	}

	// Fallback: global setting table (key = 'geolocation')
	{
		row := pool.QueryRow(ctx, `
			SELECT COALESCE(value->>'maxmind_license_key', ''), COALESCE(value->>'download_url', ''), COALESCE(value->>'download_method', '')
			FROM setting
			WHERE key = 'geolocation'
			LIMIT 1
		`)
		var globalKey, globalURL, globalMethod string
		if scanErr := row.Scan(&globalKey, &globalURL, &globalMethod); scanErr == nil {
			globalKey = strings.TrimSpace(globalKey)
			globalURL = strings.TrimSpace(globalURL)
			if globalKey != "" {
				return "", "maxmind", globalKey, "", "", nil
			}
			if globalURL != "" {
				return "", "url", "", globalURL, "", nil
			}
		}
	}

	return "", "", "", "", "", nil
}

// downloadGeoLite2IfNeeded downloads GeoLite2 database if it doesn't exist
// Tries to download from MaxMind if license_key is configured in database, otherwise skips
func downloadGeoLite2IfNeeded(cfg *config.Config, pool *pgxpool.Pool, log *logging.Logger) error {
	dbPath := filepath.Join(appdir.Base(), "GeoLite2-City.mmdb")

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
	baseDir := appdir.Base()
	dbPath := filepath.Join(baseDir, "GeoLite2-City.mmdb")

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
		client := &http.Client{
			Timeout: 5 * time.Minute,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				if len(via) >= 10 {
					return fmt.Errorf("stopped after 10 redirects")
				}
				return nil
			},
		}
		resp, err := client.Get(downloadURL)
		if err != nil {
			return fmt.Errorf("failed to download from URL: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("failed to download from URL: HTTP status %d", resp.StatusCode)
		}

		urlLower := strings.ToLower(downloadURL)
		isDirectMMDB := strings.HasSuffix(urlLower, ".mmdb") && !strings.HasSuffix(urlLower, ".mmdb.gz")

		if isDirectMMDB {
			outFile, err := os.Create(dbPath)
			if err != nil {
				return fmt.Errorf("failed to create database file: %w", err)
			}
			defer outFile.Close()
			if _, err := io.Copy(outFile, resp.Body); err != nil {
				return fmt.Errorf("failed to save database file: %w", err)
			}
		} else {
			tmpFile, err := os.CreateTemp("", "geolite2-*")
			if err != nil {
				return fmt.Errorf("failed to create temporary file: %w", err)
			}
			defer os.Remove(tmpFile.Name())
			defer tmpFile.Close()

			if _, err := io.Copy(tmpFile, resp.Body); err != nil {
				return fmt.Errorf("failed to save archive: %w", err)
			}

			if err := geolocation.ExtractFromTarGz(tmpFile.Name(), dbPath); err != nil {
				if gzErr := geolocation.ExtractFromGzip(tmpFile.Name(), dbPath); gzErr != nil {
					return fmt.Errorf("failed to extract database: tar.gz error: %w; gzip error: %v", err, gzErr)
				}
			}
		}

		if pool != nil {
			if settingsAccountID != "" {
				_ = queries.NewAccountQueries(pool).UpdateAccountGeolocationLastUpdate(context.Background(), settingsAccountID, time.Now(), "url")
			} else {
				updateGlobalGeolocationLastUpdate(context.Background(), pool, time.Now(), "url")
			}
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

		if err := geolocation.ExtractFromTarGz(tmpFile.Name(), dbPath); err != nil {
			if gzErr := geolocation.ExtractFromGzip(tmpFile.Name(), dbPath); gzErr != nil {
				return fmt.Errorf("failed to extract database from tar.gz: %w; gzip error: %v", err, gzErr)
			}
		}

		if pool != nil {
			if settingsAccountID != "" {
				_ = queries.NewAccountQueries(pool).UpdateAccountGeolocationLastUpdate(context.Background(), settingsAccountID, time.Now(), "maxmind")
			} else {
				updateGlobalGeolocationLastUpdate(context.Background(), pool, time.Now(), "maxmind")
			}
		}
		return nil
	default:
		return fmt.Errorf("unknown geolocation download method: %s", method)
	}
}
