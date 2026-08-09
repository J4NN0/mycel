package cli

import (
	"context"
	"strings"

	"charm.land/bubbles/v2/list"
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

type mode int

const (
	modeChat mode = iota
	modeResume
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

// deltaMsg is a fragment of the reply currently being generated.
type deltaMsg string

type replyMsg struct {
	reply *agent.Reply
	err   error
}

// waitForDelta feeds the next streamed fragment into the update loop. The stream
// channel is unbuffered, so the reply cannot outrun what has been rendered.
func waitForDelta(stream chan string) tea.Cmd {
	return func() tea.Msg {
		delta, ok := <-stream
		if !ok {
			return nil
		}
		return deltaMsg(delta)
	}
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

	vp         viewport.Model
	input      textinput.Model
	lines      []string
	stream     chan string
	partial    string
	width      int
	ready      bool
	busy       bool
	eyesShut   bool
	mode       mode
	resumeList list.Model
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

	case deltaMsg:
		m.partial += string(msg)
		return m.sync(), waitForDelta(m.stream)

	case replyMsg:
		m.busy = false
		m.partial = ""
		m.mode = modeChat
		if msg.err != nil {
			return m.appendLine(errorStyle.Render("error: " + msg.err.Error())), nil
		}
		if len(msg.reply.Conversations) > 0 {
			m.resumeList = newResumeList(resumeItems(msg.reply.Conversations), m.vp.Width(), m.vp.Height())
			m.mode = modeResume
			return m, nil
		}
		return m.appendLine(speakerLine(nameStyle, m.name, msg.reply.Text)), nil

	case tea.KeyPressMsg:
		return m.handleKey(msg)
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m model) handleKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	if msg.String() == keyCtrlC {
		return m, tea.Quit
	}

	if m.mode == modeResume {
		return m.handleResumeKey(msg)
	}

	switch msg.String() {
	case keyEsc:
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

func (m model) handleResumeKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case keyEsc:
		m.mode = modeChat
		return m, nil
	case keyEnter:
		item, ok := m.resumeList.SelectedItem().(resumeItem)
		if !ok {
			return m, nil
		}
		m.busy = true
		ctx, reply := m.ctx, m.handler
		return m, func() tea.Msg {
			result, err := reply(ctx, sessionID, "/resume "+item.id, nil)
			return replyMsg{reply: result, err: err}
		}
	default:
		var cmd tea.Cmd
		m.resumeList, cmd = m.resumeList.Update(msg)
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
	m.stream = make(chan string)

	ctx, reply, stream := m.ctx, m.handler, m.stream
	return m, tea.Batch(
		func() tea.Msg {
			defer close(stream)
			result, err := reply(ctx, sessionID, text, func(delta string) {
				select {
				case stream <- delta:
				case <-ctx.Done():
				}
			})
			return replyMsg{reply: result, err: err}
		},
		waitForDelta(stream),
	)
}

func (m model) appendLine(s string) model {
	m.lines = append(m.lines, s)
	if len(m.lines) > maxLines {
		m.lines = m.lines[len(m.lines)-maxLines:]
	}
	return m.sync()
}

func (m model) sync() model {
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
	if m.mode == modeResume {
		m.resumeList.SetSize(w, max(1, h-footerHeight))
	}
	return m
}
