package llm

import (
	"context"
	"fmt"

	"github.com/maximhq/bifrost/core/schemas"
)

type Message struct {
	Role    schemas.ChatMessageRole
	Content string
}

func (l *llm) Chat(ctx context.Context, messages []Message) (string, error) {
	chatMessages := make([]schemas.ChatMessage, len(messages))
	for i, m := range messages {
		chatMessages[i] = schemas.ChatMessage{
			Role: m.Role,
			Content: &schemas.ChatMessageContent{
				ContentStr: schemas.Ptr(m.Content),
			},
		}
	}

	response, err := l.bifrost.ChatCompletionRequest(schemas.NewBifrostContext(ctx, schemas.NoDeadline), &schemas.BifrostChatRequest{
		Provider: l.provider,
		Model:    l.model,
		Input:    chatMessages,
	})
	if err != nil {
		return "", fmt.Errorf("chat completion request: %v", err)
	}

	return *response.Choices[0].Message.Content.ContentStr, nil
}
