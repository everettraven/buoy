package styles

import "github.com/charmbracelet/lipgloss"

// TODO: better styling

var SelectedTitleStyle = lipgloss.NewStyle().
	Bold(true).Align(lipgloss.Left).Border(lipgloss.RoundedBorder(), true, true, false, true)

var TitleStyle = lipgloss.NewStyle().
	Bold(true).Align(lipgloss.Left).Border(lipgloss.RoundedBorder())

var ModelStyle = lipgloss.NewStyle().
	Align(lipgloss.Center, lipgloss.Center).
	BorderStyle(lipgloss.HiddenBorder())
