package ui

import (
	"fmt"
	"io"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
)

// taskDelegate wraps the default delegate and renders HeadingItems differently.
type taskDelegate struct {
	list.DefaultDelegate
}

func newTaskDelegate() taskDelegate {
	return taskDelegate{DefaultDelegate: list.NewDefaultDelegate()}
}

func (d taskDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	if h, ok := item.(HeadingItem); ok {
		fmt.Fprintf(w, "\n%s", headingStyle.Render("── "+h.label+" ──"))
		return
	}
	d.DefaultDelegate.Render(w, m, index, item)
}

func (d taskDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd {
	// Skip heading items when navigating.
	if _, ok := m.SelectedItem().(HeadingItem); ok {
		// Determine direction from key press.
		if key, ok := msg.(tea.KeyMsg); ok {
			switch key.String() {
			case "up", "k":
				m.CursorUp()
			default:
				m.CursorDown()
			}
		}
	}
	return nil
}
