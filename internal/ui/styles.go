package ui

import "charm.land/lipgloss/v2"

var (
	priorityHighStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#f7768e")) // tokyo night red
	priorityMedStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#e0af68")) // tokyo night yellow
	priorityLowStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#7aa2f7")) // tokyo night blue
	priorityNoneStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#565f89")) // tokyo night comment
	dueSoonStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("#f7768e")) // tokyo night red for overdue/today
	dueStyle          = lipgloss.NewStyle().Foreground(lipgloss.Color("#9ece6a")) // tokyo night green

	searchBarStyle   = lipgloss.NewStyle().Background(lipgloss.Color("#292e42")).Foreground(lipgloss.Color("#c0caf5"))
	searchErrorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#f7768e"))

	addBarStyle   = lipgloss.NewStyle().Background(lipgloss.Color("#292e42")).Foreground(lipgloss.Color("#c0caf5"))
	addErrorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#f7768e"))
)
