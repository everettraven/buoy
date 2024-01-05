package paneler

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/everettraven/buoy/pkg/charm/styles"
	"github.com/everettraven/buoy/pkg/types"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
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

func NewDefaultPaneler(cfg *rest.Config, theme *styles.Theme) (Paneler, error) {
	dClient, err := dynamic.NewForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("error creating dynamic client: %w", err)
	}

	kubeClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("creating kubernetes.Clientset: %w", err)
	}

	di, err := discovery.NewDiscoveryClientForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("error creating discovery client: %w", err)
	}

	gr, err := restmapper.GetAPIGroupResources(di)
	if err != nil {
		return nil, fmt.Errorf("error getting API group resources: %w", err)
	}
	rm := restmapper.NewDiscoveryRESTMapper(gr)

	return &paneler{
		panelerRegistry: map[string]Paneler{
			types.PanelTypeTable: NewTable(dClient, di, rm, theme),
			types.PanelTypeItem:  NewItem(dClient, di, rm, theme),
			types.PanelTypeLogs:  NewLog(kubeClient, dClient, di, rm, theme),
		},
	}, nil
}
