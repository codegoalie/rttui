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
	ID            string
	TaskseriesID  string
	Name          string
	Due           time.Time
	Priority      Priority
	Tags          []string
	ListID        string
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

// rtmResponse mirrors the JSON returned by rtm.tasks.getList.
type rtmResponse struct {
	Rsp struct {
		Stat  string `json:"stat"`
		Tasks struct {
			List []struct {
				ID         string `json:"id"`
				Taskseries []struct {
					ID   string `json:"id"`
					Name string `json:"name"`
					Tags rtmTags `json:"tags"`
					Task []struct {
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

// parseTasks parses a raw RTM JSON response into a slice of Tasks.
func parseTasks(data []byte) ([]Task, error) {
	var resp rtmResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parseTasks: %w", err)
	}
	if resp.Rsp.Stat != "ok" {
		return nil, fmt.Errorf("parseTasks: API stat=%q", resp.Rsp.Stat)
	}

	var tasks []Task
	for _, list := range resp.Rsp.Tasks.List {
		for _, ts := range list.Taskseries {
			for _, t := range ts.Task {
				if t.Completed != "" || t.Deleted != "" {
					continue
				}
				task := Task{
					ID:           t.ID,
					TaskseriesID: ts.ID,
					Name:         ts.Name,
					Priority:     parsePriority(t.Priority),
					Tags:         ts.Tags.Tag,
					ListID:       list.ID,
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

	return tasks, nil
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
	return parseTasks(data)
}
