package prompt

import (
	"fmt"
	"os"
	"path/filepath"
)

const (
	fileExtension = ".txt"
)

const (
	compactPrompt   = "Summarize the following conversation concisely, preserving the key points and context."
	objectivePrompt = "Work toward the objective provided below step by step. When the objective is fully complete, include OBJECTIVE_COMPLETE in your response."
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

func (m *Manager) LoadObjective() string {
	return objectivePrompt
}
