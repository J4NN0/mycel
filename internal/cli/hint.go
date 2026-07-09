package cli

import (
	"fmt"
	"strings"

	"github.com/J4NN0/mycel/internal/agent"
	"github.com/chzyer/readline"
)

// commandCompleter lets the user press Tab after "/" to complete one of agent.Commands
func commandCompleter() *readline.PrefixCompleter {
	items := make([]readline.PrefixCompleterInterface, len(agent.Commands))
	for i, cmd := range agent.Commands {
		items[i] = readline.PcItem("/" + cmd.Name)
	}
	return readline.NewPrefixCompleter(items...)
}

// commandHintListener prints a dimmed hint of matching commands right after
// the cursor as the user types "/...", then moves the cursor back so the
// hint never becomes part of the actual input. It only ever writes on the
// current line (no newlines), so it can't desync with terminal scrolling the
// way a hint rendered on a line below the prompt could.
type commandHintListener struct{}

func (commandHintListener) OnChange(line []rune, pos int, _ rune) ([]rune, int, bool) {
	hint := commandHint(line, pos)
	if hint != "" {
		fmt.Printf("\033[2m%s\033[0m\033[%dD", hint, len([]rune(hint)))
	}
	return nil, 0, false
}

func commandHint(line []rune, pos int) string {
	if pos != len(line) {
		return ""
	}

	text := string(line)
	if !strings.HasPrefix(text, "/") {
		return ""
	}
	typed := text[1:]

	var matches []string
	for _, cmd := range agent.Commands {
		if strings.HasPrefix(cmd.Name, typed) {
			matches = append(matches, cmd.Name)
		}
	}
	if len(matches) == 0 {
		return ""
	}

	return fmt.Sprintf(" (%s)", strings.Join(matches, ", "))
}
