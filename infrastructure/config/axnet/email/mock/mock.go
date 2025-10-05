package mock

import (
	"context"
	"fmt"

	"github.com/vincenzotumbiolo/infra-pulumicommons-package/infrastructure/config/axnet/email"
)

type Sender struct {
	Sent []email.Mail
}

func (m *Sender) Send(ctx context.Context, mail email.Mail) error {
	m.Sent = append(m.Sent, mail)
	fmt.Printf("[MOCK] would send email to %v, subject: %s\n", mail.To, mail.Subject)
	return nil
}
