package datastream

import (
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/api/meta"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/tools/cache"
	"sigs.k8s.io/yaml"
)

type ItemPanel interface {
	Key() types.NamespacedName
	GVK() schema.GroupVersionKind
	SetContent(string)
}

func ItemDatastreamFunc(dynamicClient *dynamic.DynamicClient, restMapper meta.RESTMapper) DatastreamFactoryFunc {
	return func(obj interface{}) (Datastream, error) {
		item, ok := obj.(ItemPanel)
		if !ok {
			return nil, &InvalidPanelType{fmt.Errorf("provided object doesn't implement the Item interface. Unable to determine namespace/name of item")}
		}

		// create informer and event handler
		infFact := dynamicinformer.NewFilteredDynamicSharedInformerFactory(dynamicClient, 1*time.Minute, item.Key().Namespace, func(lo *v1.ListOptions) {
			lo.FieldSelector = fmt.Sprintf("metadata.name=%s", item.Key().Name)
		})

		mapping, err := restMapper.RESTMapping(item.GVK().GroupKind(), item.GVK().Version)
		if err != nil {
			return nil, fmt.Errorf("error creating resource mapping: %w", err)
		}

		inf := infFact.ForResource(mapping.Resource)
		_, err = inf.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				u := obj.(*unstructured.Unstructured)
				itemJSON, err := u.MarshalJSON()
				if err != nil {
					item.SetContent(fmt.Sprintf("error marshalling item %q", item.Key().String()))
					return
				}

				itemYAML, err := yaml.JSONToYAML(itemJSON)
				if err != nil {
					item.SetContent(fmt.Sprintf("converting JSON to YAML for item %q", item.Key().String()))
					return
				}

				item.SetContent(string(itemYAML))
			},
			UpdateFunc: func(oldObj, newObj interface{}) {
				u := newObj.(*unstructured.Unstructured)
				itemJSON, err := u.MarshalJSON()
				if err != nil {
					item.SetContent(fmt.Sprintf("error marshalling item %q", item.Key().String()))
					return
				}

				itemYAML, err := yaml.JSONToYAML(itemJSON)
				if err != nil {
					item.SetContent(fmt.Sprintf("converting JSON to YAML for item %q", item.Key().String()))
					return
				}
				item.SetContent(string(itemYAML))
			},
			DeleteFunc: func(obj interface{}) {
				item.SetContent("")
			},
		})

		if err != nil {
			return nil, fmt.Errorf("adding event handler to informer: %w", err)
		}

		return inf.Informer(), nil
	}

}
