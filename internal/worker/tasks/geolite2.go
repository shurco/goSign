// Package tasks contains scheduled maintenance tasks executed by the worker.
package tasks

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/shurco/gosign/internal/queries"
	"github.com/shurco/gosign/pkg/appdir"
	"github.com/shurco/gosign/pkg/geolocation"
	"github.com/shurco/gosign/pkg/logging"
)

// SyncGeoLite2 downloads or refreshes the GeoLite2 database when needed:
// immediately when the file is missing, otherwise on Wednesday and Saturday
// (once per day). Download method and credentials are read from the database.
func SyncGeoLite2(ctx context.Context, pool *pgxpool.Pool, log *logging.Logger) {
	dbPath := appdir.GeoLite2()

	// If file is missing, we should try to download regardless of timestamps.
	_, statErr := os.Stat(dbPath)
	missing := statErr != nil

	accountID, method, licenseKey, downloadURL, lastUpdatedAt := pickGeoLite2Settings(ctx, pool)
	if method == "" {
		// Not configured; nothing to do.
		return
	}

	if !missing && !shouldForceGeoLite2UpdateToday(time.Now().UTC(), lastUpdatedAt) {
		return
	}

	if err := downloadGeoLite2(ctx, pool, log, licenseKey, downloadURL, method, accountID); err != nil {
		log.Warn().Err(err).Msg("GeoLite2 update failed")
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

// SyncOutcome is the result of an on-demand GeoLite2 sync.
type SyncOutcome int

const (
	SyncOutcomeSuccess SyncOutcome = iota
	// SyncOutcomeSkipped means the database already exists and force was not set.
	SyncOutcomeSkipped
)

// SyncParams are explicit overrides for an on-demand GeoLite2 sync.
// Zero values fall back to the settings saved in the database.
type SyncParams struct {
	Force      bool
	Method     string // "", "url" or "maxmind"
	URL        string
	LicenseKey string
	AccountID  string // account to record the last-update timestamp for
}

// SyncGeoLite2OnDemand downloads the database using explicit params, falling
// back to saved settings for anything not provided. Unlike SyncGeoLite2 it
// ignores the weekday schedule; the caller is expected to hold the
// cross-replica maintenance lock.
func SyncGeoLite2OnDemand(ctx context.Context, pool *pgxpool.Pool, log *logging.Logger, p SyncParams) (SyncOutcome, error) {
	dbPath := appdir.GeoLite2()
	if _, err := os.Stat(dbPath); err == nil && !p.Force {
		return SyncOutcomeSkipped, nil
	}

	method := p.Method
	licenseKey := p.LicenseKey
	downloadURL := p.URL
	accountID := p.AccountID

	switch method {
	case "url":
		if downloadURL == "" {
			return 0, fmt.Errorf("download method url selected but url is empty")
		}
	case "maxmind":
		if licenseKey == "" {
			_, _, settingsKey, _, _ := pickGeoLite2Settings(ctx, pool)
			licenseKey = settingsKey
		}
		if licenseKey == "" {
			return 0, fmt.Errorf("maxmind license key is not provided and not configured")
		}
	case "":
		settingsAccountID, settingsMethod, settingsKey, settingsURL, _ := pickGeoLite2Settings(ctx, pool)
		if settingsMethod == "" {
			return 0, fmt.Errorf("geolocation download is not configured")
		}
		method, licenseKey, downloadURL = settingsMethod, settingsKey, settingsURL
		if accountID == "" {
			accountID = settingsAccountID
		}
	default:
		return 0, fmt.Errorf("unknown geolocation download method: %s", method)
	}

	if err := downloadGeoLite2(ctx, pool, log, licenseKey, downloadURL, method, accountID); err != nil {
		return 0, err
	}
	return SyncOutcomeSuccess, nil
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
// Priority: MaxMind key first, then URL, then the global setting table.
func pickGeoLite2Settings(ctx context.Context, pool *pgxpool.Pool) (accountID, method, licenseKey, downloadURL, lastUpdatedAt string) {
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
			return id, "maxmind", key, "", last
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
			return id, "url", "", strings.TrimSpace(url), last
		}
	}

	// Fallback: global setting table (key = 'geolocation')
	{
		row := pool.QueryRow(ctx, `
			SELECT COALESCE(value->>'maxmind_license_key', ''), COALESCE(value->>'download_url', ''), COALESCE(value->>'last_updated_at', '')
			FROM setting
			WHERE key = 'geolocation'
			LIMIT 1
		`)
		var globalKey, globalURL, globalLast string
		if scanErr := row.Scan(&globalKey, &globalURL, &globalLast); scanErr == nil {
			globalKey = strings.TrimSpace(globalKey)
			globalURL = strings.TrimSpace(globalURL)
			if globalKey != "" {
				return "", "maxmind", globalKey, "", globalLast
			}
			if globalURL != "" {
				return "", "url", "", globalURL, globalLast
			}
		}
	}

	return "", "", "", "", ""
}

func downloadGeoLite2(ctx context.Context, pool *pgxpool.Pool, log *logging.Logger, licenseKey, downloadURL, method, settingsAccountID string) error {
	baseDir := appdir.Base()
	dbPath := appdir.GeoLite2()

	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return fmt.Errorf("failed to create base directory: %w", err)
	}

	switch method {
	case "url":
		if downloadURL == "" {
			return fmt.Errorf("download method url selected but download_url is empty")
		}
	case "maxmind":
		if licenseKey == "" {
			return fmt.Errorf("download method maxmind selected but license key is empty")
		}
		downloadURL = fmt.Sprintf("https://download.maxmind.com/app/geoip_download?edition_id=GeoLite2-City&license_key=%s&suffix=tar.gz", licenseKey)
	default:
		return fmt.Errorf("unknown geolocation download method: %s", method)
	}

	fmt.Printf("├─[🌍] Updating GeoLite2 database (%s)\n", method)

	if err := fetchGeoLite2(ctx, downloadURL, dbPath, log); err != nil {
		return err
	}

	if settingsAccountID != "" {
		_ = queries.NewAccountQueries(pool).UpdateAccountGeolocationLastUpdate(ctx, settingsAccountID, time.Now(), method)
	} else {
		updateGlobalGeolocationLastUpdate(ctx, pool, time.Now(), method)
	}
	return nil
}

