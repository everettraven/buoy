package models

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/everettraven/buoy/pkg/charm/styles"
	"github.com/everettraven/buoy/pkg/types"
	boxer "github.com/treilik/bubbleboxer"
)

type styleable interface {
	SetStyle(style lipgloss.Style)
}

// Dashboard is a tea.Model implementation
// for viewing Kubernetes information based
// on a declarative dashboard description
type Dashboard struct {
	// Title is the name of the Dashboard
	Title string
	// Tui is the bubbleboxer.Boxer to
	// use for the Dashboard
	Tui    *boxer.Boxer
	Panels []types.Panel
	state  int
}

func (d *Dashboard) Init() tea.Cmd { return nil }

func (d *Dashboard) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return d, tea.Quit
		case "tab":
			d.state++
			if d.state > len(d.Panels)-1 {
				d.state = 0
			}
		}
	case tea.WindowSizeMsg:
		d.Tui.UpdateSize(msg)
	}

	d.Tui.EditLeaf(d.Panels[d.state].Name, func(t tea.Model) (tea.Model, error) {
		if styleable, ok := t.(styleable); ok {
			styleable.SetStyle(styles.FocusedModelStyle)
		}
		tt, _ := t.Update(msg)
		return tt, nil
	})

	return d, cmd
}

func (d *Dashboard) View() string {
	return d.Tui.View()
}
