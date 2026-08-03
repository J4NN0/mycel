package llm

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/J4NN0/mycel/internal/tool"
	"github.com/maximhq/bifrost/core/schemas"
)

type Message struct {
	Role    schemas.ChatMessageRole
	Content string
}

type Response struct {
	Content          string
	PromptTokens     int
	CompletionTokens int
	TotalTokens      int
}

func (l *llm) Chat(ctx context.Context, messages []Message, tools ...tool.Tool) (Response, error) {
	chatMessages := initMessages(messages)
	chatTools, toolMap := initTools(tools)

	var params *schemas.ChatParameters
	if len(chatTools) > 0 {
		params = &schemas.ChatParameters{
			Tools: chatTools,
			ToolChoice: &schemas.ChatToolChoice{
				ChatToolChoiceStr: schemas.Ptr("auto"),
			},
		}
	}

	for {
		response, err := l.bifrost.ChatCompletionRequest(schemas.NewBifrostContext(ctx, schemas.NoDeadline), &schemas.BifrostChatRequest{
			Provider: l.provider,
			Model:    l.model,
			Input:    chatMessages,
			Params:   params,
		})
		if err != nil {
			return Response{}, fmt.Errorf("chat completion request: %v", err)
		}
		if len(response.Choices) == 0 {
			return Response{}, fmt.Errorf("chat completion returned no choices")
		}

		choice := response.Choices[0]
		if choice.ChatNonStreamResponseChoice == nil || choice.Message == nil {
			return Response{}, fmt.Errorf("chat completion returned no assistant message")
		}

		var toolCalls []schemas.ChatAssistantMessageToolCall
		if choice.Message.ChatAssistantMessage != nil {
			toolCalls = choice.Message.ChatAssistantMessage.ToolCalls
		}
		if len(toolCalls) == 0 {
			if choice.Message.Content == nil || choice.Message.Content.ContentStr == nil {
				return Response{}, fmt.Errorf("chat completion returned neither content nor tool calls")
			}
			return newResponse(*choice.Message.Content.ContentStr, response), nil
		}

		chatMessages = append(chatMessages, schemas.ChatMessage{
			Role:                 schemas.ChatMessageRoleAssistant,
			ChatAssistantMessage: &schemas.ChatAssistantMessage{ToolCalls: toolCalls},
		})

		for _, tc := range toolCalls {
			result, execErr := executeToolCall(ctx, toolMap, tc)

			chatMsgResult := result
			if execErr != nil {
				errMsg := fmt.Sprintf("tool call %s failed: %v", *tc.ID, execErr)
				chatMsgResult = errMsg
				l.log.Errorf("%s", errMsg)
			}

			chatMessages = append(chatMessages, schemas.ChatMessage{
				Role:            schemas.ChatMessageRoleTool,
				Content:         &schemas.ChatMessageContent{ContentStr: schemas.Ptr(chatMsgResult)},
				ChatToolMessage: &schemas.ChatToolMessage{ToolCallID: tc.ID},
			})
		}
	}
}

func newResponse(content string, response *schemas.BifrostChatResponse) Response {
	r := Response{Content: content}
	if response != nil && response.Usage != nil {
		r.PromptTokens = response.Usage.PromptTokens
		r.CompletionTokens = response.Usage.CompletionTokens
		r.TotalTokens = response.Usage.TotalTokens
	}
	return r
}

func initMessages(messages []Message) []schemas.ChatMessage {
	chatMessages := make([]schemas.ChatMessage, len(messages))
	for i, m := range messages {
		chatMessages[i] = schemas.ChatMessage{
			Role: m.Role,
			Content: &schemas.ChatMessageContent{
				ContentStr: schemas.Ptr(m.Content),
			},
		}
	}
	return chatMessages
}

func initTools(tools []tool.Tool) ([]schemas.ChatTool, map[string]tool.Tool) {
	var chatTools []schemas.ChatTool
	toolMap := make(map[string]tool.Tool, len(tools))

	for _, t := range tools {
		def := t.Definition()
		chatTools = append(chatTools, def)
		toolMap[def.Function.Name] = t
	}

	return chatTools, toolMap
}

func executeToolCall(ctx context.Context, toolMap map[string]tool.Tool, tc schemas.ChatAssistantMessageToolCall) (string, error) {
	t, ok := toolMap[*tc.Function.Name]
	if !ok {
		return "", fmt.Errorf("unknown tool: %s", *tc.Function.Name)
	}
	return t.Execute(ctx, json.RawMessage(tc.Function.Arguments))
}
