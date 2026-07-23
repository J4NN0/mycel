package prompt

import (
	"fmt"
	"os"
	"path/filepath"
)

const (
	fileExtension = ".txt"
	personasDir   = "personas"
	objectiveName = "objective"
	compactName   = "compact"
)

type Manager struct {
	basePath string
	persona  string
}

func NewManager(basePath, persona string) *Manager {
	return &Manager{basePath: basePath, persona: persona}
}

func (m *Manager) LoadPersona() (string, error) {
	return m.loadFile(filepath.Join(personasDir, m.persona))
}

func (m *Manager) LoadObjective() (string, error) {
	return m.loadFile(objectiveName)
}

func (m *Manager) LoadCompact() (string, error) {
	return m.loadFile(compactName)
}

func (m *Manager) loadFile(name string) (string, error) {
	path := filepath.Join(m.basePath, name+fileExtension)
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to load prompt %q: %w", name, err)
	}
	return string(data), nil
}
