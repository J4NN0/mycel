package agent

import (
	"time"

	"github.com/J4NN0/mycel/internal/llm"
	"github.com/J4NN0/mycel/internal/prompt"
	"github.com/maximhq/bifrost/core/schemas"
)

func (a *Agent) withContext(messages []llm.Message) []llm.Message {
	var injected []llm.Message

	currentTime, err := a.promptManager.LoadTime(time.Now())
	if err != nil {
		a.log.Errorf("Failed to load time prompt: %v", err)
	} else {
		injected = append(injected, llm.Message{Role: schemas.ChatMessageRoleSystem, Content: currentTime})
	}

	if len(a.tools) > 0 {
		policy, err := a.promptManager.LoadToolPolicy(a.toolInfo())
		if err != nil {
			a.log.Errorf("Failed to load tool prompt: %v", err)
		} else {
			injected = append(injected, llm.Message{Role: schemas.ChatMessageRoleSystem, Content: policy})
		}
	}

	return spliceAfterSystem(messages, injected)
}

func (a *Agent) toolInfo() []prompt.ToolInfo {
	info := make([]prompt.ToolInfo, 0, len(a.tools))

	for _, t := range a.tools {
		name, description := t.Info()
		info = append(info, prompt.ToolInfo{Name: name, Description: description})
	}

	return info
}

func spliceAfterSystem(messages, injected []llm.Message) []llm.Message {
	if len(injected) == 0 {
		return messages
	}
	if len(messages) == 0 {
		return injected
	}

	spliced := make([]llm.Message, 0, len(messages)+len(injected))
	spliced = append(spliced, messages[0])
	spliced = append(spliced, injected...)

	return append(spliced, messages[1:]...)
}
