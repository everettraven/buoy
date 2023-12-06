package styles

import "github.com/charmbracelet/lipgloss"

var DefaultColor = lipgloss.AdaptiveColor{Light: "63", Dark: "117"}

func SelectedTabStyle() lipgloss.Style {
	border := lipgloss.RoundedBorder()
	border.BottomLeft = "┘"
	border.Bottom = " "
	border.BottomRight = "└"
	return lipgloss.NewStyle().Bold(true).Align(lipgloss.Center).Border(border).BorderForeground(DefaultColor).Padding(0, 1)
}

func TabStyle() lipgloss.Style {
	border := lipgloss.RoundedBorder()
	border.BottomLeft = "┴"
	border.BottomRight = "┴"
	return lipgloss.NewStyle().Bold(true).Align(lipgloss.Center).Border(border).BorderForeground(DefaultColor).Padding(0, 1)
}

func TabGap() lipgloss.Style {
	border := lipgloss.RoundedBorder()
	return lipgloss.NewStyle().Bold(true).Align(lipgloss.Center).Border(border, false, false, true, false).BorderForeground(DefaultColor).Padding(0, 1)
}

func ContentStyle() lipgloss.Style {
	return lipgloss.NewStyle().Align(lipgloss.Center, lipgloss.Center)
}

func TableSelectedRowStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(DefaultColor)
}

func LogSearchHighlightStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(DefaultColor)
}

func LogSearchModeStyle() lipgloss.Style {
	return lipgloss.NewStyle().Italic(true).Faint(true)
}
