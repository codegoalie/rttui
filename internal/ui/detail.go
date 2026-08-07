package ui

import (
	"fmt"
	"strings"
	"time"

	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"git.codegoalie.com/rttui/internal/rtm"
)

const subtaskListHeight = 8

// newSubtaskList builds the list widget used to display a task's sub-tasks
// inside the detail drawer: no title/help/status-bar/pagination chrome, no
// fuzzy filtering, and quit keybindings disabled since it's embedded inside
// the detail view rather than being the top-level list.
func newSubtaskList() list.Model {
	l := list.New(nil, subtaskDelegate{}, 0, 0)
	l.SetShowTitle(false)
	l.SetShowHelp(false)
	l.SetShowStatusBar(false)
	l.SetShowPagination(false)
	l.SetFilteringEnabled(false)
	l.DisableQuitKeybindings()
	return l
}

// fetchSubtasksMsg carries the result of an async sub-task fetch.
type fetchSubtasksMsg struct {
	subtasks []rtm.Task
	err      error
}

// fetchSubtasksCmd returns a Cmd that fetches a task's sub-tasks off the main
// goroutine.
func fetchSubtasksCmd(client *rtm.Client, token string, parent rtm.Task) tea.Cmd {
	return func() tea.Msg {
		subtasks, err := client.GetSubtasks(token, parent)
		return fetchSubtasksMsg{subtasks: subtasks, err: err}
	}
}

// buildSubtaskItems converts tasks into list items for the sub-task list.
// Pure function — no heading grouping (unlike buildItems), just one TaskItem
// per task, since RTM already returns them sorted (SortTasks in GetSubtasks).
func buildSubtaskItems(tasks []rtm.Task) []list.Item {
	items := make([]list.Item, len(tasks))
	for i, t := range tasks {
		items[i] = TaskItem{task: t}
	}
	return items
}

// openDetail opens the full-screen task detail view for the currently
// selected task. If the cursor is on a HeadingItem rather than a task, this
// is a no-op. Kicks off an async fetch of the task's sub-tasks.
func (m Model) openDetail() (Model, tea.Cmd) {
	item, ok := m.list.SelectedItem().(TaskItem)
	if !ok {
		return m, nil
	}
	m.detailOpen = true
	m.detailTask = item.task
	m.detailErr = nil
	// Full-screen takeover: the list is hidden entirely (not sharing the
	// screen with a footer bar), so it's resized off-screen rather than
	// shrunk by footerHeight().
	m.list.SetSize(m.windowWidth, 0)

	m.subtasksLoading = true
	m.subtaskList.SetSize(m.windowWidth, subtaskListHeight)
	m.subtaskList.SetItems(nil)
	return m, tea.Batch(m.subtaskList.StartSpinner(), fetchSubtasksCmd(m.client, m.token, item.task))
}

// closeDetail closes the full-screen task detail view and restores the list
// to its full size.
func (m Model) closeDetail() Model {
	m.detailOpen = false
	m.detailErr = nil
	m.list.SetSize(m.windowWidth, m.windowHeight)
	return m
}

// completeSubtaskMsg carries the result of an async sub-task completion.
type completeSubtaskMsg struct {
	task rtm.Task
	err  error
}

// completeSubtaskCmd returns a Cmd that marks a sub-task complete off the
// main goroutine.
func completeSubtaskCmd(client *rtm.Client, token, timeline string, task rtm.Task) tea.Cmd {
	return func() tea.Msg {
		err := client.CompleteTask(token, timeline, task.ListID, task.TaskseriesID, task.ID)
		return completeSubtaskMsg{task: task, err: err}
	}
}

// fetchTimelineAndCompleteSubtaskMsg carries the result of fetching a
// timeline before completing a sub-task (when none was cached yet).
type fetchTimelineAndCompleteSubtaskMsg struct {
	timelineID string
	task       rtm.Task
	err        error
}

