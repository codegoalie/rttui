package rtm

import (
	"encoding/json"
	"fmt"
	"sort"
	"time"
)

// Priority represents a task priority level.
type Priority int

const (
	PriorityNone   Priority = 0
	PriorityHigh   Priority = 1
	PriorityMedium Priority = 2
	PriorityLow    Priority = 3
)

// Task is a single RTM task.
type Task struct {
	ID           string
	TaskseriesID string
	Name         string
	URL          string
	Due          time.Time
	Priority     Priority
	Tags         []string
	ListID       string
	Completed    bool
	ParentTaskID string // empty string = top-level task, not a sub-task
	Notes        []Note
}

// Note is a single note attached to a task.
type Note struct {
	ID, Title, Body   string
	Created, Modified time.Time
}

// rtmTags handles the RTM API quirk where tags is either [] (no tags) or {"tag":[...]}
type rtmTags struct {
	Tag []string
}

func (t *rtmTags) UnmarshalJSON(data []byte) error {
	// empty array case: "tags": []
	if len(data) > 0 && data[0] == '[' {
		t.Tag = nil
		return nil
	}
	var obj struct {
		Tag []string `json:"tag"`
	}
	if err := json.Unmarshal(data, &obj); err != nil {
		return err
	}
	t.Tag = obj.Tag
	return nil
}

// rtmNote is a single note as returned by the RTM API.
type rtmNote struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Created  string `json:"created"`
	Modified string `json:"modified"`
	// Body's JSON key ("$t") is RTM's standard convention for an XML
	// element's text content when represented as JSON, per third-party RTM
	// client docs. This is NOT verified against a live API response yet —
	// confirm the key the first time this app is pointed at a task with a
	// real note (e.g. manual `go run` testing), and fix just this tag if
	// it's wrong.
	Body string `json:"$t"`
}

// rtmNotes handles the RTM API quirk where notes is either [] (no notes),
// {"note": {...}} (a single note collapsed to a bare object), or
// {"note": [...]} (multiple notes).
type rtmNotes struct {
	Note []rtmNote
}

func (n *rtmNotes) UnmarshalJSON(data []byte) error {
	// empty array case: "notes": []
	if len(data) > 0 && data[0] == '[' {
		n.Note = nil
		return nil
	}
	var obj struct {
		Note json.RawMessage `json:"note"`
	}
	if err := json.Unmarshal(data, &obj); err != nil {
		return err
	}
	if len(obj.Note) == 0 {
		n.Note = nil
		return nil
	}
	if obj.Note[0] == '[' {
		var notes []rtmNote
		if err := json.Unmarshal(obj.Note, &notes); err != nil {
			return err
		}
		n.Note = notes
		return nil
	}
	var note rtmNote
	if err := json.Unmarshal(obj.Note, &note); err != nil {
		return err
	}
	n.Note = []rtmNote{note}
	return nil
}

// toNotes converts the raw RTM notes into the Note slice used by Task.
func (n rtmNotes) toNotes() []Note {
	var notes []Note
	for _, rn := range n.Note {
		note := Note{
			ID:    rn.ID,
			Title: rn.Title,
			Body:  rn.Body,
		}
		if rn.Created != "" {
			if created, err := time.Parse(time.RFC3339, rn.Created); err == nil {
				note.Created = created
			}
		}
		if rn.Modified != "" {
			if modified, err := time.Parse(time.RFC3339, rn.Modified); err == nil {
				note.Modified = modified
			}
		}
		notes = append(notes, note)
	}
	return notes
}

// rtmResponse mirrors the JSON returned by rtm.tasks.getList.
type rtmResponse struct {
	Rsp struct {
		Stat  string `json:"stat"`
		Tasks struct {
			List []struct {
				ID         string `json:"id"`
				Taskseries []struct {
					ID           string   `json:"id"`
					Name         string   `json:"name"`
					URL          string   `json:"url"`
					Tags         rtmTags  `json:"tags"`
					ParentTaskID string   `json:"parent_task_id"`
					Notes        rtmNotes `json:"notes"`
					Task         []struct {
						ID        string `json:"id"`
						Due       string `json:"due"`
						Priority  string `json:"priority"`
						Completed string `json:"completed"`
						Deleted   string `json:"deleted"`
					} `json:"task"`
				} `json:"taskseries"`
			} `json:"list"`
		} `json:"tasks"`
	} `json:"rsp"`
}

