package ui

import (
	"unicode"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"

	"git.codegoalie.com/rttui.git/internal/rtm"
)

type vimMode int

const (
	modeInsert vimMode = iota
	modeNormal
)

// fetchTasksMsg carries the result of an async RTM task fetch.
type fetchTasksMsg struct {
	tasks []rtm.Task
	err   error
}

// fetchTasksCmd returns a Cmd that fetches tasks off the main goroutine.
func fetchTasksCmd(client *rtm.Client, token, filter string) tea.Cmd {
	return func() tea.Msg {
		tasks, err := client.GetTasks(token, filter)
		return fetchTasksMsg{tasks: tasks, err: err}
	}
}

// openSearch activates the search bar pre-populated with the current filter.
func (m Model) openSearch() (Model, tea.Cmd) {
	m.searching = true
	m.searchMode = modeInsert
	m.searchInput.SetValue(m.currentFilter)
	m.searchInput.CursorEnd()
	cmd := m.searchInput.Focus()
	m.list.SetSize(m.windowWidth, m.windowHeight-footerBarHeight)
	return m, cmd
}

// closeSearch hides the search bar without changing the active filter.
func (m Model) closeSearch() Model {
	m.searching = false
	m.searchErr = nil
	m.list.SetSize(m.windowWidth, m.windowHeight)
	return m
}

// submitSearch fires an async fetch with the current input value.
func (m Model) submitSearch() (Model, tea.Cmd) {
	query := m.searchInput.Value()
	m.currentFilter = query
	m.loading = true
	m.searching = false
	m.list.SetSize(m.windowWidth, m.windowHeight)
	return m, tea.Batch(m.list.StartSpinner(), fetchTasksCmd(m.client, m.token, query))
}

// updateSearch routes keypresses through the vim state machine while the search bar is open.
func (m Model) updateSearch(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch m.searchMode {
	case modeInsert:
		switch msg.String() {
		case "enter":
			return m.submitSearch()
		case "esc":
			m.searchMode = modeNormal
			if m.searchInput.Position() > 0 {
				m.searchInput.SetCursor(m.searchInput.Position() - 1)
			}
			return m, nil
		default:
			var cmd tea.Cmd
			m.searchInput, cmd = m.searchInput.Update(msg)
			return m, cmd
		}

	case modeNormal:
		pos := m.searchInput.Position()
		val := []rune(m.searchInput.Value())
		switch msg.String() {
		case "enter":
			return m.submitSearch()
		case "esc":
			return m.closeSearch(), nil
		case "i":
			m.searchMode = modeInsert
		case "a":
			m.searchMode = modeInsert
			m.searchInput.SetCursor(pos + 1)
		case "A":
			m.searchMode = modeInsert
			m.searchInput.CursorEnd()
		case "I":
			m.searchMode = modeInsert
			m.searchInput.CursorStart()
		case "h":
			m.searchInput.SetCursor(pos - 1)
		case "l":
			m.searchInput.SetCursor(pos + 1)
		case "0":
			m.searchInput.CursorStart()
		case "$":
			m.searchInput.CursorEnd()
		case "w":
			m.searchInput.SetCursor(nextWordForward(val, pos))
		case "b":
			m.searchInput.SetCursor(nextWordBackward(val, pos))
		case "x":
			if pos < len(val) {
				newVal := string(append(val[:pos:pos], val[pos+1:]...))
				m.searchInput.SetValue(newVal)
				m.searchInput.SetCursor(pos)
			}
		}
		return m, nil
	}
	return m, nil
}

// nextWordForward returns the cursor position after jumping one word forward.
func nextWordForward(val []rune, pos int) int {
	n := len(val)
	for pos < n && !unicode.IsSpace(val[pos]) {
		pos++
	}
	for pos < n && unicode.IsSpace(val[pos]) {
		pos++
	}
	return pos
}

// nextWordBackward returns the cursor position after jumping one word backward.
func nextWordBackward(val []rune, pos int) int {
	if pos == 0 {
		return 0
	}
	pos--
	for pos > 0 && unicode.IsSpace(val[pos]) {
		pos--
	}
	for pos > 0 && !unicode.IsSpace(val[pos-1]) {
		pos--
	}
	return pos
}

// newSearchInput builds a configured textinput for the search bar.
func newSearchInput() textinput.Model {
	ti := textinput.New()
	ti.Prompt = "Search: "
	ti.CharLimit = 256
	return ti
}
