package job

import (
	"encoding/json"
	"fmt"
	"slices"
	"time"

	"github.com/ngmmartins/asyncq/internal/task"
)

type Status string

const (
	StatusQueued    Status = "Queued"
	StatusRunning   Status = "Running"
	StatusDone      Status = "Done"
	StatusFailed    Status = "Failed"
	StatusCancelled Status = "Cancelled"
)

var allowedStatusTransitions = map[Status][]Status{
	StatusQueued:    {StatusRunning, StatusCancelled},
	StatusRunning:   {StatusDone, StatusFailed},
	StatusDone:      {},
	StatusFailed:    {StatusQueued},
	StatusCancelled: {},
}

type Job struct {
	ID        string          `json:"id"`
	Task      task.Task       `json:"task"`
	Payload   json.RawMessage `json:"payload"`
	RunAt     time.Time       `json:"run_at"`
	Status    Status          `json:"status"`
	CreatedAt time.Time       `json:"created_at"`
}

type CreateRequest struct {
	Task    task.Task       `json:"task"`
	Payload json.RawMessage `json:"payload"`
	// If nil, run now
	RunAt *time.Time `json:"run_at,omitempty"`
}

func IsValidStatusTransition(from Status, to Status) bool {
	allowed, ok := allowedStatusTransitions[from]
	if !ok {
		return false
	}

	return slices.Contains(allowed, to)
}

type InvalidStatusTransitionError struct {
	From Status
	To   Status
}

func (e *InvalidStatusTransitionError) Error() string {
	return fmt.Sprintf("invalid status transition from %q to %q", e.From, e.To)
}
