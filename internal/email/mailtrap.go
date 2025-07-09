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

func (s *MailtrapSender) Send(ctx context.Context, to, cc, bcc []string, from, subject, body string) error {
	msg := mail.NewMsg()
	var allErrors []error

	err := msg.From(from)
	if err != nil {
		allErrors = append(allErrors, fmt.Errorf("%w: %s", ErrInvalidEmailAddress, from))
	}
	err = msg.To(to...)
	if err != nil {
		allErrors = append(allErrors, fmt.Errorf("%w: %s", ErrInvalidEmailAddress, to))
	}
	if len(cc) > 0 {
		err = msg.Cc(cc...)
		if err != nil {
			allErrors = append(allErrors, fmt.Errorf("%w: %s", ErrInvalidEmailAddress, cc))
		}
	}
	if len(bcc) > 0 {
		err = msg.Bcc(bcc...)
		if err != nil {
			allErrors = append(allErrors, fmt.Errorf("%w: %s", ErrInvalidEmailAddress, bcc))
		}
	}

	if len(allErrors) > 0 {
		return errors.Join(allErrors...)
	}

	msg.Subject(subject)
	msg.SetBodyString(mail.TypeTextPlain, body)

	return s.client.DialAndSend(msg)
}