// fetchGeoLite2 downloads the database and stores it at dbPath.
// Plain .mmdb URLs are saved as-is; anything else is treated as a tar.gz/gzip archive.
// The database is staged into a temp file next to dbPath and atomically
// renamed into place, so concurrent readers (server and other worker
// replicas via the shared volume) never observe a partially written file.
func fetchGeoLite2(ctx context.Context, downloadURL, dbPath string, log *logging.Logger) error {
	client := &http.Client{Timeout: 5 * time.Minute}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, downloadURL, nil)
	if err != nil {
		return fmt.Errorf("failed to build download request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to download GeoLite2 database: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		log.Error().Int("status", resp.StatusCode).Bytes("body", body).Msg("GeoLite2 download error")
		return fmt.Errorf("GeoLite2 download returned status %d", resp.StatusCode)
	}

	// PID suffix keeps staging files of different replicas apart.
	stagePath := fmt.Sprintf("%s.tmp-%d", dbPath, os.Getpid())
	defer os.Remove(stagePath)

	urlLower := strings.ToLower(downloadURL)
	if strings.HasSuffix(urlLower, ".mmdb") {
		if err := saveToFile(stagePath, resp.Body); err != nil {
			return err
		}
		return replaceDBFile(stagePath, dbPath)
	}

	tmpFile, err := os.CreateTemp("", "geolite2-*")
	if err != nil {
		return fmt.Errorf("failed to create temporary file: %w", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	if _, err := io.Copy(tmpFile, resp.Body); err != nil {
		return fmt.Errorf("failed to save archive: %w", err)
	}

	if err := geolocation.ExtractFromTarGz(tmpFile.Name(), stagePath); err != nil {
		if gzErr := geolocation.ExtractFromGzip(tmpFile.Name(), stagePath); gzErr != nil {
			return fmt.Errorf("failed to extract database: tar.gz error: %w; gzip error: %v", err, gzErr)
		}
	}
	return replaceDBFile(stagePath, dbPath)
}

func saveToFile(path string, r io.Reader) error {
	out, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create database file: %w", err)
	}
	defer out.Close()
	if _, err := io.Copy(out, r); err != nil {
		return fmt.Errorf("failed to save database file: %w", err)
	}
	return nil
}

// replaceDBFile atomically moves the staged database into place.
func replaceDBFile(stagePath, dbPath string) error {
	if err := os.Rename(stagePath, dbPath); err != nil {
		return fmt.Errorf("failed to replace database file: %w", err)
	}
	return nil
}
