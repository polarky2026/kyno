package email

import (
	"bytes"
	"context"
	htmltemplate "html/template"
	texttemplate "text/template"
)

type Sender interface {
	Send(ctx context.Context, msg Message) error
}

type AuthMailer struct {
	sender Sender
}

func NewAuthMailer(sender Sender) *AuthMailer {
	return &AuthMailer{sender: sender}
}

type templateData struct {
	Name string
	URL  string
}

var (
	verificationHTML = htmltemplate.Must(htmltemplate.New("verification").Parse(`<p>Hi {{.Name}},</p>
<p>Confirm your email to start your cinema archive.</p>
<p><a href="{{.URL}}">Verify email</a></p>
<p>This link expires in 24 hours. If you didn't create a Kyno account, you can ignore this email.</p>`))

	verificationText = texttemplate.Must(texttemplate.New("verification").Parse(`Hi {{.Name}},

Confirm your email to start your cinema archive:
{{.URL}}

This link expires in 24 hours. If you didn't create a Kyno account, you can ignore this email.`))

	accountExistsHTML = htmltemplate.Must(htmltemplate.New("account_exists").Parse(`<p>Hi {{.Name}},</p>
<p>Someone tried to create a Kyno account with this email address. You already have an account, so nothing changed.</p>
<p>If this was you, <a href="{{.URL}}">sign in to Kyno</a>. If it wasn't, you can ignore this email.</p>`))

	accountExistsText = texttemplate.Must(texttemplate.New("account_exists").Parse(`Hi {{.Name}},

Someone tried to create a Kyno account with this email address. You already have an account, so nothing changed.

If this was you, sign in to Kyno: {{.URL}}
If it wasn't, you can ignore this email.`))
)

func (m *AuthMailer) SendVerificationEmail(ctx context.Context, to, name, verifyURL string) error {
	return m.send(ctx, to, "Verify your email for Kyno", verificationHTML, verificationText, templateData{Name: name, URL: verifyURL})
}

func (m *AuthMailer) SendAccountExistsEmail(ctx context.Context, to, name, signInURL string) error {
	return m.send(ctx, to, "Someone tried to sign up with your email", accountExistsHTML, accountExistsText, templateData{Name: name, URL: signInURL})
}

func (m *AuthMailer) send(ctx context.Context, to, subject string, html *htmltemplate.Template, text *texttemplate.Template, data templateData) error {
	var htmlBody, textBody bytes.Buffer

	if err := html.Execute(&htmlBody, data); err != nil {
		return err
	}
	if err := text.Execute(&textBody, data); err != nil {
		return err
	}

	return m.sender.Send(ctx, Message{
		To:      to,
		Subject: subject,
		HTML:    htmlBody.String(),
		Text:    textBody.String(),
	})
}
