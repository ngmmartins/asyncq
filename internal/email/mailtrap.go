package email

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/wneessen/go-mail"
)

var (
	ErrInvalidEmailAddress = errors.New("invalid email address")
)

type MailtrapSender struct {
	client *mail.Client
}

func NewMailtrapSender(logger *slog.Logger, host string, port int, username, password string) *MailtrapSender {
	client, err := mail.NewClient(
		host,
		mail.WithSMTPAuth(mail.SMTPAuthLogin),
		mail.WithPort(port),
		mail.WithUsername(username),
		mail.WithPassword(password),
		mail.WithTimeout(5*time.Second),
	)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	ms := &MailtrapSender{
		client: client,
	}

	return ms
}

func (s *MailtrapSender) Send(ctx context.Context, to, from, subject, body string) error {
	msg := mail.NewMsg()

	err := msg.From(from)
	if err != nil {
		return fmt.Errorf("%w: %s", ErrInvalidEmailAddress, from)
	}
	err = msg.To(to)
	if err != nil {
		return fmt.Errorf("%w: %s", ErrInvalidEmailAddress, to)
	}
	msg.Subject(subject)
	msg.SetBodyString(mail.TypeTextPlain, body)

	return s.client.DialAndSend(msg)
}
