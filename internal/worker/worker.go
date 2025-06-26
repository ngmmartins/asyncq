package worker

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/ngmmartins/asyncq/internal/job"
	"github.com/ngmmartins/asyncq/internal/queue"
	"github.com/ngmmartins/asyncq/internal/store"
	"github.com/ngmmartins/asyncq/internal/task"
	"github.com/ngmmartins/asyncq/internal/worker/tasks"
)

type Worker struct {
	store         store.Store
	dispatcher    *queue.Dispatcher
	taskExecutors map[task.Task]TaskExecutor
	logger        *slog.Logger
}

func New(store store.Store, dispatcher *queue.Dispatcher, logger *slog.Logger) *Worker {
	return &Worker{
		store:      store,
		dispatcher: dispatcher,
		taskExecutors: map[task.Task]TaskExecutor{
			task.WebhookTask:   tasks.NewWebhookExecutor(logger),
			task.SendEmailTask: tasks.NewSendEmailExecutor(logger),
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

			jobIds, err := w.dispatcher.Dequeue(ctx, now)
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
	j, err := w.store.Job().Get(ctx, jobId)
	if err != nil {
		w.logger.Error("Error getting job from store", "id", jobId, "err", err.Error())
		//TODO what to do here?
		return
	}

	// update job status and save it
	j.Status = job.StatusRunning
	err = w.store.Job().Update(ctx, j)
	if err != nil {
		w.logger.Error("Error updating job status", "id", jobId, "newJobStatus", j.Status, "err", err.Error())
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

	j.Status = job.StatusDone
	err = w.store.Job().Update(ctx, j)
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
