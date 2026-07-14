package cli

import (
	"context"
	"strings"

	"charm.land/bubbles/v2/textinput"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/J4NN0/mycel/internal/agent"
)

const (
	sessionID    = "terminal"
	maxLines     = 5000
	promptSymbol = "> "
	welcomeLine  = "Terminal chat ready. Type a message and press Enter. Type / to see available commands."

	keyCtrlC  = "ctrl+c"
	keyEsc    = "esc"
	keyEnter  = "enter"
	keyTab    = "tab"
	keyPgUp   = "pgup"
	keyPgDown = "pgdown"
	keyCtrlU  = "ctrl+u"
	keyCtrlD  = "ctrl+d"
)

var (
	boxStyle    = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("63")).Padding(0, 1)
	hintStyle   = lipgloss.NewStyle().Faint(true)
	promptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("63")).Bold(true)
	nameStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Bold(true)
	errorStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
)

type logLineMsg string

type replyMsg struct {
	response string
	err      error
}

type logWriter struct {
	send func(tea.Msg)
}

func (w logWriter) Write(p []byte) (int, error) {
	for _, line := range strings.Split(strings.TrimRight(string(p), "\n"), "\n") {
		w.send(logLineMsg(line))
	}
	return len(p), nil
}

type model struct {
	ctx     context.Context
	name    string
	handler agent.MessageHandler

	vp    viewport.Model
	input textinput.Model
	lines []string
	width int
	ready bool
	busy  bool
}

func newModel(ctx context.Context, name string, handler agent.MessageHandler) model {
	ti := textinput.New()
	ti.Prompt = promptSymbol
	ti.Placeholder = "Type a message…"
	ti.SetVirtualCursor(true)
	ti.Focus()

	return model{
		ctx:     ctx,
		name:    name,
		handler: handler,
		vp:      viewport.New(),
		input:   ti,
		lines:   []string{welcomeLine},
	}
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.resize(msg.Width, msg.Height)
		m.ready = true
		return m, nil

	case logLineMsg:
		m.appendLine(string(msg))
		return m, nil

	case replyMsg:
		m.busy = false
		if msg.err != nil {
			m.appendLine(errorStyle.Render("error: " + msg.err.Error()))
		} else {
			m.appendLine(nameStyle.Render(m.name+": ") + msg.response)
		}
		return m, nil

	case tea.KeyPressMsg:
		return m.handleKey(msg)
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m model) handleKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case keyCtrlC, keyEsc:
		return m, tea.Quit
	case keyEnter:
		return m.submit()
	case keyTab:
		if c := completeCommand(m.input.Value()); c != "" {
			m.input.SetValue(c)
			m.input.CursorEnd()
		}
		return m, nil
	case keyPgUp, keyPgDown, keyCtrlU, keyCtrlD:
		var cmd tea.Cmd
		m.vp, cmd = m.vp.Update(msg)
		return m, cmd
	default:
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		return m, cmd
	}
}

func (m model) submit() (tea.Model, tea.Cmd) {
	text := strings.TrimSpace(m.input.Value())
	if text == "" || m.busy {
		return m, nil
	}

	m.appendLine(promptStyle.Render(promptSymbol) + text)
	m.input.Reset()
	m.busy = true

	ctx, handler := m.ctx, m.handler
	return m, func() tea.Msg {
		response, err := handler(ctx, sessionID, text)
		return replyMsg{response: response, err: err}
	}
}

func (m *model) appendLine(s string) {
	m.lines = append(m.lines, s)
	if len(m.lines) > maxLines {
		m.lines = m.lines[len(m.lines)-maxLines:]
	}
	if !m.ready {
		return
	}
	atBottom := m.vp.AtBottom()
	m.vp.SetContent(strings.Join(m.lines, "\n"))
	if atBottom {
		m.vp.GotoBottom()
	}
}

func (m *model) resize(w, h int) {
	m.width = w
	m.input.SetWidth(max(1, w-6))

	footerHeight := lipgloss.Height(m.footerView())
	m.vp.SetWidth(w)
	m.vp.SetHeight(max(1, h-footerHeight))
	m.vp.SetContent(strings.Join(m.lines, "\n"))
	m.vp.GotoBottom()
}

func (m model) footerView() string {
	box := boxStyle.Width(max(1, m.width)).Render(m.input.View())
	return lipgloss.JoinVertical(lipgloss.Left, box, m.hintView())
}

func (m model) hintView() string {
	if m.busy {
		return hintStyle.Render("  " + m.name + " is thinking…")
	}
	if matches := matchingCommands(m.input.Value()); len(matches) > 0 {
		hints := make([]string, len(matches))
		for i, name := range matches {
			hints[i] = "/" + name
		}
		return hintStyle.Render("  " + strings.Join(hints, "  "))
	}
	return hintStyle.Render("  / for commands · " + keyCtrlC + " to quit")
}

func (m model) View() tea.View {
	content := "\n  Starting " + m.name + "…"
	if m.ready {
		content = lipgloss.JoinVertical(lipgloss.Left, m.vp.View(), m.footerView())
	}
	v := tea.NewView(content)
	v.AltScreen = true
	return v
}
