package ui

import (
	"fmt"
	"io"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// taskDelegate renders tasks with a priority color bar and compact layout.
type taskDelegate struct {
	list.DefaultDelegate
}

func newTaskDelegate() taskDelegate {
	d := list.NewDefaultDelegate()
	d.ShowDescription = false
	d.SetSpacing(0)
	return taskDelegate{DefaultDelegate: d}
}

func (d taskDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	if h, ok := item.(HeadingItem); ok {
		fmt.Fprintf(w, "\n%s", headingStyle.Render("── "+h.label+" ──"))
		return
	}

	t, ok := item.(TaskItem)
	if !ok {
		d.DefaultDelegate.Render(w, m, index, item)
		return
	}

	selected := index == m.Index()

	// Build the task name with optional overdue date.
	title := t.task.Name
	if isOverdue(t.task.Due) {
		title += " " + overdueStyle.Render(fmt.Sprintf("[%s]", t.task.Due.Format("Jan 2")))
	}

	// Priority color bar.
	barColor := priorityColor(t.task.Priority)
	bar := lipgloss.NewStyle().Foreground(barColor).Render("▎")

	// Apply selection styling.
	if selected {
		title = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#c0caf5")).
			Bold(true).
			Render(title)
		line := fmt.Sprintf("%s %s", bar, title)
		fmt.Fprint(w, lipgloss.NewStyle().
			Background(lipgloss.Color("#292e42")).
			Width(m.Width()).
			Render(line))
	} else {
		title = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#a9b1d6")).
			Render(title)
		fmt.Fprintf(w, "%s %s", bar, title)
	}
}

func (d taskDelegate) Height() int { return 1 }

func (d taskDelegate) Spacing() int { return 0 }

func (d taskDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd {
	// Skip heading items when navigating.
	if _, ok := m.SelectedItem().(HeadingItem); ok {
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
