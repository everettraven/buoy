package datastream

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/everettraven/buoy/pkg/charm/models/panels"
	"k8s.io/apimachinery/pkg/api/meta"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/tools/cache"
)

func TableDatastreamFunc(dynamicClient *dynamic.DynamicClient, restMapper meta.RESTMapper) DatastreamFactoryFunc {
	return func(m tea.Model) (Datastream, error) {
		if _, ok := m.(*panels.Table); !ok {
			return nil, &InvalidPanelType{fmt.Errorf("model is not of type *panels.Table")}
		}
		table := m.(*panels.Table)
		tableDef := table.TableDefinition()
		// create informer and event handler
		gvk := schema.GroupVersionKind{
			Group:   tableDef.Group,
			Version: tableDef.Version,
			Kind:    tableDef.Kind,
		}
		mapping, err := restMapper.RESTMapping(gvk.GroupKind(), gvk.Version)
		if err != nil {
			return nil, fmt.Errorf("error creating resource mapping: %w", err)
		}

		ns := tableDef.Namespace
		if mapping.Scope.Name() == meta.RESTScopeNameRoot {
			ns = ""
		}
		infFact := dynamicinformer.NewFilteredDynamicSharedInformerFactory(
			dynamicClient,
			1*time.Minute,
			ns,
			dynamicinformer.TweakListOptionsFunc(func(options *v1.ListOptions) {
				ls := labels.SelectorFromSet(tableDef.LabelSelector)
				options.LabelSelector = ls.String()
			}),
		)

		inf := infFact.ForResource(mapping.Resource)
		inf.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				u := obj.(*unstructured.Unstructured)
				table.AddOrUpdate(u)
			},
			UpdateFunc: func(oldObj, newObj interface{}) {
				u := newObj.(*unstructured.Unstructured)
				table.AddOrUpdate(u)
			},
			DeleteFunc: func(obj interface{}) {
				u := obj.(*unstructured.Unstructured)
				table.DeleteRow(u.GetUID())
			},
		})

		table.SetLister(inf.Lister())
		table.SetScope(mapping.Scope.Name())
		return inf.Informer(), nil
	}
}
