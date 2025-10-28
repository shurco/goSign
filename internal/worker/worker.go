package worker

import (
	"context"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
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

// Worker manages background tasks
type Worker struct {
	cron       *cron.Cron
	jobs       map[string]cron.EntryID
	mu         sync.RWMutex
	maxRetries int
}

// NewWorker creates new worker
func NewWorker(maxRetries int) *Worker {
	return &Worker{
		cron:       cron.New(),
		jobs:       make(map[string]cron.EntryID),
		maxRetries: maxRetries,
	}
}

// AddJob adds task to schedule
func (w *Worker) AddJob(job Job) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	// Create wrapper for task with retry logic
	wrappedFunc := func() {
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

	// Add task to cron
	entryID, err := w.cron.AddFunc(job.Schedule, wrappedFunc)
	if err != nil {
		return err
	}

	w.jobs[job.Name] = entryID
	log.Info().Str("job", job.Name).Str("schedule", job.Schedule).Msg("Job registered")
	return nil
}

// RemoveJob removes task from schedule
func (w *Worker) RemoveJob(name string) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if entryID, ok := w.jobs[name]; ok {
		w.cron.Remove(entryID)
		delete(w.jobs, name)
		log.Info().Str("job", name).Msg("Job removed")
	}
}

// Start starts worker
func (w *Worker) Start() {
	w.cron.Start()
	log.Info().Int("jobs", len(w.jobs)).Msg("Worker started")
}

// Stop stops worker
func (w *Worker) Stop() {
	ctx := w.cron.Stop()
	<-ctx.Done()
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

