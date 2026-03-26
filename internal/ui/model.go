package ui

import (
	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"git.codegoalie.com/rttui.git/internal/rtm"
)

const searchBarHeight = 1

// Model is the bubbletea application model.
type Model struct {
	client        *rtm.Client
	token         string
	currentFilter string

	list         list.Model
	windowWidth  int
	windowHeight int

	searching   bool
	searchMode  vimMode
	searchInput textinput.Model
	searchErr   error

	loading bool
}

// NewModel creates a Model pre-loaded with tasks.
func NewModel(client *rtm.Client, token, filter string, tasks []rtm.Task) Model {
	items := make([]list.Item, len(tasks))
	for i, t := range tasks {
		items[i] = TaskItem{task: t}
	}

	delegate := list.NewDefaultDelegate()
	l := list.New(items, delegate, 0, 0)
	l.Title = "Remember The Milk"
	l.SetFilteringEnabled(false)

	return Model{
		client:        client,
		token:         token,
		currentFilter: filter,
		list:          l,
		searchInput:   newSearchInput(),
	}
}

// Init satisfies tea.Model; data is pre-loaded so no commands needed.
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles messages.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.windowWidth = msg.Width
		m.windowHeight = msg.Height
		if m.searching {
			m.list.SetSize(msg.Width, msg.Height-searchBarHeight)
		} else {
			m.list.SetSize(msg.Width, msg.Height)
		}
		return m, nil

	case fetchTasksMsg:
		m.loading = false
		m.list.StopSpinner()
		if msg.err != nil {
			m.searchErr = msg.err
			return m, nil
		}
		items := make([]list.Item, len(msg.tasks))
		for i, t := range msg.tasks {
			items[i] = TaskItem{task: t}
		}
		cmd := m.list.SetItems(items)
		return m, cmd

	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
		if m.loading {
			var cmd tea.Cmd
			m.list, cmd = m.list.Update(msg)
			return m, cmd
		}
		if m.searching {
			return m.updateSearch(msg)
		}
		switch msg.String() {
		case "q":
			return m, tea.Quit
		case "/":
			return m.openSearch()
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

// View renders the current state.
func (m Model) View() tea.View {
	listView := m.list.View()

	if m.searchErr != nil {
		errBar := searchErrorStyle.Render("Error: " + m.searchErr.Error() + "  (press / to retry)")
		return tea.NewView(lipgloss.JoinVertical(lipgloss.Left, listView, errBar))
	}

	if m.searching {
		label := " INSERT "
		if m.searchMode == modeNormal {
			label = " NORMAL "
		}
		bar := searchBarStyle.Render(label + m.searchInput.View())
		return tea.NewView(lipgloss.JoinVertical(lipgloss.Left, listView, bar))
	}

	return tea.NewView(listView)
}
