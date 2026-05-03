package agent

import (
	"fmt"
	"strings"
)

const (
	cmdStart = "start"
	cmdHelp  = "help"
	cmdClear = "clear"
)

type Command struct {
	Name        string
	Description string
}

var Commands = []Command{
	{Name: cmdStart, Description: "Start interacting with Mycel"},
	{Name: cmdHelp, Description: "Show available commands"},
	{Name: cmdClear, Description: "Reset the conversation history"},
}

func (a *Agent) handleCommand(sessionID, text string) (string, bool) {
	switch strings.TrimPrefix(text, "/") {
	case cmdStart:
		return "Hey, I'm Mycel. Drop me a message and let's talk.", true
	case cmdHelp:
		return buildHelp(), true
	case cmdClear:
		a.clearHistory(sessionID)
		return "Conversation history cleared.", true
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
