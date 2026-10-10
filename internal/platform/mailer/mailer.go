// Package mailer sends email: through Resend (resend.com) when an API key
// is set, otherwise it only logs what it would have sent (local
// development).
package mailer

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"
)

// Email is one message to one person.
type Email struct {
	To      string
	Subject string
	Text    string // plain-text version
	HTML    string
}

type Mailer interface {
	Send(ctx context.Context, e Email) error
	// Real reports whether mail actually goes out.
	Real() bool
}

// New returns a Resend mailer, or a log-only one when apiKey is empty.
func New(apiKey, from string) Mailer {
	if apiKey == "" {
		return logMailer{}
	}
	return &resend{
		apiKey: apiKey,
		from:   from,
		client: &http.Client{Timeout: 10 * time.Second},
		url:    "https://api.resend.com/emails",
	}
}

type resend struct {
	apiKey string
	from   string
	client *http.Client
	url    string
}

func (m *resend) Real() bool { return true }

func (m *resend) Send(ctx context.Context, e Email) error {
	body, err := json.Marshal(map[string]interface{}{
		"from":    m.from,
		"to":      []string{e.To},
		"subject": e.Subject,
		"text":    e.Text,
		"html":    e.HTML,
	})
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, m.url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+m.apiKey)
	req.Header.Set("Content-Type", "application/json")
	resp, err := m.client.Do(req)
	if err != nil {
		return fmt.Errorf("send email: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		msg, _ := io.ReadAll(io.LimitReader(resp.Body, 500))
		return fmt.Errorf("send email: resend returned %d: %s", resp.StatusCode, msg)
	}
	return nil
}

// logMailer is for local development without an API key: the message is
// written to the log instead (so a developer can read the code from it).
type logMailer struct{}

func (logMailer) Real() bool { return false }

func (logMailer) Send(_ context.Context, e Email) error {
	slog.Info("email NOT sent (no RESEND_API_KEY) — logged instead", "to", e.To, "subject", e.Subject, "text", e.Text)
	return nil
}
