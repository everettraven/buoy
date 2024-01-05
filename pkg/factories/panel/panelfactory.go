package panel

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/everettraven/buoy/pkg/charm/styles"
	"github.com/everettraven/buoy/pkg/types"
)

type PanelFactory interface {
	ModelForPanel(types.Panel) (tea.Model, error)
}

type paneler struct {
	panelerRegistry map[string]PanelFactory
}

var _ PanelFactory = &paneler{}

func (p *paneler) ModelForPanel(panel types.Panel) (tea.Model, error) {
	if p, ok := p.panelerRegistry[panel.Type]; ok {
		return p.ModelForPanel(panel)
	}
	return nil, fmt.Errorf("panel %q has unknown panel type: %q", panel.Name, panel.Type)
}

func NewPanelFactory(theme styles.Theme) PanelFactory {
	return &paneler{
		panelerRegistry: map[string]PanelFactory{
			types.PanelTypeTable: &Table{theme: theme},
			types.PanelTypeItem:  &Item{theme: theme},
			types.PanelTypeLogs:  &Log{theme: theme},
		},
	}
}
