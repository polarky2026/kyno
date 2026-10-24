package email

import (
	"context"
	"strings"
	"testing"
)

type recordingSender struct {
	messages []Message
}

func (r *recordingSender) Send(_ context.Context, msg Message) error {
	r.messages = append(r.messages, msg)
	return nil
}

func TestVerificationEmailIncludesTheLink(t *testing.T) {
	sender := &recordingSender{}
	mailer := NewAuthMailer(sender)

	err := mailer.SendVerificationEmail(context.Background(), "a@example.com", "George", "http://localhost:5173/verify-email?token=abc")
	if err != nil {
		t.Fatalf("SendVerificationEmail: %v", err)
	}

	msg := sender.messages[0]
	if msg.To != "a@example.com" {
		t.Errorf("To = %q", msg.To)
	}
	if !strings.Contains(msg.HTML, `href="http://localhost:5173/verify-email?token=abc"`) {
		t.Errorf("HTML does not contain the link: %s", msg.HTML)
	}
	if !strings.Contains(msg.Text, "http://localhost:5173/verify-email?token=abc") {
		t.Errorf("text does not contain the link: %s", msg.Text)
	}
}

func TestEmailsEscapeUserProvidedNames(t *testing.T) {
	sender := &recordingSender{}
	mailer := NewAuthMailer(sender)

	_ = mailer.SendAccountExistsEmail(context.Background(), "a@example.com", `<a href="https://evil.example">`, "http://localhost:5173/sign-in")

	if strings.Contains(sender.messages[0].HTML, `<a href="https://evil.example">`) {
		t.Fatalf("the name was not escaped: %s", sender.messages[0].HTML)
	}
}
