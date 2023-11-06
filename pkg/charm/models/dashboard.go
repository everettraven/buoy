package models

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/everettraven/buoy/pkg/charm/styles"
)

type Namer interface {
	Name() string
}

// Dashboard is a tea.Model implementation
// for viewing Kubernetes information based
// on a declarative dashboard description
type Dashboard struct {
	// Title is the name of the Dashboard
	Title  string
	Panels []tea.Model
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
	}

	d.Panels[d.state], cmd = d.Panels[d.state].Update(msg)
	return d, cmd
}

func (d *Dashboard) View() string {
	tabs := []string{}
	for i, panel := range d.Panels {
		if namer, ok := panel.(Namer); ok {
			tabText := fmt.Sprintf("  %s  ", namer.Name())
			if i == d.state {
				tabs = append(tabs, styles.SelectedTitleStyle.Render(tabText))
				continue
			}

			tabs = append(tabs, styles.TitleStyle.Render(tabText))
		}
	}
	tabBlock := lipgloss.JoinHorizontal(0, tabs...)
	return lipgloss.JoinVertical(0, tabBlock, d.Panels[d.state].View())
}
