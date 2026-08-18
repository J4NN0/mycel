package tool

import (
	"context"
	"encoding/json"
	"fmt"
	"net/mail"
	"slices"
	"strings"
	"time"

	"github.com/J4NN0/mycel/internal/logger"
	"github.com/maximhq/bifrost/core/schemas"
	"github.com/resend/resend-go/v3"
)

const (
	emailName    = "send_email"
	emailTimeout = 30 * time.Second

	emailDesc = "Send an email on the user's behalf. Only call this when the user's latest message explicitly asks for an email to be sent. Never call it to test it, to demonstrate it, or to check whether it works."
	toDesc    = "Recipient email address, exactly as the user gave it. Never invent an address and never use a placeholder such as one at example.com. If you do not have a real recipient, ask the user instead of calling this tool."
)

var reservedDomains = []string{"example", "example.com", "example.net", "example.org", "example.edu", "invalid", "test", "localhost"}

var _ Tool = (*Email)(nil)

type emailArgs struct {
	To      string `json:"to"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

type Email struct {
	log    logger.Logger
	client *resend.Client
	from   string
}

func NewEmail(log logger.Logger, apiKey, from string) Tool {
	if apiKey == "" {
		log.Debugf("Tool skipped: %s (RESEND_API_KEY not set)", emailName)
		return nil
	}
	if from == "" {
		log.Debugf("Tool skipped: %s (RESEND_FROM not set)", emailName)
		return nil
	}

	e := &Email{
		log:    log,
		client: resend.NewClient(apiKey),
		from:   from,
	}
	e.log.Debugf("Tool loaded: %s", emailName)

	return e
}

func (t *Email) Info() (string, string) {
	return emailName, emailDesc
}

func (t *Email) Definition() schemas.ChatTool {
	props := schemas.NewOrderedMapFromPairs(
		schemas.Pair{Key: "to", Value: map[string]string{"type": "string", "description": toDesc}},
		schemas.Pair{Key: "subject", Value: map[string]string{"type": "string", "description": "Email subject line"}},
		schemas.Pair{Key: "body", Value: map[string]string{"type": "string", "description": "Email body (plain text)"}},
	)

	return schemas.ChatTool{
		Type: schemas.ChatToolTypeFunction,
		Function: &schemas.ChatToolFunction{
			Name:        emailName,
			Description: schemas.Ptr(emailDesc),
			Parameters: &schemas.ToolFunctionParameters{
				Type:       "object",
				Properties: props,
				Required:   []string{"to", "subject", "body"},
			},
		},
	}
}

func (t *Email) Execute(ctx context.Context, raw json.RawMessage) (string, error) {
	var a emailArgs
	err := json.Unmarshal(raw, &a)
	if err != nil {
		return "", fmt.Errorf("parse email args: %w", err)
	}

	to, err := t.recipient(a.To)
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(a.Subject) == "" || strings.TrimSpace(a.Body) == "" {
		return "", fmt.Errorf("subject and body are both required: ask the user what the email should say rather than inventing it")
	}

	ctx, cancel := context.WithTimeout(ctx, emailTimeout)
	defer cancel()

	sent, err := t.client.Emails.SendWithContext(ctx, &resend.SendEmailRequest{
		From:    t.from,
		To:      []string{to},
		Subject: a.Subject,
		Text:    a.Body,
	})
	if err != nil {
		return "", fmt.Errorf("send email from %s to %s: %w", t.from, to, err)
	}

	t.log.Debugf("Email %s sent to %s", sent.Id, to)

	return fmt.Sprintf("Email sent to %s", to), nil
}

func (t *Email) recipient(to string) (string, error) {
	parsed, err := mail.ParseAddress(strings.TrimSpace(to))
	if err != nil {
		return "", fmt.Errorf("%q is not a valid email address: ask the user who the email should go to instead of guessing", to)
	}

	address := strings.ToLower(parsed.Address)
	if isReservedDomain(address) {
		return "", fmt.Errorf("%s is a placeholder address that cannot receive mail: ask the user for the real recipient", address)
	}

	return address, nil
}

func isReservedDomain(address string) bool {
	_, host, found := strings.Cut(address, "@")
	if !found {
		return false
	}

	return slices.ContainsFunc(reservedDomains, func(d string) bool {
		return host == d || strings.HasSuffix(host, "."+d)
	})
}
