package paneler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/alecthomas/chroma/quick"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/everettraven/buoy/pkg/charm/models/panels"
	"github.com/everettraven/buoy/pkg/types"
	"k8s.io/apimachinery/pkg/api/meta"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/tools/cache"
	"sigs.k8s.io/yaml"
)

var _ Paneler = &Item{}

type Item struct {
	dynamicClient   dynamic.Interface
	discoveryClient *discovery.DiscoveryClient
	restMapper      meta.RESTMapper
}

func NewItem(dynamicClient dynamic.Interface, discoveryClient *discovery.DiscoveryClient, restMapper meta.RESTMapper) *Item {
	return &Item{
		dynamicClient:   dynamicClient,
		discoveryClient: discoveryClient,
		restMapper:      restMapper,
	}
}

func (t *Item) Model(panel types.Panel) (tea.Model, error) {
	item := types.Item{}
	err := json.Unmarshal(panel.Blob, &item)
	if err != nil {
		return nil, fmt.Errorf("unmarshalling panel to item type: %s", err)
	}
	iw := t.modelWrapperForItemPanel(item)
	return iw, t.runInformerForItem(item, iw)
}

func (t *Item) modelWrapperForItemPanel(itemPanel types.Item) *panels.Item {
	vp := viewport.New(100, 20)
	return panels.NewItem(itemPanel.Name, vp)
}

func (t *Item) runInformerForItem(item types.Item, panel *panels.Item) error {
	theme := "nord"
	if !lipgloss.HasDarkBackground() {
		theme = "monokailight"
	}
	// create informer and event handler
	infFact := dynamicinformer.NewFilteredDynamicSharedInformerFactory(t.dynamicClient, 1*time.Minute, item.Key.Namespace, func(lo *v1.ListOptions) {
		lo.FieldSelector = fmt.Sprintf("metadata.name=%s", item.Key.Name)
	})
	gvk := schema.GroupVersionKind{
		Group:   item.Group,
		Version: item.Version,
		Kind:    item.Kind,
	}
	mapping, err := t.restMapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return fmt.Errorf("error creating resource mapping: %w", err)
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
		return fmt.Errorf("adding event handler to informer: %w", err)
	}

	go inf.Informer().Run(make(<-chan struct{}))
	return nil
}
