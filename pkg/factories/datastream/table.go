package datastream

import (
	"fmt"
	"time"

	"github.com/everettraven/buoy/pkg/charm/models/panels/table"
	"k8s.io/apimachinery/pkg/api/meta"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/tools/cache"
	"sigs.k8s.io/yaml"
)

type Table interface {
	GVK() schema.GroupVersionKind
	AddOrUpdate(*unstructured.Unstructured)
	DeleteRow(types.UID)
	Namespace() string
	LabelSelector() labels.Set
	SetViewActionFunc(table.ViewActionFunc)
}

func TableDatastreamFunc(dynamicClient *dynamic.DynamicClient, restMapper meta.RESTMapper) DatastreamFactoryFunc {
	return func(obj interface{}) (Datastream, error) {
		tbl, ok := obj.(Table)
		if !ok {
			return nil, &InvalidPanelType{fmt.Errorf("model is not of type *panels.Table")}
		}

		mapping, err := restMapper.RESTMapping(tbl.GVK().GroupKind(), tbl.GVK().Version)
		if err != nil {
			return nil, fmt.Errorf("error creating resource mapping: %w", err)
		}

		ns := tbl.Namespace()
		if mapping.Scope.Name() == meta.RESTScopeNameRoot {
			ns = ""
		}
		infFact := dynamicinformer.NewFilteredDynamicSharedInformerFactory(
			dynamicClient,
			1*time.Minute,
			ns,
			dynamicinformer.TweakListOptionsFunc(func(options *v1.ListOptions) {
				ls := labels.SelectorFromSet(tbl.LabelSelector())
				options.LabelSelector = ls.String()
			}),
		)

		inf := infFact.ForResource(mapping.Resource)
		_, err = inf.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				u := obj.(*unstructured.Unstructured)
				tbl.AddOrUpdate(u)
			},
			UpdateFunc: func(oldObj, newObj interface{}) {
				u := newObj.(*unstructured.Unstructured)
				tbl.AddOrUpdate(u)
			},
			DeleteFunc: func(obj interface{}) {
				u := obj.(*unstructured.Unstructured)
				tbl.DeleteRow(u.GetUID())
			},
		})
		if err != nil {
			return nil, err
		}
		tbl.SetViewActionFunc(func(row *table.RowInfo) (string, error) {
			name := row.Identifier.String()
			if mapping.Scope.Name() == meta.RESTScopeNameRoot {
				name = row.Identifier.Name
			}

			obj, err := inf.Lister().Get(name)
			if err != nil {
				return "", fmt.Errorf("fetching definition for %q: %w", name, err)
			}

			itemJSON, err := obj.(*unstructured.Unstructured).MarshalJSON()
			if err != nil {
				return "", fmt.Errorf("error marshalling item %q: %w", name, err)
			}

			itemYAML, err := yaml.JSONToYAML(itemJSON)
			if err != nil {
				return "", fmt.Errorf("converting JSON to YAML for item %q: %w", name, err)
			}

            return string(itemYAML), nil
		})
		return inf.Informer(), nil
	}
}
