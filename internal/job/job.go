package job

import (
	"encoding/json"
	"slices"
	"time"

	"github.com/ngmmartins/asyncq/internal/task"
	"github.com/ngmmartins/asyncq/internal/validator"
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

func ValidateCreateJob(v *validator.Validator, input *CreateRequest) {
	v.CheckRequired(input.Task != "", "task")
	v.Check(slices.Contains(task.Tasks, input.Task), "task", "unsupported task")
	v.CheckRequired(len(input.Payload) > 0, "payload")
	v.Check(input.RunAt == nil || input.RunAt.After(time.Now()), "run_at", "must be in the future")

	_, err := task.DecodeAndValidatePayload(input.Task, input.Payload, v)
	if err != nil {
		v.AddError("payload", "invalid payload for task")
	}

}
