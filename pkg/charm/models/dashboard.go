package models

import (
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

// Dashboard is a tea.Model implementation
// for viewing Kubernetes information based
// on a declarative dashboard description
type Dashboard struct {
	// Title is the name of the Dashboard
	Title    string
	Panels   []tea.Model
	state    int
	vp       viewport.Model
	vpActive bool
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
		case "esc":
			d.vpActive = !d.vpActive
		}
	case tea.WindowSizeMsg:
		d.vp = viewport.New(msg.Width, msg.Height)
	}

	if d.vpActive {
		d.vp, cmd = d.vp.Update(msg)
	} else {
		d.Panels[d.state], cmd = d.Panels[d.state].Update(msg)
	}

	return d, cmd
}

func (d *Dashboard) View() string {
	var out strings.Builder
	for _, panel := range d.Panels {
		out.WriteString(panel.View())
		out.WriteString("\n\n")
	}
	d.vp.SetContent(out.String())
	return d.vp.View()
}
