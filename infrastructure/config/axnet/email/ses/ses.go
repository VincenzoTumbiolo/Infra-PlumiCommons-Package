package ses

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsses "github.com/aws/aws-sdk-go-v2/service/sesv2"
	"github.com/aws/aws-sdk-go-v2/service/sesv2/types"
	"github.com/vincenzotumbiolo/infra-pulumicommons-package/infrastructure/config/axnet/email"
)

type Sender struct {
	client sesAPI
}

type sesAPI interface {
	SendEmail(ctx context.Context, params *awsses.SendEmailInput, optFns ...func(*awsses.Options)) (*awsses.SendEmailOutput, error)
}

func NewSender(cfg aws.Config, region string) *Sender {
	c := awsses.NewFromConfig(cfg, func(o *awsses.Options) { o.Region = region })
	return &Sender{client: c}
}

func (s *Sender) Send(ctx context.Context, m email.Mail) error {
	raw := buildRawMIME(m)
	input := &awsses.SendEmailInput{
		FromEmailAddress: aws.String(m.From),
		Destination:      &types.Destination{ToAddresses: m.To},
		Content:          &types.EmailContent{Raw: &types.RawMessage{Data: raw}},
	}
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if _, err := s.client.SendEmail(ctx, input); err != nil {
		return fmt.Errorf("ses send failed: %w", err)
	}
	return nil
}