// fetchTimelineAndCompleteSubtaskCmd returns a Cmd that fetches a fresh
// timeline off the main goroutine, to be followed by completeSubtaskCmd.
func fetchTimelineAndCompleteSubtaskCmd(client *rtm.Client, token string, task rtm.Task) tea.Cmd {
	return func() tea.Msg {
		id, err := client.GetTimeline(token)
		return fetchTimelineAndCompleteSubtaskMsg{timelineID: id, task: task, err: err}
	}
}

// completeSelectedSubtask marks the currently-selected sub-task complete in
// place (see completeSubtaskMsg handling in model.go) rather than removing it
// from the list, unlike the main list's completeSelected.
func (m Model) completeSelectedSubtask() (Model, tea.Cmd) {
	item, ok := m.subtaskList.SelectedItem().(TaskItem)
	if !ok || item.task.Completed {
		return m, nil
	}
	m.detailErr = nil
	if m.timelineID == "" {
		return m, fetchTimelineAndCompleteSubtaskCmd(m.client, m.token, item.task)
	}
	return m, completeSubtaskCmd(m.client, m.token, m.timelineID, item.task)
}

// fetchTimelineForDetailMsg carries the result of fetching a timeline for use
// by the detail drawer's sub-task/note add flows. Distinct from add.go's
// fetchTimelineMsg because that type's failure handling in model.go's
// Update() is hardcoded to the main add bar (closeAdd/addErr) — routing
// detail-drawer timeline fetches through it would incorrectly close the main
// add bar and set the wrong error field.
type fetchTimelineForDetailMsg struct {
	id  string
	err error
}

// fetchTimelineForDetailCmd returns a Cmd that fetches a fresh timeline off
// the main goroutine for the detail drawer's write flows.
func fetchTimelineForDetailCmd(client *rtm.Client, token string) tea.Cmd {
	return func() tea.Msg {
		id, err := client.GetTimeline(token)
		return fetchTimelineForDetailMsg{id: id, err: err}
	}
}

// openDetailAddSub activates the "add sub-task" input bar inside the detail
// drawer.
func (m Model) openDetailAddSub() (Model, tea.Cmd) {
	m.detailAddingSub = true
	m.detailErr = nil
	m.subtaskInput.SetValue("")
	if m.timelineID == "" {
		return m, fetchTimelineForDetailCmd(m.client, m.token)
	}
	return m, nil
}

// closeDetailAddSub hides the "add sub-task" input bar.
func (m Model) closeDetailAddSub() Model {
	m.detailAddingSub = false
	return m
}

// addSubtaskMsg carries the result of creating a sub-task and linking it to
// its parent. These are two separate RTM calls, so createdButNotLinked
// distinguishes "failed to create" from "created but failed to link".
type addSubtaskMsg struct {
	err                 error
	createdButNotLinked bool
}

// addSubtaskCmd creates a new task via smart-add and links it to parent as a
// sub-task, off the main goroutine.
func addSubtaskCmd(client *rtm.Client, token, timeline string, parent rtm.Task, transformed string) tea.Cmd {
	return func() tea.Msg {
		created, err := client.AddTaskFull(token, timeline, transformed)
		if err != nil {
			return addSubtaskMsg{err: fmt.Errorf("create sub-task: %w", err)}
		}
		if err := client.SetParentTask(token, timeline, created.ListID, created.TaskseriesID, created.ID, parent.ID); err != nil {
			return addSubtaskMsg{err: fmt.Errorf("sub-task created but not linked to parent: %w", err), createdButNotLinked: true}
		}
		return addSubtaskMsg{}
	}
}

// submitDetailAddSub submits the sub-task input's current value, creating and
// linking a new sub-task.
func (m Model) submitDetailAddSub() (Model, tea.Cmd) {
	raw := strings.TrimSpace(m.subtaskInput.Value())
	if raw == "" {
		return m.closeDetailAddSub(), nil
	}
	m.detailAddingSub = false
	m.subtasksLoading = true
	return m, tea.Batch(m.subtaskList.StartSpinner(),
		addSubtaskCmd(m.client, m.token, m.timelineID, m.detailTask, transformForRTM(raw)))
}

