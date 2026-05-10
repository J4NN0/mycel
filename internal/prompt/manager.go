package prompt

import (
	"fmt"
	"os"
	"path/filepath"
)

const (
	fileExtension = ".txt"

	compactPrompt = "Summarize the following conversation concisely, preserving the key points and context."
)

type Manager struct {
	basePath string
	persona  string
}

func NewManager(basePath, persona string) *Manager {
	return &Manager{basePath: basePath, persona: persona}
}

func (m *Manager) LoadPersona() (string, error) {
	path := filepath.Join(m.basePath, m.persona+fileExtension)
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to load persona %q: %w", m.persona, err)
	}
	return string(data), nil
}

func (m *Manager) LoadCompact() string {
	return compactPrompt
}
