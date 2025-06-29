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
			w.logger.Info("ticking", "time", now)

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
	if err != nil {
		j.Status = job.StatusFailed
		//TODO add logic here to handle retries and backoff.
		// could increment attempts and in in range enque a new job with runAt now (or with a configurable delay)

		//TODO should this be logged as error? or even logged at all?
		w.logger.Error("Job task execution failed", "id", j.ID, "err", err.Error())
		return
	}

	err = w.jobService.UpdateJobStatus(ctx, jobId, job.StatusDone)
	if err != nil {
		w.logger.Error("Error updating job status", "id", jobId, "newJobStatus", j.Status, "err", err.Error())
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
