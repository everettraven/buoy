package styles

import "github.com/charmbracelet/lipgloss"

var TitleStyle = lipgloss.NewStyle().
	Bold(true).Align(lipgloss.Left)

var ModelStyle = lipgloss.NewStyle().
	Align(lipgloss.Center, lipgloss.Center).
	BorderStyle(lipgloss.HiddenBorder())

var FocusedModelStyle = lipgloss.NewStyle().
	Align(lipgloss.Center, lipgloss.Center).
	BorderLeft(true).
	BorderLeftForeground(lipgloss.Color("78"))
