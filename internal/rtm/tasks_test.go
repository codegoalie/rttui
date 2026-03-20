package rtm

import (
	"testing"
	"time"
)

const fixtureJSON = `{
  "rsp": {
    "stat": "ok",
    "tasks": {
      "list": [
        {
          "id": "list-1",
          "taskseries": [
            {
              "id": "ts-1",
              "name": "Buy milk",
              "tags": { "tag": ["grocery", "errand"] },
              "task": [
                {
                  "id": "task-1",
                  "due": "2026-03-20T12:00:00Z",
                  "priority": "1",
                  "completed": "",
                  "deleted": ""
                }
              ]
            },
            {
              "id": "ts-2",
              "name": "Done task",
              "tags": { "tag": [] },
              "task": [
                {
                  "id": "task-2",
                  "due": "",
                  "priority": "N",
                  "completed": "2026-03-19T10:00:00Z",
                  "deleted": ""
                }
              ]
            },
            {
              "id": "ts-3",
              "name": "Read book",
              "tags": [],
              "task": [
                {
                  "id": "task-3",
                  "due": "",
                  "priority": "3",
                  "completed": "",
                  "deleted": ""
                }
              ]
            }
          ]
        }
      ]
    }
  }
}`

func TestParseTasks(t *testing.T) {
	tasks, err := parseTasks([]byte(fixtureJSON))
	if err != nil {
		t.Fatalf("parseTasks error: %v", err)
	}
	if len(tasks) != 2 {
		t.Fatalf("expected 2 tasks (completed filtered), got %d", len(tasks))
	}

	// First task: Buy milk
	milk := tasks[0]
	if milk.Name != "Buy milk" {
		t.Errorf("task[0].Name = %q, want %q", milk.Name, "Buy milk")
	}
	if milk.Priority != PriorityHigh {
		t.Errorf("task[0].Priority = %v, want PriorityHigh", milk.Priority)
	}
	if milk.ListID != "list-1" {
		t.Errorf("task[0].ListID = %q, want %q", milk.ListID, "list-1")
	}
	wantDue := time.Date(2026, 3, 20, 12, 0, 0, 0, time.UTC)
	if !milk.Due.Equal(wantDue) {
		t.Errorf("task[0].Due = %v, want %v", milk.Due, wantDue)
	}
	if len(milk.Tags) != 2 || milk.Tags[0] != "grocery" {
		t.Errorf("task[0].Tags = %v, want [grocery errand]", milk.Tags)
	}

	// Second task: Read book (no due, low priority)
	book := tasks[1]
	if book.Name != "Read book" {
		t.Errorf("task[1].Name = %q, want %q", book.Name, "Read book")
	}
	if book.Priority != PriorityLow {
		t.Errorf("task[1].Priority = %v, want PriorityLow", book.Priority)
	}
	if !book.Due.IsZero() {
		t.Errorf("task[1].Due = %v, want zero", book.Due)
	}
}
