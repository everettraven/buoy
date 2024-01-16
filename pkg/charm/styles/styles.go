package styles

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/charmbracelet/lipgloss"
)

// Theme is a collection of adaptive colors to be used when rendering the UI.
// This is the basis for customizable themes
type Theme struct {
	// TabColor is the color used to render the tabs and associated borders
	TabColor lipgloss.AdaptiveColor
	// SelectedRowHighlightColor is the color used to highlight the selected row in the table
	SelectedRowHighlightColor lipgloss.AdaptiveColor
	// LogSearchHighlightColor is the color used to highlight the search term in the log view
	LogSearchHighlightColor lipgloss.AdaptiveColor
	// SyntaxHighlightDarkTheme is the name of the syntax highlighting theme to use when the
	// terminal has a dark background. Available themes can be found at
	// https://github.com/alecthomas/chroma/tree/master/styles
	SyntaxHighlightDarkTheme string
	// SyntaxHighlightLightTheme is the name of the syntax highlighting theme to use when the
	// terminal has a light background. Available themes can be found at
	// https://github.com/alecthomas/chroma/tree/master/styles
	SyntaxHighlightLightTheme string
}

const DefaultThemePath = "~/.config/buoy/themes/default.json"

var DefaultColor = lipgloss.AdaptiveColor{Light: "63", Dark: "117"}

func LoadTheme(themePath string) (Theme, error) {
	t := Theme{
		TabColor:                  DefaultColor,
		SelectedRowHighlightColor: DefaultColor,
		LogSearchHighlightColor:   DefaultColor,
		SyntaxHighlightDarkTheme:  "nord",
		SyntaxHighlightLightTheme: "monokailight",
	}
	// If the specified theme file doesn't exist, use the default theme
	if _, err := os.Stat(themePath); err != nil {
		return t, nil
	}

	raw, err := os.ReadFile(themePath)
	if err != nil {
		return t, fmt.Errorf("reading theme file: %w", err)
	}

	customTheme := &Theme{}
	err = json.Unmarshal(raw, customTheme)
	if err != nil {
		return t, fmt.Errorf("unmarshalling theme: %w", err)
	}

	return *customTheme, nil
}

func (t *Theme) SelectedTabStyle() lipgloss.Style {
	border := lipgloss.RoundedBorder()
	border.BottomLeft = "┘"
	border.Bottom = " "
	border.BottomRight = "└"
	return lipgloss.NewStyle().Bold(true).Align(lipgloss.Center).Border(border).BorderForeground(t.TabColor).Padding(0, 1)
}

func (t *Theme) TabStyle() lipgloss.Style {
	border := lipgloss.RoundedBorder()
	border.BottomLeft = "┴"
	border.BottomRight = "┴"
	return lipgloss.NewStyle().Bold(true).Align(lipgloss.Center).Border(border).BorderForeground(t.TabColor).Padding(0, 1)
}

func (t *Theme) TabGap() lipgloss.Style {
	border := lipgloss.RoundedBorder()
	return lipgloss.NewStyle().Bold(true).Align(lipgloss.Center).Border(border, false, false, true, false).BorderForeground(t.TabColor).Padding(0, 1)
}

func (t *Theme) ContentStyle() lipgloss.Style {
	return lipgloss.NewStyle().Align(lipgloss.Center, lipgloss.Center)
}

func (t *Theme) TableSelectedRowStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(t.SelectedRowHighlightColor)
}

func (t *Theme) LogSearchHighlightStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(t.LogSearchHighlightColor)
}

func (t *Theme) LogSearchModeStyle() lipgloss.Style {
	return lipgloss.NewStyle().Italic(true).Faint(true)
}
