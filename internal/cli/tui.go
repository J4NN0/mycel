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
	userLabel    = "You"
	welcomeLine  = "Terminal chat ready. Type a message and press Enter. Type / to see available commands."
)

const (
	keyCtrlC  = "ctrl+c"
	keyEsc    = "esc"
	keyEnter  = "enter"
	keyTab    = "tab"
	keyPgUp   = "pgup"
	keyPgDown = "pgdown"
	keyCtrlU  = "ctrl+u"
	keyCtrlD  = "ctrl+d"
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

	vp       viewport.Model
	input    textinput.Model
	lines    []string
	width    int
	ready    bool
	busy     bool
	eyesShut bool
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
	return tea.Batch(textinput.Blink, openEyesCmd())
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m = m.resize(msg.Width, msg.Height)
		m.ready = true
		return m, nil

	case blinkMsg:
		m.eyesShut = !m.eyesShut
		if m.eyesShut {
			return m, blinkCmd(blinkShutFor)
		}
		return m, openEyesCmd()

	case logLineMsg:
		return m.appendLine(string(msg)), nil

	case replyMsg:
		m.busy = false
		if msg.err != nil {
			m = m.appendLine(errorStyle.Render("error: " + msg.err.Error()))
		} else {
			m = m.appendLine(speakerLine(nameStyle, m.name, msg.response))
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

	m = m.appendLine(speakerLine(userStyle, userLabel, text))
	m.input.Reset()
	m.busy = true

	ctx, handler := m.ctx, m.handler
	return m, func() tea.Msg {
		response, err := handler(ctx, sessionID, text)
		return replyMsg{response: response, err: err}
	}
}

func (m model) appendLine(s string) model {
	m.lines = append(m.lines, s)
	if len(m.lines) > maxLines {
		m.lines = m.lines[len(m.lines)-maxLines:]
	}
	if !m.ready {
		return m
	}
	atBottom := m.vp.AtBottom()
	m.vp.SetContent(m.contentView())
	if atBottom {
		m.vp.GotoBottom()
	}
	return m
}

func (m model) resize(w, h int) model {
	m.width = w
	m.input.SetWidth(max(1, m.boxWidth()-6))

	footerHeight := lipgloss.Height(m.footerView())
	m.vp.SetWidth(w)
	m.vp.SetHeight(max(1, h-footerHeight))
	m.vp.SetContent(m.contentView())
	m.vp.GotoBottom()
	return m
}
