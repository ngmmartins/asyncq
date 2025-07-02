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

const DefaultRetryDelay = 60

type Job struct {
	ID            string          `json:"id"`
	Task          task.Task       `json:"task"`
	Payload       json.RawMessage `json:"payload"`
	RunAt         *time.Time      `json:"run_at,omitempty"`
	Status        Status          `json:"status"`
	CreatedAt     time.Time       `json:"created_at"`
	FinishedAt    *time.Time      `json:"finished_at,omitempty"` // When the job finished execution (either with success or not)
	Retries       int             `json:"retries"`               // How many times the job has already been retried
	MaxRetries    int             `json:"max_retries"`           // Maximum number of retry attempts allowed for the job
	RetryDelaySec int             `json:"retry_delay_sec"`       // Interval in seconds between each retry
	LastError     *string         `json:"last_error,omitempty"`  // Stores the last error message encountered when running the job
}

type CreateRequest struct {
	Task    task.Task       `json:"task"`
	Payload json.RawMessage `json:"payload"`
	// If nil, run now
	RunAt         *time.Time `json:"run_at,omitempty"`
	MaxRetries    *int       `json:"max_retries"`
	RetryDelaySec *int       `json:"retry_delay_sec"`
}

// This type is for "internal" update requests only.
// It's not intended to be used by a client (through a handler).
// In the future when supporting that a new UpdateRequest should be created
type UpdateFields struct {
	SetRunAt bool
	RunAt    *time.Time

	SetStatus bool
	Status    *Status

	SetFinishedAt bool
	FinishedAt    *time.Time

	SetRetries bool
	Retries    *int

	SetLastError bool
	LastError    *string
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
