package email_test

import (
	"context"
	"testing"

	"github.com/vincenzotumbiolo/infra-pulumicommons-package/infrastructure/config/axnet/email"
	"github.com/vincenzotumbiolo/infra-pulumicommons-package/infrastructure/config/axnet/email/mock"
)

func TestMockSender(t *testing.T) {
	sender := &mock.Sender{}
	m := email.Mail{
		From:    "from@example.com",
		To:      []string{"to@example.com"},
		Subject: "Test",
		Text:    "Hello world",
	}
	err := sender.Send(context.Background(), m)
	if err != nil {
		t.Fatal(err)
	}
	if len(sender.Sent) != 1 {
		t.Fatalf("expected 1 email, got %d", len(sender.Sent))
	}
}
