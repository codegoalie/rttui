package ui

import "charm.land/lipgloss/v2"

var (
	priorityHighStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))  // red
	priorityMedStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("11")) // yellow
	priorityLowStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("12")) // blue
	priorityNoneStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))  // gray
	dueSoonStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))  // red for overdue/today
	dueStyle          = lipgloss.NewStyle().Foreground(lipgloss.Color("10")) // green

	searchBarStyle   = lipgloss.NewStyle().Background(lipgloss.Color("237")).Foreground(lipgloss.Color("15"))
	searchErrorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))

	addBarStyle   = lipgloss.NewStyle().Background(lipgloss.Color("237")).Foreground(lipgloss.Color("15"))
	addErrorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
)
