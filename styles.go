package main

import "github.com/charmbracelet/lipgloss"

var (
	todoColor    = lipgloss.Color("246") // gray
	doingColor   = lipgloss.Color("39")  // blue
	blockedColor = lipgloss.Color("208") // orange
	doneColor    = lipgloss.Color("34")  // green

	stateColors = map[string]lipgloss.Color{
		"todo":    todoColor,
		"doing":   doingColor,
		"blocked": blockedColor,
		"done":    doneColor,
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

	statusBarStyle = lipgloss.NewStyle().
			Faint(true).
			Padding(0, 1)

	helpStyle = lipgloss.NewStyle().
			Faint(true).
			Padding(0, 1)
)
