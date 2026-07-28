// Package worker runs scheduled maintenance tasks (Adobe trust lists,
// GeoLite2 database updates) as a process separate from the HTTP API,
// so the API stays stateless and horizontally scalable.
package worker

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/shurco/gosign/internal/config"
	"github.com/shurco/gosign/internal/queries"
	"github.com/shurco/gosign/internal/trust"
	"github.com/shurco/gosign/internal/worker/tasks"
	"github.com/shurco/gosign/pkg/appdir"
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

	fmt.Printf("└─[⚙️] Maintenance tasks scheduled every %s\n", checkInterval)

	runEvery(ctx, checkInterval, func() {
		if err := trust.Update(); err != nil {
			log.Err(err).Msg("Adobe trust list update failed")
		}
		tasks.SyncGeoLite2(ctx, pool, log)
	})

	fmt.Print("\n✍️ Shutting down worker...\n")
	return nil
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
