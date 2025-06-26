package job

import (
	"encoding/json"
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
