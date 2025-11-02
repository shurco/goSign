package app

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/gofiber/fiber/v2"
	"github.com/robfig/cron/v3"

	"github.com/shurco/gosign/internal/config"
	"github.com/shurco/gosign/internal/handlers/api"
	"github.com/shurco/gosign/internal/models"
	"github.com/shurco/gosign/internal/queries"
	"github.com/shurco/gosign/internal/routes"
	"github.com/shurco/gosign/internal/services/submission"
	"github.com/shurco/gosign/internal/trust"
	"github.com/shurco/gosign/pkg/logging"
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

func (r *simpleTemplateRepository) Update(id string, item *models.Template) error {
	// TODO: Implement proper template update
	return fmt.Errorf("not implemented")
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

	// Initialize query services
	templateQueries := &queries.TemplateQueries{Pool: pool}
	organizationQueries := queries.NewOrganizationQueries(pool)
	userQueries := queries.NewUserQueries(pool)

	// Create template repository (for now using a simple implementation)
	templateRepo := &simpleTemplateRepository{
		templateQueries: templateQueries,
	}

	// Initialize submission repository and service
	// TODO: Implement proper submission repository with database queries
	// For now, using nil repository - will return empty list for List operations
	submissionRepoImpl := &simpleSubmissionRepository{
		submissionRepo: nil, // TODO: initialize with actual repository implementation
	}
	
	submissionService := submission.NewService(nil, nil, nil)

	// update trust certs
	if err = trust.Update(cfg.Trust); err != nil {
		log.Err(err).Send()
		return err
	}

	task := cron.New()
	_, err = task.AddFunc("0 */12 * * *", func() {
		fmt.Print("cron")
		if err := trust.Update(cfg.Trust); err != nil {
			log.Err(err)
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

	// middleware.Fiber(app, log)
	// routes.SiteRoutes(app)
	app.Static("/drive/pages", "./lc_pages")
	app.Static("/drive/signed", "./lc_signed")
	app.Static("/drive/uploads", "./lc_uploads")

	// Initialize API handlers
	apiHandlers := &routes.APIHandlers{
		Submissions:    api.NewSubmissionHandler(submissionRepoImpl, submissionService),
		Submitters:     nil, // TODO: initialize with repository and service
		Templates:      api.NewTemplateHandler(templateRepo, templateQueries),
		Webhooks:       nil, // TODO: initialize with repository
		Settings:       nil, // TODO: initialize with notification service
		APIKeys:        nil, // TODO: initialize with API key service
		Stats:          api.NewStatsHandler(),
		Events:         api.NewEventHandler(),
		Organizations:  api.NewOrganizationHandler(organizationQueries),
		Members:        api.NewMemberHandler(organizationQueries),
		Invitations:    api.NewInvitationHandler(organizationQueries),
		Users:          api.NewUserHandler(userQueries),
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
