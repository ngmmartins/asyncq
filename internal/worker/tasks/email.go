package tasks

import (
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/ngmmartins/asyncq/internal/job"
	"github.com/ngmmartins/asyncq/internal/task"
)

type SendEmailExecutor struct {
	logger *slog.Logger
}

func NewSendEmailExecutor(logger *slog.Logger) *SendEmailExecutor {
	return &SendEmailExecutor{logger: logger}
}

// TODO
func (e *SendEmailExecutor) Execute(j job.Job) error {
	var payload task.SendEmailPayload
	if err := json.Unmarshal(j.Payload, &payload); err != nil {
		return fmt.Errorf("invalid email payload: %w", err)
	}
	// TODO: Replace with actual email logic
	fmt.Printf("Sending email to: %s\n", payload.To)
	return nil
}