// updateDetailAddSub routes keypresses while the "add sub-task" input bar is
// open.
func (m Model) updateDetailAddSub(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		return m.submitDetailAddSub()
	case "esc":
		return m.closeDetailAddSub(), nil
	default:
		var cmd tea.Cmd
		m.subtaskInput, cmd = m.subtaskInput.Update(msg)
		return m, cmd
	}
}

// newNoteInput builds a configured textinput for the "add note" bar. Notes
// have no smart-add syntax, so this uses a plain textinput.Model rather than
// SmartInput.
func newNoteInput() textinput.Model {
	ti := textinput.New()
	ti.Prompt = "Note: "
	ti.CharLimit = 1024
	return ti
}

// openDetailAddNote activates the "add note" input bar inside the detail
// drawer.
func (m Model) openDetailAddNote() (Model, tea.Cmd) {
	m.detailAddingNote = true
	m.detailErr = nil
	m.noteInput.SetValue("")
	cmd := m.noteInput.Focus()
	if m.timelineID == "" {
		return m, tea.Batch(cmd, fetchTimelineForDetailCmd(m.client, m.token))
	}
	return m, cmd
}

// closeDetailAddNote hides the "add note" input bar.
func (m Model) closeDetailAddNote() Model {
	m.detailAddingNote = false
	return m
}

// addNoteMsg carries the result of adding a note. text is the note body that
// was submitted, so a successful add can be appended to m.detailTask.Notes
// optimistically in model.go's Update() (RTM has no cheap single-task
// refetch endpoint).
type addNoteMsg struct {
	text string
	err  error
}

// addNoteCmd adds a note to task off the main goroutine.
func addNoteCmd(client *rtm.Client, token, timeline string, task rtm.Task, text string) tea.Cmd {
	return func() tea.Msg {
		err := client.AddNote(token, timeline, task.ListID, task.TaskseriesID, task.ID, "", text)
		return addNoteMsg{text: text, err: err}
	}
}

// submitDetailAddNote submits the note input's current value.
func (m Model) submitDetailAddNote() (Model, tea.Cmd) {
	text := strings.TrimSpace(m.noteInput.Value())
	if text == "" {
		return m.closeDetailAddNote(), nil
	}
	m.detailAddingNote = false
	return m, addNoteCmd(m.client, m.token, m.timelineID, m.detailTask, text)
}

// updateDetailAddNote routes keypresses while the "add note" input bar is
// open.
func (m Model) updateDetailAddNote(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		return m.submitDetailAddNote()
	case "esc":
		return m.closeDetailAddNote(), nil
	default:
		var cmd tea.Cmd
		m.noteInput, cmd = m.noteInput.Update(msg)
		return m, cmd
	}
}

// findSubtaskIndex returns the index of the item in items whose task matches
// target by taskKey, or -1 if not found.
func findSubtaskIndex(items []list.Item, target rtm.Task) int {
	key := taskKey(target)
	for i, it := range items {
		if t, ok := it.(TaskItem); ok && taskKey(t.task) == key {
			return i
		}
	}
	return -1
}

// updateDetail routes keypresses while the full-screen detail view is open.
// When the "add sub-task" or "add note" input bar is active, keys are routed
// there first (mirroring how the top-level Update() in model.go checks
// detailOpen before searching/adding). Otherwise: esc/q close the drawer, c
// completes the selected sub-task in place, t opens the add-sub-task bar, n
// opens the add-note bar, and everything else (arrow keys, page up/down,
// etc.) is forwarded to the sub-task list widget for free navigation.
func (m Model) updateDetail(msg tea.KeyMsg) (Model, tea.Cmd) {
	if m.detailAddingSub {
		return m.updateDetailAddSub(msg)
	}
	if m.detailAddingNote {
		return m.updateDetailAddNote(msg)
	}
	switch msg.String() {
	case "esc", "q":
		return m.closeDetail(), nil
	case "c":
		return m.completeSelectedSubtask()
	case "t":
		return m.openDetailAddSub()
	case "n":
		return m.openDetailAddNote()
	default:
		var cmd tea.Cmd
		m.subtaskList, cmd = m.subtaskList.Update(msg)
		return m, cmd
	}
}

