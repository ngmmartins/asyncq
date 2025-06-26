package store

import (
	"context"
	"errors"

	"github.com/ngmmartins/asyncq/internal/job"
)

var (
	ErrRecordNotFound = errors.New("record not found")
	ErrNoRowsAffected = errors.New("no rows affected after query execution")
)

type Store interface {
	Job() JobStore
}

type JobStore interface {
	Save(ctx context.Context, job *job.Job) error
	Get(ctx context.Context, jobId string) (*job.Job, error)
	Update(ctx context.Context, job *job.Job) error
}
