package ui

import (
	"strings"
	"time"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"

	"git.codegoalie.com/rttui.git/internal/rtm"
)

type autoRefreshMsg struct{}

func autoRefreshCmd(d time.Duration) tea.Cmd {
	return func() tea.Msg {
		time.Sleep(d)
		return autoRefreshMsg{}
	}
}

type fetchTimelineMsg struct {
	id  string
	err error
}

type addTaskMsg struct{ err error }

type completeTaskMsg struct {
	task rtm.Task
	err  error
}

func completeTaskCmd(client *rtm.Client, token, timeline string, task rtm.Task) tea.Cmd {
	return func() tea.Msg {
		err := client.CompleteTask(token, timeline, task.ListID, task.TaskseriesID, task.ID)
		return completeTaskMsg{task: task, err: err}
	}
}

func fetchTimelineCmd(client *rtm.Client, token string) tea.Cmd {
	return func() tea.Msg {
		id, err := client.GetTimeline(token)
		return fetchTimelineMsg{id: id, err: err}
	}
}

func addTaskCmd(client *rtm.Client, token, timeline, raw string) tea.Cmd {
	return func() tea.Msg {
		transformed := transformForRTM(raw)
		return addTaskMsg{err: client.AddTask(token, timeline, transformed)}
	}
}

func taskKey(t rtm.Task) string { return t.TaskseriesID + ":" + t.ID }

func (m Model) completeSelected() (Model, tea.Cmd) {
	item, ok := m.list.SelectedItem().(TaskItem)
	if !ok {
		return m, nil
	}
	key := taskKey(item.task)
	if _, dup := m.pendingCompletes[key]; dup {
		return m, nil // already completing this task
	}

	m.completeErr = nil
	if m.pendingCompletes == nil {
		m.pendingCompletes = map[string]rtm.Task{}
	}
	m.pendingCompletes[key] = item.task

	// Optimistically remove the task (and any orphaned heading) immediately.
	newItems := removeTaskAndOrphanedHeading(m.list.Items(), item)
	setCmd := m.list.SetItems(newItems)
	spinCmd := m.list.StartSpinner()
	m.list.SetSize(m.windowWidth, m.windowHeight-m.footerHeight())

	var apiCmd tea.Cmd
	if m.timelineID == "" {
		apiCmd = fetchTimelineAndCompleteCmd(m.client, m.token, item.task)
	} else {
		apiCmd = completeTaskCmd(m.client, m.token, m.timelineID, item.task)
	}
	return m, tea.Batch(setCmd, spinCmd, apiCmd)
}

// removeTaskAndOrphanedHeading removes the target task from items and prunes
// any HeadingItem that is left with no TaskItem children.
func removeTaskAndOrphanedHeading(items []list.Item, target TaskItem) []list.Item {
	targetIdx := -1
	for i, it := range items {
		if t, ok := it.(TaskItem); ok && t.task.ID == target.task.ID && t.task.TaskseriesID == target.task.TaskseriesID {
			targetIdx = i
			break
		}
	}
	if targetIdx == -1 {
		return items
	}

	result := make([]list.Item, 0, len(items)-1)
	result = append(result, items[:targetIdx]...)
	result = append(result, items[targetIdx+1:]...)

	// Find the heading immediately preceding the removed task.
	headingIdx := -1
	for i := targetIdx - 1; i >= 0; i-- {
		if _, ok := items[i].(HeadingItem); ok {
			headingIdx = i
			break
		} else if _, ok := items[i].(TaskItem); ok {
			break // another task still in this bucket — heading survives
		}
	}
	if headingIdx == -1 {
		return result
	}

	// Check if any TaskItem follows the heading in the updated result.
	hasTaskAfter := false
	for i := headingIdx + 1; i < len(result); i++ {
		if _, ok := result[i].(HeadingItem); ok {
			break
		}
		if _, ok := result[i].(TaskItem); ok {
			hasTaskAfter = true
			break
		}
	}
	if hasTaskAfter {
		return result
	}

	// Prune the orphaned heading.
	pruned := make([]list.Item, 0, len(result)-1)
	pruned = append(pruned, result[:headingIdx]...)
	pruned = append(pruned, result[headingIdx+1:]...)
	return pruned
}

// restoreTaskItems rebuilds the list items to include a task that failed to
// complete. It extracts current visible tasks, adds the restored one, re-sorts,
// and calls buildItems to regenerate headings.
func (m Model) restoreTaskItems(task rtm.Task) []list.Item {
	existing := m.list.Items()
	tasks := make([]rtm.Task, 0, len(existing)+1)
	for _, it := range existing {
		if t, ok := it.(TaskItem); ok {
			tasks = append(tasks, t.task)
		}
	}
	tasks = append(tasks, task)
	rtm.SortTasks(tasks)
	return buildItems(tasks)
}

type fetchTimelineAndCompleteMsg struct {
	timelineID string
	task       rtm.Task
	err        error
}

func fetchTimelineAndCompleteCmd(client *rtm.Client, token string, task rtm.Task) tea.Cmd {
	return func() tea.Msg {
		id, err := client.GetTimeline(token)
		return fetchTimelineAndCompleteMsg{timelineID: id, task: task, err: err}
	}
}

func (m Model) openAdd() (Model, tea.Cmd) {
	m.adding = true
	m.addErr = nil
	m.completeErr = nil
	if m.addPreset != "" {
		m.addInput.SetValue(m.addPreset + " ")
	} else {
		m.addInput.SetValue("")
	}
	m.list.SetSize(m.windowWidth, m.windowHeight-footerBarHeight)
	if m.timelineID == "" {
		return m, fetchTimelineCmd(m.client, m.token)
	}
	return m, nil
}

func (m Model) closeAdd() Model {
	m.adding = false
	m.addErr = nil
	m.list.SetSize(m.windowWidth, m.windowHeight)
	return m
}

func (m Model) submitAdd() (Model, tea.Cmd) {
	raw := strings.TrimSpace(m.addInput.Value())
	if raw == "" {
		return m.closeAdd(), nil
	}
	m.loading = true
	m.adding = false
	m.list.SetSize(m.windowWidth, m.windowHeight)
	return m, tea.Batch(m.list.StartSpinner(), addTaskCmd(m.client, m.token, m.timelineID, raw))
}

func (m Model) updateAdd(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		return m.submitAdd()
	case "esc":
		return m.closeAdd(), nil
	default:
		var cmd tea.Cmd
		m.addInput, cmd = m.addInput.Update(msg)
		return m, cmd
	}
}
