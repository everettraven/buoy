package datastream

import (
	"bytes"
	"fmt"
	"io"
	"time"

	"github.com/alecthomas/chroma/quick"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/everettraven/buoy/pkg/charm/models/panels"
	"k8s.io/apimachinery/pkg/api/meta"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/tools/cache"
	"sigs.k8s.io/yaml"
)

func ItemDatastreamFunc(dynamicClient *dynamic.DynamicClient, restMapper meta.RESTMapper) DatastreamFactoryFunc {
	return func(m tea.Model) (Datastream, error) {
		if _, ok := m.(*panels.Item); !ok {
			return nil, &InvalidPanelType{fmt.Errorf("model is not of type *panels.Item")}
		}
		panel := m.(*panels.Item)
		item := panel.ItemDefinition()
		theme := panel.Theme().SyntaxHighlightDarkTheme
		if !lipgloss.HasDarkBackground() {
			theme = panel.Theme().SyntaxHighlightLightTheme
		}
		// create informer and event handler
		infFact := dynamicinformer.NewFilteredDynamicSharedInformerFactory(dynamicClient, 1*time.Minute, item.Key.Namespace, func(lo *v1.ListOptions) {
			lo.FieldSelector = fmt.Sprintf("metadata.name=%s", item.Key.Name)
		})
		gvk := schema.GroupVersionKind{
			Group:   item.Group,
			Version: item.Version,
			Kind:    item.Kind,
		}
		mapping, err := restMapper.RESTMapping(gvk.GroupKind(), gvk.Version)
		if err != nil {
			return nil, fmt.Errorf("error creating resource mapping: %w", err)
		}

		inf := infFact.ForResource(mapping.Resource)
		_, err = inf.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				u := obj.(*unstructured.Unstructured)
				itemJSON, err := u.MarshalJSON()
				if err != nil {
					panel.SetContent(fmt.Sprintf("error marshalling item %q", item.Key.String()))
					return
				}

				itemYAML, err := yaml.JSONToYAML(itemJSON)
				if err != nil {
					panel.SetContent(fmt.Sprintf("converting JSON to YAML for item %q", item.Key.String()))
					return
				}
				rw := &bytes.Buffer{}
				err = quick.Highlight(rw, string(itemYAML), "yaml", "terminal16m", theme)
				if err != nil {
					panel.SetContent(fmt.Sprintf("highlighting YAML for item %q", item.Key.String()))
					return
				}
				highlighted, err := io.ReadAll(rw)
				if err != nil {
					panel.SetContent(fmt.Sprintf("reading highlighted YAML for item %q", item.Key.String()))
					return
				}
				panel.SetContent(string(highlighted))
			},
			UpdateFunc: func(oldObj, newObj interface{}) {
				u := newObj.(*unstructured.Unstructured)
				itemJSON, err := u.MarshalJSON()
				if err != nil {
					panel.SetContent(fmt.Sprintf("error marshalling item %q", item.Key.String()))
					return
				}

				itemYAML, err := yaml.JSONToYAML(itemJSON)
				if err != nil {
					panel.SetContent(fmt.Sprintf("converting JSON to YAML for item %q", item.Key.String()))
					return
				}
				rw := &bytes.Buffer{}
				err = quick.Highlight(rw, string(itemYAML), "yaml", "terminal16m", theme)
				if err != nil {
					panel.SetContent(fmt.Sprintf("highlighting YAML for item %q", item.Key.String()))
					return
				}
				highlighted, err := io.ReadAll(rw)
				if err != nil {
					panel.SetContent(fmt.Sprintf("reading highlighted YAML for item %q", item.Key.String()))
					return
				}
				panel.SetContent(string(highlighted))
			},
			DeleteFunc: func(obj interface{}) {
				panel.SetContent("")
			},
		})

		if err != nil {
			return nil, fmt.Errorf("adding event handler to informer: %w", err)
		}

		return inf.Informer(), nil
	}

}
