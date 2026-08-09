package agent

import (
	"context"
	"fmt"
	"strings"
)

const (
	cmdStart = "start"
	cmdHelp  = "help"
	cmdClear = "clear"
	cmdGoal  = "goal"
	cmdModel = "model"
)

type Command struct {
	Name        string
	Description string
}

var Commands = []Command{
	{Name: cmdStart, Description: "Start interacting with Mycel"},
	{Name: cmdHelp, Description: "Show available commands"},
	{Name: cmdClear, Description: "Start a new conversation"},
	{Name: cmdGoal, Description: "Give Mycel a goal to work toward autonomously"},
	{Name: cmdModel, Description: "Show which model is currently in use"},
}

func (a *Agent) handleCommand(ctx context.Context, sessionID, text string) (string, bool) {
	if !strings.HasPrefix(text, "/") {
		return "", false
	}

	name, arg, _ := strings.Cut(strings.TrimPrefix(strings.TrimSpace(text), "/"), " ")
	arg = strings.TrimSpace(arg)

	switch name {
	case cmdStart:
		return "Hey, I'm Mycel. Drop me a message and let's talk.", true
	case cmdHelp:
		return buildHelp(), true
	case cmdClear:
		err := a.startNewConversation(ctx, sessionID)
		if err != nil {
			a.log.Errorf("[%s] Failed to start new conversation: %v", sessionID, err)
			return "Failed to start a new conversation.", true
		}
		return "Started a new conversation.", true
	case cmdGoal:
		if arg == "" {
			return "Usage: /goal <what you want me to work toward>", true
		}
		go a.runGoal(ctx, arg)
		return "Goal accepted. Working on it…", true
	case cmdModel:
		return fmt.Sprintf("Currently running on %s", a.provider.Model()), true
	}

	return "", false
}

func buildHelp() string {
	var sb strings.Builder
	sb.WriteString("Available commands:\n")
	for _, c := range Commands {
		sb.WriteString(fmt.Sprintf("/%s — %s\n", c.Name, c.Description))
	}
	return sb.String()
}
