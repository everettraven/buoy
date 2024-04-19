package dashboard

import (
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/everettraven/buoy/pkg/charm/models/helper"
	"github.com/everettraven/buoy/pkg/charm/models/tabs"
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

// DashboardStyleOptions is the set of style options that can be
// used to configure the styles used by the Dashboard model
type DashboardStyleOptions struct {
	TabModelStyle tabs.TabModelStyleOptions
	DividerStyle  lipgloss.Style
}

// Dashboard is a tea.Model implementation
// for viewing Kubernetes information based
// on a declarative dashboard description
type Dashboard struct {
	tabber       *tabs.TabModel
	width        int
	help         help.Model
	keys         DashboardKeyMap
	dividerStyle lipgloss.Style
}

func New(keys DashboardKeyMap, style DashboardStyleOptions, panels ...tea.Model) *Dashboard {
	tabset := []tabs.Tab{}
	for _, panel := range panels {
		if namer, ok := panel.(Namer); ok {
			tabset = append(tabset, tabs.Tab{Name: namer.Name(), Model: panel})
		}
	}
	return &Dashboard{
		tabber:       tabs.New(tabs.DefaultTabberKeys, style.TabModelStyle, tabset...),
		help:         help.New(),
		keys:         keys,
		dividerStyle: style.DividerStyle,
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
	divider := d.dividerStyle.Render(strings.Repeat(" ", max(0, d.width-2)))
	return lipgloss.JoinVertical(0, d.tabber.View(), divider, d.help.View(d.Help()))
}

func (d *Dashboard) Help() help.KeyMap {
	return helper.NewCompositeHelpKeyMap(
		[]help.KeyMap{
			d.tabber.Help(),
			d.keys,
		}...,
	)
}
