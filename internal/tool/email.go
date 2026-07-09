package tool

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/J4NN0/mycel/internal/logger"
	"github.com/maximhq/bifrost/core/schemas"
	"github.com/resend/resend-go/v3"
)

const emailName = "send_email"

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

	e := &Email{
		log:    log,
		client: resend.NewClient(apiKey),
		from:   from,
	}
	e.log.Debugf("Tool loaded: %s", emailName)

	return e
}

func (t *Email) Definition() schemas.ChatTool {
	desc := schemas.Ptr("Send an email to a recipient")

	props := schemas.NewOrderedMapFromPairs(
		schemas.Pair{Key: "to", Value: map[string]string{"type": "string", "description": "Recipient email address"}},
		schemas.Pair{Key: "subject", Value: map[string]string{"type": "string", "description": "Email subject line"}},
		schemas.Pair{Key: "body", Value: map[string]string{"type": "string", "description": "Email body (plain text)"}},
	)

	return schemas.ChatTool{
		Type: schemas.ChatToolTypeFunction,
		Function: &schemas.ChatToolFunction{
			Name:        emailName,
			Description: desc,
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

	params := &resend.SendEmailRequest{
		From:    t.from,
		To:      []string{a.To},
		Subject: a.Subject,
		Text:    a.Body,
	}

	_, err = t.client.Emails.SendWithContext(ctx, params)
	if err != nil {
		return "", fmt.Errorf("send email: %w", err)
	}

	return fmt.Sprintf("Email sent to %s", a.To), nil
}
