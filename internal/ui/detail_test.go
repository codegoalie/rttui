package ui

import (
	"testing"
	"time"

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
