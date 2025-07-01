package job

import (
	"encoding/json"
	"slices"
	"time"

	"github.com/ngmmartins/asyncq/internal/pagination"
	"github.com/ngmmartins/asyncq/internal/task"
)

type Status string

// add to statusList when adding here a new const
const (
	StatusCreated   Status = "Created"
	StatusQueued    Status = "Queued"
	StatusRunning   Status = "Running"
	StatusDone      Status = "Done"
	StatusFailed    Status = "Failed"
	StatusCancelled Status = "Cancelled"
)

var StatusList = []Status{StatusCreated, StatusQueued, StatusRunning, StatusDone, StatusFailed, StatusCancelled}

var allowedStatusTransitions = map[Status][]Status{
	StatusCreated:   {StatusQueued},
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
	RunAt     *time.Time      `json:"run_at,omitempty"`
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

type SearchCriteria struct {
	Task      task.Task
	RunBefore *time.Time
	RunAfter  *time.Time
	Status    Status
	pagination.Params
}
