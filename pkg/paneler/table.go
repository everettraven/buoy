package paneler

import (
	"encoding/json"
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/everettraven/buoy/pkg/charm/models/panels"
	"github.com/everettraven/buoy/pkg/charm/styles"
	buoytypes "github.com/everettraven/buoy/pkg/types"
	"k8s.io/apimachinery/pkg/api/meta"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"
)

var _ Paneler = &Table{}

type Table struct {
	dynamicClient   dynamic.Interface
	discoveryClient *discovery.DiscoveryClient
	restMapper      meta.RESTMapper
	theme           *styles.Theme
}

func NewTable(dynamicClient dynamic.Interface, discoveryClient *discovery.DiscoveryClient, restMapper meta.RESTMapper, theme *styles.Theme) *Table {
	return &Table{
		dynamicClient:   dynamicClient,
		discoveryClient: discoveryClient,
		restMapper:      restMapper,
		theme:           theme,
	}
}

func (t *Table) Model(panel buoytypes.Panel) (tea.Model, error) {
	tab := &buoytypes.Table{}
	err := json.Unmarshal(panel.Blob, tab)
	if err != nil {
		return nil, fmt.Errorf("unmarshalling panel to table type: %s", err)
	}
	model, informer, err := t.modelForTablePanel(tab)
	if err != nil {
		return nil, fmt.Errorf("creating model wrapper for table panel: %w", err)
	}
	go informer.Informer().Run(make(chan struct{}))
	return model, nil
}

func (t *Table) modelForTablePanel(tablePanel *buoytypes.Table) (*panels.Table, informers.GenericInformer, error) {
	inf, scope, err := t.informerForTable(tablePanel, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("creating informer for table: %w", err)
	}
	table := panels.NewTable(panels.DefaultTableKeys, tablePanel, inf.Lister(), scope, t.theme)
	_, err = setEventHandlerForTableInformer(inf, table)
	if err != nil {
		return nil, nil, fmt.Errorf("setting event handler for table informer: %w", err)
	}
	return table, inf, nil

}

func (t *Table) informerForTable(tablePanel *buoytypes.Table, tw *panels.Table) (informers.GenericInformer, meta.RESTScopeName, error) {
	// create informer and event handler
	gvk := schema.GroupVersionKind{
		Group:   tablePanel.Group,
		Version: tablePanel.Version,
		Kind:    tablePanel.Kind,
	}
	mapping, err := t.restMapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return nil, "", fmt.Errorf("error creating resource mapping: %w", err)
	}
	ns := tablePanel.Namespace
	if mapping.Scope.Name() == meta.RESTScopeNameRoot {
		ns = ""
	}
	infFact := dynamicinformer.NewFilteredDynamicSharedInformerFactory(
		t.dynamicClient,
		1*time.Minute,
		ns,
		dynamicinformer.TweakListOptionsFunc(func(options *v1.ListOptions) {
			ls := labels.SelectorFromSet(tablePanel.LabelSelector)
			options.LabelSelector = ls.String()
		}),
	)

	inf := infFact.ForResource(mapping.Resource)

	return inf, mapping.Scope.Name(), nil
}

func setEventHandlerForTableInformer(inf informers.GenericInformer, tw *panels.Table) (cache.ResourceEventHandlerRegistration, error) {
	return inf.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			u := obj.(*unstructured.Unstructured)
			tw.AddOrUpdate(u)
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			u := newObj.(*unstructured.Unstructured)
			tw.AddOrUpdate(u)
		},
		DeleteFunc: func(obj interface{}) {
			u := obj.(*unstructured.Unstructured)
			tw.DeleteRow(u.GetUID())
		},
	})
}
