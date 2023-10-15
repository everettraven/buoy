package styles

import "github.com/charmbracelet/lipgloss"

// TODO: better styling

var TitleStyle = lipgloss.NewStyle().
	Bold(true).Align(lipgloss.Left).Underline(true)

var ModelStyle = lipgloss.NewStyle().
	Align(lipgloss.Center, lipgloss.Center).
	BorderStyle(lipgloss.HiddenBorder())

var FocusedModelStyle = lipgloss.NewStyle().
	Align(lipgloss.Center, lipgloss.Center).
	BorderStyle(lipgloss.BlockBorder()).
	BorderForeground(lipgloss.Color("78"))
