package prompt

import (
	"fmt"
	"os"
	"path/filepath"
)

const (
	personaDir    = "persona"
	fileExtension = ".txt"
)

type Manager struct {
	basePath string
}

func NewManager(basePath string) *Manager {
	return &Manager{basePath: basePath}
}

func (m *Manager) LoadPersona() (string, error) {
	path := filepath.Join(m.basePath, personaDir+fileExtension)
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to load prompt %q: %w", personaDir, err)
	}
	return string(data), nil
}
