package service

import (
	"context"
	"log/slog"
	"slices"
	"time"

	"github.com/google/uuid"
	"github.com/ngmmartins/asyncq/internal/job"
	"github.com/ngmmartins/asyncq/internal/queue"
	"github.com/ngmmartins/asyncq/internal/store"
	"github.com/ngmmartins/asyncq/internal/task"
	"github.com/ngmmartins/asyncq/internal/validator"
)

type JobService struct {
	logger *slog.Logger
	queue  queue.Queue
	store  store.Store
}

func NewJobService(logger *slog.Logger, queue queue.Queue, store store.Store) *JobService {
	return &JobService{logger: logger, queue: queue, store: store}
}

func (js *JobService) CreateJob(ctx context.Context, request *job.CreateRequest) (*job.Job, error) {
	v := validator.New()
	js.validateCreateJob(v, request)
	if !v.Valid() {
		return nil, &validator.ValidationError{Errors: v.Errors}
	}

	now := time.Now()

	runAt := now
	if request.RunAt != nil {
		runAt = *request.RunAt
	}

	job := job.Job{
		ID:        uuid.NewString(),
		Task:      request.Task,
		Payload:   request.Payload,
		RunAt:     runAt,
		Status:    job.StatusQueued,
		CreatedAt: now,
	}

	err := js.store.Job().Save(ctx, &job)
	if err != nil {
		js.logger.Error("failed to store job", "id", job.ID, "err", err.Error())
		return nil, err
	}

	err = js.queue.Enqueue(ctx, job.ID, job.RunAt)
	if err != nil {
		js.logger.Error("failed to enqueue job", "jobID", job.ID)
		return nil, err
	}

	return &job, nil
}

func (js *JobService) GetJob(ctx context.Context, jobId string) (*job.Job, error) {
	return js.store.Job().Get(ctx, jobId)
}

func (js *JobService) UpdateJobStatus(ctx context.Context, jobId string, newStatus job.Status) error {
	j, err := js.store.Job().Get(ctx, jobId)
	if err != nil {
		return err
	}

	if !job.IsValidStatusTransition(j.Status, newStatus) {
		return &job.InvalidStatusTransitionError{
			From: j.Status,
			To:   newStatus,
		}
	}

	j.Status = newStatus
	err = js.store.Job().Update(ctx, j)

	return err
}

func (js *JobService) validateCreateJob(v *validator.Validator, request *job.CreateRequest) {
	v.CheckRequired(request.Task != "", "task")
	v.Check(slices.Contains(task.Tasks, request.Task), "task", "unsupported task")
	v.CheckRequired(len(request.Payload) > 0, "payload")
	v.Check(request.RunAt == nil || request.RunAt.After(time.Now()), "run_at", "must be in the future")

	_, err := task.DecodeAndValidatePayload(request.Task, request.Payload, v)
	if err != nil {
		v.AddError("payload", "invalid payload for task")
	}
}
