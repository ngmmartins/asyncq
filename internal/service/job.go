package service

import (
	"context"
	"errors"
	"fmt"
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

var (
	ErrRecordNotFound          = store.ErrRecordNotFound
	ErrInvalidStatusTransition = errors.New("invalid status transition")
)

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

	var status job.Status
	if request.RunAt != nil {
		status = job.StatusQueued
	} else {
		status = job.StatusCreated
	}

	job := job.Job{
		ID:        uuid.NewString(),
		Task:      request.Task,
		Payload:   request.Payload,
		RunAt:     request.RunAt,
		Status:    status,
		CreatedAt: now,
	}

	err := js.store.Job().Save(ctx, &job)
	if err != nil {
		js.logger.Error("failed to store job", "id", job.ID, "err", err.Error())
		return nil, err
	}

	if job.RunAt != nil {
		err = js.queue.Enqueue(ctx, job.ID, *job.RunAt)
		if err != nil {
			js.logger.Error("failed to enqueue job", "jobID", job.ID)
			return nil, err
		}
	}

	return &job, nil
}

func (js *JobService) GetJob(ctx context.Context, jobId string) (*job.Job, error) {
	j, err := js.store.Job().Get(ctx, jobId)
	if err != nil {
		if errors.Is(err, store.ErrRecordNotFound) {
			return nil, ErrRecordNotFound
		}
		return nil, err
	}

	return j, nil
}

func (js *JobService) ScheduleJob(ctx context.Context, jobId string, runAt time.Time) error {
	j, err := js.store.Job().Get(ctx, jobId)
	if err != nil {
		if errors.Is(err, store.ErrRecordNotFound) {
			return ErrRecordNotFound
		}
		return err
	}

	if !job.IsValidStatusTransition(j.Status, job.StatusQueued) {
		return fmt.Errorf("%w from %q to %q", ErrInvalidStatusTransition, j.Status, job.StatusQueued)
	}

	j.RunAt = &runAt
	j.Status = job.StatusQueued

	err = js.store.Job().Update(ctx, j)
	if err != nil {
		return err
	}

	err = js.queue.Enqueue(ctx, j.ID, *j.RunAt)
	if err != nil {
		return err
	}

	return nil
}

func (js *JobService) UpdateJobStatus(ctx context.Context, jobId string, newStatus job.Status) error {
	j, err := js.store.Job().Get(ctx, jobId)
	if err != nil {
		if errors.Is(err, store.ErrRecordNotFound) {
			return ErrRecordNotFound
		}
		return err
	}

	if !job.IsValidStatusTransition(j.Status, newStatus) {
		return fmt.Errorf("%w from %q to %q", ErrInvalidStatusTransition, j.Status, job.StatusQueued)
	}

	j.Status = newStatus
	return js.store.Job().Update(ctx, j)
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
