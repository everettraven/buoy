package models

import (
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/everettraven/buoy/pkg/charm/styles"
)

type DashboardKeyMap struct {
	Help     key.Binding
	Quit     key.Binding
	TabRight key.Binding
	TabLeft  key.Binding
}

// ShortHelp returns keybindings to be shown in the mini help view. It's part
// of the key.Map interface.
func (k DashboardKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Help, k.Quit}
}

// FullHelp returns keybindings for the expanded help view. It's part of the
// key.Map interface.
func (k DashboardKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.TabRight, k.TabLeft}, // first column
		{k.Help, k.Quit},        // second column
	}
}

var DefaultDashboardKeys = DashboardKeyMap{
	TabRight: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "change tabs to the right"),
	),
	TabLeft: key.NewBinding(
		key.WithKeys("shift+tab"),
		key.WithHelp("shift+tab", "change tabs to the left"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "toggle help"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "esc", "ctrl+c"),
		key.WithHelp("q, esc, ctrl+c", "quit"),
	),
}

type Namer interface {
	Name() string
}

// Dashboard is a tea.Model implementation
// for viewing Kubernetes information based
// on a declarative dashboard description
type Dashboard struct {
	Panels []tea.Model
	state  int
	width  int
	help   help.Model
	keys   DashboardKeyMap
}

func NewDashboard(keys DashboardKeyMap, panels ...tea.Model) *Dashboard {
	return &Dashboard{
		Panels: panels,
		help:   help.New(),
		keys:   keys,
	}
}

func (d *Dashboard) Init() tea.Cmd { return nil }

func (d *Dashboard) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, d.keys.Quit):
			return d, tea.Quit
		case key.Matches(msg, d.keys.TabRight):
			d.state++
			if d.state > len(d.Panels)-1 {
				d.state = 0
			}
		case key.Matches(msg, d.keys.TabLeft):
			d.state--
			if d.state < 0 {
				d.state = len(d.Panels) - 1
			}
		case key.Matches(msg, d.keys.Help):
			d.help.ShowAll = !d.help.ShowAll
		}
	case tea.WindowSizeMsg:
		d.width = msg.Width
	}

	d.Panels[d.state], cmd = d.Panels[d.state].Update(msg)
	return d, cmd
}

func (d *Dashboard) View() string {
	tabs := []string{}
	for i, panel := range d.Panels {
		if namer, ok := panel.(Namer); ok {
			if i == d.state {
				tabs = append(tabs, styles.SelectedTabStyle().Render(namer.Name()))
				continue
			}

			tabs = append(tabs, styles.TabStyle().Render(namer.Name()))
		}
	}
	// TODO: This is not scrollable, so once there are more tabs than there is
	// terminal width it goes off screen. Might need to create a new model specifically.
	// for the tabs that enables some sort of scrolling/pagination.
	tabBlock := lipgloss.JoinHorizontal(lipgloss.Top, tabs...)
	// gap is a repeating of the spaces so that the bottom border continues the entire width
	// of the terminal. This allows it to look like a proper set of tabs
	gap := styles.TabGap().Render(strings.Repeat(" ", max(0, d.width-lipgloss.Width(tabBlock)-2)))
	tabsWithBorder := lipgloss.JoinHorizontal(lipgloss.Bottom, tabBlock, gap)
	content := styles.ContentStyle().Render(d.Panels[d.state].View())
	div := styles.TabGap().Render(strings.Repeat(" ", max(0, d.width-2)))
	return lipgloss.JoinVertical(0, tabsWithBorder, content, div, d.help.View(d.keys))
}
