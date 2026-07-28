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

	urlLower := strings.ToLower(downloadURL)
	if strings.HasSuffix(urlLower, ".mmdb") {
		outFile, err := os.Create(dbPath)
		if err != nil {
			return fmt.Errorf("failed to create database file: %w", err)
		}
		defer outFile.Close()
		if _, err := io.Copy(outFile, resp.Body); err != nil {
			return fmt.Errorf("failed to save database file: %w", err)
		}
		return nil
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

	if err := geolocation.ExtractFromTarGz(tmpFile.Name(), dbPath); err != nil {
		if gzErr := geolocation.ExtractFromGzip(tmpFile.Name(), dbPath); gzErr != nil {
			return fmt.Errorf("failed to extract database: tar.gz error: %w; gzip error: %v", err, gzErr)
		}
	}
	return nil
}
