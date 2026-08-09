package prompt

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	fileExtension = ".txt"
	purposeName = "purpose"
	goalName    = "goal"
	compactName = "compact"
	toolsName   = "tools"
)

const (
	personasDir = "personas"
)

type Manager struct {
	basePath string
	persona  string
}

func NewManager(basePath, persona string) *Manager {
	return &Manager{basePath: basePath, persona: persona}
}

func (m *Manager) LoadSystem() (string, error) {
	purpose, err := m.loadFile(purposeName)
	if err != nil {
		return "", err
	}

	persona, err := m.loadFile(filepath.Join(personasDir, m.persona))
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s\n\n## How you sound\n\n%s", strings.TrimSpace(purpose), strings.TrimSpace(persona)), nil
}

func (m *Manager) LoadGoal() (string, error) {
	return m.loadFile(goalName)
}

func (m *Manager) LoadCompact() (string, error) {
	return m.loadFile(compactName)
}

func (m *Manager) LoadTools() (string, error) {
	return m.loadFile(toolsName)
}

func (m *Manager) loadFile(name string) (string, error) {
	path := filepath.Join(m.basePath, name+fileExtension)
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to load prompt %q: %w", name, err)
	}
	return string(data), nil
}
