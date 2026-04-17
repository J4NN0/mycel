package llm

import (
	"context"
	"fmt"

	"github.com/maximhq/bifrost/core/schemas"
)

func (l *LLM) Chat(ctx context.Context, msg string) (string, error) {
	messages := []schemas.ChatMessage{
		{
			Role: schemas.ChatMessageRoleUser,
			Content: &schemas.ChatMessageContent{
				ContentStr: schemas.Ptr(msg),
			},
		},
	}

	response, err := l.bifrost.ChatCompletionRequest(schemas.NewBifrostContext(ctx, schemas.NoDeadline), &schemas.BifrostChatRequest{
		Provider: l.provider,
		Model:    "llama3.1:latest",
		Input:    messages,
	})
	if err != nil {
		return "", fmt.Errorf("chat completion request: %v", err)
	}

	return *response.Choices[0].Message.Content.ContentStr, nil
}
