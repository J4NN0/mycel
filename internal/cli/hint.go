package cli

import (
	"strings"

	"github.com/J4NN0/mycel/internal/agent"
)

func matchingCommands(input string) []string {
	text := strings.TrimSpace(input)
	if !strings.HasPrefix(text, "/") || strings.ContainsAny(text, " \t") {
		return nil
	}
	typed := strings.TrimPrefix(text, "/")

	var matches []string
	for _, cmd := range agent.Commands {
		if strings.HasPrefix(cmd.Name, typed) {
			matches = append(matches, cmd.Name)
		}
	}
	return matches
}

func completeCommand(input string) string {
	matches := matchingCommands(input)
	if len(matches) == 0 {
		return ""
	}
	return "/" + longestCommonPrefix(matches)
}

func longestCommonPrefix(items []string) string {
	if len(items) == 0 {
		return ""
	}
	prefix := items[0]
	for _, s := range items[1:] {
		for !strings.HasPrefix(s, prefix) {
			prefix = prefix[:len(prefix)-1]
		}
	}
	return prefix
}
