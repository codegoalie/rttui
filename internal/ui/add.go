package ui

import (
	"strings"

	tea "charm.land/bubbletea/v2"

	"git.codegoalie.com/rttui.git/internal/rtm"
)

type fetchTimelineMsg struct {
	id  string
	err error
}

type addTaskMsg struct{ err error }

type completeTaskMsg struct{ err error }

func completeTaskCmd(client *rtm.Client, token, timeline string, task rtm.Task) tea.Cmd {
	return func() tea.Msg {
		err := client.CompleteTask(token, timeline, task.ListID, task.TaskseriesID, task.ID)
		return completeTaskMsg{err: err}
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

func (m Model) completeSelected() (Model, tea.Cmd) {
	item, ok := m.list.SelectedItem().(TaskItem)
	if !ok {
		return m, nil
	}
	m.loading = true
	if m.timelineID == "" {
		// Need a timeline first; fetch it, then complete will be triggered after.
		// For simplicity, get the timeline synchronously via a two-step command.
		return m, tea.Batch(m.list.StartSpinner(), fetchTimelineAndCompleteCmd(m.client, m.token, item.task))
	}
	return m, tea.Batch(m.list.StartSpinner(), completeTaskCmd(m.client, m.token, m.timelineID, item.task))
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
	m.addInput.SetValue("")
	m.list.SetSize(m.windowWidth, m.windowHeight-addBarHeight)
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
