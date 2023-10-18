package paneler

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/everettraven/buoy/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
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
	crClient, err := client.New(cfg, client.Options{})
	if err != nil {
		return nil, fmt.Errorf("creating controller-runtime client.Client: %w", err)
	}

	kubeClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("creating kubernetes.Clientset: %w", err)
	}

	return &paneler{
		panelerRegistry: map[string]Paneler{
			types.PanelTypeTable: &Table{Client: crClient},
			types.PanelTypeItem:  &Item{Client: crClient},
			types.PanelTypeLogs:  &Log{KubeClient: kubeClient},
		},
	}, nil
}
