package ses

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"

	"github.com/vincenzotumbiolo/infra-pulumicommons-package/infrastructure/config/axnet/email"
)

func buildRawMIME(m email.Mail) []byte {
	mixed := boundary("mixed")
	alt := boundary("alt")
	var b bytes.Buffer
	w := func(s string) { _, _ = io.WriteString(&b, s) }

	w(fmt.Sprintf("From: %s\r\n", m.From))
	w(fmt.Sprintf("To: %s\r\n", join(m.To)))
	w(fmt.Sprintf("Subject: %s\r\n", m.Subject))
	w("MIME-Version: 1.0\r\n")
	w(fmt.Sprintf("Content-Type: multipart/mixed; boundary=\"%s\"\r\n\r\n", mixed))

	// part alternative
	w(fmt.Sprintf("--%s\r\n", mixed))
	w(fmt.Sprintf("Content-Type: multipart/alternative; boundary=\"%s\"\r\n\r\n", alt))

	if m.Text != "" {
		w(fmt.Sprintf("--%s\r\n", alt))
		w("Content-Type: text/plain; charset=UTF-8\r\n")
		w("Content-Transfer-Encoding: 7bit\r\n\r\n")
		w(m.Text + "\r\n\r\n")
	}
	if m.HTML != "" {
		w(fmt.Sprintf("--%s\r\n", alt))
		w("Content-Type: text/html; charset=UTF-8\r\n")
		w("Content-Transfer-Encoding: 7bit\r\n\r\n")
		w(m.HTML + "\r\n\r\n")
	}
	w(fmt.Sprintf("--%s--\r\n", alt))

	// attachments
	for _, a := range m.Attachments {
		ct := a.ContentType
		if ct == "" {
			ct = "application/octet-stream"
		}
		w(fmt.Sprintf("--%s\r\n", mixed))
		w(fmt.Sprintf("Content-Type: %s; name=\"%s\"\r\n", ct, a.Filename))
		w(fmt.Sprintf("Content-Disposition: attachment; filename=\"%s\"\r\n", a.Filename))
		w("Content-Transfer-Encoding: base64\r\n\r\n")
		enc := base64.StdEncoding.EncodeToString(a.Content)
		for i := 0; i < len(enc); i += 76 {
			end := i + 76
			if end > len(enc) {
				end = len(enc)
			}
			w(enc[i:end] + "\r\n")
		}
		w("\r\n")
	}

	w(fmt.Sprintf("--%s--\r\n", mixed))
	return b.Bytes()
}

func boundary(prefix string) string {
	var r [12]byte
	_, _ = rand.Read(r[:])
	return fmt.Sprintf("%s_%x", prefix, r)
}

func join(a []string) string {
	switch len(a) {
	case 0:
		return ""
	case 1:
		return a[0]
	default:
		s := a[0]
		for i := 1; i < len(a); i++ {
			s += ", " + a[i]
		}
		return s
	}
}
