package ui

import (
	"testing"
	"time"

	"charm.land/bubbles/v2/list"

	"git.codegoalie.com/rttui/internal/rtm"
)

func TestFormatDue(t *testing.T) {
	tests := []struct {
		name string
		due  time.Time
		want string
	}{
		{
			name: "zero time returns None",
			due:  time.Time{},
			want: "None",
		},
		{
			name: "non-zero time formats as weekday month day, year",
			due:  time.Date(2024, time.March, 15, 0, 0, 0, 0, time.UTC),
			want: "Fri Mar 15, 2024",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := formatDue(tt.due); got != tt.want {
				t.Errorf("formatDue(%v) = %q, want %q", tt.due, got, tt.want)
			}
		})
	}
}

func TestFormatTags(t *testing.T) {
	tests := []struct {
		name string
		tags []string
		want string
	}{
		{
			name: "nil slice returns None",
			tags: nil,
			want: "None",
		},
		{
			name: "empty slice returns None",
			tags: []string{},
			want: "None",
		},
		{
			name: "single tag returns tag unchanged",
			tags: []string{"errand"},
			want: "errand",
		},
		{
			name: "multiple tags are comma-and-space joined",
			tags: []string{"errand", "urgent", "home"},
			want: "errand, urgent, home",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := formatTags(tt.tags); got != tt.want {
				t.Errorf("formatTags(%v) = %q, want %q", tt.tags, got, tt.want)
			}
		})
	}
}

func TestFormatNoteHeader(t *testing.T) {
	created := time.Date(2024, time.March, 15, 15, 4, 0, 0, time.UTC)

	tests := []struct {
		name string
		note rtm.Note
		want string
	}{
		{
			// No title: header is just the created timestamp, formatted
			// "Jan 2, 2006 3:04 PM".
			name: "no title uses timestamp only",
			note: rtm.Note{Title: "", Created: created},
			want: "Mar 15, 2024 3:04 PM",
		},
		{
			// With title: header is "<title> — <timestamp>", using an
			// em-dash-with-spaces separator, timestamp formatted
			// "Jan 2, 2006 3:04 PM".
			name: "with title prefixes title and em-dash separator",
			note: rtm.Note{Title: "Follow-up", Created: created},
			want: "Follow-up — Mar 15, 2024 3:04 PM",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := formatNoteHeader(tt.note); got != tt.want {
				t.Errorf("formatNoteHeader(%+v) = %q, want %q", tt.note, got, tt.want)
			}
		})
	}
}

func TestBuildSubtaskItems(t *testing.T) {
	t.Run("nil input returns empty slice", func(t *testing.T) {
		items := buildSubtaskItems(nil)
		if len(items) != 0 {
			t.Errorf("buildSubtaskItems(nil) = %v, want empty", items)
		}
	})

	t.Run("empty input returns empty slice", func(t *testing.T) {
		items := buildSubtaskItems([]rtm.Task{})
		if len(items) != 0 {
			t.Errorf("buildSubtaskItems([]) = %v, want empty", items)
		}
	})

	t.Run("one TaskItem per task in order", func(t *testing.T) {
		tasks := []rtm.Task{
			{ID: "1", TaskseriesID: "ts1", Name: "First", Completed: false},
			{ID: "2", TaskseriesID: "ts2", Name: "Second", Completed: true},
			{ID: "3", TaskseriesID: "ts3", Name: "Third", Completed: false},
		}

		items := buildSubtaskItems(tasks)

		if len(items) != len(tasks) {
			t.Fatalf("buildSubtaskItems returned %d items, want %d", len(items), len(tasks))
		}
		for i, want := range tasks {
			ti, ok := items[i].(TaskItem)
			if !ok {
				t.Fatalf("items[%d] is not a TaskItem: %#v", i, items[i])
			}
			if ti.task.ID != want.ID {
				t.Errorf("items[%d].task.ID = %q, want %q", i, ti.task.ID, want.ID)
			}
			if ti.task.Completed != want.Completed {
				t.Errorf("items[%d].task.Completed = %v, want %v", i, ti.task.Completed, want.Completed)
			}
		}
	})
}

func TestFindSubtaskIndex(t *testing.T) {
	items := []list.Item{
		TaskItem{task: rtm.Task{ID: "1", TaskseriesID: "ts1"}},
		TaskItem{task: rtm.Task{ID: "2", TaskseriesID: "ts2"}},
		TaskItem{task: rtm.Task{ID: "3", TaskseriesID: "ts3"}},
	}

	tests := []struct {
		name   string
		target rtm.Task
		want   int
	}{
		{
			name:   "found first item",
			target: rtm.Task{ID: "1", TaskseriesID: "ts1"},
			want:   0,
		},
		{
			name:   "found second item",
			target: rtm.Task{ID: "2", TaskseriesID: "ts2"},
			want:   1,
		},
		{
			name:   "not found returns -1",
			target: rtm.Task{ID: "99", TaskseriesID: "ts99"},
			want:   -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := findSubtaskIndex(items, tt.target); got != tt.want {
				t.Errorf("findSubtaskIndex(items, %+v) = %d, want %d", tt.target, got, tt.want)
			}
		})
	}
}
