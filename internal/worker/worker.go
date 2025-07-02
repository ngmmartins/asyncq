package worker

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/ngmmartins/asyncq/internal/email"
	"github.com/ngmmartins/asyncq/internal/job"
	"github.com/ngmmartins/asyncq/internal/queue"
	"github.com/ngmmartins/asyncq/internal/service"
	"github.com/ngmmartins/asyncq/internal/store"
	"github.com/ngmmartins/asyncq/internal/task"
	"github.com/ngmmartins/asyncq/internal/worker/tasks"
)

type Worker struct {
	store         store.Store
	queue         queue.Queue
	jobService    *service.JobService
	taskExecutors map[task.Task]TaskExecutor
	logger        *slog.Logger
}

func New(store store.Store, queue queue.Queue, logger *slog.Logger,
	jobService *service.JobService, emailSender email.EmailSender) *Worker {

	return &Worker{
		store:      store,
		queue:      queue,
		jobService: jobService,
		taskExecutors: map[task.Task]TaskExecutor{
			task.WebhookTask:   tasks.NewWebhookExecutor(logger),
			task.SendEmailTask: tasks.NewSendEmailExecutor(logger, emailSender),
		},
		logger: logger,
	}
}

func (w *Worker) Run(ctx context.Context, tickInterval time.Duration) {
	ticker := time.NewTicker(tickInterval)
	defer ticker.Stop()
	w.logger.Info(fmt.Sprintf("worker configured with tick interval=%v", tickInterval))

	for {
		select {
		case <-ticker.C:
			now := time.Now()
			w.logger.Debug("ticking", "time", now)

			jobIds, err := w.queue.Dequeue(ctx, now)
			if err != nil {
				w.logger.Error("Error dequeing jobs", "err", err.Error())
				continue
			}

			for _, jobId := range jobIds {
				go w.handleJob(ctx, jobId)
			}
		case <-ctx.Done():
			w.logger.Info("Worker stopped")
			return
		}
	}
}

func (w *Worker) handleJob(ctx context.Context, jobId string) {
	w.logger.Debug("handling job", "jobId", jobId)
	// update job status and save it
	err := w.jobService.UpdateJobStatus(ctx, jobId, job.StatusRunning)
	if err != nil {
		w.logger.Error("Error updating job status", "id", jobId, "newJobStatus", job.StatusRunning, "err", err.Error())
		//TODO what to do here?
		return
	}

	j, err := w.jobService.GetJob(ctx, jobId)
	if err != nil {
		w.logger.Error("Error getting job from store", "id", jobId, "err", err.Error())
		//TODO what to do here?
		return
	}

	err = w.executeTask(j)

	now := time.Now()
	updateFields := job.UpdateFields{}

	updateFields.SetFinishedAt = true
	updateFields.FinishedAt = &now

	if err != nil {
		w.logger.Debug("job execution failed", "jobId", jobId, "err", err.Error())
		updateFields.SetLastError = true
		lastErr := err.Error()
		updateFields.LastError = &lastErr

		enqueueJob := false
		// Check if the job still has retry attempts left
		if j.Retries < j.MaxRetries {
			w.logger.Debug("job still has remaining attempts", "jobId", jobId, "retries", j.Retries, "maxRetries", j.MaxRetries)
			enqueueJob = true

			updateFields.SetRetries = true
			newRetries := j.Retries + 1
			updateFields.Retries = &newRetries

			updateFields.SetStatus = true
			status := job.StatusQueued
			updateFields.Status = &status

			updateFields.SetRunAt = true
			nextRunAt := now.Add(time.Second * time.Duration(j.RetryDelaySec))
			updateFields.RunAt = &nextRunAt

		} else {
			w.logger.Debug("job does not have remaining attempts", "jobId", jobId, "retries", j.Retries, "maxRetries", j.MaxRetries)
			updateFields.SetStatus = true
			status := job.StatusFailed
			updateFields.Status = &status
		}

		w.logger.Debug("updating job fields", "jobId", jobId, "updateFields", updateFields)
		err = w.jobService.UpdateJobFields(ctx, jobId, &updateFields)
		if err != nil {
			w.logger.Error("Error updating job fields", "id", jobId, "updateFields", updateFields, "err", err.Error())
			//TODO what to do here?
			return
		}

		if enqueueJob {
			w.logger.Debug("Enqueueing job again with new RunAt", "jobId", jobId, "RunAt", updateFields.RunAt)
			// Enqueue the job again to be retried
			err := w.queue.Enqueue(ctx, j.ID, *updateFields.RunAt)
			if err != nil {
				w.logger.Error("failed to enqueue job", "jobID", j.ID)
			}
		}

		return
	}

	w.logger.Debug("job execution succedded", "jobId", jobId)

	// Clear eventual past errors
	updateFields.SetLastError = true
	updateFields.LastError = nil

	updateFields.SetStatus = true
	status := job.StatusDone
	updateFields.Status = &status

	w.logger.Debug("updating job fields", "jobId", jobId, "updateFields", updateFields)
	err = w.jobService.UpdateJobFields(ctx, jobId, &updateFields)
	if err != nil {
		w.logger.Error("Error updating job fields", "id", jobId, "updateFields", updateFields, "err", err.Error())
		//TODO what to do here?
		return
	}

}

func (w *Worker) executeTask(j *job.Job) error {
	executor, ok := w.taskExecutors[j.Task]
	if !ok {
		//TODO change to logger
		return fmt.Errorf("unknown task: %s", j.Task)
	}
	return executor.Execute(*j)
}
