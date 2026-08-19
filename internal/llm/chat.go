package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/J4NN0/mycel/internal/tool"
	"github.com/maximhq/bifrost/core/schemas"
)

type Message struct {
	Role    schemas.ChatMessageRole
	Content string
	Images  []Image
}

type Response struct {
	Content          string
	PromptTokens     int
	CompletionTokens int
	TotalTokens      int
}

// StreamFunc receives the assistant content as it arrives. It may be nil, in
// which case the response is only returned once complete.
type StreamFunc func(delta string)

type streamRound struct {
	content   string
	toolCalls []schemas.ChatAssistantMessageToolCall
	usage     *schemas.BifrostLLMUsage
}

func (l *llm) Chat(ctx context.Context, messages []Message, onDelta StreamFunc, tools ...tool.Tool) (Response, error) {
	chatMessages, err := l.initMessages(messages)
	if err != nil {
		return Response{}, err
	}
	chatTools, toolMap := initTools(tools)
	params := initParams(chatTools)

	for {
		r, err := l.streamRequest(ctx, chatMessages, params, onDelta)
		if err != nil {
			return Response{}, err
		}

		if len(r.toolCalls) == 0 {
			if r.content == "" {
				return Response{}, fmt.Errorf("chat completion returned neither content nor tool calls")
			}
			return buildResponse(r.content, r.usage), nil
		}

		chatMessages = append(chatMessages, schemas.ChatMessage{
			Role:                 schemas.ChatMessageRoleAssistant,
			ChatAssistantMessage: &schemas.ChatAssistantMessage{ToolCalls: r.toolCalls},
		})
		chatMessages = append(chatMessages, l.runToolCalls(ctx, toolMap, r.toolCalls)...)
	}
}

func (l *llm) initMessages(messages []Message) ([]schemas.ChatMessage, error) {
	chatMessages := make([]schemas.ChatMessage, len(messages))
	for i, m := range messages {
		if len(m.Images) == 0 {
			chatMessages[i] = schemas.ChatMessage{
				Role: m.Role,
				Content: &schemas.ChatMessageContent{
					ContentStr: schemas.Ptr(m.Content),
				},
			}
			continue
		}

		if !l.supportsVision() {
			return nil, fmt.Errorf("%w: %s", ErrVisionUnsupported, l.Model())
		}

		blocks, err := contentBlocks(m)
		if err != nil {
			return nil, fmt.Errorf("build image content: %w", err)
		}
		chatMessages[i] = schemas.ChatMessage{
			Role: m.Role,
			Content: &schemas.ChatMessageContent{
				ContentBlocks: blocks,
			},
		}
	}
	return chatMessages, nil
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

func initParams(chatTools []schemas.ChatTool) *schemas.ChatParameters {
	if len(chatTools) == 0 {
		return nil
	}
	return &schemas.ChatParameters{
		Tools: chatTools,
		ToolChoice: &schemas.ChatToolChoice{
			ChatToolChoiceStr: schemas.Ptr("auto"),
		},
	}
}

func buildResponse(content string, usage *schemas.BifrostLLMUsage) Response {
	r := Response{Content: content}
	if usage != nil {
		r.PromptTokens = usage.PromptTokens
		r.CompletionTokens = usage.CompletionTokens
		r.TotalTokens = usage.TotalTokens
	}
	return r
}

func (l *llm) streamRequest(ctx context.Context, messages []schemas.ChatMessage, params *schemas.ChatParameters, onDelta StreamFunc) (streamRound, error) {
	streamCtx := schemas.NewBifrostContext(ctx, schemas.NoDeadline)
	streamCtx.SetValue(schemas.BifrostContextKeyStreamIdleTimeout, streamIdleTimeout)

	stream, err := l.bifrost.ChatCompletionStreamRequest(streamCtx, &schemas.BifrostChatRequest{
		Provider: l.provider,
		Model:    l.model,
		Input:    messages,
		Params:   params,
	})
	if err != nil {
		return streamRound{}, fmt.Errorf("chat completion stream request: %v", err)
	}

	return collectStreamRound(stream, onDelta)
}

// collectStreamRound assembles the chunks of a stream, handing content to onDelta as it arrives.
func collectStreamRound(stream chan *schemas.BifrostStreamChunk, onDelta StreamFunc) (streamRound, error) {
	var (
		r       streamRound
		content strings.Builder
	)

	for chunk := range stream {
		if chunk.BifrostError != nil {
			return streamRound{}, fmt.Errorf("chat completion stream: %v", chunk.BifrostError)
		}
		if chunk.BifrostChatResponse == nil {
			continue
		}
		// Usage come to its own chunk, at the end of the stream
		if chunk.BifrostChatResponse.Usage != nil {
			r.usage = chunk.BifrostChatResponse.Usage
		}

		delta := streamDelta(chunk.BifrostChatResponse)
		if delta == nil {
			continue
		}
		if delta.Content != nil {
			content.WriteString(*delta.Content)
			if onDelta != nil {
				onDelta(*delta.Content)
			}
		}
		r.toolCalls = append(r.toolCalls, delta.ToolCalls...)
	}

	r.content = content.String()

	return r, nil
}

// streamDelta returns the partial message a chunk carries, if it carries one.
func streamDelta(response *schemas.BifrostChatResponse) *schemas.ChatStreamResponseChoiceDelta {
	if len(response.Choices) == 0 || response.Choices[0].ChatStreamResponseChoice == nil {
		return nil
	}
	return response.Choices[0].Delta
}

func (l *llm) runToolCalls(ctx context.Context, toolMap map[string]tool.Tool, toolCalls []schemas.ChatAssistantMessageToolCall) []schemas.ChatMessage {
	messages := make([]schemas.ChatMessage, 0, len(toolCalls))

	for _, tc := range toolCalls {
		result, err := executeToolCall(ctx, toolMap, tc)
		if err != nil {
			result = fmt.Sprintf("tool call %s failed: %v", *tc.ID, err)
			l.log.Errorf("%s", result)
		}

		messages = append(messages, schemas.ChatMessage{
			Role:            schemas.ChatMessageRoleTool,
			Content:         &schemas.ChatMessageContent{ContentStr: schemas.Ptr(result)},
			ChatToolMessage: &schemas.ChatToolMessage{ToolCallID: tc.ID},
		})
	}

	return messages
}

func executeToolCall(ctx context.Context, toolMap map[string]tool.Tool, tc schemas.ChatAssistantMessageToolCall) (string, error) {
	t, ok := toolMap[*tc.Function.Name]
	if !ok {
		return "", fmt.Errorf("unknown tool: %s", *tc.Function.Name)
	}
	return t.Execute(ctx, json.RawMessage(tc.Function.Arguments))
}
