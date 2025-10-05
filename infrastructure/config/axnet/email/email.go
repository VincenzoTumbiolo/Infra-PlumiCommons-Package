package email

import (
	"context"
)

type Attachment struct {
	Filename    string
	ContentType string // es. "text/csv"
	Content     []byte
}

type Mail struct {
	From        string
	To          []string
	Subject     string
	Text        string
	HTML        string
	Attachments []Attachment
}

type Sender interface {
	Send(ctx context.Context, m Mail) error
}
