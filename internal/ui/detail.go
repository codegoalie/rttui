package ui

import (
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"git.codegoalie.com/rttui/internal/rtm"
)

// openDetail opens the full-screen task detail view for the currently
// selected task. If the cursor is on a HeadingItem rather than a task, this
// is a no-op. No network call is made this milestone — the detail view
// renders from the already-loaded task.
func (m Model) openDetail() (Model, tea.Cmd) {
	item, ok := m.list.SelectedItem().(TaskItem)
	if !ok {
		return m, nil
	}
	m.detailOpen = true
	m.detailTask = item.task
	m.detailErr = nil
	// Full-screen takeover: the list is hidden entirely (not sharing the
	// screen with a footer bar), so it's resized off-screen rather than
	// shrunk by footerHeight().
	m.list.SetSize(m.windowWidth, 0)
	return m, nil
}

// closeDetail closes the full-screen task detail view and restores the list
// to its full size.
func (m Model) closeDetail() Model {
	m.detailOpen = false
	m.detailErr = nil
	m.list.SetSize(m.windowWidth, m.windowHeight)
	return m
}

// updateDetail routes keypresses while the full-screen detail view is open.
//
// Extension seam: sub-task navigation, add-sub-task, and add-note handling
// (t/n/arrow keys) arrive in later milestones (M4/M5) — not built here.
func (m Model) updateDetail(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "q":
		return m.closeDetail(), nil
	default:
		return m, nil
	}
}

// formatDue formats a task's due date for the detail view: "None" if unset,
// otherwise "Mon Jan 2, 2006".
func formatDue(due time.Time) string {
	if due.IsZero() {
		return "None"
	}
	return due.Format("Mon Jan 2, 2006")
}

// formatTags formats a task's tags for the detail view: "None" if empty,
// otherwise a comma-and-space-joined list.
func formatTags(tags []string) string {
	if len(tags) == 0 {
		return "None"
	}
	return strings.Join(tags, ", ")
}

// formatNoteHeader formats a note's header line for the detail view.
// When the note has a title, the header is "<title> — <timestamp>"; with no
// title, it's just the timestamp. The timestamp is always formatted as
// "Jan 2, 2006 3:04 PM" (e.g. "Mar 15, 2024 3:04 PM").
func formatNoteHeader(n rtm.Note) string {
	ts := n.Created.Format("Jan 2, 2006 3:04 PM")
	if n.Title == "" {
		return ts
	}
	return n.Title + " — " + ts
}

// renderDetail renders the full-screen, read-only task detail view: header,
// fields, notes, an optional error bar, and a bottom hint line.
func (m Model) renderDetail() string {
	task := m.detailTask

	// Header: priority bar + task name, always styled as if selected since
	// it's the sole focus of a full-screen view. Mirrors delegate.go's
	// selected-title style path exactly.
	barColor := priorityColor(task.Priority)
	bar := lipgloss.NewStyle().Foreground(barColor).Render("▎")
	titleStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#c0caf5")).Bold(true)
	if task.URL != "" {
		titleStyle = titleStyle.Hyperlink(task.URL).Underline(true)
	}
	header := bar + " " + titleStyle.Render(task.Name)

	urlText := task.URL
	if urlText == "" {
		urlText = "None"
	}
	fields := lipgloss.JoinVertical(lipgloss.Left,
		detailLabelStyle.Render("Due:")+" "+formatDue(task.Due),
		detailLabelStyle.Render("List:")+" "+task.ListID,
		detailLabelStyle.Render("Tags:")+" "+formatTags(task.Tags),
		detailLabelStyle.Render("URL:")+" "+urlText,
	)

	notesLines := []string{headingStyle.Render("── Notes ──")}
	if len(task.Notes) == 0 {
		notesLines = append(notesLines, "No notes yet")
	} else {
		for _, n := range task.Notes {
			notesLines = append(notesLines,
				detailNoteHeaderStyle.Render(formatNoteHeader(n)),
				detailNoteBodyStyle.Render(n.Body),
			)
		}
	}
	notes := lipgloss.JoinVertical(lipgloss.Left, notesLines...)

	sections := []string{header, "", fields, "", notes}

	if m.detailErr != nil {
		sections = append(sections, "", addErrorStyle.Render("Error: "+m.detailErr.Error()))
	}

	// Hint fragments are joined so later milestones can append to this
	// slice without restructuring the join.
	hints := []string{"esc: close"}
	sections = append(sections, "", detailHintStyle.Render(strings.Join(hints, " • ")))

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}