func parsePriority(s string) Priority {
	switch s {
	case "1":
		return PriorityHigh
	case "2":
		return PriorityMedium
	case "3":
		return PriorityLow
	default:
		return PriorityNone
	}
}

// parseAllTasks parses a raw RTM JSON response into a slice of Tasks,
// including completed tasks (deleted tasks are still always dropped).
func parseAllTasks(data []byte) ([]Task, error) {
	var resp rtmResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parseAllTasks: %w", err)
	}
	if resp.Rsp.Stat != "ok" {
		return nil, fmt.Errorf("parseAllTasks: API stat=%q", resp.Rsp.Stat)
	}

	var tasks []Task
	for _, list := range resp.Rsp.Tasks.List {
		for _, ts := range list.Taskseries {
			for _, t := range ts.Task {
				if t.Deleted != "" {
					continue
				}
				task := Task{
					ID:           t.ID,
					TaskseriesID: ts.ID,
					Name:         ts.Name,
					URL:          ts.URL,
					Priority:     parsePriority(t.Priority),
					Tags:         ts.Tags.Tag,
					ListID:       list.ID,
					Completed:    t.Completed != "",
					ParentTaskID: ts.ParentTaskID,
					Notes:        ts.Notes.toNotes(),
				}
				if t.Due != "" {
					if due, err := time.Parse(time.RFC3339, t.Due); err == nil {
						task.Due = due
					}
				}
				tasks = append(tasks, task)
			}
		}
	}
	SortTasks(tasks)
	return tasks, nil
}

// filterMainListTasks excludes tasks that shouldn't appear in the main task list.
// Sub-tasks (ParentTaskID != "") are intentionally NOT filtered here yet —
// that's coming in a follow-up commit once the UI can show sub-tasks somewhere,
// so they don't just vanish with no way to view them.
func filterMainListTasks(all []Task) []Task {
	var out []Task
	for _, t := range all {
		if t.Completed {
			continue
		}
		out = append(out, t)
	}
	return out
}

// SortTasks sorts tasks in place using the standard display order:
// bucket (overdue, today, day-by-day, no due date) → priority → due → name → ID.
func SortTasks(tasks []Task) {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	// bucketRank assigns a sort rank to group tasks into date buckets.
	// Overdue (all past dates) = 0, Today = 1, future days by offset, no due date last.
	bucketRank := func(due time.Time) int {
		if due.IsZero() {
			return 1<<31 - 1 // no due date goes last
		}
		dueDay := time.Date(due.Year(), due.Month(), due.Day(), 0, 0, 0, 0, today.Location())
		if dueDay.Before(today) {
			return 0 // all overdue dates collapse into one bucket
		}
		return int(dueDay.Sub(today).Hours()/24) + 1
	}

	sort.Slice(tasks, func(i, j int) bool {
		bi, bj := bucketRank(tasks[i].Due), bucketRank(tasks[j].Due)

		// Sort by date bucket first.
		if bi != bj {
			return bi < bj
		}

		// Within the same bucket, sort by priority (high=1 first, none=0 last).
		pi, pj := tasks[i].Priority, tasks[j].Priority
		if pi != pj {
			if pi == PriorityNone {
				return false
			}
			if pj == PriorityNone {
				return true
			}
			return pi < pj
		}

		// Then by due date (oldest first) within the same priority.
		di, dj := tasks[i].Due, tasks[j].Due
		if !di.IsZero() && !dj.IsZero() && !di.Equal(dj) {
			return di.Before(dj)
		}

		// Stable tiebreaker: task name, then ID.
		if tasks[i].Name != tasks[j].Name {
			return tasks[i].Name < tasks[j].Name
		}
		return tasks[i].ID < tasks[j].ID
	})
}

// GetTasks fetches tasks matching the given filter for the authenticated user.
func (c *Client) GetTasks(token, filter string) ([]Task, error) {
	params := map[string]string{"auth_token": token}
	if filter != "" {
		params["filter"] = filter
	}
	data, err := c.Call("rtm.tasks.getList", params)
	if err != nil {
		return nil, err
	}
	all, err := parseAllTasks(data)
	if err != nil {
		return nil, err
	}
	return filterMainListTasks(all), nil
}
