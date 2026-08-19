package agent

import (
	"context"
	"strings"
	"time"

	"github.com/J4NN0/mycel/internal/llm"
	"github.com/maximhq/bifrost/core/schemas"
)

const (
	goalComplete   = "GOAL_COMPLETE"
	goalMaxSteps   = 20
	goalStepPause  = 1 * time.Minute
	goalStepPrompt = "Continue working toward the goal. What is the next step?"
)

func (a *Agent) runGoal(ctx context.Context, source, goal string) {
	a.log.Printf("[%s] Starting goal loop: %s", source, goal)

	systemPrompt, err := a.promptManager.LoadSystem()
	if err != nil {
		a.log.Errorf("[%s] goal: failed to load system prompt: %v", source, err)
		return
	}

	goalPrompt, err := a.promptManager.LoadGoal()
	if err != nil {
		a.log.Errorf("[%s] goal: failed to load goal prompt: %v", source, err)
		return
	}

	messages := []llm.Message{
		{Role: schemas.ChatMessageRoleSystem, Content: systemPrompt},
		{Role: schemas.ChatMessageRoleSystem, Content: goalPrompt},
		{Role: schemas.ChatMessageRoleUser, Content: goal},
	}

	for step := range goalMaxSteps {
		select {
		case <-ctx.Done():
			a.log.Printf("[%s] Goal loop cancelled at step %d", source, step+1)
			return
		default:
		}

		a.log.Debugf("[%s] Goal step %d/%d ...", source, step+1, goalMaxSteps)

		result, err := a.provider.Chat(ctx, a.withContext(messages), nil, a.tools...)
		if err != nil {
			a.log.Errorf("[%s] goal step %d: %v", source, step+1, err)
			return
		}

		a.log.Printf("[%s] Goal step %d: %s", source, step+1, result.Content)

		messages = append(messages, llm.Message{Role: schemas.ChatMessageRoleAssistant, Content: result.Content})

		if strings.Contains(result.Content, goalComplete) {
			a.log.Printf("[%s] Goal complete after %d step(s)", source, step+1)
			return
		}

		messages = append(messages, llm.Message{Role: schemas.ChatMessageRoleUser, Content: goalStepPrompt})

		select {
		case <-ctx.Done():
			return
		case <-time.After(goalStepPause):
		}
	}

	a.log.Warningf("[%s] Goal loop reached max steps (%d) without completing", source, goalMaxSteps)
}
