package agent

import (
	"context"
	"fmt"
	"strings"
)

const (
	cmdStart  = "start"
	cmdHelp   = "help"
	cmdClear  = "clear"
	cmdResume = "resume"
	cmdGoal   = "goal"
	cmdModel  = "model"
)

const (
	maxResumeChoices = 10
	previewMaxLen    = 60
)

type Command struct {
	Name        string
	Description string
}

var Commands = []Command{
	{Name: cmdStart, Description: "Start interacting with Mycel"},
	{Name: cmdHelp, Description: "Show available commands"},
	{Name: cmdClear, Description: "Start a new conversation"},
	{Name: cmdResume, Description: "Resume a past conversation"},
	{Name: cmdGoal, Description: "Give Mycel a goal to work toward autonomously"},
	{Name: cmdModel, Description: "Show which model is currently in use"},
}

func (a *Agent) handleCommand(ctx context.Context, sessionID, text string) (reply *Reply, err error) {
	name, arg, isCmd := parseCommand(text)
	if !isCmd {
		return nil, nil
	}

	switch name {
	case cmdStart:
		return &Reply{Text: "Hey, I'm Mycel. Drop me a message and let's talk."}, nil
	case cmdHelp:
		return &Reply{Text: handleHelp()}, nil
	case cmdClear:
		err := a.startNewConversation(ctx, sessionID)
		if err != nil {
			a.log.Errorf("[%s] Failed to start new conversation: %v", sessionID, err)
			return &Reply{Text: "Failed to start a new conversation."}, nil
		}
		return &Reply{Text: "Started a new conversation."}, nil
	case cmdResume:
		return a.handleResume(ctx, sessionID, arg)
	case cmdGoal:
		if arg == "" {
			return &Reply{Text: "Usage: /goal <what you want me to work toward>"}, nil
		}
		go a.runGoal(ctx, arg)
		return &Reply{Text: "Goal accepted. Working on it…"}, nil
	case cmdModel:
		return &Reply{Text: fmt.Sprintf("Currently running on %s", a.provider.Model())}, nil
	}

	return nil, nil
}

func parseCommand(text string) (name, arg string, ok bool) {
	if !strings.HasPrefix(text, "/") {
		return "", "", false
	}

	name, arg, _ = strings.Cut(strings.TrimPrefix(strings.TrimSpace(text), "/"), " ")

	return name, strings.TrimSpace(arg), true
}

func handleHelp() string {
	var sb strings.Builder
	sb.WriteString("Available commands:\n")
	for _, c := range Commands {
		sb.WriteString(fmt.Sprintf("/%s — %s\n", c.Name, c.Description))
	}
	return sb.String()
}

// Conversation is a past conversation the user can resume, with a short preview of how it started.
type Conversation struct {
	ID      string
	Preview string
}

func (a *Agent) handleResume(ctx context.Context, sessionID, arg string) (*Reply, error) {
	if arg == "" {
		conversations, err := a.listConversations(ctx, sessionID)
		if err != nil {
			return nil, err
		}
		if len(conversations) == 0 {
			return &Reply{Text: "No past conversations to resume yet."}, nil
		}
		return &Reply{Conversations: conversations}, nil
	}

	err := a.history.SetActiveConversation(ctx, sessionID, arg)
	if err != nil {
		a.log.Errorf("[%s] Failed to resume conversation %s: %v", sessionID, arg, err)
		return &Reply{Text: "Failed to resume that conversation."}, nil
	}
	return &Reply{Text: "Resumed previous conversation."}, nil
}

func (a *Agent) listConversations(ctx context.Context, sessionID string) ([]Conversation, error) {
	activeID, err := a.history.ActiveConversation(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	summaries, err := a.history.ListConversations(ctx, sessionID, activeID, maxResumeChoices)
	if err != nil {
		return nil, err
	}

	conversations := make([]Conversation, len(summaries))
	for i, s := range summaries {
		conversations[i] = Conversation{ID: s.ID, Preview: truncatePreview(s.Preview)}
	}

	return conversations, nil
}

func truncatePreview(s string) string {
	s = strings.Join(strings.Fields(s), " ")
	if len([]rune(s)) <= previewMaxLen {
		return s
	}
	return string([]rune(s)[:previewMaxLen]) + "…"
}
