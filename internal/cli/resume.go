package cli

import (
	"charm.land/bubbles/v2/list"

	"github.com/J4NN0/mycel/internal/agent"
)

type resumeItem struct {
	id      string
	preview string
}

func (r resumeItem) Title() string       { return r.preview }
func (r resumeItem) Description() string { return "" }
func (r resumeItem) FilterValue() string { return r.preview }

func resumeItems(conversations []agent.Conversation) []list.Item {
	items := make([]list.Item, len(conversations))
	for i, c := range conversations {
		items[i] = resumeItem{id: c.ID, preview: c.Preview}
	}
	return items
}

// newResumeList builds a compact, single-column picker: no filtering, status bar, help, or title —
// just the list of past conversations, navigable with the arrow keys.
func newResumeList(items []list.Item, width, height int) list.Model {
	delegate := list.NewDefaultDelegate()
	delegate.ShowDescription = false

	l := list.New(items, delegate, width, height)
	l.SetFilteringEnabled(false)
	l.SetShowStatusBar(false)
	l.SetShowHelp(false)
	l.SetShowTitle(false)

	return l
}
