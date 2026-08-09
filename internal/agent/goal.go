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

func (a *Agent) runGoal(ctx context.Context, goal string) {
	a.log.Printf("Starting goal loop: %s", goal)

	systemPrompt, err := a.promptManager.LoadSystem()
	if err != nil {
		a.log.Errorf("goal: failed to load system prompt: %v", err)
		return
	}

	goalPrompt, err := a.promptManager.LoadGoal()
	if err != nil {
		a.log.Errorf("goal: failed to load goal prompt: %v", err)
		return
	}

	messages := a.withToolPolicy([]llm.Message{
		{Role: schemas.ChatMessageRoleSystem, Content: systemPrompt},
		{Role: schemas.ChatMessageRoleSystem, Content: goalPrompt},
		{Role: schemas.ChatMessageRoleUser, Content: goal},
	})

	for step := range goalMaxSteps {
		select {
		case <-ctx.Done():
			a.log.Printf("Goal loop cancelled at step %d", step+1)
			return
		default:
		}

		a.log.Debugf("Goal step %d/%d ...", step+1, goalMaxSteps)

		result, err := a.provider.Chat(ctx, messages, nil, a.tools...)
		if err != nil {
			a.log.Errorf("goal step %d: %v", step+1, err)
			return
		}

		a.log.Printf("Goal step %d: %s", step+1, result.Content)

		messages = append(messages, llm.Message{Role: schemas.ChatMessageRoleAssistant, Content: result.Content})

		if strings.Contains(result.Content, goalComplete) {
			a.log.Printf("Goal complete after %d step(s)", step+1)
			return
		}

		messages = append(messages, llm.Message{Role: schemas.ChatMessageRoleUser, Content: goalStepPrompt})

		select {
		case <-ctx.Done():
			return
		case <-time.After(goalStepPause):
		}
	}

	a.log.Warningf("Goal loop reached max steps (%d) without completing", goalMaxSteps)
}
