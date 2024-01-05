package models

import (
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/everettraven/buoy/pkg/charm/styles"
)

type DashboardKeyMap struct {
	Help key.Binding
	Quit key.Binding
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
		{k.Help, k.Quit},
	}
}

var DefaultDashboardKeys = DashboardKeyMap{
	Help: key.NewBinding(
		key.WithKeys("ctrl+h"),
		key.WithHelp("ctrl+h", "toggle help"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q, ctrl+c", "quit"),
	),
}

type Namer interface {
	Name() string
}

// Dashboard is a tea.Model implementation
// for viewing Kubernetes information based
// on a declarative dashboard description
type Dashboard struct {
	tabber *Tabber
	width  int
	help   help.Model
	keys   DashboardKeyMap
	theme  styles.Theme
}

func NewDashboard(keys DashboardKeyMap, theme styles.Theme, panels ...tea.Model) *Dashboard {
	tabs := []Tab{}
	for _, panel := range panels {
		if namer, ok := panel.(Namer); ok {
			tabs = append(tabs, Tab{Name: namer.Name(), Model: panel})
		}
	}
	return &Dashboard{
		tabber: NewTabber(DefaultTabberKeys, theme, tabs...),
		help:   help.New(),
		keys:   keys,
		theme:  theme,
	}
}

func (d *Dashboard) Init() tea.Cmd { return nil }

func (d *Dashboard) tick() tea.Cmd {
	return tea.Tick(time.Millisecond, func(t time.Time) tea.Msg {
		return t
	})
}

func (d *Dashboard) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, d.keys.Quit):
			return d, tea.Quit
		case key.Matches(msg, d.keys.Help):
			d.help.ShowAll = !d.help.ShowAll
		}
	case tea.WindowSizeMsg:
		d.width = msg.Width
	}

	d.tabber, cmd = d.tabber.Update(msg)
	return d, tea.Batch(d.tick(), cmd)
}

func (d *Dashboard) View() string {
	div := d.theme.TabGap().Render(strings.Repeat(" ", max(0, d.width-2)))
	return lipgloss.JoinVertical(0, d.tabber.View(), div, d.help.View(d.Help()))
}

func (d *Dashboard) Help() help.KeyMap {
	return CompositeHelpKeyMap{
		helps: []help.KeyMap{
			d.tabber.Help(),
			d.keys,
		},
	}
}
