package ui

import (
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"git.codegoalie.com/rttui.git/internal/rtm"
)

const footerBarHeight = 1

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

	adding     bool
	addInput   SmartInput
	addPreset  string
	addErr     error
	timelineID string

	loading bool
}

// NewModel creates a Model pre-loaded with tasks.
func NewModel(client *rtm.Client, token, filter, addPreset string, tasks []rtm.Task) Model {
	items := buildItems(tasks)

	delegate := newTaskDelegate()
	l := list.New(items, delegate, 0, 0)
	l.Title = "Remember The Milk"
	l.SetFilteringEnabled(false)
	l.AdditionalShortHelpKeys = func() []key.Binding {
		return []key.Binding{
			key.NewBinding(key.WithKeys("/"), key.WithHelp("/", "search")),
			key.NewBinding(key.WithKeys("n"), key.WithHelp("n", "add task")),
			key.NewBinding(key.WithKeys("c"), key.WithHelp("c", "complete")),
			key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "refresh")),
		}
	}
	l.AdditionalFullHelpKeys = func() []key.Binding {
		return []key.Binding{
			key.NewBinding(key.WithKeys("/"), key.WithHelp("/", "search tasks")),
			key.NewBinding(key.WithKeys("n"), key.WithHelp("n", "add task")),
			key.NewBinding(key.WithKeys("c"), key.WithHelp("c", "complete task")),
			key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "refresh tasks")),
		}
	}

	return Model{
		client:        client,
		token:         token,
		currentFilter: filter,
		addPreset:     addPreset,
		list:          l,
		searchInput:   newSearchInput(),
		addInput:      NewSmartInput("Add: "),
	}
}

// footerHeight returns the number of rows the footer bar occupies given current state.
func (m Model) footerHeight() int {
	if m.searching || m.adding || m.searchErr != nil || m.addErr != nil {
		return footerBarHeight
	}
	return 0
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
		m.list.SetSize(msg.Width, msg.Height-m.footerHeight())
		return m, nil

	case fetchTimelineMsg:
		if msg.err != nil {
			m = m.closeAdd()
			m.addErr = msg.err // set after closeAdd so it isn't cleared
			m.list.SetSize(m.windowWidth, m.windowHeight-m.footerHeight())
		} else {
			m.timelineID = msg.id
		}
		return m, nil

	case fetchTimelineAndCompleteMsg:
		if msg.err != nil {
			m.loading = false
			m.list.StopSpinner()
			m.addErr = msg.err
			m.list.SetSize(m.windowWidth, m.windowHeight-m.footerHeight())
			return m, nil
		}
		m.timelineID = msg.timelineID
		return m, completeTaskCmd(m.client, m.token, m.timelineID, msg.task)

	case addTaskMsg:
		m.loading = false
		m.list.StopSpinner()
		if msg.err != nil {
			m.addErr = msg.err
			m.list.SetSize(m.windowWidth, m.windowHeight-m.footerHeight())
			return m, nil
		}
		return m, tea.Batch(m.list.StartSpinner(), fetchTasksCmd(m.client, m.token, m.currentFilter))

	case completeTaskMsg:
		m.loading = false
		m.list.StopSpinner()
		if msg.err != nil {
			m.addErr = msg.err
			m.list.SetSize(m.windowWidth, m.windowHeight-m.footerHeight())
			return m, nil
		}
		return m, tea.Batch(m.list.StartSpinner(), fetchTasksCmd(m.client, m.token, m.currentFilter))

	case fetchTasksMsg:
		m.loading = false
		m.list.StopSpinner()
		if msg.err != nil {
			m.searchErr = msg.err
			m.list.SetSize(m.windowWidth, m.windowHeight-m.footerHeight())
			return m, nil
		}
		items := buildItems(msg.tasks)
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
		if m.adding {
			return m.updateAdd(msg)
		}
		switch msg.String() {
		case "q":
			return m, tea.Quit
		case "/":
			return m.openSearch()
		case "n":
			return m.openAdd()
		case "c":
			return m.completeSelected()
		case "r":
			m.loading = true
			return m, tea.Batch(m.list.StartSpinner(), fetchTasksCmd(m.client, m.token, m.currentFilter))
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

// View renders the current state.
func (m Model) View() tea.View {
	listView := m.list.View()

	var content string
	if m.searchErr != nil {
		errBar := searchErrorStyle.Render("Error: " + m.searchErr.Error() + "  (press / to retry)")
		content = lipgloss.JoinVertical(lipgloss.Left, listView, errBar)
	} else if m.addErr != nil {
		errBar := addErrorStyle.Render("Add failed: " + m.addErr.Error() + "  (press n to retry)")
		content = lipgloss.JoinVertical(lipgloss.Left, listView, errBar)
	} else if m.adding {
		bar := addBarStyle.Render(m.addInput.View())
		content = lipgloss.JoinVertical(lipgloss.Left, listView, bar)
	} else if m.searching {
		label := " INSERT "
		if m.searchMode == modeNormal {
			label = " NORMAL "
		}
		bar := searchBarStyle.Render(label + m.searchInput.View())
		content = lipgloss.JoinVertical(lipgloss.Left, listView, bar)
	} else {
		content = listView
	}

	v := tea.NewView(content)
	v.AltScreen = true
	return v
}
