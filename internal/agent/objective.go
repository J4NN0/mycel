package agent

import (
	"context"
	"strings"
	"time"

	"github.com/J4NN0/mycel/internal/llm"
	"github.com/maximhq/bifrost/core/schemas"
)

const (
	objectiveComplete   = "OBJECTIVE_COMPLETE"
	objectiveMaxSteps   = 20
	objectiveStepPause  = 1 * time.Minute
	objectiveStepPrompt = "Continue working toward the objective. What is the next step?"
)

func (a *Agent) runObjective(ctx context.Context, objective string) {
	a.log.Printf("Starting objective loop: %s", objective)

	persona, err := a.promptManager.LoadPersona()
	if err != nil {
		a.log.Errorf("objective: failed to load persona: %v", err)
		return
	}

	objectivePrompt, err := a.promptManager.LoadObjective()
	if err != nil {
		a.log.Errorf("objective: failed to load objective prompt: %v", err)
		return
	}

	messages := []llm.Message{
		{Role: schemas.ChatMessageRoleSystem, Content: persona},
		{Role: schemas.ChatMessageRoleSystem, Content: objectivePrompt},
		{Role: schemas.ChatMessageRoleUser, Content: objective},
	}

	for step := range objectiveMaxSteps {
		select {
		case <-ctx.Done():
			a.log.Printf("Objective loop cancelled at step %d", step+1)
			return
		default:
		}

		a.log.Debugf("Objective step %d/%d ...", step+1, objectiveMaxSteps)

		result, err := a.provider.Chat(ctx, messages, a.tools...)
		if err != nil {
			a.log.Errorf("objective step %d: %v", step+1, err)
			return
		}

		a.log.Printf("Objective step %d: %s", step+1, result.Content)

		messages = append(messages, llm.Message{Role: schemas.ChatMessageRoleAssistant, Content: result.Content})

		if strings.Contains(result.Content, objectiveComplete) {
			a.log.Printf("Objective complete after %d step(s)", step+1)
			return
		}

		messages = append(messages, llm.Message{Role: schemas.ChatMessageRoleUser, Content: objectiveStepPrompt})

		select {
		case <-ctx.Done():
			return
		case <-time.After(objectiveStepPause):
		}
	}

	a.log.Warningf("Objective loop reached max steps (%d) without completing", objectiveMaxSteps)
}
