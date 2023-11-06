package paneler

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/everettraven/buoy/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type Paneler interface {
	Model(types.Panel) (tea.Model, error)
}

type paneler struct {
	panelerRegistry map[string]Paneler
}

func (p *paneler) Model(panel types.Panel) (tea.Model, error) {
	if p, ok := p.panelerRegistry[panel.Type]; ok {
		return p.Model(panel)
	}
	return nil, fmt.Errorf("panel %q has unknown panel type: %q", panel.Name, panel.Type)
}

func NewDefaultPaneler(cfg *rest.Config) (Paneler, error) {
	kubeClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("creating kubernetes.Clientset: %w", err)
	}

	table, err := NewTable(cfg)
	if err != nil {
		return nil, fmt.Errorf("creating table paneler: %w", err)
	}

	item, err := NewItem(cfg)
	if err != nil {
		return nil, fmt.Errorf("creating item paneler: %w", err)
	}

	return &paneler{
		panelerRegistry: map[string]Paneler{
			types.PanelTypeTable: table,
			types.PanelTypeItem:  item,
			types.PanelTypeLogs:  &Log{KubeClient: kubeClient},
		},
	}, nil
}
