package cli

import (
	"slices"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

var (
	boxStyle   = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("63")).Padding(0, 1)
	hintStyle  = lipgloss.NewStyle().Faint(true)
	userStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("63")).Bold(true)
	nameStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Bold(true)
	errorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
)

func speakerLine(style lipgloss.Style, speaker, text string) string {
	return style.Render(speaker+": ") + text
}

func (m model) View() tea.View {
	content := "\n  Starting " + m.name + "…"
	if m.ready {
		content = lipgloss.JoinVertical(lipgloss.Left, m.mainView(), m.footerView())
	}
	v := tea.NewView(content)
	v.AltScreen = true
	return v
}

func (m model) mainView() string {
	if m.mode == modeResume {
		return m.resumeList.View()
	}
	return m.vp.View()
}

func (m model) contentView() string {
	lines := m.lines
	if m.partial != "" {
		lines = append(slices.Clip(m.lines), speakerLine(nameStyle, m.name, m.partial))
	}

	width := m.vp.Width()
	if width <= 0 {
		return strings.Join(lines, "\n")
	}

	wrapped := make([]string, len(lines))
	for i, line := range lines {
		wrapped[i] = lipgloss.Wrap(line, width, "")
	}
	return strings.Join(wrapped, "\n")
}

func (m model) boxWidth() int {
	if m.width > minAvatarWidth {
		return m.width - avatarWidth
	}
	return max(1, m.width)
}

func (m model) footerView() string {
	width := m.boxWidth()
	box := boxStyle.Width(width).Render(m.input.View())
	stack := lipgloss.JoinVertical(lipgloss.Left, box, m.hintView(width))
	if m.width <= minAvatarWidth {
		return stack
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, avatarStyle.Render(blinkView(m.eyesShut)), stack)
}

func (m model) hintView(width int) string {
	style := hintStyle.MaxWidth(max(1, width))
	if m.mode == modeResume {
		return style.Render("  ↑/↓ select · enter to resume · esc to cancel")
	}
	if m.busy {
		return style.Render("  " + m.name + " is thinking…")
	}
	if matches := matchingCommands(m.input.Value()); len(matches) > 0 {
		hints := make([]string, len(matches))
		for i, name := range matches {
			hints[i] = "/" + name
		}
		return style.Render("  " + strings.Join(hints, "  "))
	}
	return style.Render("  / for commands · " + keyCtrlC + " to quit")
}
