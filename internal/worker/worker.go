// Package worker runs scheduled maintenance tasks (Adobe trust lists,
// GeoLite2 database updates) and serves internal gRPC requests coming
// from the API server. It opens no public web ports, so the API stays
// the single HTTP entry point.
package worker

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/shurco/gosign/internal/config"
	"github.com/shurco/gosign/internal/queries"
	"github.com/shurco/gosign/pkg/appdir"
	"github.com/shurco/gosign/pkg/geolocation"
	"github.com/shurco/gosign/pkg/logging"
	"github.com/shurco/gosign/pkg/storage/postgres"
)

// checkInterval is how often tasks are polled. Each task decides internally
// whether real work is due (trust: daily; GeoLite2: Wednesday/Saturday).
const checkInterval = time.Hour

// Run starts the background worker and blocks until an interrupt signal.
func Run() error {
	fmt.Print("✍️ goSign worker\n")

	appdir.Init()

	log := logging.Log

	if err := config.Load(); err != nil {
		log.Err(err).Send()
		return err
	}
	cfg := config.Data()

	// GeoLite2 database lives here; the volume is shared with the server.
	if err := os.MkdirAll(appdir.Base(), 0755); err != nil {
		log.Err(err).Str("path", appdir.Base()).Msg("Failed to create base directory")
		return err
	}

	pool, err := postgres.New(context.Background(), cfg.Postgres)
	if err != nil {
		log.Err(err).Send()
		return err
	}
	defer pool.Close()

	queries.New(pool)
	if err := queries.CheckSchema(context.Background(), pool); err != nil {
		log.Err(err).Send()
		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	// The worker owns the GeoLite2 database: it downloads updates and serves
	// IP lookups over gRPC. The reader reloads when a replica swaps the file.
	geoSvc, geoErr := geolocation.NewService(appdir.GeoLite2())
	if geoErr != nil {
		log.Warn().Err(geoErr).Str("path", appdir.GeoLite2()).Msg("GeoLite2 database is not available yet, lookups return empty results until it is downloaded")
	}
	defer geoSvc.Close()
	go watchGeoLite2(ctx, geoSvc, log)

	// Internal gRPC endpoint serving requests from the API server.
	svc := newRPCService(pool, geoSvc, log)
	grpcSrv, err := serveGRPC(cfg.WorkerGRPCAddr, cfg.WorkerGRPCToken, svc, log)
	if err != nil {
		log.Err(err).Send()
		return err
	}

	fmt.Printf("├─[🔌] gRPC: %s (internal, requests from the API server)\n", cfg.WorkerGRPCAddr)
	fmt.Printf("└─[⚙️] Maintenance tasks scheduled every %s\n", checkInterval)

	runEvery(ctx, checkInterval, func() {
		svc.runScheduled(ctx)
	})

	fmt.Print("\n✍️ Shutting down worker...\n")
	grpcSrv.GracefulStop()
	return nil
}

// geoLite2ReloadInterval is how often each replica checks whether another
// replica replaced the GeoLite2 database file on the shared volume.
const geoLite2ReloadInterval = time.Minute

// watchGeoLite2 reloads the GeoLite2 reader when the file on disk changes.
func watchGeoLite2(ctx context.Context, svc *geolocation.Service, log *logging.Logger) {
	ticker := time.NewTicker(geoLite2ReloadInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := svc.ReloadIfChanged(); err != nil {
				log.Warn().Err(err).Msg("Failed to reload GeoLite2 database")
			}
		}
	}
}

// runEvery executes fn immediately and then on every tick until ctx is done.
func runEvery(ctx context.Context, interval time.Duration, fn func()) {
	fn()

	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			fn()
		}
	}
}
