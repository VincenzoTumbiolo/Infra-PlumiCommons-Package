package ses

import (
	"context"
	"encoding/base64"
	"os"
	"path/filepath"
	"strings"
	"testing"

	awsses "github.com/aws/aws-sdk-go-v2/service/sesv2"
	"github.com/vincenzotumbiolo/infra-pulumicommons-package/infrastructure/config/axnet/email"
)

type fakeSES struct {
	lastInput *awsses.SendEmailInput
}

func (f *fakeSES) SendEmail(ctx context.Context, in *awsses.SendEmailInput, _ ...func(*awsses.Options)) (*awsses.SendEmailOutput, error) {
	f.lastInput = in
	return &awsses.SendEmailOutput{
		MessageId: awsString("fake-msg-id"),
	}, nil
}

func awsString(s string) *string { return &s }

// salva un .eml in una cartella temporanea del test e logga il percorso
func saveAsEML(t *testing.T, raw string, filename string) {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, filename)
	if err := os.WriteFile(path, []byte(raw), 0644); err != nil {
		t.Fatalf("failed to write EML file: %v", err)
	}
	t.Logf("Raw MIME saved to: %s", path)
	// Se vuoi anche vedere il MIME in log verbose:
	// t.Logf("Raw MIME:\n%s", raw)
}

func TestSender_BuildsCorrectMIMEAndCallsSES(t *testing.T) {
	// Arrange: sender con fake client (così passiamo dal nostro buildRawMIME!)
	f := &fakeSES{}
	s := &Sender{client: f}

	// Dati “nostri”: soggetto, html, testo, allegato CSV
	csv := "id,name\n1,foo\n2,bar\n"
	m := email.Mail{
		From:    "from@example.com",
		To:      []string{"to1@example.com", "to2@example.com"},
		Subject: "Test Subject",
		Text:    "Hello plain",
		HTML:    "<p>Hello <b>HTML</b></p>",
		Attachments: []email.Attachment{
			{Filename: "report.csv", ContentType: "text/csv", Content: []byte(csv)},
		},
	}

	// Act
	if err := s.Send(context.Background(), m); err != nil {
		t.Fatalf("send failed: %v", err)
	}

	// Assert: è stato chiamato SES con Raw.Data popolato
	if f.lastInput == nil || f.lastInput.Content == nil || f.lastInput.Content.Raw == nil {
		t.Fatalf("fake SES not called with raw content")
	}
	raw := string(f.lastInput.Content.Raw.Data)

	// Salva la mail per ispezione manuale
	saveAsEML(t, raw, "mail_with_csv.eml")

	// 1) multipart/mixed e multipart/alternative presenti
	if !strings.Contains(raw, "Content-Type: multipart/mixed;") {
		t.Errorf("missing multipart/mixed in raw MIME")
	}
	if !strings.Contains(raw, "Content-Type: multipart/alternative;") {
		t.Errorf("missing multipart/alternative in raw MIME")
	}

	// 2) parti text/plain e text/html
	if !strings.Contains(raw, "Content-Type: text/plain; charset=UTF-8") || !strings.Contains(raw, "Hello plain") {
		t.Errorf("missing plain text part")
	}
	if !strings.Contains(raw, "Content-Type: text/html; charset=UTF-8") || !strings.Contains(raw, "<p>Hello <b>HTML</b></p>") {
		t.Errorf("missing html part")
	}

	// 3) allegato in base64 con header corretti
	if !strings.Contains(raw, `Content-Type: text/csv; name="report.csv"`) {
		t.Errorf("attachment content-type header missing or wrong")
	}
	if !strings.Contains(raw, `Content-Disposition: attachment; filename="report.csv"`) {
		t.Errorf("attachment disposition header missing or wrong")
	}
	if !strings.Contains(raw, "Content-Transfer-Encoding: base64") {
		t.Errorf("attachment must be base64 encoded")
	}
	// check che il contenuto base64 del CSV appaia (anche parziale)
	enc := base64.StdEncoding.EncodeToString([]byte(csv))
	short := enc[:min(20, len(enc))]
	if !strings.Contains(raw, short) {
		t.Errorf("attachment base64 content not found")
	}

	// 4) intestazioni base
	if f.lastInput.FromEmailAddress == nil || *f.lastInput.FromEmailAddress != "from@example.com" {
		t.Errorf("wrong FromEmailAddress")
	}
	if f.lastInput.Destination == nil || len(f.lastInput.Destination.ToAddresses) != 2 {
		t.Errorf("wrong Destination ToAddresses")
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func TestSender_SendsPNGAttachment(t *testing.T) {
	// PNG 1x1 trasparente (bytes noti)
	png := []byte{
		0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, 0x00, 0x00, 0x00, 0x0D, 0x49, 0x48, 0x44, 0x52,
		0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01, 0x08, 0x06, 0x00, 0x00, 0x00, 0x1F, 0x15, 0xC4,
		0x89, 0x00, 0x00, 0x00, 0x0A, 0x49, 0x44, 0x41, 0x54, 0x78, 0x9C, 0x63, 0x00, 0x01, 0x00, 0x00,
		0x05, 0x00, 0x01, 0x0D, 0x0A, 0x2D, 0xB4, 0x00, 0x00, 0x00, 0x00, 0x49, 0x45, 0x4E, 0x44, 0xAE,
		0x42, 0x60, 0x82,
	}

	f := &fakeSES{}
	s := &Sender{client: f}

	m := email.Mail{
		From:    "from@example.com",
		To:      []string{"to@example.com"},
		Subject: "Image test",
		Text:    "Body text",
		HTML:    "<p>Body <b>HTML</b></p>",
		Attachments: []email.Attachment{
			{Filename: "image.png", ContentType: "image/png", Content: png},
		},
	}

	if err := s.Send(context.Background(), m); err != nil {
		t.Fatalf("send failed: %v", err)
	}

	if f.lastInput == nil || f.lastInput.Content == nil || f.lastInput.Content.Raw == nil {
		t.Fatalf("fake SES not called with raw content")
	}
	raw := string(f.lastInput.Content.Raw.Data)

	// Salva la mail per ispezione manuale
	saveAsEML(t, raw, "mail_with_png.eml")

	// Header dell'allegato PNG
	if !strings.Contains(raw, `Content-Type: image/png; name="image.png"`) {
		t.Errorf("png content-type header missing or wrong")
	}
	if !strings.Contains(raw, `Content-Disposition: attachment; filename="image.png"`) {
		t.Errorf("png disposition header missing or wrong")
	}
	if !strings.Contains(raw, "Content-Transfer-Encoding: base64") {
		t.Errorf("png must be base64 encoded")
	}

	// Il contenuto base64 dell'immagine deve comparire (anche solo un prefisso)
	enc := base64.StdEncoding.EncodeToString(png)
	prefix := enc[:min(20, len(enc))]
	if !strings.Contains(raw, prefix) {
		t.Errorf("png base64 content not found")
	}
}
