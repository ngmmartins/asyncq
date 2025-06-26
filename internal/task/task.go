package task

import (
	"encoding/json"

	"github.com/ngmmartins/asyncq/internal/validator"
)

type Task string

const (
	WebhookTask   Task = "webhook"
	SendEmailTask Task = "send_email"
)

// list of supported tasks
var Tasks = []Task{WebhookTask, SendEmailTask}

type WebhookPayload struct {
	URL     string            `json:"url"`
	Method  string            `json:"method"`
	Headers map[string]string `json:"headers,omitempty"`
	Body    json.RawMessage   `json:"body,omitempty"`
}

func ValidateWebhookPayload(v *validator.Validator, p *WebhookPayload) {
	v.CheckRequired(p.URL != "", "payload.url")
	v.CheckRequired(p.Method != "", "payload.method")
	// TODO other checks
}

type SendEmailPayload struct {
	From    string `json:"from"`
	To      string `json:"to"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

func ValidateSendEmailPayload(v *validator.Validator, p *SendEmailPayload) {
	v.CheckRequired(p.From != "", "payload.from")
	v.CheckRequired(p.To != "", "payload.to")
	v.CheckRequired(p.Subject != "", "payload.subject")
	// TODO other checks
}
