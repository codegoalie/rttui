package ui

import (
	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"

	"git.codegoalie.com/rttui.git/internal/rtm"
)

// Model is the bubbletea application model.
type Model struct {
	list  list.Model
	tasks []rtm.Task
}

// NewModel creates a Model pre-loaded with tasks.
func NewModel(tasks []rtm.Task) Model {
	items := make([]list.Item, len(tasks))
	for i, t := range tasks {
		items[i] = TaskItem{task: t}
	}

	delegate := list.NewDefaultDelegate()
	l := list.New(items, delegate, 0, 0)
	l.Title = "Remember The Milk"

	return Model{list: l, tasks: tasks}
}

// Init satisfies tea.Model; data is pre-loaded so no commands needed.
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles messages.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetSize(msg.Width, msg.Height)
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

// View renders the current state.
func (m Model) View() tea.View {
	return tea.NewView(m.list.View())
}
