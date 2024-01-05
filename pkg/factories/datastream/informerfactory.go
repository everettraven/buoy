package datastream

import (
	"errors"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
)

type Datastream interface {
	Run(stopCh <-chan struct{})
}

type DatastreamFactory interface {
	DatastreamForModel(tea.Model) (Datastream, error)
}

type InvalidPanelType struct {
	error
}

type DatastreamFactoryFunc func(tea.Model) (Datastream, error)

type datastreamFactory struct {
	// informerFactoryFuncs is a list of functions that return informers
	// for a given model. The first informer returned is used. Returning an error
	// of type InvalidPanelType will cause the next function to be called. If any other
	// error is returned, the error is returned to the caller.
	datastreamFactoryFuncs []DatastreamFactoryFunc
}

var _ DatastreamFactory = &datastreamFactory{}

func (i *datastreamFactory) DatastreamForModel(model tea.Model) (Datastream, error) {
	invalidErr := &InvalidPanelType{}
	var stream Datastream
	for _, f := range i.datastreamFactoryFuncs {
		s, err := f(model)
		if err != nil {
			if errors.As(err, &invalidErr) {
				continue
			}
			return nil, err
		}
		stream = s
		break
	}

	return stream, nil
}

func NewDatastreamFactory(cfg *rest.Config) (DatastreamFactory, error) {
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

	return &datastreamFactory{
		datastreamFactoryFuncs: []DatastreamFactoryFunc{
			ItemDatastreamFunc(dClient, rm),
			TableDatastreamFunc(dClient, rm),
			LogsDatastreamFunc(kubeClient, dClient, rm),
		},
	}, nil
}
