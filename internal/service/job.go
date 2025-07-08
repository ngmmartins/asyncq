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
	"github.com/ngmmartins/asyncq/internal/pagination"
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

func (s *JobService) CreateJob(ctx context.Context, request *job.CreateRequest) (*job.Job, error) {
	v := validator.New()
	s.validateCreateJob(v, request)
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

	maxRetries := 0
	if request.MaxRetries != nil {
		maxRetries = *request.MaxRetries
	}

	// if user didn't provided a delay we apply a default
	retryDelay := job.DefaultRetryDelay
	if request.RetryDelaySec != nil {
		retryDelay = *request.RetryDelaySec
	}

	job := job.Job{
		ID:            uuid.NewString(),
		Task:          request.Task,
		Payload:       request.Payload,
		RunAt:         request.RunAt,
		Status:        status,
		CreatedAt:     now,
		MaxRetries:    maxRetries,
		RetryDelaySec: retryDelay,
	}

	err := s.store.Job().Save(ctx, &job)
	if err != nil {
		s.logger.Error("failed to store job", "id", job.ID, "err", err.Error())
		return nil, err
	}

	if job.RunAt != nil {
		err = s.queue.Enqueue(ctx, job.ID, *job.RunAt)
		if err != nil {
			s.logger.Error("failed to enqueue job", "jobID", job.ID)
			return nil, err
		}
	}

	return &job, nil
}

func (s *JobService) SearchJobs(ctx context.Context, criteria *job.SearchCriteria) ([]*job.Job, *pagination.Metadata, error) {
	v := validator.New()
	s.validateSearchJobs(v, criteria)
	if !v.Valid() {
		return nil, nil, &validator.ValidationError{Errors: v.Errors}
	}

	jobs, metadata, err := s.store.Job().Search(ctx, criteria)
	if err != nil {
		return nil, nil, err
	}

	return jobs, metadata, nil
}

func (s *JobService) GetJob(ctx context.Context, jobId string) (*job.Job, error) {
	j, err := s.store.Job().Get(ctx, jobId)
	if err != nil {
		if errors.Is(err, store.ErrRecordNotFound) {
			return nil, ErrRecordNotFound
		}
		return nil, err
	}

	return j, nil
}

func (s *JobService) ScheduleJob(ctx context.Context, jobId string, runAt time.Time) error {
	j, err := s.store.Job().Get(ctx, jobId)
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

	err = s.store.Job().Update(ctx, j)
	if err != nil {
		return err
	}

	err = s.queue.Enqueue(ctx, j.ID, *j.RunAt)
	if err != nil {
		return err
	}

	return nil
}

func (s *JobService) UpdateJobFields(ctx context.Context, jobId string, fields *job.UpdateFields) error {
	j, err := s.store.Job().Get(ctx, jobId)
	if err != nil {
		if errors.Is(err, store.ErrRecordNotFound) {
			return ErrRecordNotFound
		}
		return err
	}

	if fields.SetRunAt {
		j.RunAt = fields.RunAt
	}
	//TODO should check here the status transiction? if so how to reply back and be handled?
	if fields.SetStatus {
		j.Status = *fields.Status
	}
	if fields.SetFinishedAt {
		j.FinishedAt = fields.FinishedAt
	}
	if fields.SetRetries {
		j.Retries = *fields.Retries
	}
	if fields.SetLastError {
		j.LastError = fields.LastError
	}

	return s.store.Job().Update(ctx, j)

}

func (s *JobService) UpdateJobStatus(ctx context.Context, jobId string, newStatus job.Status) error {
	j, err := s.store.Job().Get(ctx, jobId)
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
	return s.store.Job().Update(ctx, j)
}

func (s *JobService) validateSearchJobs(v *validator.Validator, criteria *job.SearchCriteria) {
	if criteria.Task != "" {
		v.Check(slices.Contains(task.Tasks, criteria.Task), "task", "unsupported task")
	}
	if criteria.Status != "" {
		v.Check(slices.Contains(job.StatusList, criteria.Status), "status", "unsupported status")
	}
	pagination.Validate(v, &criteria.Params, true)

}

func (s *JobService) validateCreateJob(v *validator.Validator, request *job.CreateRequest) {
	v.CheckRequired(request.Task != "", "task")
	v.Check(slices.Contains(task.Tasks, request.Task), "task", "unsupported task")
	v.CheckRequired(len(request.Payload) > 0, "payload")
	v.Check(request.RunAt == nil || request.RunAt.After(time.Now()), "run_at", "must be in the future")
	v.Check(request.MaxRetries == nil || *request.MaxRetries >= 0, "max_retries", "if set must be equal or greater than 0")
	v.Check(request.RetryDelaySec == nil || *request.RetryDelaySec > 0, "retry_delay_sec", "if set must be greater than 0")

	_, err := task.DecodeAndValidatePayload(request.Task, request.Payload, v)
	if err != nil {
		v.AddError("payload", "invalid payload for task")
	}
}
