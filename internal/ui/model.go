package ui

import (
	"time"

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

	completeErr     error
	pendingCompletes map[string]rtm.Task

	loading         bool
	refreshInterval time.Duration
}

// NewModel creates a Model pre-loaded with tasks.
func NewModel(client *rtm.Client, token, filter, addPreset string, refreshInterval time.Duration, tasks []rtm.Task) Model {
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
		client:          client,
		token:           token,
		currentFilter:   filter,
		addPreset:       addPreset,
		refreshInterval: refreshInterval,
		list:            l,
		searchInput:     newSearchInput(),
		addInput:        NewSmartInput("Add: "),
	}
}

// footerHeight returns the number of rows the footer bar occupies given current state.
func (m Model) footerHeight() int {
	if m.searching || m.adding || m.searchErr != nil || m.addErr != nil || m.completeErr != nil {
		return footerBarHeight
	}
	return 0
}

// stopSpinnerIfIdle stops the spinner when no loading or pending completions remain.
func (m *Model) stopSpinnerIfIdle() {
	if !m.loading && len(m.pendingCompletes) == 0 {
		m.list.StopSpinner()
	}
}

// Init satisfies tea.Model; schedules the first auto-refresh tick if configured.
func (m Model) Init() tea.Cmd {
	if m.refreshInterval > 0 {
		return autoRefreshCmd(m.refreshInterval)
	}
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
			key := taskKey(msg.task)
			delete(m.pendingCompletes, key)
			m.completeErr = msg.err
			setCmd := m.list.SetItems(m.restoreTaskItems(msg.task))
			m.list.SetSize(m.windowWidth, m.windowHeight-m.footerHeight())
			m.stopSpinnerIfIdle()
			return m, setCmd
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
		key := taskKey(msg.task)
		delete(m.pendingCompletes, key)
		if msg.err != nil {
			m.completeErr = msg.err
			setCmd := m.list.SetItems(m.restoreTaskItems(msg.task))
			m.list.SetSize(m.windowWidth, m.windowHeight-m.footerHeight())
			m.stopSpinnerIfIdle()
			return m, setCmd
		}
		m.stopSpinnerIfIdle()
		return m, nil

	case fetchTasksMsg:
		m.loading = false
		if msg.err != nil {
			m.searchErr = msg.err
			m.list.SetSize(m.windowWidth, m.windowHeight-m.footerHeight())
			m.stopSpinnerIfIdle()
			return m, nil
		}
		tasks := msg.tasks
		if len(m.pendingCompletes) > 0 {
			filtered := tasks[:0:0]
			for _, t := range tasks {
				if _, pending := m.pendingCompletes[taskKey(t)]; !pending {
					filtered = append(filtered, t)
				}
			}
			tasks = filtered
		}
		setCmd := m.list.SetItems(buildItems(tasks))
		m.stopSpinnerIfIdle()
		return m, setCmd

	case autoRefreshMsg:
		cmds := []tea.Cmd{autoRefreshCmd(m.refreshInterval)}
		if !m.loading {
			m.loading = true
			cmds = append(cmds, m.list.StartSpinner(), fetchTasksCmd(m.client, m.token, m.currentFilter))
		}
		return m, tea.Batch(cmds...)


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
	} else if m.completeErr != nil {
		errBar := addErrorStyle.Render("Complete failed: " + m.completeErr.Error() + "  (press c to retry)")
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
