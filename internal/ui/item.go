package ui

import (
	"fmt"
	"time"

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

// Description returns a styled one-liner with priority and due date.
func (t TaskItem) Description() string {
	pri := priorityLabel(t.task.Priority)
	due := dueLabel(t.task.Due)
	if due == "" {
		return pri
	}
	return fmt.Sprintf("%s  %s", pri, due)
}

// FilterValue is used by the list's built-in fuzzy filter.
func (t TaskItem) FilterValue() string {
	return t.task.Name
}

func priorityLabel(p rtm.Priority) string {
	switch p {
	case rtm.PriorityHigh:
		return priorityHighStyle.Render("[P1]")
	case rtm.PriorityMedium:
		return priorityMedStyle.Render("[P2]")
	case rtm.PriorityLow:
		return priorityLowStyle.Render("[P3]")
	default:
		return priorityNoneStyle.Render("[--]")
	}
}

func dueLabel(due time.Time) string {
	if due.IsZero() {
		return ""
	}
	now := time.Now()
	label := "Due: " + due.Format("Jan 2")
	if due.Before(now) || sameDay(due, now) {
		return dueSoonStyle.Render(label)
	}
	return dueStyle.Render(label)
}

func sameDay(a, b time.Time) bool {
	ay, am, ad := a.Date()
	by, bm, bd := b.Date()
	return ay == by && am == bm && ad == bd
}
