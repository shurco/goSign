package worker

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

// Task represents task for execution
type Task interface {
	// Execute performs the task
	Execute(ctx context.Context) error
	// ShouldRetry determines if task should be retried on error
	ShouldRetry(err error) bool
}

// Job represents scheduled task with schedule
type Job struct {
	Name     string
	Schedule string // cron format
	Task     Task
}

// jobRunner manages a single scheduled job
type jobRunner struct {
	job      Job
	stop     chan struct{}
	interval time.Duration
	stopOnce sync.Once
}

// Worker manages background tasks
type Worker struct {
	jobs       map[string]*jobRunner
	mu         sync.RWMutex
	maxRetries int
	ctx        context.Context
	cancel     context.CancelFunc
	wg         sync.WaitGroup
}

// NewWorker creates new worker
func NewWorker(maxRetries int) *Worker {
	ctx, cancel := context.WithCancel(context.Background())
	return &Worker{
		jobs:       make(map[string]*jobRunner),
		maxRetries: maxRetries,
		ctx:        ctx,
		cancel:     cancel,
	}
}

// parseCronSchedule parses cron expression and returns interval duration
// Supports basic cron patterns: "0 */N * * *" (every N hours), "*/N * * * *" (every N minutes)
func parseCronSchedule(schedule string) (time.Duration, error) {
	parts := strings.Fields(schedule)
	if len(parts) != 5 {
		return 0, fmt.Errorf("invalid cron format: expected 5 fields, got %d", len(parts))
	}

	// Pattern: "0 */N * * *" - every N hours
	hourPattern := regexp.MustCompile(`^\*/(\d+)$`)
	if parts[0] == "0" && hourPattern.MatchString(parts[1]) {
		matches := hourPattern.FindStringSubmatch(parts[1])
		if len(matches) == 2 {
			hours, err := strconv.Atoi(matches[1])
			if err != nil {
				return 0, fmt.Errorf("invalid hour interval: %w", err)
			}
			if hours > 0 {
				return time.Duration(hours) * time.Hour, nil
			}
		}
	}

	// Pattern: "*/N * * * *" - every N minutes
	minutePattern := regexp.MustCompile(`^\*/(\d+)$`)
	if minutePattern.MatchString(parts[0]) {
		matches := minutePattern.FindStringSubmatch(parts[0])
		if len(matches) == 2 {
			minutes, err := strconv.Atoi(matches[1])
			if err != nil {
				return 0, fmt.Errorf("invalid minute interval: %w", err)
			}
			if minutes > 0 {
				return time.Duration(minutes) * time.Minute, nil
			}
		}
	}

	// Pattern: "0 0 * * *" - daily at midnight
	if parts[0] == "0" && parts[1] == "0" {
		return 24 * time.Hour, nil
	}

	// Pattern: "0 0 * * 0" - weekly on Sunday
	if parts[0] == "0" && parts[1] == "0" && parts[4] == "0" {
		return 7 * 24 * time.Hour, nil
	}

	return 0, fmt.Errorf("unsupported cron pattern: %s", schedule)
}

// executeJobWithRetry executes job with retry logic
func (w *Worker) executeJobWithRetry(job Job) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	log.Info().Str("job", job.Name).Msg("Starting job")
	startTime := time.Now()

	// Execute task with retry
	var lastErr error
	for attempt := 0; attempt <= w.maxRetries; attempt++ {
		if attempt > 0 {
			// Exponential backoff
			backoff := time.Duration(attempt*attempt) * time.Second
			time.Sleep(backoff)
			log.Warn().
				Str("job", job.Name).
				Int("attempt", attempt+1).
				Msg("Retrying job")
		}

		err := job.Task.Execute(ctx)
		if err == nil {
			// Success
			duration := time.Since(startTime)
			log.Info().
				Str("job", job.Name).
				Dur("duration", duration).
				Msg("Job completed successfully")
			return
		}

		lastErr = err

		// Check if retry is needed
		if !job.Task.ShouldRetry(err) {
			log.Error().
				Err(err).
				Str("job", job.Name).
				Msg("Job failed, no retry")
			return
		}
	}

	// All attempts exhausted
	log.Error().
		Err(lastErr).
		Str("job", job.Name).
		Int("retries", w.maxRetries).
		Msg("Job failed after all retries")
}

// runJob runs a scheduled job in a goroutine
func (w *Worker) runJob(runner *jobRunner) {
	w.wg.Add(1)
	go func() {
		defer w.wg.Done()

		// Calculate initial delay to align with schedule
		now := time.Now()
		nextRun := now.Truncate(runner.interval).Add(runner.interval)
		if nextRun.Before(now) {
			nextRun = nextRun.Add(runner.interval)
		}
		initialDelay := time.Until(nextRun)

		// Wait for initial delay or stop signal
		select {
		case <-w.ctx.Done():
			return
		case <-runner.stop:
			return
		case <-time.After(initialDelay):
			// Execute immediately on first run
			w.executeJobWithRetry(runner.job)
		}

		// Create ticker for periodic execution
		ticker := time.NewTicker(runner.interval)
		defer ticker.Stop()

		for {
			select {
			case <-w.ctx.Done():
				return
			case <-runner.stop:
				return
			case <-ticker.C:
				w.executeJobWithRetry(runner.job)
			}
		}
	}()
}

// AddJob adds task to schedule
func (w *Worker) AddJob(job Job) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	// Check if job already exists
	if _, exists := w.jobs[job.Name]; exists {
		return fmt.Errorf("job %s already exists", job.Name)
	}

	// Parse cron schedule to interval
	interval, err := parseCronSchedule(job.Schedule)
	if err != nil {
		return fmt.Errorf("failed to parse schedule: %w", err)
	}

	// Create job runner
	runner := &jobRunner{
		job:      job,
		stop:     make(chan struct{}),
		interval: interval,
	}

	w.jobs[job.Name] = runner
	log.Info().Str("job", job.Name).Str("schedule", job.Schedule).Dur("interval", interval).Msg("Job registered")

	// Start job runner
	w.runJob(runner)

	return nil
}

// RemoveJob removes task from schedule
func (w *Worker) RemoveJob(name string) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if runner, ok := w.jobs[name]; ok {
		runner.stopOnce.Do(func() {
			close(runner.stop)
		})
		delete(w.jobs, name)
		log.Info().Str("job", name).Msg("Job removed")
	}
}

// Start starts worker
func (w *Worker) Start() {
	log.Info().Int("jobs", len(w.jobs)).Msg("Worker started")
}

// Stop stops worker
func (w *Worker) Stop() {
	w.cancel()

	// Stop all job runners
	w.mu.Lock()
	for name, runner := range w.jobs {
		runner.stopOnce.Do(func() {
			close(runner.stop)
		})
		delete(w.jobs, name)
	}
	w.mu.Unlock()

	// Wait for all goroutines to finish
	w.wg.Wait()
	log.Info().Msg("Worker stopped")
}

// RunOnce executes task once (for testing or manual run)
func (w *Worker) RunOnce(ctx context.Context, task Task) error {
	log.Info().Msg("Running task once")

	var lastErr error
	for attempt := 0; attempt <= w.maxRetries; attempt++ {
		if attempt > 0 {
			backoff := time.Duration(attempt*attempt) * time.Second
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(backoff):
			}
		}

		err := task.Execute(ctx)
		if err == nil {
			return nil
		}

		lastErr = err
		if !task.ShouldRetry(err) {
			return err
		}
	}

	return lastErr
}

