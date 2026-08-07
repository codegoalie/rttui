package ui

import (
	"fmt"
	"io"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"git.codegoalie.com/rttui/internal/rtm"
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
		fmt.Fprint(w, headingStyle.Render("── "+h.label+" ──"))
		return
	}

	t, ok := item.(TaskItem)
	if !ok {
		d.DefaultDelegate.Render(w, m, index, item)
		return
	}

	fmt.Fprint(w, renderTaskLine(t.task, index == m.Index(), m.Width(), false, true))
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

// renderTaskLine renders a single task line: priority bar + title, with an
// optional strikethrough and an optional overdue badge, applying a full-width
// background highlight when selected. Shared by taskDelegate and subtaskDelegate.
func renderTaskLine(t rtm.Task, selected bool, width int, strikethrough bool, showOverdueBadge bool) string {
	barColor := priorityColor(t.Priority)
	bar := lipgloss.NewStyle().Foreground(barColor).Render("▎")

	titleStyle := lipgloss.NewStyle()
	if selected {
		titleStyle = titleStyle.Foreground(lipgloss.Color("#c0caf5")).Bold(true)
	} else {
		titleStyle = titleStyle.Foreground(lipgloss.Color("#a9b1d6"))
	}
	if t.URL != "" {
		titleStyle = titleStyle.Hyperlink(t.URL).Underline(true)
	}
	if strikethrough {
		titleStyle = titleStyle.Strikethrough(true)
	}

	// Style the name first; append the overdue badge afterward so its own
	// ANSI codes aren't re-wrapped (and corrupted) by titleStyle's rune-by-rune
	// underline pass.
	title := titleStyle.Render(t.Name)
	if showOverdueBadge && isOverdue(t.Due) {
		title += " " + overdueStyle.Render(fmt.Sprintf("[%s]", t.Due.Format("Jan 2")))
	}

	if selected {
		line := fmt.Sprintf("%s %s", bar, title)
		return lipgloss.NewStyle().
			Background(lipgloss.Color("#292e42")).
			Width(width).
			Render(line)
	}
	return fmt.Sprintf("%s %s", bar, title)
}

// subtaskDelegate renders sub-tasks the same as taskDelegate, except
// completed sub-tasks are struck through and never show the overdue badge.
type subtaskDelegate struct{}

func (d subtaskDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	t, ok := item.(TaskItem)
	if !ok {
		return
	}
	selected := index == m.Index()
	fmt.Fprint(w, renderTaskLine(t.task, selected, m.Width(), t.task.Completed, !t.task.Completed))
}

func (d subtaskDelegate) Height() int { return 1 }

func (d subtaskDelegate) Spacing() int { return 0 }

func (d subtaskDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }
