package app

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"

	"github.com/shurco/gosign/internal/routes"
	"github.com/shurco/gosign/pkg/config"
	"github.com/shurco/gosign/pkg/logging"
	"github.com/shurco/gosign/pkg/storage/postgres"
	"github.com/shurco/gosign/pkg/storage/redis"
	"github.com/shurco/gosign/pkg/utils/fsutil"
)

func New() error {
	log := logging.Log

	godotenv.Load(".env", "./.env")

	// directories init
	if err := dirInit(); err != nil {
		log.Err(err).Send()
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// redis init
	redis.NewClient(
		ctx,
		config.GetString("REDIS_ADDR", "localhost:6379"),
		config.GetString("REDIS_PASSWORD", "redisPassword"),
	)
	defer redis.Conn.Close()
	if err := redis.Conn.Ping(); err != nil {
		log.Err(err).Send()
		return err
	}

	// pastgresql init
	err := postgres.NewClient(ctx, &postgres.PgSQLConfig{
		DSN: fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=require",
			config.GetString("POSTGRES_USER", "goSign"),
			config.GetString("POSTGRES_PASSWORD", "postgresPassword"),
			config.GetString("POSTGRES_HOST", "localhost:5432"),
			config.GetString("POSTGRES_DB", "goSign"),
		),
	})
	if err != nil {
		log.Err(err).Send()
		return err
	}
	defer postgres.Pool.Close()

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

	routes.ApiRoutes(app)
	routes.NotFoundRoute(app)

	fmt.Print("✍️ Sign documents without stress\n")

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

	appPort := config.GetString("APP_PORT", ":8080")
	StartServer(appPort, app)
	<-idleConnsClosed

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

func dirInit() error {
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
