package ui

import (
	"image/color"
	"time"

	"charm.land/bubbles/v2/list"
	"git.codegoalie.com/rttui.git/internal/rtm"
)

// TaskItem wraps an rtm.Task for use in the bubbles list.
type TaskItem struct {
	task rtm.Task
}

// Title returns the task name displayed in the list.
func (t TaskItem) Title() string {
	return t.task.Name
}

// Description is unused but required by the list.Item interface.
func (t TaskItem) Description() string { return "" }

// FilterValue is used by the list's built-in fuzzy filter.
func (t TaskItem) FilterValue() string {
	return t.task.Name
}

// priorityColor returns the color for the priority bar.
func priorityColor(p rtm.Priority) color.Color {
	switch p {
	case rtm.PriorityHigh:
		return priorityHighColor
	case rtm.PriorityMedium:
		return priorityMedColor
	case rtm.PriorityLow:
		return priorityLowColor
	default:
		return priorityNoneColor
	}
}

// isOverdue returns true if the due date is strictly before today.
func isOverdue(due time.Time) bool {
	if due.IsZero() {
		return false
	}
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	dueDay := time.Date(due.Year(), due.Month(), due.Day(), 0, 0, 0, 0, today.Location())
	return dueDay.Before(today)
}

func sameDay(a, b time.Time) bool {
	ay, am, ad := a.Date()
	by, bm, bd := b.Date()
	return ay == by && am == bm && ad == bd
}

// HeadingItem is a non-interactive separator shown between date groups.
type HeadingItem struct {
	label string
}

func (h HeadingItem) Title() string       { return h.label }
func (h HeadingItem) Description() string { return "" }
func (h HeadingItem) FilterValue() string { return "" }

// dayLabel returns a human-friendly label for the given date relative to now.
func dayLabel(due time.Time, now time.Time) string {
	if due.IsZero() {
		return "No Due Date"
	}
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	dueDay := time.Date(due.Year(), due.Month(), due.Day(), 0, 0, 0, 0, today.Location())
	if dueDay.Before(today) {
		return "Overdue"
	}
	if dueDay.Equal(today) {
		return "Today"
	}
	tomorrow := today.AddDate(0, 0, 1)
	if dueDay.Equal(tomorrow) {
		return "Tomorrow"
	}
	return dueDay.Format("Monday")
}

// buildItems creates list items from tasks, inserting HeadingItems between date groups.
func buildItems(tasks []rtm.Task) []list.Item {
	now := time.Now()
	var items []list.Item
	lastLabel := ""
	for _, t := range tasks {
		label := dayLabel(t.Due, now)
		if label != lastLabel {
			items = append(items, HeadingItem{label: label})
			lastLabel = label
		}
		items = append(items, TaskItem{task: t})
	}
	return items
}
