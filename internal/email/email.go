package email

import "context"

type EmailSender interface {
	Send(ctx context.Context, to, from, subject, body string) error
}
