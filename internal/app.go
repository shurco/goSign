package app

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/gofiber/fiber/v2"
	"github.com/robfig/cron/v3"

	"github.com/shurco/gosign/internal/config"
	"github.com/shurco/gosign/internal/middleware"
	"github.com/shurco/gosign/internal/queries"
	"github.com/shurco/gosign/internal/routes"
	"github.com/shurco/gosign/internal/trust"
	"github.com/shurco/gosign/pkg/logging"
	"github.com/shurco/gosign/pkg/storage/postgres"
	"github.com/shurco/gosign/pkg/storage/redis"
	"github.com/shurco/gosign/pkg/utils/fsutil"
)

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

	// update trust certs
	if err = trust.Update(cfg.Trust); err != nil {
		log.Err(err).Send()
		return err
	}

	task := cron.New()
	_, err = task.AddFunc("0 */12 * * *", func() {
		fmt.Print("cron")
		if err := trust.Update(cfg.Trust); err != nil {
			log.Err(err).Send()
		}
	})
	if err != nil {
		log.Err(err).Send()
		return err
	}
	task.Start()

	// web init
	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
		BodyLimit:             50 * 1024 * 1024,
	})

	// Enable middleware based on mode
	if cfg.DevMode {
		middleware.Fiber(app, log)
		routes.SiteRoutes(app)
	}
	
	app.Static("/drive/pages", "./lc_pages")
	app.Static("/drive/signed", "./lc_signed")
	app.Static("/drive/uploads", "./lc_uploads")

	routes.ApiRoutes(app)
	routes.NotFoundRoute(app)

	fmt.Printf("├─[🚀] Admin UI: http://%s/_/\n", cfg.HTTPAddr)
	fmt.Printf("├─[🚀] Public UI: http://%s/\n", cfg.HTTPAddr)
	fmt.Printf("└─[🚀] Public API: http://%s/api/\n", cfg.HTTPAddr)

	if cfg.DevMode {
		StartServer(cfg.HTTPAddr, app)
	} else {
		idleConnsClosed := make(chan struct{})

		go func() {
			sigint := make(chan os.Signal, 1)
			signal.Notify(sigint, os.Interrupt)
			<-sigint

			if err := app.Shutdown(); err != nil {
				log.Err(err).Send()
			}

			close(idleConnsClosed)
		}()

		StartServer(cfg.HTTPAddr, app)
		<-idleConnsClosed
	}

	return nil
}

// StartServer is ...
func StartServer(addr string, a *fiber.App) {
	log := logging.Log

	if err := a.Listen(addr); err != nil {
		log.Err(err).Send()
		os.Exit(1)
	}
}

func createDirs() error {
	dirsToCheck := []struct {
		path string
		name string
	}{
		{"./lc_pages", "lc_pages"},
		{"./lc_signed", "lc_signed"},
		{"./lc_tmp", "lc_tmp"},
		{"./lc_uploads", "lc_uploads"},
	}

	for _, dir := range dirsToCheck {
		if err := fsutil.MkDirs(0o775, dir.path); err != nil {
			return err
		}
	}

	return nil
}
