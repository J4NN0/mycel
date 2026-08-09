package prompt

import (
	"embed"
	"fmt"
	"io/fs"
	"path"
	"strings"
)

//go:embed data
var embedded embed.FS

const dataDir = "data"

const (
	fileExtension = ".txt"
	purposeName   = "purpose"
	goalName      = "goal"
	compactName   = "compact"
	toolsName     = "tools"
)

const (
	personasDir = "personas"
)

type Manager struct {
	fs      fs.FS
	persona string
}

func NewManager(persona string) *Manager {
	data, err := fs.Sub(embedded, dataDir)
	if err != nil {
		panic(fmt.Sprintf("prompt: embedded data missing: %v", err))
	}
	return &Manager{fs: data, persona: persona}
}

func (m *Manager) LoadSystem() (string, error) {
	purpose, err := m.loadFile(purposeName)
	if err != nil {
		return "", err
	}

	persona, err := m.loadFile(path.Join(personasDir, m.persona))
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
	data, err := fs.ReadFile(m.fs, name+fileExtension)
	if err != nil {
		return "", fmt.Errorf("failed to load prompt %q: %w", name, err)
	}
	return string(data), nil
}
