package tasks

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/ngmmartins/asyncq/internal/job"
	"github.com/ngmmartins/asyncq/internal/task"
)

type WebhookExecutor struct {
	logger *slog.Logger
}

func NewWebhookExecutor(logger *slog.Logger) *WebhookExecutor {
	return &WebhookExecutor{logger: logger}
}

// TODO
func (e *WebhookExecutor) Execute(j job.Job) error {
	var payload task.WebhookPayload
	if err := json.Unmarshal(j.Payload, &payload); err != nil {
		return fmt.Errorf("invalid webhook payload: %w", err)
	}
	req, err := http.NewRequest(payload.Method, payload.URL, bytes.NewReader(payload.Body))
	if err != nil {
		return err
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	fmt.Printf("Webhook call to %s returned %d\n", payload.URL, resp.StatusCode)
	return nil
}
