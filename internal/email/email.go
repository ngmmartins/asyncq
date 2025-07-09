package email

import "context"

type EmailSender interface {
	Send(ctx context.Context, to, cc, bcc []string, from, subject, body string) error
}
