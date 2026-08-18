package prompt

import (
	"embed"
	"fmt"
	"io/fs"
	"path"
	"strings"
	"text/template"
	"time"
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
	timeName      = "time"
)

const (
	personasDir = "personas"
)

const (
	humanTimeLayout = "Monday, 2 January 2006 at 15:04 MST"
	isoTimeLayout   = time.RFC3339
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

func (m *Manager) LoadTime(now time.Time) (string, error) {
	return m.render(timeName, struct{ Human, ISO string }{
		Human: now.Format(humanTimeLayout),
		ISO:   now.Format(isoTimeLayout),
	})
}

type ToolInfo struct {
	Name        string
	Description string
}

func (m *Manager) LoadToolPolicy(tools []ToolInfo) (string, error) {
	return m.render(toolsName, struct{ Tools []ToolInfo }{Tools: tools})
}

func (m *Manager) render(name string, data any) (string, error) {
	raw, err := m.loadFile(name)
	if err != nil {
		return "", err
	}

	tmpl, err := template.New(name).Parse(raw)
	if err != nil {
		return "", fmt.Errorf("failed to parse prompt %q: %w", name, err)
	}

	var rendered strings.Builder
	err = tmpl.Execute(&rendered, data)
	if err != nil {
		return "", fmt.Errorf("failed to render prompt %q: %w", name, err)
	}

	return strings.TrimSpace(rendered.String()), nil
}

func (m *Manager) loadFile(name string) (string, error) {
	data, err := fs.ReadFile(m.fs, name+fileExtension)
	if err != nil {
		return "", fmt.Errorf("failed to load prompt %q: %w", name, err)
	}
	return string(data), nil
}
