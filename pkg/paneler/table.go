package paneler

import (
	"encoding/json"
	"fmt"
	"time"

	tbl "github.com/calyptia/go-bubble-table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/everettraven/buoy/pkg/charm/models/panels"
	"github.com/everettraven/buoy/pkg/charm/styles"
	buoytypes "github.com/everettraven/buoy/pkg/types"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/cache"
)

var _ Paneler = &Table{}

type Table struct {
	dynamicClient   dynamic.Interface
	discoveryClient *discovery.DiscoveryClient
	restMapper      meta.RESTMapper
}

func NewTable(cfg *rest.Config) (*Table, error) {
	client, err := dynamic.NewForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("error creating dynamic client: %w", err)
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
	return &Table{
		dynamicClient:   client,
		discoveryClient: di,
		restMapper:      rm,
	}, nil
}

func (t *Table) Model(panel buoytypes.Panel) (tea.Model, error) {
	tab := buoytypes.Table{}
	err := json.Unmarshal(panel.Blob, &tab)
	if err != nil {
		return nil, fmt.Errorf("unmarshalling panel to table type: %s", err)
	}
	tw := t.modelWrapperForTablePanel(tab)
	return tw, t.runInformerForTable(tab, tw)
}

func (t *Table) modelWrapperForTablePanel(tablePanel buoytypes.Table) *panels.Table {
	columns := []string{}
	width := 0
	for _, column := range tablePanel.Columns {
		columns = append(columns, column.Header)
		width += column.Width
	}

	tab := tbl.New(columns, 100, 10)
	tab.Styles.SelectedRow = styles.TableSelectedRowStyle()
	return panels.NewTable(tablePanel.Name, tab, tablePanel.Columns)
}

func (t *Table) runInformerForTable(tablePanel buoytypes.Table, tw *panels.Table) error {
	// create informer and event handler
	infFact := dynamicinformer.NewDynamicSharedInformerFactory(t.dynamicClient, 1*time.Minute)
	gvk := schema.GroupVersionKind{
		Group:   tablePanel.Group,
		Version: tablePanel.Version,
		Kind:    tablePanel.Kind,
	}
	mapping, err := t.restMapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return fmt.Errorf("error creating resource mapping: %w", err)
	}

	inf := infFact.ForResource(mapping.Resource)
	_, err = inf.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			u := obj.(*unstructured.Unstructured)
			row := tbl.SimpleRow{}
			for _, column := range tw.Columns() {
				val, err := getDotNotationValue(u.Object, column.Path)
				if err != nil {
					//TODO: Log some kind of info here
					continue
				}
				switch val := val.(type) {
				case string:
					row = append(row, val)
				case map[string]interface{}:
					data, err := json.Marshal(val)
					if err != nil {
						//TODO: log some kind of info here
						continue
					}
					row = append(row, string(data))
				default:
					row = append(row, fmt.Sprint(val))
				}
			}

			tw.AddOrUpdateRow(u.GetUID(), row)
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			u := newObj.(*unstructured.Unstructured)
			row := tbl.SimpleRow{}
			for _, column := range tw.Columns() {
				val, err := getDotNotationValue(u.Object, column.Path)
				if err != nil {
					//TODO: Log some kind of info here
					continue
				}
				switch val := val.(type) {
				case string:
					row = append(row, val)
				case map[string]interface{}:
					data, err := json.Marshal(val)
					if err != nil {
						//TODO: log some kind of info here
						continue
					}
					row = append(row, string(data))
				default:
					row = append(row, fmt.Sprint(val))
				}
			}

			tw.AddOrUpdateRow(u.GetUID(), row)
		},
		DeleteFunc: func(obj interface{}) {
			u := obj.(*unstructured.Unstructured)
			tw.DeleteRow(u.GetUID())
		},
	})
	if err != nil {
		return fmt.Errorf("adding event handler to informer: %w", err)
	}

	go inf.Informer().Run(make(<-chan struct{}))
	return nil
}
