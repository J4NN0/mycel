package cli

import (
	"fmt"
	"math/rand/v2"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

const (
	avatarTemplate = " .-o-00-o-.\n" +
		"(__________)\n" +
		"   |%s|\n" +
		"   |____|"

	avatarEyesOpen = "●  ●"
	avatarEyesShut = "-  -"
)

const (
	blinkShutFor    = 150 * time.Millisecond
	blinkMinOpenFor = 2 * time.Second
	blinkMaxOpenFor = 5 * time.Second
)

var (
	avatarStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("63")).MarginRight(1)

	avatarWidth    = lipgloss.Width(avatarStyle.Render(avatarView(avatarEyesOpen)))
	minAvatarWidth = avatarWidth + 24
)

type blinkMsg struct{}

func avatarView(eyes string) string {
	return fmt.Sprintf(avatarTemplate, eyes)
}

func blinkView(eyesShut bool) string {
	if eyesShut {
		return avatarView(avatarEyesShut)
	}
	return avatarView(avatarEyesOpen)
}

func blinkCmd(d time.Duration) tea.Cmd {
	return tea.Tick(d, func(time.Time) tea.Msg { return blinkMsg{} })
}

func openEyesCmd() tea.Cmd {
	return blinkCmd(blinkMinOpenFor + rand.N(blinkMaxOpenFor-blinkMinOpenFor))
}
