package ui

import "charm.land/lipgloss/v2"

var (
	priorityHighColor = lipgloss.Color("#ff9e64") // tokyo night orange
	priorityMedColor  = lipgloss.Color("#7aa2f7") // tokyo night blue
	priorityLowColor  = lipgloss.Color("#565f89") // tokyo night comment/grey
	priorityNoneColor = lipgloss.Color("#3b4261") // tokyo night dim

	overdueStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#f7768e")) // red for overdue dates

	searchBarStyle   = lipgloss.NewStyle().Background(lipgloss.Color("#292e42")).Foreground(lipgloss.Color("#c0caf5"))
	searchErrorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#f7768e"))

	addBarStyle   = lipgloss.NewStyle().Background(lipgloss.Color("#292e42")).Foreground(lipgloss.Color("#c0caf5"))
	addErrorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#f7768e"))

	headingStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7aa2f7")). // tokyo night blue
			Bold(true)
)
