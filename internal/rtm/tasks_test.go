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
              "url": "https://example.com/milk",
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

const parentTaskFixtureJSON = `{
  "rsp": {
    "stat": "ok",
    "tasks": {
      "list": [
        {
          "id": "list-1",
          "taskseries": [
            {
              "id": "ts-parent",
              "name": "Sub-task",
              "parent_task_id": "ts-parent-1",
              "task": [
                {
                  "id": "task-parent",
                  "due": "",
                  "priority": "N",
                  "completed": "",
                  "deleted": ""
                }
              ]
            },
            {
              "id": "ts-noparent",
              "name": "Top-level task",
              "task": [
                {
                  "id": "task-noparent",
                  "due": "",
                  "priority": "N",
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

const notesFixtureJSON = `{
  "rsp": {
    "stat": "ok",
    "tasks": {
      "list": [
        {
          "id": "list-1",
          "taskseries": [
            {
              "id": "ts-nonotes",
              "name": "No notes key",
              "task": [
                { "id": "task-nonotes", "due": "", "priority": "N", "completed": "", "deleted": "" }
              ]
            },
            {
              "id": "ts-emptynotes",
              "name": "Empty notes",
              "notes": [],
              "task": [
                { "id": "task-emptynotes", "due": "", "priority": "N", "completed": "", "deleted": "" }
              ]
            },
            {
              "id": "ts-singlenote",
              "name": "Single note",
              "notes": { "note": {"id":"n1","title":"T","$t":"body text","created":"2026-03-19T10:00:00Z","modified":"2026-03-19T10:00:00Z"} },
              "task": [
                { "id": "task-singlenote", "due": "", "priority": "N", "completed": "", "deleted": "" }
              ]
            },
            {
              "id": "ts-multinote",
              "name": "Multi note",
              "notes": { "note": [
                {"id":"n2","title":"A","$t":"body A","created":"2026-03-19T10:00:00Z","modified":"2026-03-19T10:00:00Z"},
                {"id":"n3","title":"B","$t":"body B","created":"2026-03-19T11:00:00Z","modified":"2026-03-19T11:00:00Z"}
              ] },
              "task": [
                { "id": "task-multinote", "due": "", "priority": "N", "completed": "", "deleted": "" }
              ]
            }
          ]
        }
      ]
    }
  }
}`

// findTask locates a task by ID, failing the test if it's not present.
// Tasks are sorted by SortTasks, so tests must not rely on slice index order.
func findTask(t *testing.T, tasks []Task, id string) Task {
	t.Helper()
	for _, task := range tasks {
		if task.ID == id {
			return task
		}
	}
	t.Fatalf("task with ID %q not found in %+v", id, tasks)
	return Task{}
}

func TestParseAllTasks(t *testing.T) {
	tasks, err := parseAllTasks([]byte(fixtureJSON))
	if err != nil {
		t.Fatalf("parseAllTasks error: %v", err)
	}
	if len(tasks) != 3 {
		t.Fatalf("expected 3 tasks (completed retained), got %d", len(tasks))
	}

	milk := findTask(t, tasks, "task-1")
	if milk.Name != "Buy milk" {
		t.Errorf("milk.Name = %q, want %q", milk.Name, "Buy milk")
	}
	if milk.Completed {
		t.Errorf("milk.Completed = true, want false")
	}
	if milk.Priority != PriorityHigh {
		t.Errorf("milk.Priority = %v, want PriorityHigh", milk.Priority)
	}
	if milk.ListID != "list-1" {
		t.Errorf("milk.ListID = %q, want %q", milk.ListID, "list-1")
	}
	wantDue := time.Date(2026, 3, 20, 12, 0, 0, 0, time.UTC)
	if !milk.Due.Equal(wantDue) {
		t.Errorf("milk.Due = %v, want %v", milk.Due, wantDue)
	}
	if len(milk.Tags) != 2 || milk.Tags[0] != "grocery" {
		t.Errorf("milk.Tags = %v, want [grocery errand]", milk.Tags)
	}
	if milk.URL != "https://example.com/milk" {
		t.Errorf("milk.URL = %q, want %q", milk.URL, "https://example.com/milk")
	}

	done := findTask(t, tasks, "task-2")
	if done.Name != "Done task" {
		t.Errorf("done.Name = %q, want %q", done.Name, "Done task")
	}
	if !done.Completed {
		t.Errorf("done.Completed = false, want true")
	}

	book := findTask(t, tasks, "task-3")
	if book.Name != "Read book" {
		t.Errorf("book.Name = %q, want %q", book.Name, "Read book")
	}
	if book.Completed {
		t.Errorf("book.Completed = true, want false")
	}
	if book.Priority != PriorityLow {
		t.Errorf("book.Priority = %v, want PriorityLow", book.Priority)
	}
	if !book.Due.IsZero() {
		t.Errorf("book.Due = %v, want zero", book.Due)
	}
	if book.URL != "" {
		t.Errorf("book.URL = %q, want empty", book.URL)
	}
}

func TestParseAddTaskResponse(t *testing.T) {
	t.Run("no due date, no priority", func(t *testing.T) {
		const addFixtureJSON = `{"rsp":{"stat":"ok","list":{"id":"list-1","taskseries":{"id":"ts-1","name":"New task","url":"","task":{"id":"task-1","due":"","priority":"N"}}}}}`

		task, err := parseAddTaskResponse([]byte(addFixtureJSON))
		if err != nil {
			t.Fatalf("parseAddTaskResponse error: %v", err)
		}
		if task.ID != "task-1" {
			t.Errorf("task.ID = %q, want %q", task.ID, "task-1")
		}
		if task.TaskseriesID != "ts-1" {
			t.Errorf("task.TaskseriesID = %q, want %q", task.TaskseriesID, "ts-1")
		}
		if task.ListID != "list-1" {
			t.Errorf("task.ListID = %q, want %q", task.ListID, "list-1")
		}
		if task.Name != "New task" {
			t.Errorf("task.Name = %q, want %q", task.Name, "New task")
		}
		if task.Priority != PriorityNone {
			t.Errorf("task.Priority = %v, want PriorityNone", task.Priority)
		}
	})

	t.Run("with due date and priority", func(t *testing.T) {
		const addFixtureJSON = `{"rsp":{"stat":"ok","list":{"id":"list-2","taskseries":{"id":"ts-2","name":"Urgent task","url":"https://example.com/urgent","task":{"id":"task-2","due":"2026-03-20T12:00:00Z","priority":"1"}}}}}`

		task, err := parseAddTaskResponse([]byte(addFixtureJSON))
		if err != nil {
			t.Fatalf("parseAddTaskResponse error: %v", err)
		}
		wantDue := time.Date(2026, 3, 20, 12, 0, 0, 0, time.UTC)
		if !task.Due.Equal(wantDue) {
			t.Errorf("task.Due = %v, want %v", task.Due, wantDue)
		}
		if task.Priority != PriorityHigh {
			t.Errorf("task.Priority = %v, want PriorityHigh", task.Priority)
		}
	})
}

func TestFilterSubtasksOf(t *testing.T) {
	all := []Task{
		{ID: "1", Name: "Child one", ParentTaskID: "parent-1"},
		{ID: "2", Name: "Unrelated", ParentTaskID: "parent-2"},
		{ID: "3", Name: "Top level", ParentTaskID: ""},
		{ID: "4", Name: "Child two", ParentTaskID: "parent-1"},
	}

	got := filterSubtasksOf(all, "parent-1")

	if len(got) != 2 {
		t.Fatalf("expected 2 subtasks, got %d: %+v", len(got), got)
	}
	ids := map[string]bool{got[0].ID: true, got[1].ID: true}
	if !ids["1"] || !ids["4"] {
		t.Errorf("filterSubtasksOf returned wrong tasks: got %+v", got)
	}
}

func TestFilterMainListTasks(t *testing.T) {
	all := []Task{
		{ID: "1", Name: "Active one", Completed: false},
		{ID: "2", Name: "Completed one", Completed: true},
		{ID: "3", Name: "Active two", Completed: false},
		{ID: "4", Name: "Sub-task", ParentTaskID: "parent-1"},
	}

	got := filterMainListTasks(all)

	if len(got) != 2 {
		t.Fatalf("expected 2 tasks after filtering, got %d", len(got))
	}
	for _, task := range got {
		if task.ID == "2" {
			t.Errorf("filterMainListTasks retained completed task %q", task.ID)
		}
		if task.ID == "4" {
			t.Errorf("filterMainListTasks retained sub-task %q", task.ID)
		}
	}
	if got[0].ID != "1" || got[1].ID != "3" {
		t.Errorf("filterMainListTasks changed order or content: got %+v", got)
	}
}

func TestParseAllTasks_ParentTaskID(t *testing.T) {
	tasks, err := parseAllTasks([]byte(parentTaskFixtureJSON))
	if err != nil {
		t.Fatalf("parseAllTasks error: %v", err)
	}

	sub := findTask(t, tasks, "task-parent")
	if sub.ParentTaskID != "ts-parent-1" {
		t.Errorf("sub.ParentTaskID = %q, want %q", sub.ParentTaskID, "ts-parent-1")
	}

	topLevel := findTask(t, tasks, "task-noparent")
	if topLevel.ParentTaskID != "" {
		t.Errorf("topLevel.ParentTaskID = %q, want empty", topLevel.ParentTaskID)
	}
}

func TestParseAllTasks_Notes(t *testing.T) {
	tasks, err := parseAllTasks([]byte(notesFixtureJSON))
	if err != nil {
		t.Fatalf("parseAllTasks error: %v", err)
	}

	noNotes := findTask(t, tasks, "task-nonotes")
	if len(noNotes.Notes) != 0 {
		t.Errorf("noNotes.Notes = %v, want empty (no notes key present)", noNotes.Notes)
	}

	emptyNotes := findTask(t, tasks, "task-emptynotes")
	if len(emptyNotes.Notes) != 0 {
		t.Errorf("emptyNotes.Notes = %v, want empty (notes: [])", emptyNotes.Notes)
	}

	single := findTask(t, tasks, "task-singlenote")
	if len(single.Notes) != 1 {
		t.Fatalf("single.Notes = %+v, want exactly 1 note", single.Notes)
	}
	note := single.Notes[0]
	if note.ID != "n1" {
		t.Errorf("note.ID = %q, want %q", note.ID, "n1")
	}
	if note.Title != "T" {
		t.Errorf("note.Title = %q, want %q", note.Title, "T")
	}
	if note.Body != "body text" {
		t.Errorf("note.Body = %q, want %q", note.Body, "body text")
	}
	wantTime := time.Date(2026, 3, 19, 10, 0, 0, 0, time.UTC)
	if !note.Created.Equal(wantTime) {
		t.Errorf("note.Created = %v, want %v", note.Created, wantTime)
	}
	if !note.Modified.Equal(wantTime) {
		t.Errorf("note.Modified = %v, want %v", note.Modified, wantTime)
	}

	multi := findTask(t, tasks, "task-multinote")
	if len(multi.Notes) != 2 {
		t.Fatalf("multi.Notes = %+v, want exactly 2 notes", multi.Notes)
	}
	if multi.Notes[0].ID != "n2" || multi.Notes[0].Body != "body A" {
		t.Errorf("multi.Notes[0] = %+v, want ID=n2 Body=%q", multi.Notes[0], "body A")
	}
	if multi.Notes[1].ID != "n3" || multi.Notes[1].Body != "body B" {
		t.Errorf("multi.Notes[1] = %+v, want ID=n3 Body=%q", multi.Notes[1], "body B")
	}
}
