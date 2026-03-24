package main

import "github.com/charmbracelet/lipgloss"

var (
	todoColor  = lipgloss.Color("246") // gray
	doingColor = lipgloss.Color("39")  // blue
	doneColor  = lipgloss.Color("34")  // green

	blockedTagStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("208")).
			Bold(true)

	blockedFlashStyle = lipgloss.NewStyle().
				Padding(0, 1).
				Bold(true).
				Background(lipgloss.Color("208")).
				Foreground(lipgloss.Color("0"))

	stateColors = map[string]lipgloss.Color{
		"todo":  todoColor,
		"doing": doingColor,
		"done":  doneColor,
	}

	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Padding(0, 1)

	taskStyle = lipgloss.NewStyle().
			Padding(0, 1)

	selectedTaskStyle = lipgloss.NewStyle().
				Padding(0, 1).
				Bold(true).
				Reverse(true)

	projectStyle = lipgloss.NewStyle().
			Faint(true)

	doneDateStyle = lipgloss.NewStyle().
			Faint(true)

	statusBarStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("229")).
			Padding(0, 1)

	helpStyle = lipgloss.NewStyle().
			Faint(true).
			Padding(0, 1)
)