// formatDue formats a task's due date for the detail view: "None" if unset,
// otherwise "Mon Jan 2, 2006".
func formatDue(due time.Time) string {
	if due.IsZero() {
		return "None"
	}
	return due.Format("Mon Jan 2, 2006")
}

// formatTags formats a task's tags for the detail view: "None" if empty,
// otherwise a comma-and-space-joined list.
func formatTags(tags []string) string {
	if len(tags) == 0 {
		return "None"
	}
	return strings.Join(tags, ", ")
}

// formatNoteHeader formats a note's header line for the detail view.
// When the note has a title, the header is "<title> — <timestamp>"; with no
// title, it's just the timestamp. The timestamp is always formatted as
// "Jan 2, 2006 3:04 PM" (e.g. "Mar 15, 2024 3:04 PM").
func formatNoteHeader(n rtm.Note) string {
	ts := n.Created.Format("Jan 2, 2006 3:04 PM")
	if n.Title == "" {
		return ts
	}
	return n.Title + " — " + ts
}

// renderDetail renders the full-screen, read-only task detail view: header,
// fields, notes, an optional error bar, and a bottom hint line.
func (m Model) renderDetail() string {
	task := m.detailTask

	// Header: priority bar + task name, always styled as if selected since
	// it's the sole focus of a full-screen view. Mirrors delegate.go's
	// selected-title style path exactly.
	barColor := priorityColor(task.Priority)
	bar := lipgloss.NewStyle().Foreground(barColor).Render("▎")
	titleStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#c0caf5")).Bold(true)
	if task.URL != "" {
		titleStyle = titleStyle.Hyperlink(task.URL).Underline(true)
	}
	header := bar + " " + titleStyle.Render(task.Name)

	urlText := task.URL
	if urlText == "" {
		urlText = "None"
	}
	fields := lipgloss.JoinVertical(lipgloss.Left,
		detailLabelStyle.Render("Due:")+" "+formatDue(task.Due),
		detailLabelStyle.Render("List:")+" "+task.ListID,
		detailLabelStyle.Render("Tags:")+" "+formatTags(task.Tags),
		detailLabelStyle.Render("URL:")+" "+urlText,
	)

	subtaskLines := []string{headingStyle.Render("── Sub-tasks ──")}
	switch {
	case len(m.subtaskList.Items()) > 0:
		subtaskLines = append(subtaskLines, m.subtaskList.View())
	case m.subtasksLoading:
		subtaskLines = append(subtaskLines, "Loading…")
	default:
		subtaskLines = append(subtaskLines, "No sub-tasks yet")
	}
	subtasks := lipgloss.JoinVertical(lipgloss.Left, subtaskLines...)

	notesLines := []string{headingStyle.Render("── Notes ──")}
	if len(task.Notes) == 0 {
		notesLines = append(notesLines, "No notes yet")
	} else {
		for _, n := range task.Notes {
			notesLines = append(notesLines,
				detailNoteHeaderStyle.Render(formatNoteHeader(n)),
				detailNoteBodyStyle.Render(n.Body),
			)
		}
	}
	notes := lipgloss.JoinVertical(lipgloss.Left, notesLines...)

	sections := []string{header, "", fields, "", subtasks, "", notes}

	if m.detailAddingSub {
		sections = append(sections, "", addBarStyle.Render(m.subtaskInput.View()))
	} else if m.detailAddingNote {
		sections = append(sections, "", addBarStyle.Render(m.noteInput.View()))
	}

	if m.detailErr != nil {
		sections = append(sections, "", addErrorStyle.Render("Error: "+m.detailErr.Error()))
	}

	hints := []string{"esc: close", "↑/↓: navigate", "c: complete", "t: add sub-task", "n: add note"}
	sections = append(sections, "", detailHintStyle.Render(strings.Join(hints, " • ")))

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}
