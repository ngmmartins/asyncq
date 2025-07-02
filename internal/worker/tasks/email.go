package tasks

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/ngmmartins/asyncq/internal/email"
	"github.com/ngmmartins/asyncq/internal/job"
	"github.com/ngmmartins/asyncq/internal/task"
)

type SendEmailExecutor struct {
	logger      *slog.Logger
	emailSender email.EmailSender
}

func NewSendEmailExecutor(logger *slog.Logger, emailSender email.EmailSender) *SendEmailExecutor {
	return &SendEmailExecutor{logger: logger, emailSender: emailSender}
}

// TODO
func (e *SendEmailExecutor) Execute(j job.Job) error {
	var payload task.SendEmailPayload
	if err := json.Unmarshal(j.Payload, &payload); err != nil {
		return fmt.Errorf("invalid email payload: %w", err)
	}

	return e.emailSender.Send(
		context.Background(),
		payload.To,
		payload.From,
		payload.Subject,
		payload.Body,
	)
}
