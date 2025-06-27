package queue

import (
	"context"
	"time"
)

type Queue interface {
	Enqueue(ctx context.Context, jobId string, runAt time.Time) error
	Dequeue(ctx context.Context, timeThreshold time.Time) ([]string, error)
	Remove(ctx context.Context, jobId string) error
}
